package usecase

import (
	"context"
	"fmt"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/repository"
	"github.com/google/uuid"
)

// GetAPIKeysInput represents the input for getting API keys
type GetAPIKeysInput struct {
	AccountID uuid.UUID `json:"account_id" validate:"required"`
	Limit     int       `json:"limit" validate:"min=1,max=100"`
	Offset    int       `json:"offset" validate:"min=0"`
}

// GetAPIKeysOutput represents the output of getting API keys
type GetAPIKeysOutput struct {
	APIKeys []*domain.ApiKey `json:"api_keys"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
	Total   int              `json:"total"`
}

// GetAPIKeys handles the business logic for retrieving API keys
type GetAPIKeys struct {
	accountRepo repository.AppRepository
	apiKeyRepo  repository.ApiKeyRepository
}

// NewGetAPIKeys creates a new GetAPIKeys use case
func NewGetAPIKeys(accountRepo repository.AppRepository, apiKeyRepo repository.ApiKeyRepository) *GetAPIKeys {
	return &GetAPIKeys{
		accountRepo: accountRepo,
		apiKeyRepo:  apiKeyRepo,
	}
}

// Execute retrieves API keys for an account
func (uc *GetAPIKeys) Execute(ctx context.Context, input GetAPIKeysInput) (*GetAPIKeysOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Verify account exists and is active
	account, err := uc.accountRepo.GetByID(ctx, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil || !account.IsValid() {
		return nil, fmt.Errorf("account not found or inactive")
	}

	// Get API keys for the account
	apiKeys, err := uc.apiKeyRepo.List(ctx, input.AccountID, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}

	// Get all API keys for the account to calculate total
	allApiKeys, err := uc.apiKeyRepo.GetByAccountID(ctx, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all API keys for total count: %w", err)
	}

	// Create output
	output := &GetAPIKeysOutput{
		APIKeys: apiKeys,
		Limit:   input.Limit,
		Offset:  input.Offset,
		Total:   len(allApiKeys),
	}

	return output, nil
}

// validateInput validates the get API keys input
func (uc *GetAPIKeys) validateInput(input GetAPIKeysInput) error {
	if input.AccountID == uuid.Nil {
		return fmt.Errorf("account_id is required")
	}

	if input.Limit <= 0 || input.Limit > 100 {
		return fmt.Errorf("limit must be between 1 and 100")
	}

	if input.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	return nil
}
