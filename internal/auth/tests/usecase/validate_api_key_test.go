package usecase_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/tests/mocks"
	"github.com/aws-payment-gateway/internal/auth/usecase"
)

func TestValidateApiKey_Execute_WithRawKey(t *testing.T) {
	// Setup
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	mockAppRepo := mocks.NewMockAppRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	// Test data
	accountID := uuid.New()
	apiKeyID := uuid.New()
	now := time.Now()

	// Use the special pattern that mock repository recognizes for ValidateByKey
	rawKey := "raw-api-key-" + apiKeyID.String()

	// For the mock, we need to use SHA256 hash since ValidateByKey uses that
	hash := sha256.Sum256([]byte(rawKey))
	hashStr := hex.EncodeToString(hash[:])

	apiKey := &domain.ApiKey{
		ID:          apiKeyID,
		AccountID:   accountID,
		Name:        "Test API Key",
		KeyHash:     hashStr,
		Permissions: domain.ApiKeyPermissions{"read", "write"},
		Status:      domain.ApiKeyStatusActive,
		ExpiresAt:   now.Add(24 * time.Hour),
		CreatedAt:   now,
		LastUsedAt:  &now,
	}

	account := &domain.Account{
		ID:     accountID,
		Name:   "Test Account",
		Status: domain.AccountStatusActive,
	}

	// Setup mock data
	mockApiKeyRepo.AddApiKey(apiKey)
	mockAppRepo.AddAccount(account)

	// Execute
	input := usecase.ValidateApiKeyInput{
		RawKey: rawKey,
	}
	result, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Equal(t, accountID, *result.AccountID)
	assert.Equal(t, apiKeyID, *result.APIKeyID)
	assert.Equal(t, "Test API Key", *result.Name)
	assert.Equal(t, domain.ApiKeyPermissions{"read", "write"}, result.Permissions)
	assert.Equal(t, "Test Account", *result.AccountName)
	assert.Equal(t, string(domain.AccountStatusActive), *result.AccountStatus)
}

func TestValidateApiKey_Execute_WithKeyHash(t *testing.T) {
	// Setup
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	mockAppRepo := mocks.NewMockAppRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	// Test data
	accountID := uuid.New()
	apiKeyID := uuid.New()
	now := time.Now()
	hashedKey := "hashed-key-12345"

	apiKey := &domain.ApiKey{
		ID:          apiKeyID,
		AccountID:   accountID,
		Name:        "Test API Key",
		KeyHash:     hashedKey,
		Permissions: domain.ApiKeyPermissions{"read"},
		Status:      domain.ApiKeyStatusActive,
		ExpiresAt:   now.Add(24 * time.Hour),
		CreatedAt:   now,
		LastUsedAt:  &now,
	}

	account := &domain.Account{
		ID:     accountID,
		Name:   "Test Account",
		Status: domain.AccountStatusActive,
	}

	// Setup mock data
	mockApiKeyRepo.AddApiKey(apiKey)
	mockAppRepo.AddAccount(account)

	// Execute
	input := usecase.ValidateApiKeyInput{
		KeyHash: hashedKey,
	}
	result, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Equal(t, accountID, *result.AccountID)
	assert.Equal(t, apiKeyID, *result.APIKeyID)
}

func TestValidateApiKey_Execute_WithInvalidKey(t *testing.T) {
	// Setup
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	mockAppRepo := mocks.NewMockAppRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	// Test data - this key won't match any stored keys
	rawKey := "raw-api-key-invalid"

	// Execute
	input := usecase.ValidateApiKeyInput{
		RawKey: rawKey,
	}
	result, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Nil(t, result.AccountID)
	assert.Nil(t, result.APIKeyID)
}

func TestValidateApiKey_Execute_WithExpiredKey(t *testing.T) {
	// Setup
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	mockAppRepo := mocks.NewMockAppRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	// Test data
	accountID := uuid.New()
	apiKeyID := uuid.New()
	now := time.Now()

	// Use the special pattern that mock repository recognizes for ValidateByKey
	rawKey := "raw-api-key-" + apiKeyID.String()

	// For the mock, we need to use SHA256 hash since ValidateByKey uses that
	hash := sha256.Sum256([]byte(rawKey))
	hashStr := hex.EncodeToString(hash[:])

	apiKey := &domain.ApiKey{
		ID:          apiKeyID,
		AccountID:   accountID,
		Name:        "Expired API Key",
		KeyHash:     hashStr,
		Permissions: domain.ApiKeyPermissions{"read"},
		Status:      domain.ApiKeyStatusActive,
		ExpiresAt:   now.Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt:   now.Add(-24 * time.Hour),
		LastUsedAt:  &now,
	}

	account := &domain.Account{
		ID:     accountID,
		Name:   "Test Account",
		Status: domain.AccountStatusActive,
	}

	// Setup mock data
	mockApiKeyRepo.AddApiKey(apiKey)
	mockAppRepo.AddAccount(account)

	// Execute
	input := usecase.ValidateApiKeyInput{
		RawKey: rawKey,
	}
	result, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid) // Should be false due to expiration
	assert.Equal(t, accountID, *result.AccountID)
	assert.Equal(t, apiKeyID, *result.APIKeyID)
}

func TestValidateApiKey_Execute_WithInactiveAccount(t *testing.T) {
	// Setup
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	mockAppRepo := mocks.NewMockAppRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	// Test data
	accountID := uuid.New()
	apiKeyID := uuid.New()
	now := time.Now()

	// Use the special pattern that mock repository recognizes for ValidateByKey
	rawKey := "raw-api-key-" + apiKeyID.String()

	// For the mock, we need to use SHA256 hash since ValidateByKey uses that
	hash := sha256.Sum256([]byte(rawKey))
	hashStr := hex.EncodeToString(hash[:])

	apiKey := &domain.ApiKey{
		ID:          apiKeyID,
		AccountID:   accountID,
		Name:        "Test API Key",
		KeyHash:     hashStr,
		Permissions: domain.ApiKeyPermissions{"read"},
		Status:      domain.ApiKeyStatusActive,
		ExpiresAt:   now.Add(24 * time.Hour),
		CreatedAt:   now,
		LastUsedAt:  &now,
	}

	account := &domain.Account{
		ID:     accountID,
		Name:   "Inactive Account",
		Status: domain.AccountStatusSuspended, // Inactive account
	}

	// Setup mock data
	mockApiKeyRepo.AddApiKey(apiKey)
	mockAppRepo.AddAccount(account)

	// Execute
	input := usecase.ValidateApiKeyInput{
		RawKey: rawKey,
	}
	result, err := uc.Execute(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid) // Should be false due to inactive account
	assert.Equal(t, accountID, *result.AccountID)
	assert.Equal(t, apiKeyID, *result.APIKeyID)
}

func TestValidateApiKey_Execute_WithNoInput(t *testing.T) {
	// Setup
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	mockAppRepo := mocks.NewMockAppRepository()
	uc := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)

	// Execute with empty input
	input := usecase.ValidateApiKeyInput{}
	result, err := uc.Execute(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "either raw_key or key_hash must be provided")
}
