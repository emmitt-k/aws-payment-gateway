package domain

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a specific error code for auth operations
type ErrorCode string

const (
	// Authentication errors
	ErrCodeMissingAPIKey    ErrorCode = "missing_api_key"
	ErrCodeInvalidAPIKey    ErrorCode = "invalid_api_key"
	ErrCodeExpiredAPIKey    ErrorCode = "expired_api_key"
	ErrCodeInactiveAPIKey   ErrorCode = "inactive_api_key"
	ErrCodeInactiveAccount  ErrorCode = "inactive_account"
	ErrCodeValidationFailed ErrorCode = "validation_failed"

	// Rate limiting errors
	ErrCodeRateLimitExceeded    ErrorCode = "rate_limit_exceeded"
	ErrCodeRateLimitCheckFailed ErrorCode = "rate_limit_check_failed"

	// Idempotency errors
	ErrCodeIdempotencyKeyPending     ErrorCode = "idempotency_key_pending"
	ErrCodeIdempotencyKeyExpired     ErrorCode = "idempotency_key_expired"
	ErrCodeIdempotencyCheckFailed    ErrorCode = "idempotency_check_failed"
	ErrCodeIdempotencyCreateFailed   ErrorCode = "idempotency_create_failed"
	ErrCodeIdempotencyCompleteFailed ErrorCode = "idempotency_complete_failed"

	// Permission errors
	ErrCodeInsufficientPermissions ErrorCode = "insufficient_permissions"
	ErrCodeNotAuthenticated        ErrorCode = "not_authenticated"

	// System errors
	ErrCodeInternalError      ErrorCode = "internal_error"
	ErrCodeDatabaseError      ErrorCode = "database_error"
	ErrCodeServiceUnavailable ErrorCode = "service_unavailable"
)

// AuthError represents a structured error with code and details
type AuthError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
}

// Error implements the error interface
func (e *AuthError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %s (details: %+v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAuthError creates a new AuthError
func NewAuthError(code ErrorCode, message string) *AuthError {
	return &AuthError{
		Code:       code,
		Message:    message,
		StatusCode: getHTTPStatusForError(code),
	}
}

// NewAuthErrorWithDetails creates a new AuthError with details
func NewAuthErrorWithDetails(code ErrorCode, message string, details map[string]interface{}) *AuthError {
	return &AuthError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: getHTTPStatusForError(code),
	}
}

// getHTTPStatusForError returns appropriate HTTP status code for error
func getHTTPStatusForError(code ErrorCode) int {
	switch code {
	case ErrCodeMissingAPIKey, ErrCodeInvalidAPIKey, ErrCodeExpiredAPIKey, ErrCodeInactiveAPIKey:
		return http.StatusUnauthorized
	case ErrCodeInactiveAccount:
		return http.StatusForbidden
	case ErrCodeInsufficientPermissions:
		return http.StatusForbidden
	case ErrCodeNotAuthenticated:
		return http.StatusUnauthorized
	case ErrCodeRateLimitExceeded:
		return http.StatusTooManyRequests
	case ErrCodeIdempotencyKeyPending:
		return http.StatusConflict
	case ErrCodeValidationFailed:
		return http.StatusBadRequest
	case ErrCodeInternalError, ErrCodeDatabaseError:
		return http.StatusInternalServerError
	case ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// Common error instances
var (
	ErrMissingAPIKey           = NewAuthError(ErrCodeMissingAPIKey, "API key is required")
	ErrInvalidAPIKey           = NewAuthError(ErrCodeInvalidAPIKey, "API key is invalid or expired")
	ErrExpiredAPIKey           = NewAuthError(ErrCodeExpiredAPIKey, "API key has expired")
	ErrInactiveAccount         = NewAuthError(ErrCodeInactiveAccount, "Account is not active")
	ErrRateLimitExceeded       = NewAuthError(ErrCodeRateLimitExceeded, "Rate limit exceeded")
	ErrInsufficientPermissions = NewAuthError(ErrCodeInsufficientPermissions, "Insufficient permissions")
	ErrNotAuthenticated        = NewAuthError(ErrCodeNotAuthenticated, "Authentication required")
	ErrInternalError           = NewAuthError(ErrCodeInternalError, "Internal server error")
)
