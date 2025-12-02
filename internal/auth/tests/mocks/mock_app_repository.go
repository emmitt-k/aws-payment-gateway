package mocks

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/aws-payment-gateway/internal/auth/domain"
)

// MockAppRepository is a mock implementation of the AppRepository interface
type MockAppRepository struct {
	accounts       map[uuid.UUID]*domain.Account
	accountsByName map[string]*domain.Account
	createError    error
	getError       error
	updateError    error
	deleteError    error
	listError      error
}

// NewMockAppRepository creates a new mock app repository
func NewMockAppRepository() *MockAppRepository {
	return &MockAppRepository{
		accounts:       make(map[uuid.UUID]*domain.Account),
		accountsByName: make(map[string]*domain.Account),
	}
}

// Create stores an account in the mock repository
func (m *MockAppRepository) Create(ctx context.Context, account *domain.Account) error {
	if m.createError != nil {
		return m.createError
	}

	m.accounts[account.ID] = account
	m.accountsByName[account.Name] = account
	return nil
}

// GetByID retrieves an account by ID from the mock repository
func (m *MockAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	account, exists := m.accounts[id]
	if !exists {
		return nil, nil // Not found
	}
	return account, nil
}

// GetByName retrieves an account by name from the mock repository
func (m *MockAppRepository) GetByName(ctx context.Context, name string) (*domain.Account, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	account, exists := m.accountsByName[name]
	if !exists {
		return nil, nil // Not found
	}
	return account, nil
}

// Update updates an account in the mock repository
func (m *MockAppRepository) Update(ctx context.Context, account *domain.Account) error {
	if m.updateError != nil {
		return m.updateError
	}

	if _, exists := m.accounts[account.ID]; !exists {
		return errors.New("account not found")
	}

	m.accounts[account.ID] = account
	m.accountsByName[account.Name] = account
	return nil
}

// Delete soft deletes an account in the mock repository
func (m *MockAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}

	account, exists := m.accounts[id]
	if !exists {
		return errors.New("account not found")
	}

	account.Status = domain.AccountStatusDeleted
	m.accounts[id] = account
	return nil
}

// List retrieves accounts with pagination from the mock repository
func (m *MockAppRepository) List(ctx context.Context, limit, offset int) ([]*domain.Account, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	accounts := make([]*domain.Account, 0, len(m.accounts))
	for _, account := range m.accounts {
		accounts = append(accounts, account)
	}

	// Apply pagination
	if offset >= len(accounts) {
		return []*domain.Account{}, nil
	}

	end := offset + limit
	if end > len(accounts) {
		end = len(accounts)
	}

	return accounts[offset:end], nil
}

// Helper methods for testing

// SetCreateError sets an error to be returned by Create
func (m *MockAppRepository) SetCreateError(err error) {
	m.createError = err
}

// SetGetError sets an error to be returned by Get methods
func (m *MockAppRepository) SetGetError(err error) {
	m.getError = err
}

// SetUpdateError sets an error to be returned by Update
func (m *MockAppRepository) SetUpdateError(err error) {
	m.updateError = err
}

// SetDeleteError sets an error to be returned by Delete
func (m *MockAppRepository) SetDeleteError(err error) {
	m.deleteError = err
}

// SetListError sets an error to be returned by List
func (m *MockAppRepository) SetListError(err error) {
	m.listError = err
}

// AddAccount adds an account directly to the mock repository
func (m *MockAppRepository) AddAccount(account *domain.Account) {
	m.accounts[account.ID] = account
	m.accountsByName[account.Name] = account
}

// Clear removes all accounts from the mock repository
func (m *MockAppRepository) Clear() {
	m.accounts = make(map[uuid.UUID]*domain.Account)
	m.accountsByName = make(map[string]*domain.Account)
	m.createError = nil
	m.getError = nil
	m.updateError = nil
	m.deleteError = nil
	m.listError = nil
}

// Count returns the number of accounts in the mock repository
func (m *MockAppRepository) Count() int {
	return len(m.accounts)
}
