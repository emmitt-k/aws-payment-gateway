package fixtures

import (
	"context"
	"fmt"
	"log"

	"github.com/aws-payment-gateway/internal/auth/tests/integration/utils"
	"github.com/aws-payment-gateway/internal/common/db"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// CleanupTestData provides utilities for cleaning up test data
type CleanupTestData struct {
	postgresClient *db.PostgreSQLClient
	dynamoClient   *db.DynamoDBClient
	auditClient    *db.DynamoDBClient
}

// NewCleanupTestData creates a new cleanup utility
func NewCleanupTestData(postgresClient *db.PostgreSQLClient, dynamoClient, auditClient *db.DynamoDBClient) *CleanupTestData {
	return &CleanupTestData{
		postgresClient: postgresClient,
		dynamoClient:   dynamoClient,
		auditClient:    auditClient,
	}
}

// CleanAll cleans up all test data from all databases
func (c *CleanupTestData) CleanAll(ctx context.Context) error {
	if err := c.CleanDynamoDBTables(ctx); err != nil {
		return fmt.Errorf("failed to clean DynamoDB tables: %w", err)
	}

	if err := c.CleanPostgresTables(ctx); err != nil {
		return fmt.Errorf("failed to clean PostgreSQL tables: %w", err)
	}

	return nil
}

// CleanDynamoDBTables cleans up all DynamoDB test tables
func (c *CleanupTestData) CleanDynamoDBTables(ctx context.Context) error {
	tables := []string{
		"test-auth-service",
		"test-audit_logs",
		"test-idempotency_keys",
		"test-rate_limits",
	}

	for _, tableName := range tables {
		if err := c.CleanDynamoDBTable(ctx, tableName); err != nil {
			log.Printf("Failed to clean DynamoDB table %s: %v", tableName, err)
		}
	}

	return nil
}

// CleanDynamoDBTable removes all items from a DynamoDB table
func (c *CleanupTestData) CleanDynamoDBTable(ctx context.Context, tableName string) error {
	// Use existing DynamoDB setup utility
	dynamoSetup, err := utils.NewDynamoDBSetup("http://localhost:8001", "us-west-2")
	if err != nil {
		return fmt.Errorf("failed to create DynamoDB setup: %w", err)
	}

	return dynamoSetup.CleanupTable(ctx, tableName)
}

// CleanPostgresTables cleans up all PostgreSQL test tables
func (c *CleanupTestData) CleanPostgresTables(ctx context.Context) error {
	if c.postgresClient == nil {
		return nil
	}

	// Clean up tables in reverse order of dependencies
	tables := []string{
		"api_keys",
		"accounts",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		_, err := c.postgresClient.ExecContext(ctx, query)
		if err != nil {
			log.Printf("Failed to clean table %s: %v", table, err)
		}
	}

	return nil
}

// CleanAccountByID cleans up a specific account and its related data
func (c *CleanupTestData) CleanAccountByID(ctx context.Context, accountID string) error {
	if c.postgresClient != nil {
		// Delete API keys first (foreign key constraint)
		_, err := c.postgresClient.ExecContext(ctx, "DELETE FROM api_keys WHERE account_id = $1", accountID)
		if err != nil {
			log.Printf("Failed to delete API keys for account %s: %v", accountID, err)
		}

		// Delete the account
		_, err = c.postgresClient.ExecContext(ctx, "DELETE FROM accounts WHERE id = $1", accountID)
		if err != nil {
			log.Printf("Failed to delete account %s: %v", accountID, err)
		}
	}

	// Clean up related DynamoDB items
	if c.dynamoClient != nil {
		// Delete API keys from DynamoDB
		if err := c.DeleteAPIKeysByAccountID(ctx, accountID); err != nil {
			log.Printf("Failed to delete DynamoDB API keys for account %s: %v", accountID, err)
		}
	}

	// Clean up audit logs for the account
	if c.auditClient != nil {
		if err := c.DeleteAuditLogsByAccountID(ctx, accountID); err != nil {
			log.Printf("Failed to delete audit logs for account %s: %v", accountID, err)
		}
	}

	return nil
}

// DeleteAPIKeysByAccountID deletes all API keys for a specific account from DynamoDB
func (c *CleanupTestData) DeleteAPIKeysByAccountID(ctx context.Context, accountID string) error {
	// Create a new DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-west-2"),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{
					URL:           "http://localhost:8001",
					SigningRegion: region,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
		})),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	// Query by account_id index
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String("test-auth-service"),
		IndexName:              aws.String("account_id_index"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account_id": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	result, err := client.Query(ctx, queryInput)
	if err != nil {
		return fmt.Errorf("failed to query API keys for account %s: %w", accountID, err)
	}

	for _, item := range result.Items {
		if id, ok := item["id"]; ok {
			_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
				TableName: aws.String("test-auth-service"),
				Key: map[string]types.AttributeValue{
					"id": id,
				},
			})
			if err != nil {
				log.Printf("Failed to delete API key item: %v", err)
			}
		}
	}

	return nil
}

// DeleteAuditLogsByAccountID deletes all audit logs for a specific account
func (c *CleanupTestData) DeleteAuditLogsByAccountID(ctx context.Context, accountID string) error {
	// Create a new DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-west-2"),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{
					URL:           "http://localhost:8001",
					SigningRegion: region,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
		})),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	// Scan for audit logs containing the account ID
	scanInput := &dynamodb.ScanInput{
		TableName:        aws.String("test-audit_logs"),
		FilterExpression: aws.String("contains(partition_key, :account_id)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account_id": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	result, err := client.Scan(ctx, scanInput)
	if err != nil {
		return fmt.Errorf("failed to scan audit logs for account %s: %w", accountID, err)
	}

	for _, item := range result.Items {
		if partitionKey, ok := item["partition_key"]; ok {
			if sortKey, ok := item["sort_key"]; ok {
				_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
					TableName: aws.String("test-audit_logs"),
					Key: map[string]types.AttributeValue{
						"partition_key": partitionKey,
						"sort_key":      sortKey,
					},
				})
				if err != nil {
					log.Printf("Failed to delete audit log item: %v", err)
				}
			}
		}
	}

	return nil
}
