## MODIFIED Requirements

### Requirement: API Authentication Table
The system SHALL provide DynamoDB table for API key authentication and management with optimized query patterns.

#### Scenario: API key validation
- **WHEN** validating API requests
- **THEN** api_keys table provides fast lookup by api_key_hash using GSI without table scans

#### Scenario: API key lifecycle management
- **WHEN** managing API key expiration
- **THEN** TTL attribute automatically expires disabled keys based on expires_at timestamp

#### Scenario: Efficient key retrieval
- **WHEN** accessing API keys by ID or account
- **THEN** composite key structure (PK=ACCOUNT#id, SK=APIKEY#id) enables efficient queries

## ADDED Requirements

### Requirement: Optimized DynamoDB Schema
The system SHALL provide optimized DynamoDB table structure for efficient authentication operations.

#### Scenario: Composite key design
- **WHEN** storing API keys in DynamoDB
- **THEN** the system uses PK=ACCOUNT#id and SK=APIKEY#id for account grouping and direct access

#### Scenario: GSI for key hash lookup
- **WHEN** validating API keys by hash
- **THEN** the system uses GSI1 with gsi1pk=KEYHASH#hash for O(1) lookup performance

#### Scenario: TTL-based expiration
- **WHEN** API keys reach expiration time
- **THEN** the system automatically removes expired items using DynamoDB TTL configuration

### Requirement: Idempotency Key Storage
The system SHALL provide DynamoDB table for preventing duplicate request processing with automatic expiration.

#### Scenario: Duplicate request prevention
- **WHEN** processing identical requests within TTL window
- **THEN** idempotency_keys table prevents duplicate processing using request hash as primary key

#### Scenario: Account-scoped idempotency
- **WHEN** querying requests by account
- **THEN** GSI on account_id enables efficient account-specific request tracking

#### Scenario: Automatic cleanup
- **WHEN** idempotency keys reach 24-hour retention
- **THEN** TTL automatically removes expired keys to control storage costs

### Requirement: Enhanced Audit Logging Storage
The system SHALL provide DynamoDB table for structured audit logging with configurable retention.

#### Scenario: High-volume audit storage
- **WHEN** logging authentication events at scale
- **THEN** audit_logs table provides efficient storage with time-based partitioning

#### Scenario: Compliance retention
- **WHEN** audit logs reach 90-day retention period
- **THEN** TTL automatically removes expired logs while maintaining compliance requirements

#### Scenario: Efficient audit querying
- **WHEN** investigating security incidents
- **THEN** GSIs on event_type and timestamp enable efficient audit log queries

### Requirement: Rate Limiting Storage
The system SHALL provide DynamoDB table for rate limiting authentication operations.

#### Scenario: Distributed rate limiting
- **WHEN** multiple service instances enforce rate limits
- **THEN** rate_limits table provides consistent rate limiting across the distributed system

#### Scenario: Sliding window rate limiting
- **WHEN** implementing sophisticated rate limits
- **THEN** the system supports sliding window algorithms with time-based key expiration

#### Scenario: Rate limit cleanup
- **WHEN** rate limit records expire
- **THEN** TTL automatically removes outdated rate limit records

## MODIFIED Requirements

### Requirement: Performance Optimization
The system SHALL provide optimized DynamoDB configurations for cost-effective authentication operations.

#### Scenario: Query optimization
- **WHEN** accessing frequently queried authentication data
- **THEN** GSIs provide efficient access patterns without full table scans

#### Scenario: Cost management
- **WHEN** managing operational costs
- **THEN** TTL settings and appropriate capacity modes minimize storage and compute costs

#### Scenario: Provisioned capacity planning
- **WHEN** planning for authentication load
- **THEN** the system supports both provisioned and on-demand capacity modes based on usage patterns