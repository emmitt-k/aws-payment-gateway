package usecase

import (
	"context"
	"fmt"

	"github.com/aws-payment-gateway/internal/auth/repository"
	"github.com/google/uuid"
)

// RevokeApiKeyInput represents the input for revoking an API key
type RevokeApiKeyInput struct {
	APIKeyID uuid.UUID `json:"api_key_id" validate:"required"`
}

// RevokeApiKeyOutput represents the output of API key revocation
type RevokeApiKeyOutput struct {
	Success bool `json:"success"`
}

// RevokeApiKey handles the business logic for revoking API keys
type RevokeApiKey struct {
	apiKeyRepo repository.ApiKeyRepository
}

// NewRevokeApiKey creates a new RevokeApiKey use case
func NewRevokeApiKey(apiKeyRepo repository.ApiKeyRepository) *RevokeApiKey {
	return &RevokeApiKey{
		apiKeyRepo: apiKeyRepo,
	}
}

// Execute revokes an API key
func (uc *RevokeApiKey) Execute(ctx context.Context, input RevokeApiKeyInput) (*RevokeApiKeyOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Check if API key exists
	apiKey, err := uc.apiKeyRepo.GetByID(ctx, input.APIKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	if apiKey == nil {
		return nil, fmt.Errorf("API key not found")
	}

	// Revoke the API key
	if err := uc.apiKeyRepo.Revoke(ctx, input.APIKeyID); err != nil {
		return nil, fmt.Errorf("failed to revoke API key: %w", err)
	}

	// Create output
	output := &RevokeApiKeyOutput{
		Success: true,
	}

	return output, nil
}

// validateInput validates the revoke API key input
func (uc *RevokeApiKey) validateInput(input RevokeApiKeyInput) error {
	if input.APIKeyID == uuid.Nil {
		return fmt.Errorf("api_key_id is required")
	}

	return nil
}
