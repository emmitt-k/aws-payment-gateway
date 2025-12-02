## MODIFIED Requirements
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