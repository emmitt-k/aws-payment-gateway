# Change: Add Unit Tests for Auth Service

## Why
The auth service currently lacks unit tests, which increases the risk of regressions and makes it difficult to verify the correctness of authentication logic. Adding comprehensive unit tests will improve code reliability, make refactoring safer, and provide better documentation of expected behavior.

## What Changes
- Create a dedicated `tests` directory inside the auth service (`internal/auth/tests/`)
- Add unit tests for all use cases, domain entities, and repository implementations
- Add test utilities and mocks for external dependencies
- Implement table-driven tests following Go conventions
- Add test coverage reporting

## Impact
- Affected specs: auth
- Affected code: internal/auth/ (all modules)
- New files: internal/auth/tests/ directory with test files