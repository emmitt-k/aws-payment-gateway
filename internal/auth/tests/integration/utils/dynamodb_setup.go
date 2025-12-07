package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBSetup handles DynamoDB table creation for tests
type DynamoDBSetup struct {
	client *dynamodb.Client
}

// NewDynamoDBSetup creates a new DynamoDB setup instance
func NewDynamoDBSetup(endpoint, region string) (*DynamoDBSetup, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{
					URL:           endpoint,
					SigningRegion: region,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	return &DynamoDBSetup{client: client}, nil
}

// SetupTables creates the required DynamoDB tables for testing
func (d *DynamoDBSetup) SetupTables() error {
	ctx := context.Background()

	// Create API keys table
	if err := d.createAPIKeysTable(ctx); err != nil {
		return fmt.Errorf("failed to create API keys table: %w", err)
	}

	// Create audit logs table
	if err := d.createAuditLogsTable(ctx); err != nil {
		return fmt.Errorf("failed to create audit logs table: %w", err)
	}

	// Create idempotency keys table
	if err := d.createIdempotencyKeysTable(ctx); err != nil {
		return fmt.Errorf("failed to create idempotency keys table: %w", err)
	}

	// Create rate limit table
	if err := d.createRateLimitTable(ctx); err != nil {
		return fmt.Errorf("failed to create rate limit table: %w", err)
	}

	return nil
}

// createAPIKeysTable creates the API keys DynamoDB table
func (d *DynamoDBSetup) createAPIKeysTable(ctx context.Context) error {
	tableName := "test-auth-service"

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("account_id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("account_id_index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("account_id"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := d.client.CreateTable(ctx, input)
	if err != nil {
		// Check if table already exists
		if isResourceExistsError(err) {
			log.Printf("Table %s already exists", tableName)
			return nil
		}
		return err
	}

	// Wait for table to be created
	return d.waitForTableActive(ctx, tableName)
}

// createAuditLogsTable creates the audit logs DynamoDB table
func (d *DynamoDBSetup) createAuditLogsTable(ctx context.Context) error {
	tableName := "test-audit_logs"

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("partition_key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("sort_key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("partition_key"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("sort_key"),
				KeyType:       types.KeyTypeRange,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := d.client.CreateTable(ctx, input)
	if err != nil {
		if isResourceExistsError(err) {
			log.Printf("Table %s already exists", tableName)
			return nil
		}
		return err
	}

	return d.waitForTableActive(ctx, tableName)
}

// createIdempotencyKeysTable creates the idempotency keys DynamoDB table
func (d *DynamoDBSetup) createIdempotencyKeysTable(ctx context.Context) error {
	tableName := "test-idempotency_keys"

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("idempotency_key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("account_id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("idempotency_key"),
				KeyType:       types.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("account_id_index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("account_id"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := d.client.CreateTable(ctx, input)
	if err != nil {
		if isResourceExistsError(err) {
			log.Printf("Table %s already exists", tableName)
			return nil
		}
		return err
	}

	return d.waitForTableActive(ctx, tableName)
}

// createRateLimitTable creates the rate limit DynamoDB table
func (d *DynamoDBSetup) createRateLimitTable(ctx context.Context) error {
	tableName := "test-rate_limits"

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("key"),
				KeyType:       types.KeyTypeHash,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := d.client.CreateTable(ctx, input)
	if err != nil {
		if isResourceExistsError(err) {
			log.Printf("Table %s already exists", tableName)
			return nil
		}
		return err
	}

	return d.waitForTableActive(ctx, tableName)
}

// waitForTableActive waits for a DynamoDB table to become active
func (d *DynamoDBSetup) waitForTableActive(ctx context.Context, tableName string) error {
	for i := 0; i < 30; i++ {
		resp, err := d.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return err
		}

		if resp.Table.TableStatus == types.TableStatusActive {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for table %s to become active", tableName)
}

// isResourceExistsError checks if the error indicates a resource already exists
func isResourceExistsError(err error) bool {
	var resourceInUseException *types.ResourceInUseException
	var resourceNotFoundException *types.ResourceNotFoundException

	return resourceInUseException != nil || resourceNotFoundException != nil
}

// CleanupTable removes all items from a DynamoDB table
func (d *DynamoDBSetup) CleanupTable(ctx context.Context, tableName string) error {
	// Scan and delete all items
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := d.client.Scan(ctx, scanInput)
	if err != nil {
		return err
	}

	for _, item := range result.Items {
		// Extract key attributes based on table schema
		key := make(map[string]types.AttributeValue)

		if id, ok := item["id"]; ok {
			key["id"] = id
		} else if idempotencyKey, ok := item["idempotency_key"]; ok {
			key["idempotency_key"] = idempotencyKey
		} else if partitionKey, ok := item["partition_key"]; ok {
			key["partition_key"] = partitionKey
			if sortKey, ok := item["sort_key"]; ok {
				key["sort_key"] = sortKey
			}
		} else if rateLimitKey, ok := item["key"]; ok {
			key["key"] = rateLimitKey
		}

		if len(key) > 0 {
			_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
				TableName: aws.String(tableName),
				Key:       key,
			})
			if err != nil {
				log.Printf("Failed to delete item from table %s: %v", tableName, err)
			}
		}
	}

	return nil
}
