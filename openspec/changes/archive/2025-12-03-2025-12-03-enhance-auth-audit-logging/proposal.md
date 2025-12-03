# Change: Enhance Auth Service Audit Logging with DynamoDB Integration

## Why
The current auth service implementation has audit logging code but it's not properly integrated with DynamoDB as specified in the requirements. The middleware and handlers are using the basic file-based AuditLogger instead of the DynamoDBAuditLogger, which means audit events are not being stored in DynamoDB for compliance and security monitoring as required.

## What Changes
- Update auth middleware to use DynamoDBAuditLogger instead of basic AuditLogger
- Update auth handlers to use DynamoDBAuditLogger for API key lifecycle events
- Fix DynamoDB audit_logs table schema to match the implementation requirements
- Add proper audit logging for API key creation and revocation events
- Ensure all authentication events are logged to DynamoDB with proper TTL

## Impact
- Affected specs: auth, dynamodb
- Affected code: internal/auth/adapter/http/middleware.go, internal/auth/adapter/http/handlers.go, cmd/auth-svc/main.go
- Compliance: Ensures audit trail is properly stored in DynamoDB for security monitoring