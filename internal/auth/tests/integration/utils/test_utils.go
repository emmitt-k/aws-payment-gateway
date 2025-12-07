package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws-payment-gateway/internal/auth/adapter/http"
	"github.com/aws-payment-gateway/internal/auth/audit"
	"github.com/aws-payment-gateway/internal/auth/repository"
	"github.com/aws-payment-gateway/internal/auth/tests/integration/config"
	"github.com/aws-payment-gateway/internal/auth/usecase"
	"github.com/aws-payment-gateway/internal/common/db"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"
)

// TestSuite provides a base suite for integration tests
type TestSuite struct {
	suite.Suite
	Config            *config.TestConfig
	App               *fiber.App
	AuthHandler       *http.AuthHandler
	AuthMiddleware    *http.AuthMiddleware
	PostgresClient    *db.PostgreSQLClient
	DynamoClient      *db.DynamoDBClient
	AuditDynamoClient *db.DynamoDBClient
	CleanupFunc       func()
}

// SetupSuite initializes the test environment
func (s *TestSuite) SetupSuite() {
	s.Config = config.LoadTestConfig()

	// Initialize PostgreSQL
	postgresClient, err := db.NewPostgreSQLClient(context.Background(),
		s.Config.PostgreSQLHost,
		fmt.Sprintf("%d", s.Config.PostgreSQLPort),
		s.Config.PostgreSQLUser,
		s.Config.PostgreSQLPassword,
		s.Config.PostgreSQLDBName,
	)
	s.Require().NoError(err)
	s.PostgresClient = postgresClient

	// Initialize DynamoDB client for API keys
	dynamoClient, err := db.NewDynamoDBClient(context.Background(),
		s.Config.DynamoDBRegion,
		s.Config.DynamoDBTable)
	s.Require().NoError(err)
	s.DynamoClient = dynamoClient

	// Initialize DynamoDB client for audit logs
	auditDynamoClient, err := db.NewDynamoDBClient(context.Background(),
		s.Config.DynamoDBRegion,
		s.Config.AuditLogsTable)
	s.Require().NoError(err)
	s.AuditDynamoClient = auditDynamoClient

	// Initialize repositories
	appRepo := repository.NewPostgreSQLAppRepository(postgresClient)
	apiKeyRepo := repository.NewDynamoDBApiKeyRepository(dynamoClient)

	// Initialize audit logger
	auditLogger := audit.NewDynamoDBAuditLogger(auditDynamoClient)

	// Initialize use cases
	registerApp := usecase.NewRegisterApp(appRepo, apiKeyRepo)
	issueApiKey := usecase.NewIssueApiKey(appRepo, apiKeyRepo)
	validateApiKey := usecase.NewValidateApiKey(apiKeyRepo, appRepo)
	getAPIKeys := usecase.NewGetAPIKeys(appRepo, apiKeyRepo)
	revokeApiKey := usecase.NewRevokeApiKey(apiKeyRepo)

	// Initialize handlers
	s.AuthHandler = http.NewAuthHandler(registerApp, issueApiKey, validateApiKey, getAPIKeys, revokeApiKey, auditLogger)
	s.AuthMiddleware = http.NewAuthMiddleware(validateApiKey, apiKeyRepo, auditLogger)

	// Initialize Fiber app
	s.App = fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(struct {
				Error   string `json:"error"`
				Message string `json:"message"`
				Details string `json:"details,omitempty"`
			}{
				Error:   "internal_error",
				Message: "An internal error occurred",
				Details: err.Error(),
			})
		},
	})

	// Setup routes
	s.setupRoutes()

	// Setup cleanup function
	s.CleanupFunc = func() {
		s.cleanup()
	}
}

// TearDownSuite cleans up the test environment
func (s *TestSuite) TearDownSuite() {
	if s.CleanupFunc != nil {
		s.CleanupFunc()
	}
}

// SetupTest runs before each test
func (s *TestSuite) SetupTest() {
	// Clean up any test data before each test
	s.cleanupTestData()
}

// TearDownTest runs after each test
func (s *TestSuite) TearDownTest() {
	if s.Config.CleanupAfterTest {
		s.cleanupTestData()
	}
}

// setupRoutes configures the test routes
func (s *TestSuite) setupRoutes() {
	// Health check endpoint
	s.App.Get("/health", s.AuthHandler.HealthCheck)

	// API routes
	api := s.App.Group("/api/v1")
	auth := api.Group("/auth")

	// Public routes
	auth.Post("/register", s.AuthHandler.RegisterApp)
	auth.Post("/api-keys", s.AuthHandler.IssueApiKey)
	auth.Post("/validate", s.AuthHandler.ValidateApiKey)

	// Protected routes
	protected := auth.Group("/")
	protected.Use(s.AuthMiddleware.RequireAuth())

	// Account-specific routes (require authentication)
	protected.Get("/accounts/:account_id/api-keys", s.AuthMiddleware.RequirePermission("read:keys"), s.AuthHandler.GetAPIKeys)
	protected.Delete("/api-keys/:api_key_id", s.AuthMiddleware.RequirePermission("write:keys"), s.AuthHandler.RevokeApiKey)
}

// cleanup performs final cleanup
func (s *TestSuite) cleanup() {
	if s.PostgresClient != nil {
		s.PostgresClient.Close()
	}
}

// cleanupTestData cleans up test data from databases
func (s *TestSuite) cleanupTestData() {
	ctx := context.Background()

	// Clean up DynamoDB tables
	if s.DynamoClient != nil {
		s.cleanupDynamoDBTable(ctx, s.Config.DynamoDBTable)
	}

	if s.AuditDynamoClient != nil {
		s.cleanupDynamoDBTable(ctx, s.Config.AuditLogsTable)
	}

	// Clean up PostgreSQL tables
	if s.PostgresClient != nil {
		s.cleanupPostgresTables(ctx)
	}
}

// cleanupDynamoDBTable removes all items from a DynamoDB table
func (s *TestSuite) cleanupDynamoDBTable(ctx context.Context, tableName string) {
	// Scan and delete all items
	// Note: This is a simple approach for test cleanup
	// In production, you might want to use table recreation
	log.Printf("Cleaning up DynamoDB table: %s", tableName)

	// Implementation would depend on your DynamoDB client interface
	// This is a placeholder for the actual cleanup logic
}

// cleanupPostgresTables removes all data from PostgreSQL tables
func (s *TestSuite) cleanupPostgresTables(ctx context.Context) {
	// Clean up tables in reverse order of dependencies
	tables := []string{
		"api_keys",
		"accounts",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		_, err := s.PostgresClient.ExecContext(ctx, query)
		if err != nil {
			log.Printf("Failed to clean table %s: %v", table, err)
		}
	}
}

// WaitForService waits for a service to be ready
func (s *TestSuite) WaitForService(serviceName string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		// Check if service is ready
		// This is a placeholder - implement actual health checks
		time.Sleep(1 * time.Second)
	}
	return nil
}
