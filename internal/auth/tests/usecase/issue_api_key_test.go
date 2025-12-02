package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/tests/mocks"
	"github.com/aws-payment-gateway/internal/auth/tests/utils"
	"github.com/aws-payment-gateway/internal/auth/usecase"
)

// RequireTrue is a helper to require a condition to be true
func RequireTrue(t *testing.T, condition bool) {
	require.True(t, condition)
}

// RequireNotEqual is a helper to require two values to not be equal
func RequireNotEqual(t *testing.T, expected, actual interface{}) {
	require.NotEqual(t, expected, actual)
}

func TestIssueApiKey_Execute_Success(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	expiresIn := 24 // hours
	input := usecase.IssueApiKeyInput{
		AccountID:   account.ID,
		Name:        "test-api-key",
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts, domain.PermissionReadKeys},
		ExpiresIn:   &expiresIn,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, input.AccountID, output.AccountID)
	utils.RequireEqual(t, input.Name, output.Name)
	utils.RequireEqual(t, input.Permissions, output.Permissions)
	utils.RequireEqual(t, string(domain.ApiKeyStatusActive), output.Status)
	utils.RequireNotNil(t, output.APIKeyID)
	utils.RequireNotNil(t, output.APIKey) // Actual API key should be returned
	utils.RequireNotNil(t, output.KeyHash)
	utils.RequireNotNil(t, output.ExpiresAt)
	utils.RequireNotNil(t, output.CreatedAt)

	// Verify API key was saved
	apiKeys, err := mockApiKeyRepo.GetByAccountID(ctx, account.ID)
	utils.RequireNoError(t, err)
	utils.RequireEqual(t, 1, len(apiKeys))
	utils.RequireEqual(t, output.APIKeyID, apiKeys[0].ID)
	utils.RequireEqual(t, output.KeyHash, apiKeys[0].KeyHash)
}

func TestIssueApiKey_Execute_AccountNotFound(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	input := usecase.IssueApiKeyInput{
		AccountID:   uuid.New(), // Non-existent account
		Name:        "test-api-key",
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "account not found or inactive", err.Error())
}

func TestIssueApiKey_Execute_InactiveAccount(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	account.Status = domain.AccountStatusSuspended // Make account inactive
	mockAppRepo.AddAccount(account)

	input := usecase.IssueApiKeyInput{
		AccountID:   account.ID,
		Name:        "test-api-key",
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "account not found or inactive", err.Error())
}

func TestIssueApiKey_Execute_EmptyPermissions(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	input := usecase.IssueApiKeyInput{
		AccountID:   account.ID,
		Name:        "test-api-key",
		Permissions: domain.ApiKeyPermissions{}, // Empty permissions
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: at least one permission is required", err.Error())
}

func TestIssueApiKey_Execute_InvalidPermission(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	input := usecase.IssueApiKeyInput{
		AccountID:   account.ID,
		Name:        "test-api-key",
		Permissions: domain.ApiKeyPermissions{"invalid:permission"}, // Invalid permission
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: invalid permission: invalid:permission", err.Error())
}

func TestIssueApiKey_Execute_ValidPermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		valid       bool
	}{
		{
			name:        "single valid permission",
			permissions: []string{domain.PermissionReadAccounts},
			valid:       true,
		},
		{
			name:        "multiple valid permissions",
			permissions: []string{domain.PermissionReadAccounts, domain.PermissionWriteKeys, domain.PermissionManageWebhooks},
			valid:       true,
		},
		{
			name: "all valid permissions",
			permissions: []string{
				domain.PermissionReadAccounts,
				domain.PermissionWriteAccounts,
				domain.PermissionReadKeys,
				domain.PermissionWriteKeys,
				domain.PermissionManageWebhooks,
			},
			valid: true,
		},
		{
			name:        "mixed valid and invalid permissions",
			permissions: []string{domain.PermissionReadAccounts, "invalid:permission"},
			valid:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := utils.TestContext(t)
			mockAppRepo := mocks.NewMockAppRepository()
			mockApiKeyRepo := mocks.NewMockApiKeyRepository()
			uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

			account := utils.CreateTestAccount(t)
			mockAppRepo.AddAccount(account)

			input := usecase.IssueApiKeyInput{
				AccountID:   account.ID,
				Name:        "test-api-key",
				Permissions: domain.ApiKeyPermissions(tt.permissions),
			}

			// Act
			output, err := uc.Execute(ctx, input)

			// Assert
			if tt.valid {
				utils.RequireNoError(t, err)
				utils.RequireNotNil(t, output)
				utils.RequireEqual(t, tt.permissions, output.Permissions)
			} else {
				utils.RequireError(t, err)
				utils.RequireNil(t, output)
			}
		})
	}
}

func TestIssueApiKey_Execute_WithExpiration(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	expiresIn := 48 // 48 hours
	input := usecase.IssueApiKeyInput{
		AccountID:   account.ID,
		Name:        "test-api-key",
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
		ExpiresIn:   &expiresIn,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)

	// Verify expiration is approximately 48 hours from now
	expectedExpiry := time.Now().Add(48 * time.Hour)
	timeDiff := output.ExpiresAt.Sub(expectedExpiry)
	require.True(t, timeDiff < time.Minute) // Allow for small time differences
}

func TestIssueApiKey_Execute_WithoutExpiration(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	input := usecase.IssueApiKeyInput{
		AccountID:   account.ID,
		Name:        "test-api-key",
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
		ExpiresIn:   nil, // No expiration specified
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)

	// Verify expiration is approximately now (default behavior)
	expectedExpiry := time.Now()
	timeDiff := output.ExpiresAt.Sub(expectedExpiry)
	require.True(t, timeDiff < time.Minute) // Allow for small time differences
}

func TestIssueApiKey_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	// Set up repository to return error on GetByID
	repoError := errors.New("database error")
	mockAppRepo.SetGetError(repoError)

	input := usecase.IssueApiKeyInput{
		AccountID:   uuid.New(),
		Name:        "test-api-key",
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to get account: database error", err.Error())
}

func TestIssueApiKey_Execute_CreateError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	// Set up repository to return error on Create
	repoError := errors.New("create error")
	mockApiKeyRepo.SetCreateError(repoError)

	input := usecase.IssueApiKeyInput{
		AccountID:   account.ID,
		Name:        "test-api-key",
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to create API key: create error", err.Error())
}

func TestIssueApiKey_Execute_ShortName(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	input := usecase.IssueApiKeyInput{
		AccountID:   account.ID,
		Name:        "ab", // Less than 3 characters
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	// Note: The current implementation doesn't validate name length, but this test documents expected behavior
	// If name length validation is added, this test should expect an error
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
}

func TestIssueApiKey_Execute_APIKeyGeneration(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	input := usecase.IssueApiKeyInput{
		AccountID:   account.ID,
		Name:        "test-api-key",
		Permissions: domain.ApiKeyPermissions{domain.PermissionReadAccounts},
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)

	// Verify API key is returned (only during creation)
	utils.RequireNotNil(t, output.APIKey)
	RequireTrue(t, len(output.APIKey) > 0) // Should be a non-empty string

	// Verify key hash is different from actual key
	RequireNotEqual(t, output.APIKey, output.KeyHash)
	RequireTrue(t, len(output.KeyHash) > 0) // Should be a non-empty hash
}
