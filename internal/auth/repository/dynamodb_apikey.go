package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/common/db"
)

// DynamoDBApiKeyRepository implements ApiKeyRepository using DynamoDB
type DynamoDBApiKeyRepository struct {
	client *db.DynamoDBClient
}

// NewDynamoDBApiKeyRepository creates a new DynamoDBApiKeyRepository
func NewDynamoDBApiKeyRepository(client *db.DynamoDBClient) *DynamoDBApiKeyRepository {
	return &DynamoDBApiKeyRepository{
		client: client,
	}
}

// DynamoDBApiKey represents the ApiKey entity in DynamoDB
type DynamoDBApiKey struct {
	domain.ApiKey
	PK     string `dynamodbav:"pk" json:"pk"`
	SK     string `dynamodbav:"sk" json:"sk"`
	GSI1PK string `dynamodbav:"gsi1pk" json:"gsi1pk"` // For lookup by key hash
}

// Create creates a new API key
func (r *DynamoDBApiKeyRepository) Create(ctx context.Context, apiKey *domain.ApiKey) error {
	// Set timestamps before creation
	now := time.Now()
	apiKey.CreatedAt = now

	// Create DynamoDB entity with composite key
	dynamoApiKey := &DynamoDBApiKey{
		ApiKey: *apiKey,
		PK:     fmt.Sprintf("ACCOUNT#%s", apiKey.AccountID.String()),
		SK:     fmt.Sprintf("APIKEY#%s", apiKey.ID.String()),
		GSI1PK: fmt.Sprintf("KEYHASH#%s", apiKey.KeyHash),
	}

	return r.client.PutItem(ctx, dynamoApiKey)
}

// GetByID retrieves an API key by its ID
func (r *DynamoDBApiKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ApiKey, error) {
	// We need the account ID to create the full key, so we'll need to query
	// In a real implementation, you might have a different key structure or GSI
	// For now, we'll scan for the API key by ID (not efficient for production)
	input := &dynamodb.ScanInput{
		TableName:        aws.String(r.client.GetTableName()),
		FilterExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: id.String()},
		},
		Limit: aws.Int32(1),
	}

	var results []DynamoDBApiKey
	err := r.client.ScanItems(ctx, input, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to scan for API key: %w", err)
	}

	if len(results) == 0 {
		return nil, nil // API key not found
	}

	return &results[0].ApiKey, nil
}

// GetByKeyHash retrieves an API key by its hash
func (r *DynamoDBApiKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*domain.ApiKey, error) {
	// Query using GSI on key hash
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.client.GetTableName()),
		IndexName:              aws.String("gsi1"), // Assuming GSI1 on key hash
		KeyConditionExpression: aws.String("gsi1pk = :gsi1pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("KEYHASH#%s", keyHash)},
		},
		Limit: aws.Int32(1),
	}

	var results []DynamoDBApiKey
	err := r.client.QueryItems(ctx, input, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to query API key by hash: %w", err)
	}

	if len(results) == 0 {
		return nil, nil // API key not found
	}

	// Update last used at
	now := time.Now()
	results[0].LastUsedAt = &now

	// Update the last used timestamp
	key, err := db.CreateCompositeKey("pk", results[0].PK, "sk", results[0].SK)
	if err != nil {
		return nil, fmt.Errorf("failed to create key for update: %w", err)
	}

	updateExpr := "SET last_used_at = :l"
	exprAttrValues := map[string]types.AttributeValue{
		":l": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
	}

	err = r.client.UpdateItem(ctx, key, updateExpr, nil, exprAttrValues, nil)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to update last_used_at for API key: %v\n", err)
	}

	return &results[0].ApiKey, nil
}

// GetByAccountID retrieves all API keys for an account
func (r *DynamoDBApiKeyRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.ApiKey, error) {
	// Query all API keys for an account
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.client.GetTableName()),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("ACCOUNT#%s", accountID.String())},
			":sk_prefix": &types.AttributeValueMemberS{Value: "APIKEY#"},
		},
	}

	var results []DynamoDBApiKey
	err := r.client.QueryItems(ctx, input, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys by account: %w", err)
	}

	apiKeys := make([]*domain.ApiKey, len(results))
	for i, result := range results {
		apiKeys[i] = &result.ApiKey
	}

	return apiKeys, nil
}

