package audit

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/aws-payment-gateway/internal/common/db"
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

// DynamoDBAuditLogger handles logging of audit events to DynamoDB
type DynamoDBAuditLogger struct {
	client *db.DynamoDBClient
}

// NewDynamoDBAuditLogger creates a new DynamoDBAuditLogger
func NewDynamoDBAuditLogger(client *db.DynamoDBClient) *DynamoDBAuditLogger {
	return &DynamoDBAuditLogger{
		client: client,
	}
}

// DynamoDBAuditEvent represents the audit event in DynamoDB
type DynamoDBAuditEvent struct {
	AuditEvent
	PK  string `dynamodbav:"pk" json:"pk"`
	SK  string `dynamodbav:"sk" json:"sk"`
	TTL int64  `dynamodbav:"ttl" json:"ttl"` // For automatic cleanup (90 days)
}

// LogAuthentication logs an authentication event to DynamoDB
func (a *DynamoDBAuditLogger) LogAuthentication(ctx context.Context, accountID, apiKeyID *uuid.UUID, apiKeyName *string, ipAddress, userAgent string, success bool, details map[string]string) {
	// Create DynamoDB event
	event := &DynamoDBAuditEvent{
		AuditEvent: AuditEvent{
			Timestamp:  time.Now(),
			EventType:  "authentication",
			AccountID:  accountID,
			APIKeyID:   apiKeyID,
			APIKeyName: apiKeyName,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Success:    success,
			Details:    details,
		},
		PK:  a.createPartitionKey("authentication", time.Now()),
		SK:  a.createSortKey(time.Now()),
		TTL: time.Now().Add(90 * 24 * time.Hour).Unix(), // 90-day TTL
	}

	// Store in DynamoDB with error handling
	if err := a.storeAuditEvent(ctx, event); err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to store authentication audit event in DynamoDB: %v", err)
	}
}

// LogAPIKeyCreation logs an API key creation event to DynamoDB
func (a *DynamoDBAuditLogger) LogAPIKeyCreation(ctx context.Context, accountID, apiKeyID *uuid.UUID, apiKeyName *string, ipAddress, userAgent string, details map[string]string) {
	// Create DynamoDB event
	event := &DynamoDBAuditEvent{
		AuditEvent: AuditEvent{
			Timestamp:  time.Now(),
			EventType:  "api_key_created",
			AccountID:  accountID,
			APIKeyID:   apiKeyID,
			APIKeyName: apiKeyName,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Success:    true,
			Details:    details,
		},
		PK:  a.createPartitionKey("api_key_created", time.Now()),
		SK:  a.createSortKey(time.Now()),
		TTL: time.Now().Add(90 * 24 * time.Hour).Unix(), // 90-day TTL
	}

	// Store in DynamoDB with error handling
	if err := a.storeAuditEvent(ctx, event); err != nil {
		// Log error but don't fail request
		log.Printf("Failed to store account creation audit event in DynamoDB: %v", err)
	}
}

// LogAPIKeyRevocation logs an API key revocation event to DynamoDB
func (a *DynamoDBAuditLogger) LogAPIKeyRevocation(ctx context.Context, accountID, apiKeyID *uuid.UUID, apiKeyName *string, ipAddress, userAgent string, details map[string]string) {
	// Create DynamoDB event
	event := &DynamoDBAuditEvent{
		AuditEvent: AuditEvent{
			Timestamp:  time.Now(),
			EventType:  "api_key_revoked",
			AccountID:  accountID,
			APIKeyID:   apiKeyID,
			APIKeyName: apiKeyName,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Success:    true,
			Details:    details,
		},
		PK:  a.createPartitionKey("api_key_revoked", time.Now()),
		SK:  a.createSortKey(time.Now()),
		TTL: time.Now().Add(90 * 24 * time.Hour).Unix(), // 90-day TTL
	}

	// Store in DynamoDB with error handling
	if err := a.storeAuditEvent(ctx, event); err != nil {
		// Log error but don't fail request
		log.Printf("Failed to store API key creation audit event in DynamoDB: %v", err)
	}
}

