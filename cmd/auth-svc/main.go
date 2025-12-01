package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/aws-payment-gateway/internal/auth/adapter/http"
	"github.com/aws-payment-gateway/internal/auth/repository"
	"github.com/aws-payment-gateway/internal/auth/usecase"
	"github.com/aws-payment-gateway/internal/common/db"
)

func main() {
	// Load configuration
	config := loadConfig()

	// Initialize database client
	dbClient, err := db.NewDynamoDBClient(context.Background(), config.AWSRegion, config.DynamoDBTable)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	appRepo := repository.NewDynamoDBAppRepository(dbClient)
	apiKeyRepo := repository.NewDynamoDBApiKeyRepository(dbClient)

	// Initialize use cases
	registerApp := usecase.NewRegisterApp(appRepo, apiKeyRepo)
	issueApiKey := usecase.NewIssueApiKey(appRepo, apiKeyRepo)
	validateApiKey := usecase.NewValidateApiKey(apiKeyRepo)
	getAPIKeys := usecase.NewGetAPIKeys(appRepo, apiKeyRepo)
	revokeApiKey := usecase.NewRevokeApiKey(apiKeyRepo)

	// Initialize handlers
	authHandler := http.NewAuthHandler(registerApp, issueApiKey, validateApiKey, getAPIKeys, revokeApiKey)
	authMiddleware := http.NewAuthMiddleware(validateApiKey, apiKeyRepo)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
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

	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,x-api-key",
	}))

	// Health check endpoint
	app.Get("/health", authHandler.HealthCheck)

	// API routes
	api := app.Group("/api/v1")
	auth := api.Group("/auth")

	// Public routes
	auth.Post("/register", authHandler.RegisterApp)
	auth.Post("/api-keys", authHandler.IssueApiKey)
	auth.Post("/validate", authHandler.ValidateApiKey)

	// Protected routes
	protected := auth.Group("/")
	protected.Use(authMiddleware.RequireAuth())

	// Account-specific routes (require authentication)
	protected.Get("/accounts/:account_id/api-keys", authMiddleware.RequirePermission("read:keys"), authHandler.GetAPIKeys)
	protected.Delete("/api-keys/:api_key_id", authMiddleware.RequirePermission("write:keys"), authHandler.RevokeApiKey)

	// Start server
	go func() {
		if err := app.Listen(":" + config.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// Config represents the application configuration
type Config struct {
	Port          string
	AWSRegion     string
	DynamoDBTable string
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	config := &Config{
		Port:          getEnv("PORT", "8080"),
		AWSRegion:     getEnv("AWS_REGION", "us-west-2"),
		DynamoDBTable: getEnv("DYNAMODB_TABLE", "auth-service"),
	}

	return config
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