// ValidateByKey validates an API key by comparing the raw key with stored hashes
func (r *DynamoDBApiKeyRepository) ValidateByKey(ctx context.Context, rawKey string) (*domain.ApiKey, error) {
	// In a production system, you would want to optimize this with a GSI or caching
	// For now, we'll scan all active API keys and compare hashes
	input := &dynamodb.ScanInput{
		TableName:        aws.String(r.client.GetTableName()),
		FilterExpression: aws.String("contains(sk, :sk_prefix) AND #s = :status"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sk_prefix": &types.AttributeValueMemberS{Value: "APIKEY#"},
			":status":    &types.AttributeValueMemberS{Value: string(domain.ApiKeyStatusActive)},
		},
	}

	var results []DynamoDBApiKey
	err := r.client.ScanItems(ctx, input, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to scan API keys: %w", err)
	}

	// Check each API key hash against the provided raw key
	for _, result := range results {
		if err := bcrypt.CompareHashAndPassword([]byte(result.KeyHash), []byte(rawKey)); err == nil {
			// Found a match, update last used timestamp
			now := time.Now()
			result.LastUsedAt = &now

			// Update the last used timestamp
			key, err := db.CreateCompositeKey("pk", result.PK, "sk", result.SK)
			if err != nil {
				return nil, fmt.Errorf("failed to create key for update: %w", err)
			}

			updateExpr := "SET last_used_at = :l"
			exprAttrValues := map[string]types.AttributeValue{
				":l": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			}

			err = r.client.UpdateItem(ctx, key, updateExpr, nil, exprAttrValues, nil)
			if err != nil {
				// Log error but don't fail the request
				fmt.Printf("Failed to update last_used_at for API key: %v\n", err)
			}

			// Check if the key is expired
			if result.IsExpired() {
				return nil, nil // Key is expired
			}

			return &result.ApiKey, nil
		}
	}

	return nil, nil // No matching key found
}

// Update updates an existing API key
func (r *DynamoDBApiKeyRepository) Update(ctx context.Context, apiKey *domain.ApiKey) error {
	key, err := db.CreateCompositeKey("pk", fmt.Sprintf("ACCOUNT#%s", apiKey.AccountID.String()), "sk", fmt.Sprintf("APIKEY#%s", apiKey.ID.String()))
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	updateExpr := "SET #n = :n, #p = :p, #s = :s, #e = :e"
	exprAttrNames := map[string]string{
		"#n": "name",
		"#p": "permissions",
		"#s": "status",
		"#e": "expires_at",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":n": &types.AttributeValueMemberS{Value: apiKey.Name},
		":p": &types.AttributeValueMemberSS{Value: apiKey.Permissions},
		":s": &types.AttributeValueMemberS{Value: string(apiKey.Status)},
		":e": &types.AttributeValueMemberS{Value: apiKey.ExpiresAt.Format(time.RFC3339)},
	}

	var updatedApiKey DynamoDBApiKey
	err = r.client.UpdateItem(ctx, key, updateExpr, exprAttrNames, exprAttrValues, &updatedApiKey)
	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	return nil
}

// Delete soft deletes an API key by setting status to inactive
func (r *DynamoDBApiKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// First get the API key to get account ID
	apiKey, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get API key for deletion: %w", err)
	}
	if apiKey == nil {
		return fmt.Errorf("API key not found")
	}

	key, err := db.CreateCompositeKey("pk", fmt.Sprintf("ACCOUNT#%s", apiKey.AccountID.String()), "sk", fmt.Sprintf("APIKEY#%s", id.String()))
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	updateExpr := "SET #s = :s"
	exprAttrNames := map[string]string{
		"#s": "status",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":s": &types.AttributeValueMemberS{Value: string(domain.ApiKeyStatusInactive)},
	}

	err = r.client.UpdateItem(ctx, key, updateExpr, exprAttrNames, exprAttrValues, nil)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

// Revoke revokes an API key immediately
func (r *DynamoDBApiKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	// Revoking is the same as deleting in this implementation
	return r.Delete(ctx, id)
}

// List retrieves API keys with pagination
func (r *DynamoDBApiKeyRepository) List(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*domain.ApiKey, error) {
	// Query API keys for an account with pagination
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.client.GetTableName()),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("ACCOUNT#%s", accountID.String())},
			":sk_prefix": &types.AttributeValueMemberS{Value: "APIKEY#"},
		},
		Limit: aws.Int32(int32(limit)),
	}

	// Handle offset by using ExclusiveStartKey if needed
	if offset > 0 {
		// In a real implementation, you would need to store and use the last evaluated key
		// For simplicity, we're not implementing offset here
	}

	var results []DynamoDBApiKey
	err := r.client.QueryItems(ctx, input, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	apiKeys := make([]*domain.ApiKey, len(results))
	for i, result := range results {
		apiKeys[i] = &result.ApiKey
	}

	return apiKeys, nil
}
