package utils

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aws-payment-gateway/internal/auth/domain"
)

// TestContext returns a context with timeout for testing
func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// RequireNoError is a helper to require no error and fail the test if there is one
func RequireNoError(t *testing.T, err error) {
	require.NoError(t, err)
}

// RequireError is a helper to require an error and fail the test if there isn't one
func RequireError(t *testing.T, err error) {
	require.Error(t, err)
}

// RequireEqual is a helper to require equality and fail the test if not equal
func RequireEqual(t *testing.T, expected, actual interface{}) {
	require.Equal(t, expected, actual)
}

// RequireNotNil is a helper to require non-nil value and fail the test if nil
func RequireNotNil(t *testing.T, value interface{}) {
	require.NotNil(t, value)
}

// RequireNil is a helper to require nil value and fail the test if not nil
func RequireNil(t *testing.T, value interface{}) {
	require.Nil(t, value)
}

// CreateTestAccount creates a test account with default values
func CreateTestAccount(t *testing.T) *domain.Account {
	accountID := uuid.New()
	return &domain.Account{
		ID:         accountID,
		Name:       "test-account-" + accountID.String()[:8],
		Status:     domain.AccountStatusActive,
		WebhookURL: stringPtr("https://example.com/webhook"),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// CreateTestApiKey creates a test API key with default values
func CreateTestApiKey(t *testing.T, accountID uuid.UUID) *domain.ApiKey {
	apiKeyID := uuid.New()
	return &domain.ApiKey{
		ID:          apiKeyID,
		AccountID:   accountID,
		Name:        "test-api-key-" + apiKeyID.String()[:8],
		KeyHash:     "test-key-hash-" + apiKeyID.String(),
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts, domain.PermissionReadKeys},
		Status:      domain.ApiKeyStatusActive,
		LastUsedAt:  nil,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}
}

// CreateExpiredTestApiKey creates an expired test API key
func CreateExpiredTestApiKey(t *testing.T, accountID uuid.UUID) *domain.ApiKey {
	apiKeyID := uuid.New()
	return &domain.ApiKey{
		ID:          apiKeyID,
		AccountID:   accountID,
		Name:        "expired-api-key-" + apiKeyID.String()[:8],
		KeyHash:     "expired-key-hash-" + apiKeyID.String(),
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
		Status:      domain.ApiKeyStatusActive,
		LastUsedAt:  nil,
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt:   time.Now().Add(-2 * time.Hour),
	}
}

// CreateInactiveTestApiKey creates an inactive test API key
func CreateInactiveTestApiKey(t *testing.T, accountID uuid.UUID) *domain.ApiKey {
	apiKeyID := uuid.New()
	return &domain.ApiKey{
		ID:          apiKeyID,
		AccountID:   accountID,
		Name:        "inactive-api-key-" + apiKeyID.String()[:8],
		KeyHash:     "inactive-key-hash-" + apiKeyID.String(),
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
		Status:      domain.ApiKeyStatusInactive,
		LastUsedAt:  nil,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// timePtr returns a pointer to a time
func timePtr(t time.Time) *time.Time {
	return &t
}

// AssertAccountEquals compares two accounts for equality
func AssertAccountEquals(t *testing.T, expected, actual *domain.Account) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Status, actual.Status)
	require.Equal(t, expected.WebhookURL, actual.WebhookURL)
	// Don't compare timestamps as they might differ by milliseconds
	require.NotNil(t, actual.CreatedAt)
	require.NotNil(t, actual.UpdatedAt)
}

// AssertApiKeyEquals compares two API keys for equality
func AssertApiKeyEquals(t *testing.T, expected, actual *domain.ApiKey) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.AccountID, actual.AccountID)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.KeyHash, actual.KeyHash)
	require.Equal(t, expected.Permissions, actual.Permissions)
	require.Equal(t, expected.Status, actual.Status)
	require.Equal(t, expected.LastUsedAt, actual.LastUsedAt)
	// Don't compare timestamps as they might differ by milliseconds
	require.NotNil(t, actual.ExpiresAt)
	require.NotNil(t, actual.CreatedAt)
}

// MockPostgreSQLClient is a mock implementation of PostgreSQLClient for testing
type MockPostgreSQLClient struct {
	// Mock methods that implement the PostgreSQLClient interface
}

func (m *MockPostgreSQLClient) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return &MockSQLResult{}, nil
}

func (m *MockPostgreSQLClient) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Return nil since we can't easily mock sql.Row
	// In a real implementation, we'd use interfaces or different testing approach
	return nil
}

func (m *MockPostgreSQLClient) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Return nil rows since we can't mock sql.Rows directly
	// This is a limitation of the current design - in practice, we'd use interfaces
	return nil, nil
}

// MockSQLResult implements sql.Result for testing
type MockSQLResult struct{}

func (m *MockSQLResult) LastInsertId() (int64, error) { return 1, nil }
func (m *MockSQLResult) RowsAffected() (int64, error) { return 1, nil }

// MockSQLRow implements sql.Row behavior for testing
type MockSQLRow struct{}

func (m *MockSQLRow) Scan(dest ...interface{}) error {
	return nil
}

// MockSQLRows implements sql.Rows behavior for testing
type MockSQLRows struct{}

func (m *MockSQLRows) Close() error                   { return nil }
func (m *MockSQLRows) Columns() ([]string, error)     { return []string{}, nil }
func (m *MockSQLRows) Next() bool                     { return false }
func (m *MockSQLRows) Err() error                     { return nil }
func (m *MockSQLRows) Scan(dest ...interface{}) error { return nil }
