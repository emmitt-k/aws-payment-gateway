# Change: Refactor Auth Service Storage Architecture

## Why
The current auth service implementation incorrectly stores both account data and API key data in DynamoDB, which violates the established architectural pattern where account data should be stored in RDS PostgreSQL and API keys should remain in DynamoDB.

## What Changes
- **BREAKING**: Refactor account storage from DynamoDB to RDS PostgreSQL
- Update account repository implementation to use PostgreSQL instead of DynamoDB
- Keep API key storage in DynamoDB as originally intended
- Update auth service to work with mixed storage backends
- Add database migration for accounts table if not already applied

## Impact
- Affected specs: auth, database
- Affected code: internal/auth/repository/, internal/auth/usecase/
- Requires database migration execution
- Requires updates to service initialization and configuration