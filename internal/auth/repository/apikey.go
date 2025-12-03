package repository

import (
	"context"
	"time"

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

// IdempotencyKeyRepository defines the interface for idempotency key persistence operations
type IdempotencyKeyRepository interface {
	// Create creates a new idempotency key
	Create(ctx context.Context, key *domain.IdempotencyKey) error

	// GetByID retrieves an idempotency key by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.IdempotencyKey, error)

	// GetByRequestHash retrieves an idempotency key by request hash
	GetByRequestHash(ctx context.Context, requestHash string) (*domain.IdempotencyKey, error)

	// GetByAccountID retrieves all idempotency keys for an account
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.IdempotencyKey, error)

	// Update updates an existing idempotency key
	Update(ctx context.Context, key *domain.IdempotencyKey) error

	// Delete soft deletes an idempotency key by setting status to expired
	Delete(ctx context.Context, id uuid.UUID) error

	// CleanupExpired removes expired idempotency keys
	CleanupExpired(ctx context.Context) error
}

// RateLimitRepository defines the interface for rate limiting operations
type RateLimitRepository interface {
	// CheckRateLimit checks if a request exceeds the rate limit
	CheckRateLimit(ctx context.Context, key string, requests int, window time.Duration) (bool, int, int64, error)

	// IncrementRateLimit increments the counter for a key
	IncrementRateLimit(ctx context.Context, key string, window time.Duration) error

	// ResetRateLimit resets the counter for a key
	ResetRateLimit(ctx context.Context, key string) error
}
