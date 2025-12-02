package http_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"github.com/aws-payment-gateway/internal/auth/adapter/http"
	"github.com/aws-payment-gateway/internal/auth/adapter/http/dto"
	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/tests/mocks"
	"github.com/aws-payment-gateway/internal/auth/tests/utils"
	"github.com/aws-payment-gateway/internal/auth/usecase"
)

func setupTestAppWithMiddleware(middleware fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			c.Status(code).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "Internal server error",
				"details": err.Error(),
			})
			return nil
		},
	})

	// Add middleware and a test endpoint
	app.Use(middleware)
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	return app
}

// readResponseBody reads response body and returns it as a byte slice
func readResponseBody(body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	return io.ReadAll(body)
}

func TestAuthMiddleware_RequireAuth_MissingAPIKey(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	validateApiKey := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)
	middleware := http.NewAuthMiddleware(validateApiKey, mockApiKeyRepo)

	app := setupTestAppWithMiddleware(middleware.RequireAuth())

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ErrorResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "missing_api_key", response.Error)
}

func TestAuthMiddleware_RequireAuth_InvalidAPIKey(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	validateApiKey := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)
	middleware := http.NewAuthMiddleware(validateApiKey, mockApiKeyRepo)

	app := setupTestAppWithMiddleware(middleware.RequireAuth())

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-api-key", "invalid-api-key")
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ErrorResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "invalid_api_key", response.Error)
}

func TestAuthMiddleware_RequireAuth_ValidAPIKey(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	validateApiKey := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)
	middleware := http.NewAuthMiddleware(validateApiKey, mockApiKeyRepo)

	// Add valid API key to mock
	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateTestApiKey(t, account.ID)
	mockApiKeyRepo.AddApiKey(apiKey)

	app := setupTestAppWithMiddleware(middleware.RequireAuth())

	// Act - Use key hash directly for validation
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-api-key", apiKey.KeyHash) // Use the hash directly
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "success", response["message"])
}

func TestAuthMiddleware_RequireAuth_BearerToken(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	validateApiKey := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)
	middleware := http.NewAuthMiddleware(validateApiKey, mockApiKeyRepo)

	// Add valid API key to mock
	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateTestApiKey(t, account.ID)
	mockApiKeyRepo.AddApiKey(apiKey)

	app := setupTestAppWithMiddleware(middleware.RequireAuth())

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey.KeyHash)
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "success", response["message"])
}

func TestAuthMiddleware_RequirePermission_MissingAuthentication(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	validateApiKey := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)
	middleware := http.NewAuthMiddleware(validateApiKey, mockApiKeyRepo)

	app := setupTestAppWithMiddleware(middleware.RequirePermission(domain.PermissionReadAccounts))

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ErrorResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "not_authenticated", response.Error)
}

func TestAuthMiddleware_RequirePermission_InsufficientPermissions(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	validateApiKey := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)
	middleware := http.NewAuthMiddleware(validateApiKey, mockApiKeyRepo)

	// Add API key with limited permissions
	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateTestApiKey(t, account.ID)
	apiKey.Permissions = domain.ApiKeyPermissions{domain.PermissionReadAccounts} // Only read permission
	mockApiKeyRepo.AddApiKey(apiKey)

	// Chain RequireAuth and RequirePermission
	app := setupTestAppWithMiddleware(func(c *fiber.Ctx) error {
		if err := middleware.RequireAuth()(c); err != nil {
			return err
		}
		return middleware.RequirePermission(domain.PermissionWriteAccounts)(c)
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-api-key", apiKey.KeyHash)
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ErrorResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "insufficient_permissions", response.Error)
}

func TestAuthMiddleware_RequirePermission_Success(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	validateApiKey := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)
	middleware := http.NewAuthMiddleware(validateApiKey, mockApiKeyRepo)

	// Add API key with required permission
	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateTestApiKey(t, account.ID)
	apiKey.Permissions = domain.ApiKeyPermissions{domain.PermissionReadAccounts, domain.PermissionWriteAccounts}
	mockApiKeyRepo.AddApiKey(apiKey)

	// Chain RequireAuth and RequirePermission
	app := setupTestApp()
	app.Use(middleware.RequireAuth())
	app.Use(middleware.RequirePermission(domain.PermissionWriteAccounts))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-api-key", apiKey.KeyHash)
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "success", response["message"])
}

