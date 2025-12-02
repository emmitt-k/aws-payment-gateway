package dto_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aws-payment-gateway/internal/auth/adapter/http/dto"
	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/tests/utils"
)

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}

func TestRegisterAppRequest_Validate_Success(t *testing.T) {
	// Arrange
	tests := []struct {
		name    string
		request dto.RegisterAppRequest
	}{
		{
			name: "valid request with webhook",
			request: dto.RegisterAppRequest{
				Name:       "test-app",
				WebhookURL: stringPtr("https://example.com/webhook"),
			},
		},
		{
			name: "valid request without webhook",
			request: dto.RegisterAppRequest{
				Name:       "test-app",
				WebhookURL: nil,
			},
		},
		{
			name: "valid request with HTTP webhook",
			request: dto.RegisterAppRequest{
				Name:       "test-app",
				WebhookURL: stringPtr("http://example.com/webhook"),
			},
		},
		{
			name: "valid request with minimum name length",
			request: dto.RegisterAppRequest{
				Name:       "abc",
				WebhookURL: nil,
			},
		},
		{
			name: "valid request with maximum name length",
			request: dto.RegisterAppRequest{
				Name:       string(make([]rune, 100)), // 100 characters
				WebhookURL: stringPtr("https://example.com/webhook"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.request.Validate()

			// Assert
			utils.RequireNoError(t, err)
		})
	}
}

func TestRegisterAppRequest_Validate_Error(t *testing.T) {
	// Arrange
	tests := []struct {
		name          string
		request       dto.RegisterAppRequest
		expectedError string
	}{
		{
			name: "empty name",
			request: dto.RegisterAppRequest{
				Name:       "",
				WebhookURL: nil,
			},
			expectedError: "name is required",
		},
		{
			name: "name too short",
			request: dto.RegisterAppRequest{
				Name:       "ab",
				WebhookURL: nil,
			},
			expectedError: "name must be at least 3 characters",
		},
		{
			name: "name too long",
			request: dto.RegisterAppRequest{
				Name:       string(make([]rune, 101)), // 101 characters
				WebhookURL: nil,
			},
			expectedError: "name must be at most 100 characters",
		},
		{
			name: "invalid webhook URL",
			request: dto.RegisterAppRequest{
				Name:       "test-app",
				WebhookURL: stringPtr("invalid-url"),
			},
			expectedError: "invalid webhook URL",
		},
		{
			name: "empty webhook URL",
			request: dto.RegisterAppRequest{
				Name:       "test-app",
				WebhookURL: stringPtr(""),
			},
			expectedError: "invalid webhook URL",
		},
		{
			name: "webhook URL with spaces",
			request: dto.RegisterAppRequest{
				Name:       "test-app",
				WebhookURL: stringPtr("https://example .com/webhook"),
			},
			expectedError: "invalid webhook URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.request.Validate()

			// Assert
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestIssueApiKeyRequest_Validate_Success(t *testing.T) {
	// Arrange
	tests := []struct {
		name    string
		request dto.IssueApiKeyRequest
	}{
		{
			name: "valid request with expiration",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   intPtr(24),
			},
		},
		{
			name: "valid request without expiration",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   nil,
			},
		},
		{
			name: "valid request with multiple permissions",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{domain.PermissionReadAccounts, domain.PermissionWriteKeys},
				ExpiresIn:   intPtr(1),
			},
		},
		{
			name: "valid request with minimum expiration",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   intPtr(1),
			},
		},
		{
			name: "valid request with maximum expiration",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   intPtr(8760),
			},
		},
		{
			name: "valid request with minimum name length",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "abc",
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   nil,
			},
		},
		{
			name: "valid request with maximum name length",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        string(make([]rune, 100)), // 100 characters
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.request.Validate()

			// Assert
			utils.RequireNoError(t, err)
		})
	}
}

