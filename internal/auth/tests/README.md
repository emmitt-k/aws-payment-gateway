# Auth Service Tests

This directory contains comprehensive unit tests for the auth service, organized by layer:

## Directory Structure

```
tests/
├── domain/           # Domain entity tests
│   ├── account_test.go
│   └── apikey_test.go
├── usecase/          # Use case tests
│   ├── register_app_test.go
│   ├── issue_api_key_test.go
│   ├── validate_api_key_test.go
│   ├── revoke_api_key_test.go
│   └── get_api_keys_test.go
├── repository/        # Repository tests (when implemented)
├── adapter/           # Adapter tests (when implemented)
│   ├── http/
│   └── dto/
├── mocks/            # Mock implementations
│   ├── mock_app_repository.go
│   └── mock_apikey_repository.go
└── utils/            # Test utilities and helpers
    └── test_helpers.go
```

## Running Tests

To run all auth service tests:

```bash
go test ./internal/auth/tests/...
```

To run tests for a specific layer:

```bash
# Domain tests
go test ./internal/auth/tests/domain/...

# Use case tests
go test ./internal/auth/tests/usecase/...

# Repository tests
go test ./internal/auth/tests/repository/...

# Adapter tests
go test ./internal/auth/tests/adapter/...
```

## Test Coverage

To generate test coverage report:

```bash
go test -coverprofile=coverage.out ./internal/auth/tests/...
go tool cover -html=coverage.html coverage.out
```

The coverage report will be generated in `coverage.html`.

## Test Guidelines

1. **Table-driven tests**: All tests use table-driven approach for comprehensive scenario testing
2. **Mock dependencies**: External dependencies are mocked to isolate units under test
3. **Arrange-Act-Assert pattern**: Tests follow clear structure for readability
4. **Edge cases**: Tests cover success cases, error cases, and edge conditions
5. **Helpers**: Common test utilities are centralized in `utils/test_helpers.go`

## Test Categories

### Domain Tests
- Entity validation logic
- Business method behavior
- Status and permission checks

### Use Case Tests
- Input validation
- Business logic flows
- Error handling
- Integration with repositories

### Repository Tests
- CRUD operations
- Query methods
- Error scenarios

### Adapter Tests
- HTTP request/response handling
- DTO validation
- Middleware behavior

## Mock Implementations

The `mocks/` directory provides mock implementations of:
- `AppRepository` - For account persistence
- `ApiKeyRepository` - For API key persistence

These mocks include helper methods for:
- Setting up test data
- Simulating errors
- Verifying interactions