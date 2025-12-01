package dto

import (
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// RegisterAppRequest represents a registration request
type RegisterAppRequest struct {
	Name       string  `json:"name" validate:"required,min=3,max=100"`
	WebhookURL *string `json:"webhook_url,omitempty" validate:"omitempty,url"`
}

// Validate validates the registration request
func (r *RegisterAppRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(r.Name) < 3 {
		return fmt.Errorf("name must be at least 3 characters")
	}

	if len(r.Name) > 100 {
		return fmt.Errorf("name must be at most 100 characters")
	}

	if r.WebhookURL != nil {
		if _, err := url.ParseRequestURI(*r.WebhookURL); err != nil {
			return fmt.Errorf("invalid webhook URL: %w", err)
		}
	}

	return nil
}

// RegisterAppResponse represents a registration response
type RegisterAppResponse struct {
	AccountID uuid.UUID `json:"account_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// IssueApiKeyRequest represents an API key issuance request
type IssueApiKeyRequest struct {
	AccountID   uuid.UUID `json:"account_id" validate:"required"`
	Name        string    `json:"name" validate:"required,min=3,max=100"`
	Permissions []string  `json:"permissions" validate:"required,dive,required,min=1"`
	ExpiresIn   *int      `json:"expires_in,omitempty" validate:"omitempty,min=1,max=8760"` // hours
}

// Validate validates the API key issuance request
func (r *IssueApiKeyRequest) Validate() error {
	if r.AccountID == uuid.Nil {
		return fmt.Errorf("account_id is required")
	}

	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(r.Name) < 3 {
		return fmt.Errorf("name must be at least 3 characters")
	}

	if len(r.Name) > 100 {
		return fmt.Errorf("name must be at most 100 characters")
	}

	if len(r.Permissions) == 0 {
		return fmt.Errorf("at least one permission is required")
	}

	for _, perm := range r.Permissions {
		if perm == "" {
			return fmt.Errorf("permission cannot be empty")
		}
	}

	if r.ExpiresIn != nil {
		if *r.ExpiresIn < 1 {
			return fmt.Errorf("expires_in must be at least 1 hour")
		}

		if *r.ExpiresIn > 8760 {
			return fmt.Errorf("expires_in must be at most 8760 hours (1 year)")
		}
	}

	return nil
}

// IssueApiKeyResponse represents an API key issuance response
type IssueApiKeyResponse struct {
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

// ValidateApiKeyRequest represents an API key validation request
type ValidateApiKeyRequest struct {
	KeyHash string `json:"key_hash" validate:"required"`
}

// Validate validates the API key validation request
func (r *ValidateApiKeyRequest) Validate() error {
	if r.KeyHash == "" {
		return fmt.Errorf("key_hash is required")
	}

	return nil
}

// ValidateApiKeyResponse represents an API key validation response
type ValidateApiKeyResponse struct {
	Valid       bool       `json:"valid"`
	AccountID   *uuid.UUID `json:"account_id,omitempty"`
	APIKeyID    *uuid.UUID `json:"api_key_id,omitempty"`
	Name        *string    `json:"name,omitempty"`
	Permissions []string   `json:"permissions,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ApiKeyResponse represents an API key in list responses
type ApiKeyResponse struct {
	APIKeyID    uuid.UUID  `json:"api_key_id"`
	Name        string     `json:"name"`
	Permissions []string   `json:"permissions"`
	Status      string     `json:"status"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// GetAPIKeysResponse represents a get API keys response
type GetAPIKeysResponse struct {
	APIKeys []ApiKeyResponse `json:"api_keys"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
	Total   int              `json:"total"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
}
