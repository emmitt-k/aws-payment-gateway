# Tron USDT Custodial Payment Gateway

## Overview
A cloud-native, custodial payment gateway for Tron USDT (TRC20) built on AWS. This system manages user wallets, processes deposits (payins) and withdrawals (payouts), and ensures financial integrity through double-entry bookkeeping.

## Requirements
- **Wallet Management**: Generate and manage deposit addresses for users.
- **Payin (Deposit)**: Detect and credit USDT transfers to user addresses.
- **Payout (Withdrawal)**: Securely sign and broadcast USDT transfer transactions.
- **Financial Integrity**: Double-entry bookkeeping for all balance changes.
- **Scalability**: Auto-scaling on AWS to handle high transaction throughput.
- **Security**: Secure key management (AWS KMS) and strict network policies.

## Technology Stack
- **Language**: Go (Golang) 1.22+
- **Frameworks**: `gofiber/fiber` (HTTP Router), `gorm` (ORM)
- **Migrations**: `golang-migrate` (CLI/Lib)
- **Communication**: 
    - **HTTP (REST)**: For synchronous internal requests (e.g., `GetBalance`, `GetDepositAddress`).
    - **SQS**: For asynchronous, durable event processing (e.g., `DepositDetected`, `PayoutRequested`).
- **Database**: 
    - **PostgreSQL (RDS)**: Core relational data (Accounts, Transactions, Ledger).
    - **DynamoDB**: High-volume audit logs, raw webhook events, and idempotency keys.
- **Messaging**: Amazon SQS
- **Blockchain Interaction**: Native Go Client (e.g., `gotron-sdk`) or direct HTTP to TronGrid.
    - *Note*: Avoid `tronweb` (JS) to keep the backend stack 100% Go.
- **Infrastructure**: AWS (VPC, EC2, ASG, ALB, KMS, Secrets Manager)
- **Observability**: 
    - **Tracing**: AWS X-Ray
    - **Metrics**: Amazon Managed Service for Prometheus (AMP)
    - **Dashboards**: Amazon Managed Grafana (AMG)
    - **Logs**: Amazon CloudWatch Logs
- **CI/CD**: GitHub Actions

## Architecture

### Style: Microservices in a Monorepo
We utilize a **Monorepo** structure to house multiple microservices. This allows for shared code (utils, domain models) while maintaining independent deployability.

### Pattern: Go Clean Architecture
Each service follows **Clean Architecture** principles to separate concerns:
1.  **Entities**: Core business objects (e.g., `Transaction`, `Account`).
2.  **Usecases**: Business logic (e.g., `ProcessDeposit`, `CreateWallet`).
3.  **Interfaces/Adapters**: Gateways to external systems (e.g., `PostgresRepo`, `TronClient`, `SQSPublisher`).
4.  **Drivers/Infrastructure**: Frameworks and tools (e.g., `HTTP Server`, `Cobra CLI`).

### Services
1.  **Auth Service**: App registration and API key issuance/validation for external clients.
2.  **Payin Service**: Stores payin requests and generates deposit addresses (via KMS).
3.  **Observer Service**: Scans Tron blockchain for events and publishes to SQS.
4.  **Ledger Service**: Consumes events, manages double-entry accounting, and updates balances.
5.  **Payout Service**: Handles withdrawal requests, signing, and broadcasting.

## Infrastructure (AWS)
- **VPC**: Isolated network with Public/Private subnets.
- **RDS**: PostgreSQL Multi-AZ for persistence.
- **SQS**:
    - `deposit-events.fifo`: Ordered queue for incoming deposits.
    - `payout-requests.fifo`: Ordered queue for withdrawals.
- **ASG**: Auto Scaling Groups for stateless services (Observer, API).
- **KMS**: Hardware security module for managing Tron private keys.

## File Structure (Monorepo)
This project follows the **Standard Go Project Layout** adapted for a monorepo with Clean Architecture.

