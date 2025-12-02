package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQLClient wraps the PostgreSQL database connection
type PostgreSQLClient struct {
	db *sql.DB
}

// NewPostgreSQLClient creates a new PostgreSQL client
func NewPostgreSQLClient(ctx context.Context, host, port, user, password, dbname string) (*PostgreSQLClient, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgreSQLClient{
		db: db,
	}, nil
}

// GetDB returns the underlying database connection
func (p *PostgreSQLClient) GetDB() *sql.DB {
	return p.db
}

// Close closes the database connection
func (p *PostgreSQLClient) Close() error {
	return p.db.Close()
}

// BeginTx begins a transaction
func (p *PostgreSQLClient) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, opts)
}

// ExecContext executes a query without returning any rows
func (p *PostgreSQLClient) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows
func (p *PostgreSQLClient) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row
func (p *PostgreSQLClient) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}
