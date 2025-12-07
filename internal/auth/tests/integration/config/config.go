package config

import (
	"fmt"
	"os"
	"strconv"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	// PostgreSQL configuration
	PostgreSQLHost     string
	PostgreSQLPort     int
	PostgreSQLUser     string
	PostgreSQLPassword string
	PostgreSQLDBName   string

	// DynamoDB configuration
	DynamoDBEndpoint string
	DynamoDBRegion   string
	DynamoDBTable    string
	AuditLogsTable   string

	// Test configuration
	TestTimeout      int
	CleanupAfterTest bool
}

// LoadTestConfig loads test configuration from environment variables
func LoadTestConfig() *TestConfig {
	config := &TestConfig{
		// PostgreSQL configuration with test defaults
		PostgreSQLHost:     getEnv("TEST_POSTGRES_HOST", "localhost"),
		PostgreSQLPort:     getEnvInt("TEST_POSTGRES_PORT", 5432),
		PostgreSQLUser:     getEnv("TEST_POSTGRES_USER", "test_user"),
		PostgreSQLPassword: getEnv("TEST_POSTGRES_PASSWORD", "test_password"),
		PostgreSQLDBName:   getEnv("TEST_POSTGRES_DB", "test_payment_gateway"),

		// DynamoDB configuration with test defaults
		DynamoDBEndpoint: getEnv("TEST_DYNAMODB_ENDPOINT", "http://localhost:8000"),
		DynamoDBRegion:   getEnv("TEST_DYNAMODB_REGION", "us-west-2"),
		DynamoDBTable:    getEnv("TEST_DYNAMODB_TABLE", "test-auth-service"),
		AuditLogsTable:   getEnv("TEST_AUDIT_LOGS_TABLE", "test-audit_logs"),

		// Test configuration
		TestTimeout:      getEnvInt("TEST_TIMEOUT", 30),
		CleanupAfterTest: getEnvBool("TEST_CLEANUP", true),
	}

	return config
}

// GetPostgresConnectionString returns the PostgreSQL connection string
func (c *TestConfig) GetPostgresConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.PostgreSQLHost, c.PostgreSQLPort, c.PostgreSQLUser, c.PostgreSQLPassword, c.PostgreSQLDBName)
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as integer with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool gets environment variable as boolean with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