// LogAccountCreation logs an account creation event to DynamoDB
func (a *DynamoDBAuditLogger) LogAccountCreation(ctx context.Context, accountID *uuid.UUID, accountName *string, ipAddress, userAgent string, details map[string]string) {
	// Create DynamoDB event
	event := &DynamoDBAuditEvent{
		AuditEvent: AuditEvent{
			Timestamp: time.Now(),
			EventType: "account_created",
			AccountID: accountID,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Success:   true,
			Details:   details,
		},
		PK:  a.createPartitionKey("account_created", time.Now()),
		SK:  a.createSortKey(time.Now()),
		TTL: time.Now().Add(90 * 24 * time.Hour).Unix(), // 90-day TTL
	}

	// Store in DynamoDB with error handling
	if err := a.storeAuditEvent(ctx, event); err != nil {
		// Log error but don't fail request
		log.Printf("Failed to store API key revocation audit event in DynamoDB: %v", err)
	}
}

// QueryAuditLogs queries audit logs with filtering options
func (a *DynamoDBAuditLogger) QueryAuditLogs(ctx context.Context, eventType string, accountID *uuid.UUID, startTime, endTime time.Time, limit int) ([]*AuditEvent, error) {
	// Build query expression
	var keyCondition string
	var exprValues map[string]types.AttributeValue

	if eventType != "" && accountID != nil {
		// Query by both event type and account
		keyCondition = "pk = :pk AND begins_with(sk, :sk_prefix)"
		exprValues = map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("AUDIT#%s", eventType)},
			":sk_prefix": &types.AttributeValueMemberS{Value: fmt.Sprintf("%s#", startTime.Format("2006-01-02"))},
		}
	} else if eventType != "" {
		// Query by event type only
		keyCondition = "pk = :pk"
		exprValues = map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("AUDIT#%s", eventType)},
		}
	} else if accountID != nil {
		// Query by account only
		keyCondition = "pk = :pk"
		exprValues = map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("ACCOUNT#%s", accountID.String())},
		}
	} else {
		return nil, fmt.Errorf("at least one of eventType or accountID must be provided")
	}

	// Build query input
	input := &dynamodb.QueryInput{
		TableName:                 aws.String(a.client.GetTableName()),
		KeyConditionExpression:    aws.String(keyCondition),
		ExpressionAttributeValues: exprValues,
		Limit:                     aws.Int32(int32(limit)),
	}

	// Add time filter if specified
	if !startTime.IsZero() && !endTime.IsZero() {
		input.FilterExpression = aws.String("timestamp BETWEEN :start AND :end")
		if exprValues == nil {
			exprValues = make(map[string]types.AttributeValue)
		}
		exprValues[":start"] = &types.AttributeValueMemberS{Value: startTime.Format(time.RFC3339)}
		exprValues[":end"] = &types.AttributeValueMemberS{Value: endTime.Format(time.RFC3339)}
		input.ExpressionAttributeValues = exprValues
	}

	var results []DynamoDBAuditEvent
	err := a.client.QueryItems(ctx, input, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}

	// Convert to AuditEvent slice
	events := make([]*AuditEvent, len(results))
	for i, result := range results {
		events[i] = &result.AuditEvent
	}

	return events, nil
}

// createPartitionKey creates a partition key for audit events
func (a *DynamoDBAuditLogger) createPartitionKey(eventType string, timestamp time.Time) string {
	switch eventType {
	case "authentication":
		return fmt.Sprintf("AUDIT#AUTH#%s", timestamp.Format("2006-01-02"))
	case "api_key_created", "api_key_revoked":
		return fmt.Sprintf("AUDIT#APIKEY#%s", timestamp.Format("2006-01-02"))
	case "account_created":
		return fmt.Sprintf("AUDIT#ACCOUNT#%s", timestamp.Format("2006-01-02"))
	default:
		return fmt.Sprintf("AUDIT#%s#%s", eventType, timestamp.Format("2006-01-02"))
	}
}

// createSortKey creates a sort key for audit events
func (a *DynamoDBAuditLogger) createSortKey(timestamp time.Time) string {
	return fmt.Sprintf("%s#%s", timestamp.Format("2006-01-02"), timestamp.Unix())
}

// storeAuditEvent stores an audit event in DynamoDB with comprehensive error handling
func (a *DynamoDBAuditLogger) storeAuditEvent(ctx context.Context, event *DynamoDBAuditEvent) error {
	// Store in DynamoDB
	err := a.client.PutItem(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to store audit event in DynamoDB: %w", err)
	}
	return nil
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
