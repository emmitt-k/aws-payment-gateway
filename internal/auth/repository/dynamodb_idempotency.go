package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/common/db"
)

// DynamoDBIdempotencyKeyRepository implements IdempotencyKeyRepository using DynamoDB
type DynamoDBIdempotencyKeyRepository struct {
	client *db.DynamoDBClient
}

// NewDynamoDBIdempotencyKeyRepository creates a new DynamoDBIdempotencyKeyRepository
func NewDynamoDBIdempotencyKeyRepository(client *db.DynamoDBClient) *DynamoDBIdempotencyKeyRepository {
	return &DynamoDBIdempotencyKeyRepository{
		client: client,
	}
}

// DynamoDBIdempotencyKey represents the IdempotencyKey entity in DynamoDB
type DynamoDBIdempotencyKey struct {
	domain.IdempotencyKey
	PK  string `dynamodbav:"pk" json:"pk"`
	SK  string `dynamodbav:"sk" json:"sk"`
	TTL int64  `dynamodbav:"ttl" json:"ttl"` // For automatic expiration
}

// Create creates a new idempotency key
func (r *DynamoDBIdempotencyKeyRepository) Create(ctx context.Context, key *domain.IdempotencyKey) error {
	// Set timestamps before creation
	now := time.Now()
	key.CreatedAt = now

	// Set 24-hour expiration
	key.ExpiresAt = now.Add(24 * time.Hour)

	// Create DynamoDB entity with composite key and TTL
	dynamoKey := &DynamoDBIdempotencyKey{
		IdempotencyKey: *key,
		PK:             fmt.Sprintf("IDEMPOTENCY#%s", key.ID.String()),
		SK:             fmt.Sprintf("KEY#%s", key.ID.String()),
		TTL:            key.ExpiresAt.Unix(), // Set TTL to expiration time
	}

	return r.client.PutItem(ctx, dynamoKey)
}

// GetByID retrieves an idempotency key by its ID
func (r *DynamoDBIdempotencyKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.IdempotencyKey, error) {
	// Use composite key for direct lookup
	key, err := db.CreateCompositeKey("pk", fmt.Sprintf("IDEMPOTENCY#%s", id.String()), "sk", fmt.Sprintf("KEY#%s", id.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.client.GetTableName()),
		Key:       key,
	}

	var result DynamoDBIdempotencyKey
	err = r.client.GetItem(ctx, input.Key, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get idempotency key: %w", err)
	}

	if result.ID == uuid.Nil {
		return nil, nil // Idempotency key not found
	}

	return &result.IdempotencyKey, nil
}

// GetByRequestHash retrieves an idempotency key by request hash
func (r *DynamoDBIdempotencyKeyRepository) GetByRequestHash(ctx context.Context, requestHash string) (*domain.IdempotencyKey, error) {
	// Use GSI for efficient request hash lookup
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.client.GetTableName()),
		IndexName:              aws.String("gsi1"), // GSI for request hash lookup
		KeyConditionExpression: aws.String("gsi1pk = :gsi1pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("REQUEST#%s", requestHash)},
		},
		Limit: aws.Int32(1),
	}

	var results []DynamoDBIdempotencyKey
	err := r.client.QueryItems(ctx, input, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to query idempotency key by request hash: %w", err)
	}

	if len(results) == 0 {
		return nil, nil // Idempotency key not found
	}

	return &results[0].IdempotencyKey, nil
}

// GetByAccountID retrieves all idempotency keys for an account
func (r *DynamoDBIdempotencyKeyRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.IdempotencyKey, error) {
	// Query all idempotency keys for an account
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.client.GetTableName()),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("IDEMPOTENCY#%s", accountID.String())},
			":sk_prefix": &types.AttributeValueMemberS{Value: "KEY#"},
		},
	}

	var results []DynamoDBIdempotencyKey
	err := r.client.QueryItems(ctx, input, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to query idempotency keys by account: %w", err)
	}

	keys := make([]*domain.IdempotencyKey, len(results))
	for i, result := range results {
		keys[i] = &result.IdempotencyKey
	}

	return keys, nil
}

// Update updates an existing idempotency key
func (r *DynamoDBIdempotencyKeyRepository) Update(ctx context.Context, key *domain.IdempotencyKey) error {
	compositeKey, err := db.CreateCompositeKey("pk", fmt.Sprintf("IDEMPOTENCY#%s", key.ID.String()), "sk", fmt.Sprintf("KEY#%s", key.ID.String()))
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	updateExpr := "SET #s = :s, #r = :r"
	exprAttrNames := map[string]string{
		"#s": "status",
		"#r": "response",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":s": &types.AttributeValueMemberS{Value: string(key.Status)},
		":r": &types.AttributeValueMemberS{Value: key.Response},
	}

	var updatedKey DynamoDBIdempotencyKey
	err = r.client.UpdateItem(ctx, compositeKey, updateExpr, exprAttrNames, exprAttrValues, &updatedKey)
	if err != nil {
		return fmt.Errorf("failed to update idempotency key: %w", err)
	}

	return nil
}

// Delete soft deletes an idempotency key by setting status to expired
func (r *DynamoDBIdempotencyKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	compositeKey, err := db.CreateCompositeKey("pk", fmt.Sprintf("IDEMPOTENCY#%s", id.String()), "sk", fmt.Sprintf("KEY#%s", id.String()))
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	updateExpr := "SET #s = :s"
	exprAttrNames := map[string]string{
		"#s": "status",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":s": &types.AttributeValueMemberS{Value: string(domain.IdempotencyKeyStatusExpired)},
	}

	err = r.client.UpdateItem(ctx, compositeKey, updateExpr, exprAttrNames, exprAttrValues, nil)
	if err != nil {
		return fmt.Errorf("failed to delete idempotency key: %w", err)
	}

	return nil
}

// CleanupExpired removes expired idempotency keys
func (r *DynamoDBIdempotencyKeyRepository) CleanupExpired(ctx context.Context) error {
	// Query for expired keys
	input := &dynamodb.ScanInput{
		TableName:        aws.String(r.client.GetTableName()),
		FilterExpression: aws.String("expires_at < :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":now": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	}

	var results []DynamoDBIdempotencyKey
	err := r.client.ScanItems(ctx, input, &results)
	if err != nil {
		return fmt.Errorf("failed to scan for expired idempotency keys: %w", err)
	}

	// Delete expired keys
	for _, result := range results {
		compositeKey, err := db.CreateCompositeKey("pk", result.PK, "sk", result.SK)
		if err != nil {
			continue // Skip if we can't create key
		}

		deleteErr := r.client.DeleteItem(ctx, compositeKey)
		if deleteErr != nil {
			// Log error but continue with cleanup
			fmt.Printf("Failed to delete expired idempotency key %s: %v\n", result.ID, deleteErr)
		}
	}

	return nil
}