func TestAuthMiddleware_RequireAnyPermission_Success(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	validateApiKey := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)
	middleware := http.NewAuthMiddleware(validateApiKey, mockApiKeyRepo)

	// Add API key with one of the required permissions
	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateTestApiKey(t, account.ID)
	apiKey.Permissions = domain.ApiKeyPermissions{domain.PermissionReadAccounts} // Only read permission
	mockApiKeyRepo.AddApiKey(apiKey)

	// Chain RequireAuth and RequireAnyPermission
	app := setupTestApp()
	app.Use(middleware.RequireAuth())
	app.Use(middleware.RequireAnyPermission(domain.PermissionWriteAccounts, domain.PermissionReadAccounts))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-api-key", apiKey.KeyHash)
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "success", response["message"])
}

func TestAuthMiddleware_RequireAnyPermission_InsufficientPermissions(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	validateApiKey := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)
	middleware := http.NewAuthMiddleware(validateApiKey, mockApiKeyRepo)

	// Add API key with none of the required permissions
	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateTestApiKey(t, account.ID)
	apiKey.Permissions = domain.ApiKeyPermissions{domain.PermissionReadKeys} // Different permission
	mockApiKeyRepo.AddApiKey(apiKey)

	// Chain RequireAuth and RequireAnyPermission
	app := setupTestApp()
	app.Use(middleware.RequireAuth())
	app.Use(middleware.RequireAnyPermission(domain.PermissionWriteAccounts, domain.PermissionReadAccounts))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-api-key", apiKey.KeyHash)
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ErrorResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "insufficient_permissions", response.Error)
}

func TestAuthMiddleware_ContextHelpers(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	validateApiKey := usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo)
	middleware := http.NewAuthMiddleware(validateApiKey, mockApiKeyRepo)

	// Add valid API key to mock
	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateTestApiKey(t, account.ID)
	mockApiKeyRepo.AddApiKey(apiKey)

	app := fiber.New()
	app.Use(middleware.RequireAuth())

	// Test endpoint that uses context helpers
	app.Get("/test", func(c *fiber.Ctx) error {
		// Test GetAccountID
		accountID, err := http.GetAccountID(c)
		utils.RequireNoError(t, err)
		utils.RequireEqual(t, account.ID, accountID)

		// Test GetAPIKeyID
		apiKeyID, err := http.GetAPIKeyID(c)
		utils.RequireNoError(t, err)
		utils.RequireEqual(t, apiKey.ID, apiKeyID)

		// Test GetAPIKeyName
		apiKeyName, err := http.GetAPIKeyName(c)
		utils.RequireNoError(t, err)
		utils.RequireEqual(t, apiKey.Name, apiKeyName)

		// Test GetPermissions
		permissions, err := http.GetPermissions(c)
		utils.RequireNoError(t, err)
		utils.RequireEqual(t, []string(apiKey.Permissions), permissions)

		// Test HasPermission
		hasReadAccounts := http.HasPermission(c, domain.PermissionReadAccounts)
		utils.RequireEqual(t, true, hasReadAccounts)

		hasWriteAccounts := http.HasPermission(c, domain.PermissionWriteAccounts)
		utils.RequireEqual(t, false, hasWriteAccounts)

		return c.JSON(fiber.Map{"success": true})
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-api-key", apiKey.KeyHash)
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, true, response["success"])
}

func TestAuthMiddleware_ContextHelpers_MissingContext(t *testing.T) {
	// Arrange
	app := fiber.New()

	// Test endpoint that tries to use context helpers without authentication
	app.Get("/test", func(c *fiber.Ctx) error {
		// Test GetAccountID without context
		_, err := http.GetAccountID(c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "account_id not found in context")

		// Test GetAPIKeyID without context
		_, err = http.GetAPIKeyID(c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "api_key_id not found in context")

		// Test GetAPIKeyName without context
		_, err = http.GetAPIKeyName(c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "api_key_name not found in context")

		// Test GetPermissions without context
		_, err = http.GetPermissions(c)
		require.Error(t, err)
		require.Contains(t, err.Error(), "permissions not found in context")

		// Test HasPermission without context
		hasPermission := http.HasPermission(c, domain.PermissionReadAccounts)
		require.Equal(t, false, hasPermission)

		return c.JSON(fiber.Map{"success": true})
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, true, response["success"])
}
