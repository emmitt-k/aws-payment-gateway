package db

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// DynamoDBClient wraps the AWS DynamoDB client
type DynamoDBClient struct {
	client *dynamodb.Client
	table  string
}

// GetTableName returns the table name
func (d *DynamoDBClient) GetTableName() string {
	return d.table
}

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClient(ctx context.Context, region, table string) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	// Test connection
	_, err = client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(table),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe table %s: %w", table, err)
	}

	return &DynamoDBClient{
		client: client,
		table:  table,
	}, nil
}

// PutItem puts an item into DynamoDB
func (d *DynamoDBClient) PutItem(ctx context.Context, item interface{}) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.table),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

// GetItem gets an item from DynamoDB by key
func (d *DynamoDBClient) GetItem(ctx context.Context, key map[string]types.AttributeValue, result interface{}) error {
	resp, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.table),
		Key:       key,
	})
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	if len(resp.Item) == 0 {
		return nil // Item not found
	}

	err = attributevalue.UnmarshalMap(resp.Item, result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return nil
}

// UpdateItem updates an item in DynamoDB
func (d *DynamoDBClient) UpdateItem(ctx context.Context, key map[string]types.AttributeValue, updateExpr string, exprAttrNames map[string]string, exprAttrValues map[string]types.AttributeValue, result interface{}) error {
	input := &dynamodb.UpdateItemInput{
		TableName:        aws.String(d.table),
		Key:              key,
		UpdateExpression: aws.String(updateExpr),
		ReturnValues:     types.ReturnValueUpdatedNew,
	}

	if exprAttrNames != nil {
		input.ExpressionAttributeNames = exprAttrNames
	}

	if exprAttrValues != nil {
		input.ExpressionAttributeValues = exprAttrValues
	}

	resp, err := d.client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	if result != nil {
		err = attributevalue.UnmarshalMap(resp.Attributes, result)
		if err != nil {
			return fmt.Errorf("failed to unmarshal updated item: %w", err)
		}
	}

	return nil
}

// QueryItems queries items from DynamoDB
func (d *DynamoDBClient) QueryItems(ctx context.Context, input *dynamodb.QueryInput, results interface{}) error {
	resp, err := d.client.Query(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to query items: %w", err)
	}

	err = attributevalue.UnmarshalListOfMaps(resp.Items, results)
	if err != nil {
		return fmt.Errorf("failed to unmarshal query results: %w", err)
	}

	return nil
}

// ScanItems scans items from DynamoDB
func (d *DynamoDBClient) ScanItems(ctx context.Context, input *dynamodb.ScanInput, results interface{}) error {
	resp, err := d.client.Scan(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to scan items: %w", err)
	}

	err = attributevalue.UnmarshalListOfMaps(resp.Items, results)
	if err != nil {
		return fmt.Errorf("failed to unmarshal scan results: %w", err)
	}

	return nil
}

// DeleteItem deletes an item from DynamoDB
func (d *DynamoDBClient) DeleteItem(ctx context.Context, key map[string]types.AttributeValue) error {
	_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(d.table),
		Key:       key,
	})
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// CreateKey creates a key map for DynamoDB operations
func CreateKey(name string, value interface{}) (map[string]types.AttributeValue, error) {
	av, err := attributevalue.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key value: %w", err)
	}

	return map[string]types.AttributeValue{
		name: av,
	}, nil
}

// CreateCompositeKey creates a composite key map for DynamoDB operations
func CreateCompositeKey(partitionKey, partitionValue, sortKey, sortValue string) (map[string]types.AttributeValue, error) {
	pkAv, err := attributevalue.Marshal(partitionValue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal partition key: %w", err)
	}

	skAv, err := attributevalue.Marshal(sortValue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sort key: %w", err)
	}

	return map[string]types.AttributeValue{
		partitionKey: pkAv,
		sortKey:      skAv,
	}, nil
}

// BaseModel represents a base model for DynamoDB entities
type BaseModel struct {
	ID        uuid.UUID `dynamodbav:"id" json:"id"`
	CreatedAt time.Time `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt time.Time `dynamodbav:"updated_at" json:"updated_at"`
}

// BeforeCreate sets the ID and timestamps before creating
func (b *BaseModel) BeforeCreate() {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
}

// BeforeUpdate updates the timestamp before updating
func (b *BaseModel) BeforeUpdate() {
	b.UpdatedAt = time.Now()
}
