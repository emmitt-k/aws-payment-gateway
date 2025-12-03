package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws-payment-gateway/internal/common/db"
)

// DynamoDBRateLimitRepository implements RateLimitRepository using DynamoDB
type DynamoDBRateLimitRepository struct {
	client *db.DynamoDBClient
}

// NewDynamoDBRateLimitRepository creates a new DynamoDBRateLimitRepository
func NewDynamoDBRateLimitRepository(client *db.DynamoDBClient) *DynamoDBRateLimitRepository {
	return &DynamoDBRateLimitRepository{
		client: client,
	}
}

// DynamoDBRateLimit represents the rate limit entity in DynamoDB
type DynamoDBRateLimit struct {
	Key       string `dynamodbav:"key" json:"key"`
	Count     int64  `dynamodbav:"count" json:"count"`
	Window    int64  `dynamodbav:"window" json:"window"`
	ExpiresAt int64  `dynamodbav:"expires_at" json:"expires_at"`
	TTL       int64  `dynamodbav:"ttl" json:"ttl"`
}

// CheckRateLimit checks if a request exceeds the rate limit
func (r *DynamoDBRateLimitRepository) CheckRateLimit(ctx context.Context, key string, requests int, window time.Duration) (bool, int, int64, error) {
	// Calculate window expiry time
	now := time.Now()
	resetTime := now.Add(window).Unix()

	// Get current rate limit entry
	keyMap, err := db.CreateKey("key", key)
	if err != nil {
		return false, requests, resetTime, fmt.Errorf("failed to create key: %w", err)
	}

	var result DynamoDBRateLimit
	err = r.client.GetItem(ctx, keyMap, &result)
	if err != nil {
		return false, requests, resetTime, fmt.Errorf("failed to get rate limit: %w", err)
	}

	// If no entry exists, allow the request
	if result.Key == "" {
		// Create new entry
		newEntry := &DynamoDBRateLimit{
			Key:       key,
			Count:     1,
			Window:    int64(window.Seconds()),
			ExpiresAt: resetTime,
			TTL:       resetTime,
		}

		err = r.client.PutItem(ctx, newEntry)
		if err != nil {
			return false, requests, resetTime, fmt.Errorf("failed to create rate limit entry: %w", err)
		}

		return true, requests - 1, resetTime, nil
	}

	// Check if the entry is still within the current window
	if result.ExpiresAt < now.Unix() {
		// Entry expired, reset count
		updatedEntry := &DynamoDBRateLimit{
			Key:       key,
			Count:     1,
			Window:    result.Window,
			ExpiresAt: resetTime,
			TTL:       resetTime,
		}

		err = r.client.PutItem(ctx, updatedEntry)
		if err != nil {
			return false, requests, resetTime, fmt.Errorf("failed to reset rate limit entry: %w", err)
		}

		return true, requests - 1, resetTime, nil
	}

	// Check if rate limit exceeded
	if result.Count >= int64(requests) {
		remaining := 0
		if result.Count > int64(requests) {
			remaining = int(result.Count - int64(requests))
		}
		return false, remaining, resetTime, nil
	}

	// Increment count
	newCount := result.Count + 1
	updatedEntry := &DynamoDBRateLimit{
		Key:       result.Key,
		Count:     newCount,
		Window:    result.Window,
		ExpiresAt: result.ExpiresAt,
		TTL:       result.ExpiresAt,
	}

	err = r.client.PutItem(ctx, updatedEntry)
	if err != nil {
		return false, requests, resetTime, fmt.Errorf("failed to update rate limit entry: %w", err)
	}

	remaining := requests - int(newCount)
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, resetTime, nil
}

// IncrementRateLimit increments the counter for a key
func (r *DynamoDBRateLimitRepository) IncrementRateLimit(ctx context.Context, key string, window time.Duration) error {
	now := time.Now()
	resetTime := now.Add(window).Unix()

	// Get current entry
	keyMap, err := db.CreateKey("key", key)
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	var result DynamoDBRateLimit
	err = r.client.GetItem(ctx, keyMap, &result)
	if err != nil {
		return fmt.Errorf("failed to get rate limit: %w", err)
	}

	// Create or update entry
	if result.Key == "" {
		// New entry
		newEntry := &DynamoDBRateLimit{
			Key:       key,
			Count:     1,
			Window:    int64(window.Seconds()),
			ExpiresAt: resetTime,
			TTL:       resetTime,
		}

		return r.client.PutItem(ctx, newEntry)
	}

	// Update existing entry
	updatedEntry := &DynamoDBRateLimit{
		Key:       result.Key,
		Count:     result.Count + 1,
		Window:    result.Window,
		ExpiresAt: result.ExpiresAt,
		TTL:       result.ExpiresAt,
	}

	return r.client.PutItem(ctx, updatedEntry)
}

// ResetRateLimit resets the counter for a key
func (r *DynamoDBRateLimitRepository) ResetRateLimit(ctx context.Context, key string) error {
	// Delete the rate limit entry
	keyMap, err := db.CreateKey("key", key)
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	err = r.client.DeleteItem(ctx, keyMap)
	if err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	return nil
}
