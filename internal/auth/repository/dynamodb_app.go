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

// DynamoDBAppRepository implements AppRepository using DynamoDB
type DynamoDBAppRepository struct {
	client *db.DynamoDBClient
}

// NewDynamoDBAppRepository creates a new DynamoDBAppRepository
func NewDynamoDBAppRepository(client *db.DynamoDBClient) *DynamoDBAppRepository {
	return &DynamoDBAppRepository{
		client: client,
	}
}

// DynamoDBAccount represents the Account entity in DynamoDB
type DynamoDBAccount struct {
	domain.Account
	PK string `dynamodbav:"pk" json:"pk"`
	SK string `dynamodbav:"sk" json:"sk"`
}

// Create creates a new account
func (r *DynamoDBAppRepository) Create(ctx context.Context, account *domain.Account) error {
	// Set timestamps before creation
	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	// Create DynamoDB entity with composite key
	dynamoAccount := &DynamoDBAccount{
		Account: *account,
		PK:      fmt.Sprintf("ACCOUNT#%s", account.ID.String()),
		SK:      "ACCOUNT",
	}

	return r.client.PutItem(ctx, dynamoAccount)
}

// GetByID retrieves an account by its ID
func (r *DynamoDBAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	key, err := db.CreateCompositeKey("pk", fmt.Sprintf("ACCOUNT#%s", id.String()), "sk", "ACCOUNT")
	if err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}

	var dynamoAccount DynamoDBAccount
	err = r.client.GetItem(ctx, key, &dynamoAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	if dynamoAccount.ID == uuid.Nil {
		return nil, nil // Account not found
	}

	return &dynamoAccount.Account, nil
}

// GetByName retrieves an account by its name
func (r *DynamoDBAppRepository) GetByName(ctx context.Context, name string) (*domain.Account, error) {
	// Query using GSI on name (assuming GSI exists)
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.client.GetTableName()),
		IndexName:              aws.String("gsi1"), // Assuming GSI1 on name
		KeyConditionExpression: aws.String("gsi1pk = :gsi1pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("NAME#%s", name)},
		},
		Limit: aws.Int32(1),
	}

	var results []DynamoDBAccount
	err := r.client.QueryItems(ctx, input, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to query account by name: %w", err)
	}

	if len(results) == 0 {
		return nil, nil // Account not found
	}

	return &results[0].Account, nil
}

// Update updates an existing account
func (r *DynamoDBAppRepository) Update(ctx context.Context, account *domain.Account) error {
	// Update timestamp
	account.UpdatedAt = time.Now()

	key, err := db.CreateCompositeKey("pk", fmt.Sprintf("ACCOUNT#%s", account.ID.String()), "sk", "ACCOUNT")
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	updateExpr := "SET #n = :n, #s = :s, #w = :w, #u = :u"
	exprAttrNames := map[string]string{
		"#n": "name",
		"#s": "status",
		"#w": "webhook_url",
		"#u": "updated_at",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":n": &types.AttributeValueMemberS{Value: account.Name},
		":s": &types.AttributeValueMemberS{Value: string(account.Status)},
		":w": &types.AttributeValueMemberNULL{Value: account.WebhookURL == nil},
		":u": &types.AttributeValueMemberS{Value: account.UpdatedAt.Format(time.RFC3339)},
	}

	// Handle webhook URL if present
	if account.WebhookURL != nil {
		exprAttrValues[":webhook_url"] = &types.AttributeValueMemberS{Value: *account.WebhookURL}
	}

	var updatedAccount DynamoDBAccount
	err = r.client.UpdateItem(ctx, key, updateExpr, exprAttrNames, exprAttrValues, &updatedAccount)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// Delete soft deletes an account by setting status to deleted
func (r *DynamoDBAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	key, err := db.CreateCompositeKey("pk", fmt.Sprintf("ACCOUNT#%s", id.String()), "sk", "ACCOUNT")
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	updateExpr := "SET #s = :s, #u = :u"
	exprAttrNames := map[string]string{
		"#s": "status",
		"#u": "updated_at",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":s": &types.AttributeValueMemberS{Value: string(domain.AccountStatusDeleted)},
		":u": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
	}

	err = r.client.UpdateItem(ctx, key, updateExpr, exprAttrNames, exprAttrValues, nil)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

// List retrieves accounts with pagination
func (r *DynamoDBAppRepository) List(ctx context.Context, limit, offset int) ([]*domain.Account, error) {
	// Query all accounts with pagination
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.client.GetTableName()),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "ACCOUNT"},
		},
		Limit: aws.Int32(int32(limit)),
	}

	// Handle offset by using ExclusiveStartKey if needed
	if offset > 0 {
		// In a real implementation, you would need to store and use the last evaluated key
		// For simplicity, we're not implementing offset here
	}

	var results []DynamoDBAccount
	err := r.client.QueryItems(ctx, input, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	accounts := make([]*domain.Account, len(results))
	for i, result := range results {
		accounts[i] = &result.Account
	}

	return accounts, nil
}
