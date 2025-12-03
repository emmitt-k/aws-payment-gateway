package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/repository"
	"github.com/google/uuid"
)

// ValidateApiKeyInput represents the input for API key validation
type ValidateApiKeyInput struct {
	// RawKey is the raw API key provided by the client
	RawKey string `json:"raw_key" validate:"required"`
	// KeyHash is the pre-hashed API key (deprecated, use RawKey instead)
	KeyHash string `json:"key_hash,omitempty"`
}

// ValidateApiKeyOutput represents the output of API key validation
type ValidateApiKeyOutput struct {
	Valid         bool                     `json:"valid"`
	AccountID     *uuid.UUID               `json:"account_id,omitempty"`
	APIKeyID      *uuid.UUID               `json:"api_key_id,omitempty"`
	Name          *string                  `json:"name,omitempty"`
	Permissions   domain.ApiKeyPermissions `json:"permissions,omitempty"`
	LastUsedAt    *time.Time               `json:"last_used_at,omitempty"`
	ExpiresAt     *time.Time               `json:"expires_at,omitempty"`
	AccountName   *string                  `json:"account_name,omitempty"`
	AccountStatus *string                  `json:"account_status,omitempty"`
}

// ValidateApiKey handles the business logic for validating API keys
type ValidateApiKey struct {
	apiKeyRepo repository.ApiKeyRepository
	appRepo    repository.AppRepository
}

// NewValidateApiKey creates a new ValidateApiKey use case
func NewValidateApiKey(apiKeyRepo repository.ApiKeyRepository, appRepo repository.AppRepository) *ValidateApiKey {
	return &ValidateApiKey{
		apiKeyRepo: apiKeyRepo,
		appRepo:    appRepo,
	}
}

// Execute validates an API key and returns the result
func (uc *ValidateApiKey) Execute(ctx context.Context, input ValidateApiKeyInput) (*ValidateApiKeyOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	var apiKey *domain.ApiKey
	var err error

	// Handle both raw key and hash for backward compatibility
	if input.RawKey != "" {
		// Use the new validation method that accepts raw keys
		apiKey, err = uc.apiKeyRepo.ValidateByKey(ctx, input.RawKey)
		if err != nil {
			return nil, fmt.Errorf("failed to validate API key: %w", err)
		}
	} else if input.KeyHash != "" {
		// Legacy support for pre-hashed keys
		apiKey, err = uc.apiKeyRepo.GetByKeyHash(ctx, input.KeyHash)
		if err != nil {
			return nil, fmt.Errorf("failed to get API key: %w", err)
		}
	} else {
		return nil, fmt.Errorf("either raw_key or key_hash must be provided")
	}

	// Create output
	output := &ValidateApiKeyOutput{
		Valid:       apiKey != nil && apiKey.IsValid() && !apiKey.IsExpired(),
		Permissions: domain.ApiKeyPermissions{}, // Initialize with empty permissions
	}

	if apiKey != nil {
		output.AccountID = &apiKey.AccountID
		output.APIKeyID = &apiKey.ID
		output.Name = &apiKey.Name
		output.Permissions = apiKey.Permissions
		output.LastUsedAt = apiKey.LastUsedAt
		output.ExpiresAt = &apiKey.ExpiresAt

		// Get account information from PostgreSQL
		account, err := uc.appRepo.GetByID(ctx, apiKey.AccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get account: %w", err)
		}

		if account != nil {
			accountName := account.Name
			accountStatus := string(account.Status)
			output.AccountName = &accountName
			output.AccountStatus = &accountStatus

			// Account must be active for API key to be valid
			if !account.IsValid() {
				output.Valid = false
			}
		}
	}

	return output, nil
}

// validateInput validates the API key validation input
func (uc *ValidateApiKey) validateInput(input ValidateApiKeyInput) error {
	if input.RawKey == "" && input.KeyHash == "" {
		return fmt.Errorf("either raw_key or key_hash must be provided")
	}

	return nil
}
