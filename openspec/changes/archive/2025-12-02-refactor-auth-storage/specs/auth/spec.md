## MODIFIED Requirements
### Requirement: Account Registration
The system SHALL provide account registration functionality for external clients with account data stored in PostgreSQL.

#### Scenario: New account registration
- **WHEN** a company requests to register an account
- **THEN** the system creates a new account with unique identifier and default settings in PostgreSQL

#### Scenario: Account validation
- **WHEN** registering an account with invalid data
- **THEN** the system rejects the registration with appropriate error details

#### Scenario: Account persistence
- **WHEN** storing or retrieving account data
- **THEN** the system uses PostgreSQL as the primary storage backend

### Requirement: API Key Management
The system SHALL provide API key issuance and lifecycle management for authenticated access with API keys stored in DynamoDB.

#### Scenario: API key generation
- **WHEN** an account owner requests a new API key
- **THEN** the system generates a secure API key with specified permissions and expiration in DynamoDB

#### Scenario: API key validation
- **WHEN** a client makes a request with an API key
- **THEN** the system validates the key from DynamoDB, checks expiration, and returns account context from PostgreSQL

#### Scenario: API key revocation
- **WHEN** an account owner revokes an API key
- **THEN** the system immediately invalidates the key in DynamoDB for all subsequent requests

### Requirement: Secure Key Storage
The system SHALL store API keys securely in DynamoDB and account data in PostgreSQL with proper access controls.

#### Scenario: Key hashing
- **WHEN** storing API keys in DynamoDB
- **THEN** the system hashes keys using secure algorithm before storage

#### Scenario: Account data storage
- **WHEN** storing account information
- **THEN** the system persists account data in PostgreSQL with proper constraints

#### Scenario: Key lookup
- **WHEN** validating an API request
- **THEN** the system performs efficient lookup by hashed key in DynamoDB

#### Scenario: Account lookup
- **WHEN** retrieving account information
- **THEN** the system queries PostgreSQL for account data