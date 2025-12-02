.PHONY: test test-coverage test-coverage-html clean-coverage

# Default target
all: test

# Run all tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...

# Generate HTML coverage report
test-coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html

# Show coverage in terminal
test-coverage-show: test-coverage
	go tool cover -func=coverage.out

# Run tests with coverage threshold check (80% minimum)
test-coverage-threshold:
	go test -coverprofile=coverage.out ./...
	@echo "Checking coverage threshold..."
	@go tool cover -func=coverage.out | grep "total:" | awk '{if ($$3+0 < 80.0) {print "Coverage " $$3 " is below 80% threshold"; exit 1} else {print "Coverage " $$3 " meets 80% threshold"}}'

# Clean coverage files
clean-coverage:
	rm -f coverage.out coverage.html

# Run tests for specific layers
test-domain:
	go test ./domain/...

test-usecase:
	go test ./usecase/...

test-repository:
	go test ./repository/...

test-adapter:
	go test ./adapter/...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -race ./...