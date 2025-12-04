package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/repository"
)

// CheckIdempotencyInput represents the input for checking idempotency
type CheckIdempotencyInput struct {
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
	RequestHash    string `json:"request_hash" validate:"required"`
}

// CheckIdempotencyOutput represents the output of checking idempotency
type CheckIdempotencyOutput struct {
	Exists    bool       `json:"exists"`
	Status    string     `json:"status,omitempty"`
	Response  string     `json:"response,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// CheckIdempotency handles checking if an idempotency key exists and its status
type CheckIdempotency struct {
	idempotencyRepo repository.IdempotencyKeyRepository
}

// NewCheckIdempotency creates a new CheckIdempotency use case
func NewCheckIdempotency(idempotencyRepo repository.IdempotencyKeyRepository) *CheckIdempotency {
	return &CheckIdempotency{
		idempotencyRepo: idempotencyRepo,
	}
}

// Execute checks if an idempotency key exists and returns its status
func (uc *CheckIdempotency) Execute(ctx context.Context, input CheckIdempotencyInput) (*CheckIdempotencyOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Check if idempotency key exists
	key, err := uc.idempotencyRepo.GetByRequestHash(ctx, input.RequestHash)
	if err != nil {
		return nil, fmt.Errorf("failed to check idempotency key: %w", err)
	}

	if key == nil {
		// No existing key, this is a new request
		return &CheckIdempotencyOutput{
			Exists: false,
		}, nil
	}

	// Check if the key has expired
	if key.IsExpired() {
		return &CheckIdempotencyOutput{
			Exists: true,
			Status: string(domain.IdempotencyKeyStatusExpired),
		}, nil
	}

	// Return the key status and response if completed
	if key.Status == domain.IdempotencyKeyStatusCompleted {
		return &CheckIdempotencyOutput{
			Exists:    true,
			Status:    string(key.Status),
			Response:  key.Response,
			CreatedAt: &key.CreatedAt,
		}, nil
	}

	// Key exists but is still pending
	return &CheckIdempotencyOutput{
		Exists:    true,
		Status:    string(key.Status),
		CreatedAt: &key.CreatedAt,
	}, nil
}

// validateInput validates the input for checking idempotency
func (uc *CheckIdempotency) validateInput(input CheckIdempotencyInput) error {
	if input.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	if input.RequestHash == "" {
		return fmt.Errorf("request_hash is required")
	}
	return nil
}

// CreateIdempotencyInput represents the input for creating idempotency
type CreateIdempotencyInput struct {
	IdempotencyKey string    `json:"idempotency_key" validate:"required"`
	RequestHash    string    `json:"request_hash" validate:"required"`
	Response       string    `json:"response,omitempty"`
	AccountID      uuid.UUID `json:"account_id,omitempty"` // Optional: can be extracted from context
}

// CreateIdempotencyOutput represents the output of creating idempotency
type CreateIdempotencyOutput struct {
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// CreateIdempotency handles creating new idempotency keys
type CreateIdempotency struct {
	idempotencyRepo repository.IdempotencyKeyRepository
}

// NewCreateIdempotency creates a new CreateIdempotency use case
func NewCreateIdempotency(idempotencyRepo repository.IdempotencyKeyRepository) *CreateIdempotency {
	return &CreateIdempotency{
		idempotencyRepo: idempotencyRepo,
	}
}

// Execute creates a new idempotency key
func (uc *CreateIdempotency) Execute(ctx context.Context, input CreateIdempotencyInput) (*CreateIdempotencyOutput, error) {
	// Validate input
	if err := uc.validateCreateInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Create idempotency key
	now := time.Now()
	accountID := input.AccountID
	if accountID == uuid.Nil {
		accountID = uuid.New() // Fallback for testing/unauthenticated contexts
	}

	key := &domain.IdempotencyKey{
		ID:          uuid.New(),
		AccountID:   accountID,
		RequestHash: input.RequestHash,
		Status:      domain.IdempotencyKeyStatusPending,
		Response:    input.Response,
		CreatedAt:   now,
		ExpiresAt:   now.Add(24 * time.Hour), // 24-hour TTL
	}

	err := uc.idempotencyRepo.Create(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create idempotency key: %w", err)
	}

	return &CreateIdempotencyOutput{
		IdempotencyKey: key.ID.String(),
		CreatedAt:      key.CreatedAt,
		ExpiresAt:      key.ExpiresAt,
	}, nil
}

// CompleteIdempotencyInput represents the input for completing idempotency
type CompleteIdempotencyInput struct {
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
	Response       string `json:"response" validate:"required"`
}

// CompleteIdempotencyOutput represents the output of completing idempotency
type CompleteIdempotencyOutput struct {
	IdempotencyKey string    `json:"idempotency_key"`
	Status         string    `json:"status"`
	CompletedAt    time.Time `json:"completed_at"`
}

// CompleteIdempotency handles completing idempotency keys
type CompleteIdempotency struct {
	idempotencyRepo repository.IdempotencyKeyRepository
}

// NewCompleteIdempotency creates a new CompleteIdempotency use case
func NewCompleteIdempotency(idempotencyRepo repository.IdempotencyKeyRepository) *CompleteIdempotency {
	return &CompleteIdempotency{
		idempotencyRepo: idempotencyRepo,
	}
}

// Execute completes an idempotency key
func (uc *CompleteIdempotency) Execute(ctx context.Context, input CompleteIdempotencyInput) (*CompleteIdempotencyOutput, error) {
	// Validate input
	if err := uc.validateCompleteInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Get the existing idempotency key by ID (not request hash)
	keyUUID, err := uuid.Parse(input.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("invalid idempotency key format: %w", err)
	}

	key, err := uc.idempotencyRepo.GetByID(ctx, keyUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get idempotency key: %w", err)
	}

	if key == nil {
		return nil, fmt.Errorf("idempotency key not found")
	}

	// Check if key is still pending
	if key.Status != domain.IdempotencyKeyStatusPending {
		return nil, fmt.Errorf("idempotency key is not in pending status")
	}

	// Update key to completed status
	now := time.Now()
	key.Status = domain.IdempotencyKeyStatusCompleted
	key.Response = input.Response

	err = uc.idempotencyRepo.Update(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to complete idempotency key: %w", err)
	}

	return &CompleteIdempotencyOutput{
		IdempotencyKey: key.ID.String(),
		Status:         string(key.Status),
		CompletedAt:    now,
	}, nil
}

// validateCreateInput validates the input for creating idempotency
func (uc *CreateIdempotency) validateCreateInput(input CreateIdempotencyInput) error {
	if input.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	if input.RequestHash == "" {
		return fmt.Errorf("request_hash is required")
	}
	return nil
}

// validateCompleteInput validates the input for completing idempotency
func (uc *CompleteIdempotency) validateCompleteInput(input CompleteIdempotencyInput) error {
	if input.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	if input.Response == "" {
		return fmt.Errorf("response is required")
	}
	return nil
}
