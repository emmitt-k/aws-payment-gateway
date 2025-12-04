# Change: Remove File Logger from Auth Service

## Why
The auth service currently uses a dual logging approach with both file-based logging and DynamoDB logging. The file logger serves as a backup but adds complexity, storage overhead, and operational burden. Removing the file logger simplifies the architecture, reduces dependencies, and aligns with the cloud-native approach of using DynamoDB as the single source of truth for audit logs.

## What Changes
- Remove file-based audit logger from the auth service
- Update DynamoDB logger to remove file logger backup functionality
- Modify audit logging to use only DynamoDB for persistence
- Remove file logger dependencies and imports
- **BREAKING**: File-based audit logs will no longer be generated

## Impact
- Affected specs: auth
- Affected code: internal/auth/audit/logger.go, internal/auth/audit/dynamodb_logger.go, cmd/auth-svc/main.go
- Operational impact: Reduced local disk usage and simplified log management
- Security impact: All audit logs will be centrally stored in DynamoDB with proper TTL and access controls