# Change: Fix Auth Service Gaps and Missing Features

## Why
The current auth service implementation has several critical gaps that prevent full compliance with openspec specifications, including inefficient API key validation, missing TTL implementation, and security vulnerabilities in key lookup patterns.

## What Changes
- Fix API key validation flow to accept raw keys instead of pre-hashed keys
- Implement efficient DynamoDB query patterns for API key lookups
- Add TTL configuration for automatic API key expiration
- Implement idempotency key management for auth operations
- Add rate limiting for authentication endpoints
- Enhance audit logging with DynamoDB integration
- Implement proper error handling with specific error codes

## Impact
- Affected specs: auth, dynamodb
- Affected code: internal/auth/usecase/validate_api_key.go, internal/auth/repository/dynamodb_apikey.go, internal/auth/adapter/http/middleware.go
- Breaking changes: API key validation flow will change (middleware will remain compatible)