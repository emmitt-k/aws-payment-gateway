# Project Context

## Purpose
A cloud-native, custodial payment gateway for Tron USDT (TRC20) built on AWS. This system manages user wallets, processes deposits (payins) and withdrawals (payouts), and ensures financial integrity through double-entry bookkeeping.

## Tech Stack
- **Language**: Go (Golang) 1.22+
- **Frameworks**: `gofiber/fiber` (HTTP Router), `gorm` (ORM)
- **Migrations**: `golang-migrate` (CLI/Lib)
- **Communication**:
    - **HTTP (REST)**: For synchronous internal requests (e.g., `GetBalance`, `GetDepositAddress`)
    - **SQS**: For asynchronous, durable event processing (e.g., `DepositDetected`, `PayoutRequested`)
- **Database**:
    - **PostgreSQL (RDS)**: Core relational data (Accounts, Transactions, Ledger)
    - **DynamoDB**: High-volume audit logs, raw webhook events, and idempotency keys
- **Messaging**: Amazon SQS
- **Blockchain Interaction**: Native Go Client (e.g., `gotron-sdk`) or direct HTTP to TronGrid
- **Infrastructure**: AWS (VPC, EC2, ASG, ALB, KMS, Secrets Manager)
- **Observability**:
    - **Tracing**: AWS X-Ray
    - **Metrics**: Amazon Managed Service for Prometheus (AMP)
    - **Dashboards**: Amazon Managed Grafana (AMG)
    - **Logs**: Amazon CloudWatch Logs
- **CI/CD**: GitHub Actions

## Project Conventions

### Code Style
- Follow Go standard formatting and naming conventions
- Use Clean Architecture principles with clear separation between entities, use cases, and adapters
- Organize code in a monorepo structure with shared utilities
- Implement structured logging throughout the system

### Architecture Patterns
- **Microservices in a Monorepo**: Multiple services in a single repository with shared code
- **Clean Architecture**: Each service follows layered architecture with entities, use cases, interfaces/adapters, and drivers
- **Domain-Driven Design**: Clear domain boundaries for Auth, Payin, Observer, Ledger, and Payout services
- **Event-Driven Architecture**: SQS queues for asynchronous processing of deposits and payouts

### Testing Strategy
- Unit tests for business logic in use cases and entities
- Integration tests for adapters and repositories
- End-to-end tests for critical flows (deposit and payout)
- Mock external dependencies (TronGrid, AWS services) in test environments

### Git Workflow
- Feature branches for new development
- Pull requests for code review
- Semantic versioning for releases
- Automated CI/CD pipeline with GitHub Actions

## Domain Context
- **Custodial Payment Gateway**: System holds and manages user funds on their behalf
- **Tron USDT (TRC20)**: Specific focus on USDT transactions on the Tron blockchain
- **Double-Entry Bookkeeping**: Financial integrity through proper accounting principles
- **Wallet Management**: Hierarchical wallet system with Cold, Hot, and Gas wallets
- **Microservices**: Auth, Payin, Observer, Ledger, and Payout services working together
- **Blockchain Events**: Monitoring and processing Tron blockchain transfer events
- **API Authentication**: API key-based authentication for client applications

## Important Constraints
- **Security**: Private keys managed through AWS KMS, never exposed in code
- **Financial Integrity**: All balance changes must follow double-entry accounting principles
- **Scalability**: Auto-scaling architecture to handle high transaction throughput
- **Idempotency**: All operations must be idempotent to prevent duplicate processing
- **Regulatory Compliance**: Audit trails and proper financial record-keeping
- **Gas Optimization**: Strategic staking of TRX for Energy and Bandwidth to minimize transaction costs

## External Dependencies
- **AWS Services**: VPC, EC2, ASG, ALB, RDS, DynamoDB, SQS, KMS, Secrets Manager, X-Ray, CloudWatch
- **TronGrid**: External API for blockchain interaction and event monitoring
- **Tron Network**: Main blockchain network for USDT transactions
- **Client Webhooks**: External endpoints for transaction notifications
- **gotron-sdk**: Go SDK for Tron blockchain interaction
