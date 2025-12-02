# Change: Implement Auth Service

## Why
We need to implement the Auth Service as the first microservice in the payment gateway system. This service will handle app registration, API key issuance, and request validation for external clients.

## What Changes
- Implement Auth Service with Clean Architecture pattern
- Create domain entities for App and ApiKey
- Build use cases for RegisterApp, IssueApiKey, ValidateApiKey
- Develop HTTP adapters and repository implementations
- Add middleware for API key validation across services
- Integrate with DynamoDB for API key storage

## Impact
- Affected specs: auth (new capability)
- Affected code: cmd/auth-svc/, internal/auth/, pkg/auth/
- Dependencies: DynamoDB tables (api_keys), AWS KMS for key generation