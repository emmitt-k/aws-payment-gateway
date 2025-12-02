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

func TestApiKey_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		status    domain.ApiKeyStatus
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "active and not expired key is valid",
			status:    domain.ApiKeyStatusActive,
			expiresAt: time.Now().Add(1 * time.Hour),
			expected:  true,
		},
		{
			name:      "inactive key is invalid",
			status:    domain.ApiKeyStatusInactive,
			expiresAt: time.Now().Add(1 * time.Hour),
			expected:  false,
		},
		{
			name:      "expired key is invalid",
			status:    domain.ApiKeyStatusActive,
			expiresAt: time.Now().Add(-1 * time.Hour),
			expected:  false,
		},
		{
			name:      "inactive and expired key is invalid",
			status:    domain.ApiKeyStatusInactive,
			expiresAt: time.Now().Add(-1 * time.Hour),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey := &domain.ApiKey{
				ID:        uuid.New(),
				AccountID: uuid.New(),
				Name:      "test-key",
				KeyHash:   "test-hash",
				Status:    tt.status,
				ExpiresAt: tt.expiresAt,
			}

			result := apiKey.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApiKey_HasPermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions domain.ApiKeyPermissions
		permission  string
		expected    bool
	}{
		{
			name:        "has permission when it exists in list",
			permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts, domain.PermissionWriteKeys},
			permission:  domain.PermissionReadAccounts,
			expected:    true,
		},
		{
			name:        "has permission when it's the only one",
			permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
			permission:  domain.PermissionReadAccounts,
			expected:    true,
		},
		{
			name:        "does not have permission when not in list",
			permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
			permission:  domain.PermissionWriteKeys,
			expected:    false,
		},
		{
			name:        "empty permissions list has no permissions",
			permissions: domain.ApiKeyPermissions{},
			permission:  domain.PermissionReadAccounts,
			expected:    false,
		},
		{
			name:        "case sensitive permission check",
			permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
			permission:  "read:accounts", // exact match
			expected:    true,
		},
		{
			name:        "case sensitive permission check fails",
			permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
			permission:  "Read:Accounts", // different case
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey := &domain.ApiKey{
				ID:          uuid.New(),
				AccountID:   uuid.New(),
				Name:        "test-key",
				KeyHash:     "test-hash",
				Permissions: tt.permissions,
				Status:      domain.ApiKeyStatusActive,
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			}

			result := apiKey.HasPermission(tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApiKey_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "not expired when expires in future",
			expiresAt: time.Now().Add(1 * time.Hour),
			expected:  false,
		},
		{
			name:      "expired when expires in past",
			expiresAt: time.Now().Add(-1 * time.Hour),
			expected:  true,
		},
		{
			name:      "expired when expires exactly now",
			expiresAt: time.Now(),
			expected:  true, // IsExpired uses After(), so equal time is considered expired
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey := &domain.ApiKey{
				ID:        uuid.New(),
				AccountID: uuid.New(),
				Name:      "test-key",
				KeyHash:   "test-hash",
				Status:    domain.ApiKeyStatusActive,
				ExpiresAt: tt.expiresAt,
			}

			result := apiKey.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApiKey_Constants(t *testing.T) {
	utils.RequireEqual(t, domain.ApiKeyStatus("active"), domain.ApiKeyStatusActive)
	utils.RequireEqual(t, domain.ApiKeyStatus("inactive"), domain.ApiKeyStatusInactive)

	utils.RequireEqual(t, "read:accounts", domain.PermissionReadAccounts)
	utils.RequireEqual(t, "write:accounts", domain.PermissionWriteAccounts)
	utils.RequireEqual(t, "read:keys", domain.PermissionReadKeys)
	utils.RequireEqual(t, "write:keys", domain.PermissionWriteKeys)
	utils.RequireEqual(t, "manage:webhooks", domain.PermissionManageWebhooks)
}

func TestApiKey_Creation(t *testing.T) {
	accountID := uuid.New()
	apiKey := utils.CreateTestApiKey(t, accountID)

	utils.RequireNotNil(t, apiKey)
	utils.RequireEqual(t, accountID, apiKey.AccountID)
	utils.RequireEqual(t, domain.ApiKeyStatusActive, apiKey.Status)
	utils.RequireNotNil(t, apiKey.ID)
	utils.RequireNotNil(t, apiKey.CreatedAt)
	utils.RequireNotNil(t, apiKey.ExpiresAt)
}

func TestApiKey_ExpiredCreation(t *testing.T) {
	accountID := uuid.New()
	apiKey := utils.CreateExpiredTestApiKey(t, accountID)

	utils.RequireNotNil(t, apiKey)
	utils.RequireEqual(t, accountID, apiKey.AccountID)
	utils.RequireEqual(t, domain.ApiKeyStatusActive, apiKey.Status)
	utils.RequireEqual(t, true, apiKey.IsExpired())
	utils.RequireEqual(t, false, apiKey.IsValid()) // Expired keys are invalid
}

func TestApiKey_InactiveCreation(t *testing.T) {
	accountID := uuid.New()
	apiKey := utils.CreateInactiveTestApiKey(t, accountID)

	utils.RequireNotNil(t, apiKey)
	utils.RequireEqual(t, accountID, apiKey.AccountID)
	utils.RequireEqual(t, domain.ApiKeyStatusInactive, apiKey.Status)
	utils.RequireEqual(t, false, apiKey.IsExpired()) // Not expired, just inactive
	utils.RequireEqual(t, false, apiKey.IsValid())   // Inactive keys are invalid
}

func TestApiKey_TimeFields(t *testing.T) {
	beforeCreation := time.Now()
	apiKey := utils.CreateTestApiKey(t, uuid.New())
	afterCreation := time.Now()

	// Verify timestamps are within reasonable range
	require.True(t, apiKey.CreatedAt.After(beforeCreation) || apiKey.CreatedAt.Equal(beforeCreation))
	require.True(t, apiKey.CreatedAt.Before(afterCreation) || apiKey.CreatedAt.Equal(afterCreation))
	require.True(t, apiKey.ExpiresAt.After(beforeCreation))
	require.True(t, apiKey.ExpiresAt.After(apiKey.CreatedAt))
}

func TestApiKey_Equality(t *testing.T) {
	accountID := uuid.New()
	apiKey1 := utils.CreateTestApiKey(t, accountID)
	apiKey2 := utils.CreateTestApiKey(t, accountID)

	// Different API keys should not be equal
	require.NotEqual(t, apiKey1.ID, apiKey2.ID)
	require.NotEqual(t, apiKey1.Name, apiKey2.Name)

	// Test equality helper
	apiKeyCopy := &domain.ApiKey{
		ID:          apiKey1.ID,
		AccountID:   apiKey1.AccountID,
		Name:        apiKey1.Name,
		KeyHash:     apiKey1.KeyHash,
		Permissions: apiKey1.Permissions,
		Status:      apiKey1.Status,
		LastUsedAt:  apiKey1.LastUsedAt,
		ExpiresAt:   apiKey1.ExpiresAt,
		CreatedAt:   apiKey1.CreatedAt,
	}

	utils.AssertApiKeyEquals(t, apiKey1, apiKeyCopy)
}

func TestApiKey_PermissionsArray(t *testing.T) {
	accountID := uuid.New()
	apiKey := utils.CreateTestApiKey(t, accountID)

	// Verify permissions are correctly set
	require.Contains(t, apiKey.Permissions, domain.PermissionReadAccounts)
	require.Contains(t, apiKey.Permissions, domain.PermissionReadKeys)
	require.Equal(t, 2, len(apiKey.Permissions))
}

func TestApiKey_LastUsedAt(t *testing.T) {
	accountID := uuid.New()
	apiKey := utils.CreateTestApiKey(t, accountID)

	// Initially LastUsedAt should be nil
	require.Nil(t, apiKey.LastUsedAt)

	// Set LastUsedAt and verify
	now := time.Now()
	apiKey.LastUsedAt = &now
	require.NotNil(t, apiKey.LastUsedAt)
	require.Equal(t, now, *apiKey.LastUsedAt)
}
