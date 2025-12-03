package audit

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
)

// AuditLoggerInterface defines the interface for audit logging
type AuditLoggerInterface interface {
	LogAuthentication(ctx context.Context, accountID, apiKeyID *uuid.UUID, apiKeyName *string, ipAddress, userAgent string, success bool, details map[string]string)
	LogAPIKeyCreation(ctx context.Context, accountID, apiKeyID *uuid.UUID, apiKeyName *string, ipAddress, userAgent string, details map[string]string)
	LogAPIKeyRevocation(ctx context.Context, accountID, apiKeyID *uuid.UUID, apiKeyName *string, ipAddress, userAgent string, details map[string]string)
	LogAccountCreation(ctx context.Context, accountID *uuid.UUID, accountName *string, ipAddress, userAgent string, details map[string]string)
}

// AuditEvent represents an audit log event
type AuditEvent struct {
	Timestamp  time.Time         `json:"timestamp"`
	EventType  string            `json:"event_type"`
	AccountID  *uuid.UUID        `json:"account_id,omitempty"`
	APIKeyID   *uuid.UUID        `json:"api_key_id,omitempty"`
	APIKeyName *string           `json:"api_key_name,omitempty"`
	IPAddress  string            `json:"ip_address"`
	UserAgent  string            `json:"user_agent"`
	Success    bool              `json:"success"`
	Details    map[string]string `json:"details,omitempty"`
}

// AuditLogger handles logging of authentication events
type AuditLogger struct {
	logger *log.Logger
}

// NewAuditLogger creates a new AuditLogger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logger: log.New(log.Writer(), "[AUDIT] ", log.LstdFlags),
	}
}

// LogAuthentication logs an authentication event
func (a *AuditLogger) LogAuthentication(ctx context.Context, accountID, apiKeyID *uuid.UUID, apiKeyName *string, ipAddress, userAgent string, success bool, details map[string]string) {
	event := AuditEvent{
		Timestamp:  time.Now(),
		EventType:  "authentication",
		AccountID:  accountID,
		APIKeyID:   apiKeyID,
		APIKeyName: apiKeyName,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Success:    success,
		Details:    details,
	}

	a.logEvent(event)
}

// LogAPIKeyCreation logs an API key creation event
func (a *AuditLogger) LogAPIKeyCreation(ctx context.Context, accountID, apiKeyID *uuid.UUID, apiKeyName *string, ipAddress, userAgent string, details map[string]string) {
	event := AuditEvent{
		Timestamp:  time.Now(),
		EventType:  "api_key_created",
		AccountID:  accountID,
		APIKeyID:   apiKeyID,
		APIKeyName: apiKeyName,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Success:    true,
		Details:    details,
	}

	a.logEvent(event)
}

// LogAPIKeyRevocation logs an API key revocation event
func (a *AuditLogger) LogAPIKeyRevocation(ctx context.Context, accountID, apiKeyID *uuid.UUID, apiKeyName *string, ipAddress, userAgent string, details map[string]string) {
	event := AuditEvent{
		Timestamp:  time.Now(),
		EventType:  "api_key_revoked",
		AccountID:  accountID,
		APIKeyID:   apiKeyID,
		APIKeyName: apiKeyName,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Success:    true,
		Details:    details,
	}

	a.logEvent(event)
}

// LogAccountCreation logs an account creation event
func (a *AuditLogger) LogAccountCreation(ctx context.Context, accountID *uuid.UUID, accountName *string, ipAddress, userAgent string, details map[string]string) {
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: "account_created",
		AccountID: accountID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
		Details:   details,
	}

	if details == nil {
		details = make(map[string]string)
	}
	details["account_name"] = *accountName
	event.Details = details

	a.logEvent(event)
}

// logEvent logs an audit event
func (a *AuditLogger) logEvent(event AuditEvent) {
	// Convert to JSON for structured logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		a.logger.Printf("Failed to marshal audit event: %v", err)
		return
	}

	a.logger.Printf("%s", string(eventJSON))
}

// GetEventDescription returns a human-readable description of an event type
func GetEventDescription(eventType string) string {
	descriptions := map[string]string{
		"authentication":  "API key authentication attempt",
		"api_key_created": "API key created",
		"api_key_revoked": "API key revoked",
		"account_created": "Account created",
	}

	if desc, exists := descriptions[eventType]; exists {
		return desc
	}
	return eventType
}