func TestIssueApiKeyRequest_Validate_Error(t *testing.T) {
	// Arrange
	tests := []struct {
		name          string
		request       dto.IssueApiKeyRequest
		expectedError string
	}{
		{
			name: "nil account ID",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.Nil,
				Name:        "test-api-key",
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   nil,
			},
			expectedError: "account_id is required",
		},
		{
			name: "empty name",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "",
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   nil,
			},
			expectedError: "name is required",
		},
		{
			name: "name too short",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "ab",
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   nil,
			},
			expectedError: "name must be at least 3 characters",
		},
		{
			name: "name too long",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        string(make([]rune, 101)), // 101 characters
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   nil,
			},
			expectedError: "name must be at most 100 characters",
		},
		{
			name: "empty permissions",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{},
				ExpiresIn:   nil,
			},
			expectedError: "at least one permission is required",
		},
		{
			name: "nil permissions",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: nil,
				ExpiresIn:   nil,
			},
			expectedError: "at least one permission is required",
		},
		{
			name: "empty permission in list",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{""},
				ExpiresIn:   nil,
			},
			expectedError: "permission cannot be empty",
		},
		{
			name: "mixed empty and valid permissions",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{domain.PermissionReadAccounts, ""},
				ExpiresIn:   nil,
			},
			expectedError: "permission cannot be empty",
		},
		{
			name: "expiration too short",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   intPtr(0),
			},
			expectedError: "expires_in must be at least 1 hour",
		},
		{
			name: "expiration too long",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{domain.PermissionReadAccounts},
				ExpiresIn:   intPtr(8761),
			},
			expectedError: "expires_in must be at most 8760 hours (1 year)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.request.Validate()

			// Assert
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestValidateApiKeyRequest_Validate_Success(t *testing.T) {
	// Arrange
	tests := []struct {
		name    string
		request dto.ValidateApiKeyRequest
	}{
		{
			name: "valid request with hash",
			request: dto.ValidateApiKeyRequest{
				KeyHash: "valid-key-hash",
			},
		},
		{
			name: "valid request with long hash",
			request: dto.ValidateApiKeyRequest{
				KeyHash: string(make([]rune, 100)), // 100 characters
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.request.Validate()

			// Assert
			utils.RequireNoError(t, err)
		})
	}
}

