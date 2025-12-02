package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/tests/utils"
)

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

func TestAccount_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   domain.AccountStatus
		expected bool
	}{
		{
			name:     "active account is valid",
			status:   domain.AccountStatusActive,
			expected: true,
		},
		{
			name:     "suspended account is invalid",
			status:   domain.AccountStatusSuspended,
			expected: false,
		},
		{
			name:     "deleted account is invalid",
			status:   domain.AccountStatusDeleted,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &domain.Account{
				ID:     uuid.New(),
				Name:   "test-account",
				Status: tt.status,
			}

			result := account.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAccount_Fields(t *testing.T) {
	accountID := uuid.New()
	webhookURL := "https://example.com/webhook"
	now := time.Now()

	account := &domain.Account{
		ID:         accountID,
		Name:       "test-account",
		Status:     domain.AccountStatusActive,
		WebhookURL: &webhookURL,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Test all fields are set correctly
	utils.RequireEqual(t, accountID, account.ID)
	utils.RequireEqual(t, "test-account", account.Name)
	utils.RequireEqual(t, domain.AccountStatusActive, account.Status)
	utils.RequireEqual(t, &webhookURL, account.WebhookURL)
	utils.RequireEqual(t, now, account.CreatedAt)
	utils.RequireEqual(t, now, account.UpdatedAt)
}

func TestAccount_WithNilWebhookURL(t *testing.T) {
	account := &domain.Account{
		ID:         uuid.New(),
		Name:       "test-account",
		Status:     domain.AccountStatusActive,
		WebhookURL: nil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	utils.RequireNil(t, account.WebhookURL)
	utils.RequireEqual(t, true, account.IsValid())
}

func TestAccount_Constants(t *testing.T) {
	utils.RequireEqual(t, domain.AccountStatus("active"), domain.AccountStatusActive)
	utils.RequireEqual(t, domain.AccountStatus("suspended"), domain.AccountStatusSuspended)
	utils.RequireEqual(t, domain.AccountStatus("deleted"), domain.AccountStatusDeleted)
}

func TestAccount_Creation(t *testing.T) {
	// Test creating a valid account
	account := utils.CreateTestAccount(t)

	utils.RequireNotNil(t, account)
	utils.RequireEqual(t, true, account.IsValid())
	utils.RequireNotNil(t, account.ID)
	utils.RequireEqual(t, domain.AccountStatusActive, account.Status)
	utils.RequireNotNil(t, account.CreatedAt)
	utils.RequireNotNil(t, account.UpdatedAt)
}

func TestAccount_WebhookURLValidation(t *testing.T) {
	tests := []struct {
		name       string
		webhookURL *string
		expected   bool
	}{
		{
			name:       "valid webhook URL",
			webhookURL: stringPtr("https://example.com/webhook"),
			expected:   true,
		},
		{
			name:       "nil webhook URL",
			webhookURL: nil,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &domain.Account{
				ID:         uuid.New(),
				Name:       "test-account",
				Status:     domain.AccountStatusActive,
				WebhookURL: tt.webhookURL,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			// The account should be valid regardless of webhook URL
			// (webhook URL validation is handled at the use case level)
			result := account.IsValid()
			assert.Equal(t, true, result)
		})
	}
}

func TestAccount_TimeFields(t *testing.T) {
	beforeCreation := time.Now()
	account := utils.CreateTestAccount(t)
	afterCreation := time.Now()

	// Verify timestamps are within reasonable range
	require.True(t, account.CreatedAt.After(beforeCreation) || account.CreatedAt.Equal(beforeCreation))
	require.True(t, account.CreatedAt.Before(afterCreation) || account.CreatedAt.Equal(afterCreation))
	require.True(t, account.UpdatedAt.After(beforeCreation) || account.UpdatedAt.Equal(beforeCreation))
	require.True(t, account.UpdatedAt.Before(afterCreation) || account.UpdatedAt.Equal(afterCreation))
}

func TestAccount_Equality(t *testing.T) {
	account1 := utils.CreateTestAccount(t)
	account2 := utils.CreateTestAccount(t)

	// Different accounts should not be equal
	require.NotEqual(t, account1.ID, account2.ID)
	require.NotEqual(t, account1.Name, account2.Name)

	// Test equality helper
	accountCopy := &domain.Account{
		ID:         account1.ID,
		Name:       account1.Name,
		Status:     account1.Status,
		WebhookURL: account1.WebhookURL,
		CreatedAt:  account1.CreatedAt,
		UpdatedAt:  account1.UpdatedAt,
	}

	utils.AssertAccountEquals(t, account1, accountCopy)
}
