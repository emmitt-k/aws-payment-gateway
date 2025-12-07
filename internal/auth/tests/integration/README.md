# Integration Tests

This directory contains integration tests for the auth service that verify end-to-end functionality with real database connections.

## Structure

```
integration/
├── auth_integration_test.go    # Main integration test suite
├── config/
│   └── config.go            # Test configuration
├── fixtures/
│   ├── accounts.go           # Test fixtures for accounts and API keys
│   └── cleanup.go            # Test data cleanup utilities
└── utils/
    ├── dynamodb_setup.go     # DynamoDB test setup
    └── test_utils.go         # Test utilities and base suite
```

## Running Tests Locally

### Prerequisites

1. Docker and Docker Compose installed
2. Go 1.23 or later
3. Make tool available

### Quick Start

```bash
# Start test dependencies
make docker-test-up

# Run integration tests
make test-integration

# Clean up test containers
make docker-test-down
```

### Manual Setup

If you prefer to run tests manually without Makefile:

1. Start PostgreSQL and DynamoDB Local:
```bash
docker-compose -f docker-compose.test.yml up -d
```

2. Wait for services to be ready (approximately 10-15 seconds)

3. Run tests with environment variables:
```bash
TEST_POSTGRES_HOST=localhost \
TEST_POSTGRES_PORT=5433 \
TEST_POSTGRES_USER=test_user \
TEST_POSTGRES_PASSWORD=test_password \
TEST_POSTGRES_DB=test_payment_gateway \
TEST_DYNAMODB_ENDPOINT=http://localhost:8001 \
TEST_DYNAMODB_REGION=us-west-2 \
TEST_DYNAMODB_TABLE=test-auth-service \
TEST_AUDIT_LOGS_TABLE=test-audit_logs \
TEST_TIMEOUT=30 \
TEST_CLEANUP=true \
go test -v -race -tags=integration ./internal/auth/tests/integration/... -coverprofile=coverage-integration.out
```

## Test Coverage

The integration tests cover:

- Account registration and validation
- API key issuance, validation, and revocation
- Authentication middleware
- Audit logging verification
- Database persistence
- Concurrency scenarios
- Error handling and edge cases

## Test Data Management

### Fixtures

Test fixtures are provided in `fixtures/accounts.go`:
- Valid account requests
- Invalid account requests (for error testing)
- Valid API key requests
- Invalid API key requests (for error testing)

### Cleanup

Test data is automatically cleaned up after each test run. The cleanup utilities in `fixtures/cleanup.go` handle:
- PostgreSQL table cleanup
- DynamoDB table cleanup
- Account-specific cleanup
- Audit log cleanup

## Troubleshooting

### Common Issues

1. **Connection refused errors**
   - Ensure Docker containers are running: `docker-compose -f docker-compose.test.yml ps`
   - Check port conflicts (5433 for PostgreSQL, 8001 for DynamoDB)

2. **Timeout errors**
   - Increase TEST_TIMEOUT environment variable
   - Check system resources (CPU, memory)

3. **Migration failures**
   - Verify PostgreSQL is fully ready before running tests
   - Check migration files in `migrations/` directory

4. **DynamoDB table creation failures**
   - Ensure DynamoDB Local is running
   - Check AWS credentials (can use dummy values for local testing)

### Debug Mode

For detailed debugging, you can:
1. Add `-v` flag for verbose test output
2. Check Docker logs: `docker-compose -f docker-compose.test.yml logs`
3. Examine test coverage: `go tool cover -html=coverage-integration.out`

## CI/CD Integration

The integration tests are automatically run in GitHub Actions on:
- Push to main/develop branches
- Pull requests to main branch

See `.github/workflows/integration-tests.yml` for the complete CI configuration.