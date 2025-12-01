package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/repository"
	"github.com/aws-payment-gateway/pkg/auth"
	"github.com/google/uuid"
)

// IssueApiKeyInput represents the input for issuing a new API key
type IssueApiKeyInput struct {
	AccountID   uuid.UUID `json:"account_id" validate:"required"`
	Name        string    `json:"name" validate:"required,min=3,max=100"`
	Permissions []string  `json:"permissions" validate:"required,dive,keys,required,min=1"`
	ExpiresIn   *int      `json:"expires_in,omitempty" validate:"omitempty,min=1,max=8760"` // hours
}

// IssueApiKeyOutput represents the output of API key issuance
type IssueApiKeyOutput struct {
	APIKeyID    uuid.UUID `json:"api_key_id"`
	APIKey      string    `json:"api_key"` // The actual API key (only returned once)
	KeyHash     string    `json:"key_hash"`
	AccountID   uuid.UUID `json:"account_id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// IssueApiKey handles the business logic for issuing a new API key
type IssueApiKey struct {
	accountRepo repository.AppRepository
	apiKeyRepo  repository.ApiKeyRepository
}

// NewIssueApiKey creates a new IssueApiKey use case
func NewIssueApiKey(accountRepo repository.AppRepository, apiKeyRepo repository.ApiKeyRepository) *IssueApiKey {
	return &IssueApiKey{
		accountRepo: accountRepo,
		apiKeyRepo:  apiKeyRepo,
	}
}

// Execute issues a new API key and returns the result
func (uc *IssueApiKey) Execute(ctx context.Context, input IssueApiKeyInput) (*IssueApiKeyOutput, error) {
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

	// Generate API key and hash
	apiKey, hashedKey, err := auth.GenerateAPIKeyWithHash()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Calculate expiration
	expiresAt := time.Now()
	if input.ExpiresIn != nil {
		expiresAt = expiresAt.Add(time.Duration(*input.ExpiresIn) * time.Hour)
	}

	// Create API key entity
	apiKeyEntity := &domain.ApiKey{
		ID:          uuid.New(),
		AccountID:   input.AccountID,
		Name:        input.Name,
		KeyHash:     string(hashedKey),
		Permissions: input.Permissions,
		Status:      domain.ApiKeyStatusActive,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}

	// Save to repository
	if err := uc.apiKeyRepo.Create(ctx, apiKeyEntity); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Create output
	output := &IssueApiKeyOutput{
		APIKeyID:    apiKeyEntity.ID,
		APIKey:      apiKey, // Only return the actual key once during creation
		KeyHash:     hashedKey,
		AccountID:   input.AccountID,
		Name:        input.Name,
		Permissions: input.Permissions,
		Status:      string(apiKeyEntity.Status),
		ExpiresAt:   apiKeyEntity.ExpiresAt,
		CreatedAt:   apiKeyEntity.CreatedAt,
	}

	return output, nil
}

// validateInput validates the API key issuance input
func (uc *IssueApiKey) validateInput(input IssueApiKeyInput) error {
	if len(input.Permissions) == 0 {
		return fmt.Errorf("at least one permission is required")
	}

	for _, perm := range input.Permissions {
		if !isValidPermission(perm) {
			return fmt.Errorf("invalid permission: %s", perm)
		}
	}

	return nil
}

// isValidPermission checks if a permission is valid
func isValidPermission(permission string) bool {
	validPermissions := []string{
		domain.PermissionReadAccounts,
		domain.PermissionWriteAccounts,
		domain.PermissionReadKeys,
		domain.PermissionWriteKeys,
		domain.PermissionManageWebhooks,
	}

	for _, valid := range validPermissions {
		if permission == valid {
			return true
		}
	}
	return false
}
