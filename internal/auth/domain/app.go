package domain

import (
	"time"

	"github.com/google/uuid"
)

// AccountStatus represents the status of an account
type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusDeleted   AccountStatus = "deleted"
)

// Account represents a company account in the system
type Account struct {
	ID         uuid.UUID     `json:"id" db:"id"`
	Name       string        `json:"name" db:"name"`
	Status     AccountStatus `json:"status" db:"status"`
	WebhookURL *string       `json:"webhook_url,omitempty" db:"webhook_url"`
	CreatedAt  time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at" db:"updated_at"`
}

// IsValid checks if the account is in a valid state
func (a *Account) IsValid() bool {
	return a.Status == AccountStatusActive
}

// ApiKeyStatus represents the status of an API key
type ApiKeyStatus string

const (
	ApiKeyStatusActive   ApiKeyStatus = "active"
	ApiKeyStatusInactive ApiKeyStatus = "inactive"
)

// ApiKeyPermissions represents the permissions granted to an API key
type ApiKeyPermissions []string

const (
	PermissionReadAccounts   = "read:accounts"
	PermissionWriteAccounts  = "write:accounts"
	PermissionReadKeys       = "read:keys"
	PermissionWriteKeys      = "write:keys"
	PermissionManageWebhooks = "manage:webhooks"
)

// ApiKey represents an API key for external client access
type ApiKey struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	AccountID   uuid.UUID         `json:"account_id" db:"account_id"`
	Name        string            `json:"name" db:"name"`
	KeyHash     string            `json:"key_hash" db:"key_hash"`
	Permissions ApiKeyPermissions `json:"permissions" db:"permissions"`
	Status      ApiKeyStatus      `json:"status" db:"status"`
	LastUsedAt  *time.Time        `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt   time.Time         `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
}

// IsValid checks if the API key is in a valid state
func (k *ApiKey) IsValid() bool {
	return k.Status == ApiKeyStatusActive && time.Now().Before(k.ExpiresAt)
}

// HasPermission checks if the API key has a specific permission
func (k *ApiKey) HasPermission(permission string) bool {
	for _, p := range k.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsExpired checks if the API key has expired
func (k *ApiKey) IsExpired() bool {
	return time.Now().After(k.ExpiresAt)
}
