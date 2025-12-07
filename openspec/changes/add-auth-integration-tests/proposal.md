# Change: Add Integration Tests to Auth Service

## Why
Integration tests are needed to verify that the auth service works correctly with its dependencies (PostgreSQL and DynamoDB) before deploying to AWS infrastructure. Currently, the auth service lacks comprehensive integration testing, which increases the risk of deployment issues and makes it difficult to validate end-to-end functionality.

## What Changes
- Add integration test suite for auth service using local resources (Docker containers for PostgreSQL and DynamoDB Local)
- Implement test scenarios covering all major auth service endpoints
- Create test utilities for setting up and tearing down test environments
- Add mock configurations for AWS services to avoid actual AWS dependencies during testing
- Implement test data fixtures for consistent test scenarios
- Create GitHub Actions workflow to automatically run integration tests on every push and pull request
- Configure CI/CD pipeline with test result reporting and coverage analysis

## Impact
- Affected specs: auth
- Affected code: internal/auth/, cmd/auth-svc/
- New files: internal/auth/tests/integration/
- Dependencies: testcontainers-go, DynamoDB Local, Docker