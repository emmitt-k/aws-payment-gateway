package usecase_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/tests/mocks"
	"github.com/aws-payment-gateway/internal/auth/tests/utils"
	"github.com/aws-payment-gateway/internal/auth/usecase"
)

func TestGetAPIKeys_Execute_Success(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	// Add some API keys to the account
	apiKey1 := utils.CreateTestApiKey(t, account.ID)
	apiKey2 := utils.CreateTestApiKey(t, account.ID)
	apiKey3 := utils.CreateTestApiKey(t, account.ID)
	mockApiKeyRepo.AddApiKey(apiKey1)
	mockApiKeyRepo.AddApiKey(apiKey2)
	mockApiKeyRepo.AddApiKey(apiKey3)

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     10,
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, 10, output.Limit)
	utils.RequireEqual(t, 0, output.Offset)
	utils.RequireEqual(t, 3, output.Total)
	utils.RequireEqual(t, 3, len(output.APIKeys))
}

func TestGetAPIKeys_Execute_Pagination(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	// Add 5 API keys to the account
	for i := 0; i < 5; i++ {
		apiKey := utils.CreateTestApiKey(t, account.ID)
		mockApiKeyRepo.AddApiKey(apiKey)
	}

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     2, // Request only 2 at a time
		Offset:    1, // Start from index 1
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, 2, output.Limit)
	utils.RequireEqual(t, 1, output.Offset)
	utils.RequireEqual(t, 5, output.Total)
	utils.RequireEqual(t, 2, len(output.APIKeys)) // Should return only 2 keys starting from offset 1
}

func TestGetAPIKeys_Execute_EmptyResult(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     10,
		Offset:    10, // Offset beyond available keys
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, 10, output.Limit)
	utils.RequireEqual(t, 10, output.Offset)
	utils.RequireEqual(t, 0, output.Total) // No API keys should be returned
	utils.RequireEqual(t, 0, len(output.APIKeys))
}

func TestGetAPIKeys_Execute_AccountNotFound(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	input := usecase.GetAPIKeysInput{
		AccountID: uuid.New(), // Non-existent account
		Limit:     10,
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "account not found or inactive", err.Error())
}

func TestGetAPIKeys_Execute_InactiveAccount(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	account.Status = domain.AccountStatusSuspended // Make account inactive
	mockAppRepo.AddAccount(account)

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     10,
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "account not found or inactive", err.Error())
}

func TestGetAPIKeys_Execute_EmptyAccountID(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	input := usecase.GetAPIKeysInput{
		AccountID: uuid.Nil, // Empty UUID
		Limit:     10,
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: account_id is required", err.Error())
}

func TestGetAPIKeys_Execute_InvalidLimit(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     0, // Invalid limit (too low)
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: limit must be between 1 and 100", err.Error())
}

func TestGetAPIKeys_Execute_InvalidHighLimit(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     101, // Invalid limit (too high)
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: limit must be between 1 and 100", err.Error())
}

func TestGetAPIKeys_Execute_NegativeOffset(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     10,
		Offset:    -1, // Negative offset
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: offset must be non-negative", err.Error())
}

func TestGetAPIKeys_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	// Set up repository to return error on GetByID
	repoError := errors.New("database error")
	mockAppRepo.SetGetError(repoError)

	input := usecase.GetAPIKeysInput{
		AccountID: uuid.New(),
		Limit:     10,
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to get account: database error", err.Error())
}

func TestGetAPIKeys_Execute_ListError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	// Set up repository to return error on List
	repoError := errors.New("list error")
	mockApiKeyRepo.SetListError(repoError)

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     10,
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to get API keys: list error", err.Error())
}

func TestGetAPIKeys_Execute_GetByAccountIDError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	// Set up repository to return error on GetByAccountID
	repoError := errors.New("get all error")
	mockApiKeyRepo.SetGetError(repoError)

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     10,
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to get all API keys for total count: get all error", err.Error())
}

func TestGetAPIKeys_Execute_MixedStatusKeys(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	// Add API keys with different statuses
	activeKey := utils.CreateTestApiKey(t, account.ID)
	inactiveKey := utils.CreateInactiveTestApiKey(t, account.ID)
	expiredKey := utils.CreateExpiredTestApiKey(t, account.ID)

	mockApiKeyRepo.AddApiKey(activeKey)
	mockApiKeyRepo.AddApiKey(inactiveKey)
	mockApiKeyRepo.AddApiKey(expiredKey)

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     10,
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, 10, output.Limit)
	utils.RequireEqual(t, 0, output.Offset)
	utils.RequireEqual(t, 3, output.Total)
	utils.RequireEqual(t, 3, len(output.APIKeys))

	// Verify all keys are returned regardless of status
	var foundActive, foundInactive, foundExpired bool
	for _, key := range output.APIKeys {
		if key.ID == activeKey.ID {
			foundActive = true
		}
		if key.ID == inactiveKey.ID {
			foundInactive = true
		}
		if key.ID == expiredKey.ID {
			foundExpired = true
		}
	}
	require.True(t, foundActive)
	require.True(t, foundInactive)
	require.True(t, foundExpired)
}

func TestGetAPIKeys_Execute_MaxLimit(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo)

	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	// Add 150 API keys to the account
	for i := 0; i < 150; i++ {
		apiKey := utils.CreateTestApiKey(t, account.ID)
		mockApiKeyRepo.AddApiKey(apiKey)
	}

	input := usecase.GetAPIKeysInput{
		AccountID: account.ID,
		Limit:     100, // Maximum allowed limit
		Offset:    0,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, 100, output.Limit)
	utils.RequireEqual(t, 0, output.Offset)
	utils.RequireEqual(t, 150, output.Total)
	utils.RequireEqual(t, 100, len(output.APIKeys)) // Should return only 100 keys due to limit
}
