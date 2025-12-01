package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aws-payment-gateway/internal/auth/domain"
)

// ApiKeyRepository defines the interface for API key persistence operations
type ApiKeyRepository interface {
	// Create creates a new API key
	Create(ctx context.Context, apiKey *domain.ApiKey) error

	// GetByID retrieves an API key by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ApiKey, error)

	// GetByKeyHash retrieves an API key by its hash
	GetByKeyHash(ctx context.Context, keyHash string) (*domain.ApiKey, error)

	// GetByAccountID retrieves all API keys for an account
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.ApiKey, error)

	// ValidateByKey validates an API key by comparing the raw key with stored hashes
	ValidateByKey(ctx context.Context, rawKey string) (*domain.ApiKey, error)

	// Update updates an existing API key
	Update(ctx context.Context, apiKey *domain.ApiKey) error

	// Delete soft deletes an API key by setting status to inactive
	Delete(ctx context.Context, id uuid.UUID) error

	// Revoke revokes an API key immediately
	Revoke(ctx context.Context, id uuid.UUID) error

	// List retrieves API keys with pagination
	List(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*domain.ApiKey, error)
}
