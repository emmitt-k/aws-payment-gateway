package domain

import (
	"time"

	"github.com/google/uuid"
)

// IdempotencyKeyStatus represents the status of an idempotency key
type IdempotencyKeyStatus string

const (
	IdempotencyKeyStatusPending   IdempotencyKeyStatus = "pending"
	IdempotencyKeyStatusCompleted IdempotencyKeyStatus = "completed"
	IdempotencyKeyStatusExpired   IdempotencyKeyStatus = "expired"
)

// IdempotencyKey represents an idempotency key for preventing duplicate processing
type IdempotencyKey struct {
	ID          uuid.UUID            `json:"id" db:"id"`
	AccountID   uuid.UUID            `json:"account_id" db:"account_id"`
	RequestHash string               `json:"request_hash" db:"request_hash"`
	Status      IdempotencyKeyStatus `json:"status" db:"status"`
	Response    string               `json:"response,omitempty" db:"response,omitempty"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	ExpiresAt   time.Time            `json:"expires_at" db:"expires_at"`
}

// IsExpired checks if the idempotency key has expired
func (k *IdempotencyKey) IsExpired() bool {
	return time.Now().After(k.ExpiresAt)
}

// IsValid checks if the idempotency key is in a valid state
func (k *IdempotencyKey) IsValid() bool {
	return k.Status == IdempotencyKeyStatusPending && !k.IsExpired()
}
