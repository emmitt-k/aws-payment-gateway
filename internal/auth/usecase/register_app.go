package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/repository"
	"github.com/google/uuid"
)

// RegisterAppInput represents the input for registering a new app
type RegisterAppInput struct {
	Name       string  `json:"name" validate:"required,min=3,max=100"`
	WebhookURL *string `json:"webhook_url,omitempty" validate:"omitempty,url"`
}

// RegisterAppOutput represents the output of app registration
type RegisterAppOutput struct {
	AccountID uuid.UUID `json:"account_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterApp handles the business logic for registering a new app
type RegisterApp struct {
	appRepo     repository.AppRepository
	accountRepo repository.ApiKeyRepository
}

// NewRegisterApp creates a new RegisterApp use case
func NewRegisterApp(appRepo repository.AppRepository, accountRepo repository.ApiKeyRepository) *RegisterApp {
	return &RegisterApp{
		appRepo:     appRepo,
		accountRepo: accountRepo,
	}
}

// Execute registers a new app and returns the result
func (uc *RegisterApp) Execute(ctx context.Context, input RegisterAppInput) (*RegisterAppOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Check if app name already exists
	existing, err := uc.appRepo.GetByName(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing app: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("app with name '%s' already exists", input.Name)
	}

	// Create new account
	account := &domain.Account{
		ID:        uuid.New(),
		Name:      input.Name,
		Status:    domain.AccountStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.appRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Create output
	output := &RegisterAppOutput{
		AccountID: account.ID,
		Name:      account.Name,
		Status:    string(account.Status),
		CreatedAt: account.CreatedAt,
	}

	return output, nil
}

// validateInput validates the registration input
func (uc *RegisterApp) validateInput(input RegisterAppInput) error {
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(input.Name) < 3 {
		return fmt.Errorf("name must be at least 3 characters")
	}

	if input.WebhookURL != nil && !isValidURL(*input.WebhookURL) {
		return fmt.Errorf("invalid webhook URL format")
	}

	return nil
}

// isValidURL performs basic URL validation
func isValidURL(url string) bool {
	// Basic URL validation - in production, use proper URL validation library
	return len(url) > 10 && (url[:7] == "http://" || url[:8] == "https://")
}
