package mocks

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"

	"github.com/aws-payment-gateway/internal/auth/domain"
)

// MockApiKeyRepository is a mock implementation of the ApiKeyRepository interface
type MockApiKeyRepository struct {
	apiKeys       map[uuid.UUID]*domain.ApiKey
	apiKeysByHash map[string]*domain.ApiKey
	apiKeysByAcc  map[uuid.UUID][]*domain.ApiKey
	createError   error
	getError      error
	updateError   error
	deleteError   error
	revokeError   error
	listError     error
	validateError error
}

// NewMockApiKeyRepository creates a new mock API key repository
func NewMockApiKeyRepository() *MockApiKeyRepository {
	return &MockApiKeyRepository{
		apiKeys:       make(map[uuid.UUID]*domain.ApiKey),
		apiKeysByHash: make(map[string]*domain.ApiKey),
		apiKeysByAcc:  make(map[uuid.UUID][]*domain.ApiKey),
	}
}

// Create stores an API key in the mock repository
func (m *MockApiKeyRepository) Create(ctx context.Context, apiKey *domain.ApiKey) error {
	if m.createError != nil {
		return m.createError
	}

	m.apiKeys[apiKey.ID] = apiKey
	m.apiKeysByHash[apiKey.KeyHash] = apiKey

	if _, exists := m.apiKeysByAcc[apiKey.AccountID]; !exists {
		m.apiKeysByAcc[apiKey.AccountID] = []*domain.ApiKey{}
	}
	m.apiKeysByAcc[apiKey.AccountID] = append(m.apiKeysByAcc[apiKey.AccountID], apiKey)

	return nil
}

// GetByID retrieves an API key by ID from the mock repository
func (m *MockApiKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ApiKey, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	apiKey, exists := m.apiKeys[id]
	if !exists {
		return nil, nil // Not found
	}
	return apiKey, nil
}

// GetByKeyHash retrieves an API key by hash from the mock repository
func (m *MockApiKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*domain.ApiKey, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	apiKey, exists := m.apiKeysByHash[keyHash]
	if !exists {
		return nil, nil // Not found
	}
	return apiKey, nil
}

// GetByAccountID retrieves all API keys for an account from the mock repository
func (m *MockApiKeyRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.ApiKey, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	apiKeys, exists := m.apiKeysByAcc[accountID]
	if !exists {
		return []*domain.ApiKey{}, nil // Not found
	}
	return apiKeys, nil
}

// ValidateByKey validates an API key by comparing raw key with stored hashes
func (m *MockApiKeyRepository) ValidateByKey(ctx context.Context, rawKey string) (*domain.ApiKey, error) {
	if m.validateError != nil {
		return nil, m.validateError
	}

	// Special case for testing: if the raw key matches our test pattern,
	// extract the API key ID and look up by that ID instead
	if len(rawKey) > 14 && rawKey[:14] == "raw-api-key-" {
		// Extract UUID from the raw key
		uuidStr := rawKey[14:]
		if id, err := uuid.Parse(uuidStr); err == nil {
			if apiKey, exists := m.apiKeys[id]; exists {
				return apiKey, nil
			}
		}
	}

	// Hash the raw key
	hash := sha256.Sum256([]byte(rawKey))
	hashStr := hex.EncodeToString(hash[:])

	// Look up by hash
	apiKey, exists := m.apiKeysByHash[hashStr]
	if !exists {
		return nil, nil // Not found
	}
	return apiKey, nil
}

// Update updates an API key in the mock repository
func (m *MockApiKeyRepository) Update(ctx context.Context, apiKey *domain.ApiKey) error {
	if m.updateError != nil {
		return m.updateError
	}

	if _, exists := m.apiKeys[apiKey.ID]; !exists {
		return errors.New("API key not found")
	}

	m.apiKeys[apiKey.ID] = apiKey
	m.apiKeysByHash[apiKey.KeyHash] = apiKey

	// Update in account mapping
	for i, existingKey := range m.apiKeysByAcc[apiKey.AccountID] {
		if existingKey.ID == apiKey.ID {
			m.apiKeysByAcc[apiKey.AccountID][i] = apiKey
			break
		}
	}

	return nil
}

