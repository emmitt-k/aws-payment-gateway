## MODIFIED Requirements
### Requirement: Authentication Middleware
The system SHALL provide middleware for API key validation across all services with comprehensive audit logging to DynamoDB.

#### Scenario: Request authentication
- **WHEN** a request includes a valid x-api-key header
- **THEN** the middleware validates the key, attaches account_id to request context, and logs the authentication event to DynamoDB with timestamp and IP address

#### Scenario: Invalid key handling
- **WHEN** a request includes an invalid or expired API key
- **THEN** the middleware returns 401 Unauthorized with appropriate error details and logs the failed authentication attempt to DynamoDB

#### Scenario: Audit logging integration
- **WHEN** any authentication event occurs (success or failure)
- **THEN** the middleware stores the audit event in DynamoDB audit_logs table with proper TTL and partition keys

### Requirement: API Key Management
The system SHALL provide API key issuance and lifecycle management for authenticated access with API keys stored in DynamoDB and comprehensive audit logging.

#### Scenario: API key generation
- **WHEN** an account owner requests a new API key
- **THEN** the system generates a secure API key with specified permissions and expiration in DynamoDB and logs the key creation event to DynamoDB audit_logs

#### Scenario: API key validation
- **WHEN** a client makes a request with an API key
- **THEN** the system validates the raw API key by hashing it and performing efficient lookup in DynamoDB, checks expiration, returns account context from PostgreSQL, and logs the validation attempt to DynamoDB

#### Scenario: API key revocation
- **WHEN** an account owner revokes an API key
- **THEN** the system immediately invalidates the key in DynamoDB for all subsequent requests and logs the revocation event to DynamoDB audit_logs

#### Scenario: API key expiration
- **WHEN** an API key reaches its expiration time
- **THEN** the system automatically removes the key via TTL configuration without manual intervention and logs the expiration event

### Requirement: Enhanced Audit Logging
The system SHALL provide comprehensive audit logging with DynamoDB integration for compliance and security monitoring.

#### Scenario: Structured audit events
- **WHEN** authentication-related actions occur
- **THEN** the system logs structured events to DynamoDB with TTL for automatic retention management using proper partition key format (AUDIT#EVENTTYPE#YYYY-MM-DD)

#### Scenario: Audit log querying
- **WHEN** compliance reports or security investigations are needed
- **THEN** the system provides efficient querying capabilities for audit events within retention period using partition keys and sort keys

#### Scenario: Audit log retention
- **WHEN** audit logs reach retention period (90 days)
- **THEN** the system automatically removes expired logs via TTL configuration

#### Scenario: Failed authentication tracking
- **WHEN** authentication attempts fail for any reason
- **THEN** the system logs detailed failure information including IP address, user agent, and failure reason to DynamoDB for security analysis