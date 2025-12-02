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

func TestValidateApiKey_Execute_ValidKey(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateTestApiKey(t, account.ID)
	mockAppRepo.AddAccount(account)
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.ValidateApiKeyInput{
		KeyHash: apiKey.KeyHash,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, true, output.Valid)
	utils.RequireEqual(t, &apiKey.AccountID, output.AccountID)
	utils.RequireEqual(t, &apiKey.ID, output.APIKeyID)
	utils.RequireEqual(t, &apiKey.Name, output.Name)
	utils.RequireEqual(t, apiKey.Permissions, output.Permissions)
	utils.RequireEqual(t, apiKey.LastUsedAt, output.LastUsedAt)
	utils.RequireEqual(t, &apiKey.ExpiresAt, output.ExpiresAt)
	utils.RequireEqual(t, account.Name, *output.AccountName)
	utils.RequireEqual(t, string(account.Status), *output.AccountStatus)
}

func TestValidateApiKey_Execute_InvalidKey(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	input := usecase.ValidateApiKeyInput{
		KeyHash: "non-existent-hash",
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, false, output.Valid)
	utils.RequireNil(t, output.AccountID)
	utils.RequireNil(t, output.APIKeyID)
	utils.RequireNil(t, output.Name)
	utils.RequireEqual(t, domain.ApiKeyPermissions{}, output.Permissions)
	utils.RequireNil(t, output.LastUsedAt)
	utils.RequireNil(t, output.ExpiresAt)
	utils.RequireNil(t, output.AccountName)
	utils.RequireNil(t, output.AccountStatus)
}

func TestValidateApiKey_Execute_EmptyKeyHash(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	input := usecase.ValidateApiKeyInput{
		KeyHash: "", // Empty key hash
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: key_hash is required", err.Error())
}

func TestValidateApiKey_Execute_ExpiredKey(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateExpiredTestApiKey(t, account.ID) // Expired key
	mockAppRepo.AddAccount(account)
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.ValidateApiKeyInput{
		KeyHash: apiKey.KeyHash,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, false, output.Valid) // Expired keys are invalid
	utils.RequireEqual(t, &apiKey.AccountID, output.AccountID)
	utils.RequireEqual(t, &apiKey.ID, output.APIKeyID)
	utils.RequireEqual(t, &apiKey.Name, output.Name)
	utils.RequireEqual(t, apiKey.Permissions, output.Permissions)
	utils.RequireEqual(t, &apiKey.ExpiresAt, output.ExpiresAt)
}

func TestValidateApiKey_Execute_InactiveKey(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateInactiveTestApiKey(t, account.ID) // Inactive key
	mockAppRepo.AddAccount(account)
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.ValidateApiKeyInput{
		KeyHash: apiKey.KeyHash,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, false, output.Valid) // Inactive keys are invalid
	utils.RequireEqual(t, &apiKey.AccountID, output.AccountID)
	utils.RequireEqual(t, &apiKey.ID, output.APIKeyID)
	utils.RequireEqual(t, &apiKey.Name, output.Name)
	utils.RequireEqual(t, apiKey.Permissions, output.Permissions)
	utils.RequireEqual(t, &apiKey.ExpiresAt, output.ExpiresAt)
}

func TestValidateApiKey_Execute_InactiveAccount(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	account := utils.CreateTestAccount(t)
	account.Status = domain.AccountStatusSuspended // Make account inactive
	mockAppRepo.AddAccount(account)

	apiKey := utils.CreateTestApiKey(t, account.ID)
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.ValidateApiKeyInput{
		KeyHash: apiKey.KeyHash,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, false, output.Valid) // Keys with inactive accounts are invalid
	utils.RequireEqual(t, &apiKey.AccountID, output.AccountID)
	utils.RequireEqual(t, &apiKey.ID, output.APIKeyID)
	utils.RequireEqual(t, &apiKey.Name, output.Name)
	utils.RequireEqual(t, apiKey.Permissions, output.Permissions)
	utils.RequireEqual(t, account.Name, *output.AccountName)
	utils.RequireEqual(t, string(account.Status), *output.AccountStatus)
}

func TestValidateApiKey_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	// Set up repository to return error on GetByKeyHash
	repoError := errors.New("database error")
	mockApiKeyRepo.SetGetError(repoError)

	input := usecase.ValidateApiKeyInput{
		KeyHash: "test-hash",
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to get API key: database error", err.Error())
}

func TestValidateApiKey_Execute_AccountRepositoryError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateTestApiKey(t, account.ID)
	mockAppRepo.AddAccount(account)
	mockApiKeyRepo.AddApiKey(apiKey)

	// Set up app repository to return error on GetByID
	repoError := errors.New("database error")
	mockAppRepo.SetGetError(repoError)

	input := usecase.ValidateApiKeyInput{
		KeyHash: apiKey.KeyHash,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to get account: database error", err.Error())
}

func TestValidateApiKey_Execute_NonExistentAccount(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	// Don't add account to repository (simulating non-existent account)
	apiKey := utils.CreateTestApiKey(t, uuid.New())
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.ValidateApiKeyInput{
		KeyHash: apiKey.KeyHash,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, true, output.Valid) // Key is valid but account info is missing
	utils.RequireEqual(t, &apiKey.AccountID, output.AccountID)
	utils.RequireEqual(t, &apiKey.ID, output.APIKeyID)
	utils.RequireEqual(t, &apiKey.Name, output.Name)
	utils.RequireEqual(t, apiKey.Permissions, output.Permissions)
	utils.RequireNil(t, output.AccountName)   // Account not found
	utils.RequireNil(t, output.AccountStatus) // Account not found
}

func TestValidateApiKey_Execute_KeyWithLastUsedAt(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	// Create API key with LastUsedAt set
	apiKey := utils.CreateTestApiKey(t, account.ID)
	lastUsedAt := time.Now().Add(-1 * time.Hour)
	apiKey.LastUsedAt = &lastUsedAt
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.ValidateApiKeyInput{
		KeyHash: apiKey.KeyHash,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, true, output.Valid)
	utils.RequireEqual(t, &lastUsedAt, output.LastUsedAt)
}

func TestValidateApiKey_Execute_MultiplePermissions(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	// Create API key with multiple permissions
	apiKey := utils.CreateTestApiKey(t, account.ID)
	apiKey.Permissions = domain.ApiKeyPermissions{
		domain.PermissionReadAccounts,
		domain.PermissionWriteKeys,
		domain.PermissionManageWebhooks,
	}
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.ValidateApiKeyInput{
		KeyHash: apiKey.KeyHash,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, true, output.Valid)
	utils.RequireEqual(t, 3, len(output.Permissions))
	require.Contains(t, output.Permissions, domain.PermissionReadAccounts)
	require.Contains(t, output.Permissions, domain.PermissionWriteKeys)
	require.Contains(t, output.Permissions, domain.PermissionManageWebhooks)
}
