package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aws-payment-gateway/internal/auth/domain"
)

// AppRepository defines the interface for account persistence operations
type AppRepository interface {
	// Create creates a new account
	Create(ctx context.Context, account *domain.Account) error

	// GetByID retrieves an account by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)

	// GetByName retrieves an account by its name
	GetByName(ctx context.Context, name string) (*domain.Account, error)

	// Update updates an existing account
	Update(ctx context.Context, account *domain.Account) error

	// Delete soft deletes an account by setting status to deleted
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves accounts with pagination
	List(ctx context.Context, limit, offset int) ([]*domain.Account, error)
}
