package usecase_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/tests/mocks"
	"github.com/aws-payment-gateway/internal/auth/tests/utils"
	"github.com/aws-payment-gateway/internal/auth/usecase"
)

func TestRevokeApiKey_Execute_Success(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRevokeApiKey(mockApiKeyRepo)

	apiKey := utils.CreateTestApiKey(t, uuid.New())
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.RevokeApiKeyInput{
		APIKeyID: apiKey.ID,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, true, output.Success)

	// Verify API key was revoked
	revokedApiKey, err := mockApiKeyRepo.GetByID(ctx, apiKey.ID)
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, revokedApiKey)
	utils.RequireEqual(t, domain.ApiKeyStatusInactive, revokedApiKey.Status)
}

func TestRevokeApiKey_Execute_NonExistentKey(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRevokeApiKey(mockApiKeyRepo)

	input := usecase.RevokeApiKeyInput{
		APIKeyID: uuid.New(), // Non-existent API key
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "API key not found", err.Error())
}

func TestRevokeApiKey_Execute_EmptyID(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRevokeApiKey(mockApiKeyRepo)

	input := usecase.RevokeApiKeyInput{
		APIKeyID: uuid.Nil, // Empty UUID
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: api_key_id is required", err.Error())
}

func TestRevokeApiKey_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRevokeApiKey(mockApiKeyRepo)

	// Set up repository to return error on GetByID
	repoError := errors.New("database error")
	mockApiKeyRepo.SetGetError(repoError)

	input := usecase.RevokeApiKeyInput{
		APIKeyID: uuid.New(),
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to get API key: database error", err.Error())
}

func TestRevokeApiKey_Execute_RevokeError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRevokeApiKey(mockApiKeyRepo)

	apiKey := utils.CreateTestApiKey(t, uuid.New())
	mockApiKeyRepo.AddApiKey(apiKey)

	// Set up repository to return error on Revoke
	repoError := errors.New("revoke error")
	mockApiKeyRepo.SetRevokeError(repoError)

	input := usecase.RevokeApiKeyInput{
		APIKeyID: apiKey.ID,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to revoke API key: revoke error", err.Error())
}

func TestRevokeApiKey_Execute_AlreadyRevokedKey(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRevokeApiKey(mockApiKeyRepo)

	// Create an already inactive API key
	apiKey := utils.CreateTestApiKey(t, uuid.New())
	apiKey.Status = domain.ApiKeyStatusInactive // Already revoked
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.RevokeApiKeyInput{
		APIKeyID: apiKey.ID,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, true, output.Success)

	// Verify API key is still inactive
	revokedApiKey, err := mockApiKeyRepo.GetByID(ctx, apiKey.ID)
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, revokedApiKey)
	utils.RequireEqual(t, domain.ApiKeyStatusInactive, revokedApiKey.Status)
}

func TestRevokeApiKey_Execute_ExpiredKey(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRevokeApiKey(mockApiKeyRepo)

	// Create an expired API key
	apiKey := utils.CreateExpiredTestApiKey(t, uuid.New())
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.RevokeApiKeyInput{
		APIKeyID: apiKey.ID,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, true, output.Success)

	// Verify API key was revoked
	revokedApiKey, err := mockApiKeyRepo.GetByID(ctx, apiKey.ID)
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, revokedApiKey)
	utils.RequireEqual(t, domain.ApiKeyStatusInactive, revokedApiKey.Status)
}

func TestRevokeApiKey_Execute_InactiveKey(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRevokeApiKey(mockApiKeyRepo)

	// Create an inactive API key
	apiKey := utils.CreateInactiveTestApiKey(t, uuid.New())
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.RevokeApiKeyInput{
		APIKeyID: apiKey.ID,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, true, output.Success)

	// Verify API key is still inactive
	revokedApiKey, err := mockApiKeyRepo.GetByID(ctx, apiKey.ID)
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, revokedApiKey)
	utils.RequireEqual(t, domain.ApiKeyStatusInactive, revokedApiKey.Status)
}

func TestRevokeApiKey_Execute_MultipleRevoke(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRevokeApiKey(mockApiKeyRepo)

	apiKey := utils.CreateTestApiKey(t, uuid.New())
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.RevokeApiKeyInput{
		APIKeyID: apiKey.ID,
	}

	// Act - First revoke
	output1, err1 := uc.Execute(ctx, input)

	// Assert - First revoke
	utils.RequireNoError(t, err1)
	utils.RequireNotNil(t, output1)
	utils.RequireEqual(t, true, output1.Success)

	// Act - Second revoke (should still succeed)
	output2, err2 := uc.Execute(ctx, input)

	// Assert - Second revoke
	utils.RequireNoError(t, err2)
	utils.RequireNotNil(t, output2)
	utils.RequireEqual(t, true, output2.Success)

	// Verify API key is still inactive
	revokedApiKey, err := mockApiKeyRepo.GetByID(ctx, apiKey.ID)
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, revokedApiKey)
	utils.RequireEqual(t, domain.ApiKeyStatusInactive, revokedApiKey.Status)
}

func TestRevokeApiKey_Execute_WithPermissions(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRevokeApiKey(mockApiKeyRepo)

	// Create API key with multiple permissions
	apiKey := utils.CreateTestApiKey(t, uuid.New())
	apiKey.Permissions = domain.ApiKeyPermissions{
		domain.PermissionReadAccounts,
		domain.PermissionWriteKeys,
		domain.PermissionManageWebhooks,
	}
	mockApiKeyRepo.AddApiKey(apiKey)

	input := usecase.RevokeApiKeyInput{
		APIKeyID: apiKey.ID,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, true, output.Success)

	// Verify API key was revoked
	revokedApiKey, err := mockApiKeyRepo.GetByID(ctx, apiKey.ID)
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, revokedApiKey)
	utils.RequireEqual(t, domain.ApiKeyStatusInactive, revokedApiKey.Status)
}
