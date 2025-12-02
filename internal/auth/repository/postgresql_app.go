package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/common/db"
)

// PostgreSQLAppRepository implements AppRepository using PostgreSQL
type PostgreSQLAppRepository struct {
	client *db.PostgreSQLClient
}

// NewPostgreSQLAppRepository creates a new PostgreSQLAppRepository
func NewPostgreSQLAppRepository(client *db.PostgreSQLClient) *PostgreSQLAppRepository {
	return &PostgreSQLAppRepository{
		client: client,
	}
}

// Create creates a new account
func (r *PostgreSQLAppRepository) Create(ctx context.Context, account *domain.Account) error {
	// Set timestamps before creation
	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	query := `
		INSERT INTO accounts (id, name, status, webhook_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.client.ExecContext(ctx, query,
		account.ID,
		account.Name,
		string(account.Status),
		account.WebhookURL,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetByID retrieves an account by its ID
func (r *PostgreSQLAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	query := `
		SELECT id, name, status, webhook_url, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	var account domain.Account
	var webhookURL sql.NullString

	err := r.client.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.Name,
		&account.Status,
		&webhookURL,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Account not found
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Handle nullable webhook URL
	if webhookURL.Valid {
		account.WebhookURL = &webhookURL.String
	}

	return &account, nil
}

// GetByName retrieves an account by its name
func (r *PostgreSQLAppRepository) GetByName(ctx context.Context, name string) (*domain.Account, error) {
	query := `
		SELECT id, name, status, webhook_url, created_at, updated_at
		FROM accounts
		WHERE name = $1
	`

	var account domain.Account
	var webhookURL sql.NullString

	err := r.client.QueryRowContext(ctx, query, name).Scan(
		&account.ID,
		&account.Name,
		&account.Status,
		&webhookURL,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Account not found
		}
		return nil, fmt.Errorf("failed to get account by name: %w", err)
	}

	// Handle nullable webhook URL
	if webhookURL.Valid {
		account.WebhookURL = &webhookURL.String
	}

	return &account, nil
}

// Update updates an existing account
func (r *PostgreSQLAppRepository) Update(ctx context.Context, account *domain.Account) error {
	// Update timestamp
	account.UpdatedAt = time.Now()

	query := `
		UPDATE accounts
		SET name = $2, status = $3, webhook_url = $4, updated_at = $5
		WHERE id = $1
	`

	_, err := r.client.ExecContext(ctx, query,
		account.ID,
		account.Name,
		string(account.Status),
		account.WebhookURL,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// Delete soft deletes an account by setting status to deleted
func (r *PostgreSQLAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE accounts
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.client.ExecContext(ctx, query,
		id,
		string(domain.AccountStatusDeleted),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

// List retrieves accounts with pagination
func (r *PostgreSQLAppRepository) List(ctx context.Context, limit, offset int) ([]*domain.Account, error) {
	query := `
		SELECT id, name, status, webhook_url, created_at, updated_at
		FROM accounts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.client.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.Account

	for rows.Next() {
		var account domain.Account
		var webhookURL sql.NullString

		err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Status,
			&webhookURL,
			&account.CreatedAt,
			&account.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}

		// Handle nullable webhook URL
		if webhookURL.Valid {
			account.WebhookURL = &webhookURL.String
		}

		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate accounts: %w", err)
	}

	return accounts, nil
}

// CreateTx creates a new account within a transaction
func (r *PostgreSQLAppRepository) CreateTx(ctx context.Context, tx *sql.Tx, account *domain.Account) error {
	// Set timestamps before creation
	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	query := `
		INSERT INTO accounts (id, name, status, webhook_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := tx.ExecContext(ctx, query,
		account.ID,
		account.Name,
		string(account.Status),
		account.WebhookURL,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create account in transaction: %w", err)
	}

	return nil
}

// UpdateTx updates an existing account within a transaction
func (r *PostgreSQLAppRepository) UpdateTx(ctx context.Context, tx *sql.Tx, account *domain.Account) error {
	// Update timestamp
	account.UpdatedAt = time.Now()

	query := `
		UPDATE accounts
		SET name = $2, status = $3, webhook_url = $4, updated_at = $5
		WHERE id = $1
	`

	_, err := tx.ExecContext(ctx, query,
		account.ID,
		account.Name,
		string(account.Status),
		account.WebhookURL,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update account in transaction: %w", err)
	}

	return nil
}
