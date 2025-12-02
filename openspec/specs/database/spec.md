# database Specification

## Purpose
TBD - created by archiving change add-database-migrations. Update Purpose after archive.
## Requirements
### Requirement: Database Schema Management
The system SHALL provide database migration files to establish and manage the PostgreSQL schema for the payment gateway.

#### Scenario: Initial database setup
- **WHEN** the system is first deployed
- **THEN** migration files create all required tables with proper constraints and indexes

#### Scenario: Schema evolution
- **WHEN** database schema needs to be updated
- **THEN** new migration files can be applied to evolve the schema without data loss

### Requirement: Core Financial Tables
The system SHALL provide database tables for double-entry bookkeeping and financial transaction management, including account management.

#### Scenario: Account management
- **WHEN** storing company account information
- **THEN** accounts table in PostgreSQL stores id, name, status, webhook_url, and timestamps

#### Scenario: Account authentication integration
- **WHEN** validating API requests
- **THEN** auth service queries PostgreSQL for account data while using DynamoDB for API key validation

#### Scenario: Ledger account management
- **WHEN** managing double-entry accounts
- **THEN** ledger_accounts table stores account-scoped ledger entries with type, code, and status

#### Scenario: Transaction recording
- **WHEN** recording financial transactions
- **THEN** journal_entries and postings tables store immutable business events with proper debit/credit balancing

### Requirement: Payment Processing Tables
The system SHALL provide database tables for managing payin and payout operations.

#### Scenario: Deposit address management
- **WHEN** generating deposit addresses for users
- **THEN** payin_addresses table stores address mappings with KMS key bindings

#### Scenario: Transaction observation
- **WHEN** monitoring blockchain transactions
- **THEN** observed_transactions table stores detected transfers with status tracking

#### Scenario: Payout processing
- **WHEN** processing withdrawal requests
- **THEN** payout_requests table tracks withdrawal lifecycle with state management

### Requirement: System Configuration Tables
The system SHALL provide database tables for system configuration and operational data.

#### Scenario: Wallet management
- **WHEN** managing system wallets
- **THEN** system_wallets table stores wallet roles and KMS key associations

#### Scenario: Blockchain scanning
- **WHEN** tracking blockchain scanning progress
- **THEN** chain_cursors table stores scanning checkpoints

### Requirement: Database Constraints and Performance
The system SHALL provide proper constraints and indexes for data integrity and performance, including account authentication queries.

#### Scenario: Data integrity
- **WHEN** inserting or updating account data
- **THEN** foreign key, unique, and check constraints in PostgreSQL ensure data consistency

#### Scenario: Query performance
- **WHEN** querying financial data or account information
- **THEN** indexes optimize frequently accessed queries and joins in PostgreSQL

#### Scenario: Authentication performance
- **WHEN** validating API requests
- **THEN** PostgreSQL account queries are optimized for frequent auth service lookups