// Delete soft deletes an API key by setting status to inactive
func (m *MockApiKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}

	apiKey, exists := m.apiKeys[id]
	if !exists {
		return errors.New("API key not found")
	}

	apiKey.Status = domain.ApiKeyStatusInactive
	m.apiKeys[id] = apiKey
	m.apiKeysByHash[apiKey.KeyHash] = apiKey

	return nil
}

// Revoke revokes an API key immediately
func (m *MockApiKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if m.revokeError != nil {
		return m.revokeError
	}

	apiKey, exists := m.apiKeys[id]
	if !exists {
		return errors.New("API key not found")
	}

	apiKey.Status = domain.ApiKeyStatusInactive
	m.apiKeys[id] = apiKey
	m.apiKeysByHash[apiKey.KeyHash] = apiKey

	return nil
}

// List retrieves API keys with pagination from the mock repository
func (m *MockApiKeyRepository) List(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*domain.ApiKey, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	apiKeys, exists := m.apiKeysByAcc[accountID]
	if !exists {
		return []*domain.ApiKey{}, nil
	}

	// Apply pagination
	if offset >= len(apiKeys) {
		return []*domain.ApiKey{}, nil
	}

	end := offset + limit
	if end > len(apiKeys) {
		end = len(apiKeys)
	}

	return apiKeys[offset:end], nil
}

// Helper methods for testing

// SetCreateError sets an error to be returned by Create
func (m *MockApiKeyRepository) SetCreateError(err error) {
	m.createError = err
}

// SetGetError sets an error to be returned by Get methods
func (m *MockApiKeyRepository) SetGetError(err error) {
	m.getError = err
}

// SetUpdateError sets an error to be returned by Update
func (m *MockApiKeyRepository) SetUpdateError(err error) {
	m.updateError = err
}

// SetDeleteError sets an error to be returned by Delete
func (m *MockApiKeyRepository) SetDeleteError(err error) {
	m.deleteError = err
}

// SetRevokeError sets an error to be returned by Revoke
func (m *MockApiKeyRepository) SetRevokeError(err error) {
	m.revokeError = err
}

// SetListError sets an error to be returned by List
func (m *MockApiKeyRepository) SetListError(err error) {
	m.listError = err
}

// SetValidateError sets an error to be returned by ValidateByKey
func (m *MockApiKeyRepository) SetValidateError(err error) {
	m.validateError = err
}

// AddApiKey adds an API key directly to the mock repository
func (m *MockApiKeyRepository) AddApiKey(apiKey *domain.ApiKey) {
	m.apiKeys[apiKey.ID] = apiKey
	m.apiKeysByHash[apiKey.KeyHash] = apiKey

	if _, exists := m.apiKeysByAcc[apiKey.AccountID]; !exists {
		m.apiKeysByAcc[apiKey.AccountID] = []*domain.ApiKey{}
	}
	m.apiKeysByAcc[apiKey.AccountID] = append(m.apiKeysByAcc[apiKey.AccountID], apiKey)
}

// Clear removes all API keys from the mock repository
func (m *MockApiKeyRepository) Clear() {
	m.apiKeys = make(map[uuid.UUID]*domain.ApiKey)
	m.apiKeysByHash = make(map[string]*domain.ApiKey)
	m.apiKeysByAcc = make(map[uuid.UUID][]*domain.ApiKey)
	m.createError = nil
	m.getError = nil
	m.updateError = nil
	m.deleteError = nil
	m.revokeError = nil
	m.listError = nil
	m.validateError = nil
}

// Count returns the number of API keys in the mock repository
func (m *MockApiKeyRepository) Count() int {
	return len(m.apiKeys)
}

// CountByAccountID returns the number of API keys for a specific account
func (m *MockApiKeyRepository) CountByAccountID(accountID uuid.UUID) int {
	if apiKeys, exists := m.apiKeysByAcc[accountID]; exists {
		return len(apiKeys)
	}
	return 0
}