```text
.
├── api/                                # API Definitions (Contracts)
│
├── cmd/                                # Application Entry Points (Main)
│   ├── payin-svc/                      # Payin Microservice
│   │   └── main.go                     # Wires up deps (DB, Config) and starts HTTP server
│   ├── auth-svc/                       # Auth Microservice (API keys)
│   │   └── main.go                     # Issues/validates API keys, HTTP server
│   ├── observer-svc/                   # Blockchain Observer Service
│   │   └── main.go                     # Starts the block scanning loop
│   └── ledger-svc/                     # Ledger Microservice
│       └── main.go                     # Starts the double-entry accounting engine
│
├── internal/                           # Private Application Code (The Core)
│   ├── common/                         # Shared utilities across services
│   │   ├── config/                     # Configuration loading (Env vars, SSM)
│   │   ├── db/                         # Database connection helpers (Postgres)
│   │   └── logger/                     # Structured logging setup (Zap/Logrus)
│   │
│   ├── auth/                           # Auth Service (Clean Architecture)
│   │   ├── domain/                     # App, ApiKey entities
│   │   ├── usecase/                    # RegisterApp, IssueApiKey, ValidateApiKey
│   │   └── adapter/                    # Repository, HTTP handlers, interceptors
│   │
│   ├── payin/                          # Payin Service (Clean Architecture)
│   │   ├── domain/                     # Enterprise Business Rules (Pure Go structs)
│   │   │   └── account.go              # Account entity, value objects
│   │   ├── usecase/                    # Application Business Rules
│   │   │   └── create_address.go       # Logic flow: Validate -> DB -> KMS -> Return
│   │   └── adapter/                    # Interface Adapters (Talks to outside world)
│   │       ├── repository/             # Database implementations (SQL queries)
│   │       └── http/                   # HTTP handlers (maps JSON -> Usecase)
│   │
│   ├── ledger/                         # Ledger Service (Double Entry Engine)
│   │   ├── domain/                     # Transaction, Posting entities
│   │   └── ...
│   │
│   └── observer/                       # Tron Observer Logic
│       ├── scanner.go                  # Polls TronGrid for new blocks
│       └── publisher.go                # Pushes detected events to SQS
│
├── pkg/                                # Public Library Code (Safe to import by others)
│   └── tron/                           # Custom Tron Client Wrapper
│       ├── client.go                   # Wrapper around gotron-sdk
│       └── address.go                  # Tron address validation/conversion utils
│
├── migrations/                         # Database Migration Files (SQL)
│
│
├── deploy/                             # Infrastructure as Code
│   ├── terraform/                      # AWS Infrastructure (VPC, RDS, SQS)
│   │   ├── main.tf
│   │   └── variables.tf
│   └── docker/                         # Dockerfiles for local dev/production
│
├── go.mod                              # Go Module definition
├── Makefile                            # Build scripts (build, test, run)
└── README.md                           # This documentation
```

## System Flows

### 0. App Registration & Authentication
1.  **Register Account**: Company registers an `account` with `Auth Service`.
2.  **Issue API Key**: `Auth Service` issues API keys tied to `account_id` (stored hashed in DynamoDB).
3.  **Authenticate Requests**: Clients include `x-api-key`; interceptors/middleware validate and attach `account_id`.

### 1. Deposit Flow (Payin)
1.  **Address Provisioning (Authenticated)**: Client calls `Payin Service` (`CreateAddress`) with valid API key; receives deposit address mapped via KMS within the `account_id` scope.
2.  **User Transfer**: User sends USDT to the provided deposit address.
3.  **Scan**: `Observer Service` polls TronGrid and detects the `Transfer` event.
4.  **Queue**: Event is pushed to `deposit-events.fifo` SQS queue.
5.  **Process**: `Ledger Service` consumes the event.
    *   Checks idempotency (have we processed this TxHash?).
    *   **Transaction**: Debit `System_Hot_Wallet_Asset`, Credit `User_Liability`.
6.  **Notify**: System sends a webhook to the client's backend confirming the deposit.

### 2. Payout Flow (Withdrawal)
1.  **Request (Authenticated)**: Client requests withdrawal via API with API key.
2.  **Validation**: `Payout Service` checks user balance via `Ledger Service`.
3.  **Lock Funds**: `Ledger Service` records a "Pending Withdrawal" transaction.
    *   Debit `User_Liability`, Credit `Clearing_Account`.
4.  **Queue**: Request pushed to `payout-requests.fifo`.
5.  **Sign & Broadcast**: `Payout Service` worker:
    *   Constructs TRC20 transfer.
    *   Signs using **AWS KMS**.
    *   Broadcasts to Tron Network.
6.  **Finalize**:
    *   **Success**: `Ledger Service` moves funds from `Clearing_Account` to `System_Hot_Wallet_Asset` (Credit).
    *   **Failure**: `Ledger Service` refunds user (Debit `Clearing_Account`, Credit `User_Liability`).

## Wallet Architecture

### Roles
1.  **Cold Wallet** (Offline): Ultimate destination for funds. High security.
    *   *Strategy*: Periodic sweeping from Payin addresses.
2.  **Hot Wallet** (Online/KMS): Signs Payout transactions.
    *   *Strategy*: Staked with high **Energy** to make payouts free. Monitored for low balance.
3.  **Gas Wallet** (Online/KMS): Funds Payin addresses with TRX for sweeping.
    *   *Strategy*: Staked with high **Bandwidth** (TRX transfers are cheap).

### Optimization Strategy
*   **Gas Pump**: Payin addresses start with 0 TRX. To sweep USDT, the `Gas Wallet` sends ~15-30 TRX to the Payin address first.
*   **Batched Sweeping**: We do **not** sweep every small deposit. We wait until a Payin address accumulates a threshold (e.g., > $50) to minimize gas costs and operational noise.
*   **Resource Staking**: We stake TRX on Hot and Gas wallets to generate Energy and Bandwidth, aiming for zero-fee operations.