func TestValidateApiKeyRequest_Validate_Error(t *testing.T) {
	// Arrange
	tests := []struct {
		name          string
		request       dto.ValidateApiKeyRequest
		expectedError string
	}{
		{
			name: "empty key hash",
			request: dto.ValidateApiKeyRequest{
				KeyHash: "",
			},
			expectedError: "key_hash is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.request.Validate()

			// Assert
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestDTO_Structure(t *testing.T) {
	// Test ErrorResponse structure
	t.Run("ErrorResponse", func(t *testing.T) {
		errorResp := dto.ErrorResponse{
			Error:   "test_error",
			Message: "Test error message",
			Details: "Test error details",
		}

		require.Equal(t, "test_error", errorResp.Error)
		require.Equal(t, "Test error message", errorResp.Message)
		require.Equal(t, "Test error details", errorResp.Details)
	})

	// Test RegisterAppResponse structure
	t.Run("RegisterAppResponse", func(t *testing.T) {
		accountID := uuid.New()
		now := time.Now()
		response := dto.RegisterAppResponse{
			AccountID: accountID,
			Name:      "test-app",
			Status:    "active",
			CreatedAt: now,
		}

		require.Equal(t, accountID, response.AccountID)
		require.Equal(t, "test-app", response.Name)
		require.Equal(t, "active", response.Status)
		require.Equal(t, now, response.CreatedAt)
	})

	// Test IssueApiKeyResponse structure
	t.Run("IssueApiKeyResponse", func(t *testing.T) {
		accountID := uuid.New()
		apiKeyID := uuid.New()
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)
		response := dto.IssueApiKeyResponse{
			APIKeyID:    apiKeyID,
			APIKey:      "test-api-key",
			KeyHash:     "test-key-hash",
			AccountID:   accountID,
			Name:        "test-api-key",
			Permissions: []string{domain.PermissionReadAccounts},
			Status:      "active",
			ExpiresAt:   expiresAt,
			CreatedAt:   now,
		}

		require.Equal(t, apiKeyID, response.APIKeyID)
		require.Equal(t, "test-api-key", response.APIKey)
		require.Equal(t, "test-key-hash", response.KeyHash)
		require.Equal(t, accountID, response.AccountID)
		require.Equal(t, "test-api-key", response.Name)
		require.Equal(t, []string{domain.PermissionReadAccounts}, response.Permissions)
		require.Equal(t, "active", response.Status)
		require.Equal(t, expiresAt, response.ExpiresAt)
		require.Equal(t, now, response.CreatedAt)
	})

	// Test ValidateApiKeyResponse structure
	t.Run("ValidateApiKeyResponse", func(t *testing.T) {
		accountID := uuid.New()
		apiKeyID := uuid.New()
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)
		name := "test-api-key"
		response := dto.ValidateApiKeyResponse{
			Valid:       true,
			AccountID:   &accountID,
			APIKeyID:    &apiKeyID,
			Name:        &name,
			Permissions: []string{domain.PermissionReadAccounts},
			LastUsedAt:  &now,
			ExpiresAt:   &expiresAt,
		}

		require.Equal(t, true, response.Valid)
		require.Equal(t, &accountID, response.AccountID)
		require.Equal(t, &apiKeyID, response.APIKeyID)
		require.Equal(t, &name, response.Name)
		require.Equal(t, []string{domain.PermissionReadAccounts}, response.Permissions)
		require.Equal(t, &now, response.LastUsedAt)
		require.Equal(t, &expiresAt, response.ExpiresAt)
	})

	// Test ApiKeyResponse structure
	t.Run("ApiKeyResponse", func(t *testing.T) {
		apiKeyID := uuid.New()
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)
		response := dto.ApiKeyResponse{
			APIKeyID:    apiKeyID,
			Name:        "test-api-key",
			Permissions: []string{domain.PermissionReadAccounts},
			Status:      "active",
			LastUsedAt:  &now,
			ExpiresAt:   expiresAt,
			CreatedAt:   now,
		}

		require.Equal(t, apiKeyID, response.APIKeyID)
		require.Equal(t, "test-api-key", response.Name)
		require.Equal(t, []string{domain.PermissionReadAccounts}, response.Permissions)
		require.Equal(t, "active", response.Status)
		require.Equal(t, &now, response.LastUsedAt)
		require.Equal(t, expiresAt, response.ExpiresAt)
		require.Equal(t, now, response.CreatedAt)
	})

	// Test GetAPIKeysResponse structure
	t.Run("GetAPIKeysResponse", func(t *testing.T) {
		apiKeyID := uuid.New()
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)
		apiKeys := []dto.ApiKeyResponse{
			{
				APIKeyID:    apiKeyID,
				Name:        "test-api-key",
				Permissions: []string{domain.PermissionReadAccounts},
				Status:      "active",
				ExpiresAt:   expiresAt,
				CreatedAt:   now,
			},
		}
		response := dto.GetAPIKeysResponse{
			APIKeys: apiKeys,
			Limit:   10,
			Offset:  0,
			Total:   1,
		}

		require.Equal(t, apiKeys, response.APIKeys)
		require.Equal(t, 10, response.Limit)
		require.Equal(t, 0, response.Offset)
		require.Equal(t, 1, response.Total)
	})

	// Test HealthResponse structure
	t.Run("HealthResponse", func(t *testing.T) {
		now := time.Now()
		response := dto.HealthResponse{
			Status:    "healthy",
			Timestamp: now,
			Service:   "auth-service",
			Version:   "1.0.0",
		}

		require.Equal(t, "healthy", response.Status)
		require.Equal(t, now, response.Timestamp)
		require.Equal(t, "auth-service", response.Service)
		require.Equal(t, "1.0.0", response.Version)
	})
}