## Database Schema Design

### Relational (PostgreSQL)
- `accounts`: Company account registry (one company = one account)
  - `id (uuid PK)`, `name (text)`, `status (enum: active|suspended|deleted)`, `webhook_url (text|null)`, `created_at`, `updated_at`
- `ledger_accounts`: Double-entry accounts scoped by `account_id`
  - `id (uuid PK)`, `account_id (uuid FK)`, `type (enum: asset|liability|equity|revenue|expense)`, `code (text)`, `name (text)`, `status (enum: active|archived)`, `created_at`
- `journal_entries`: Immutable business events
  - `id (uuid PK)`, `account_id (uuid FK)`, `event_type (enum: payin|payout|adjustment)`, `external_ref (text|null)`, `created_at`
- `postings`: Line items for `journal_entries` (double-entry)
  - `id (uuid PK)`, `journal_entry_id (uuid FK)`, `ledger_account_id (uuid FK)`, `currency (text)`, `amount_minor (bigint)`, `side (enum: debit|credit)`, `created_at`
  - Constraint: Sum(amount_minor where debit) = Sum(amount_minor where credit) per `journal_entry_id`
- `balances`: Materialized balances for fast reads (multi-currency)
  - `ledger_account_id (uuid)`, `currency (text)`, `amount_minor (bigint)`, `updated_at`
  - Primary Key: `(ledger_account_id, currency)`
- `payin_addresses`: Payin addresses and KMS bindings
  - `id (uuid PK)`, `account_id (uuid FK)`, `ledger_account_id (uuid FK|null)`, `tron_address_base58 (text)`, `tron_address_hex (text)`, `kms_key_id (text)`, `status (enum: active|disabled)`, `created_at`
- `payin_requests`: Provisioned payin intents
  - `id (uuid PK)`, `account_id (uuid FK)`, `ledger_account_id (uuid FK|null)`, `address_id (uuid FK)`, `client_reference (text|null)`, `state (enum: pending|active|disabled)`, `created_at`
- `observed_transactions`: Raw chain transactions mapped to addresses
  - `id (uuid PK)`, `account_id (uuid FK)`, `address_id (uuid FK)`, `tx_hash (text unique)`, `block_number (bigint)`, `amount_minor (bigint)`, `status (enum: detected|confirmed)`, `raw_event (jsonb)`, `journal_entry_id (uuid FK|null)`, `created_at`
- `payout_requests`: Outbound transfer lifecycle (payouts)
  - `id (uuid PK)`, `account_id (uuid FK)`, `from_ledger_account_id (uuid FK)`, `to_address_base58 (text)`, `amount_minor (bigint)`, `state (enum: requested|pending_sign|broadcasted|confirmed|failed)`, `tron_tx_hash (text|null)`, `created_at`, `updated_at`
- `system_wallets`: Internal wallet configuration
  - `id (uuid PK)`, `role (enum: hot|cold|gas)`, `address (text)`, `kms_key_id (text|null)`, `status (enum: active|inactive)`, `created_at`
- `chain_cursors`: Blockchain scanning checkpoints
  - `id (text PK)`, `last_block_number (bigint)`, `updated_at`


Indexes & constraints (essentials):
- Unique: `observed_transactions(tx_hash)`
- Unique: `ledger_accounts(account_id, code)`
- Index: `payin_addresses(account_id, tron_address_hex)`
- Primary Key: `balances(ledger_account_id, currency)`
- Amounts stored as `amount_minor (bigint)` in minor units per currency

### NoSQL (DynamoDB)
High-volume operational data with TTL support. Tables are created via Terraform in `deploy/terraform/dynamodb.tf`.

- `api_keys` (Auth Service)
  - PK: `api_key_hash`, GSI: `account_id`
  - Attributes: `account_id`, `name`, `permissions`, `status`, `last_used_at`, `expires_at`, `ttl`
  - Purpose: API authentication with auto-expiration
  
- `idempotency_keys` (All Services)
  - PK: `key`, GSI: `account_id`
  - Attributes: `account_id`, `scope`, `request_hash`, `response`, `status`, `created_at`, `ttl`
  - Purpose: Prevent duplicate request processing (24-hour TTL)
  
- `webhook_events` (Notification Service)
  - PK: `account_id`, SK: `event_id`, GSI: `status` + `next_retry_at`
  - Attributes: `event_type`, `payload`, `status`, `attempts`, `last_attempt_at`, `last_error`, `next_retry_at`, `ttl`
  - Purpose: Track webhook delivery with retry logic (30-day TTL)
  
- `audit_logs` (Compliance)
  - PK: `account_id`, SK: `timestamp`
  - Attributes: `action`, `actor_id`, `actor_type`, `data`, `ip_address`, `ttl`
  - Purpose: Audit trail of admin/system actions (90-day TTL)
