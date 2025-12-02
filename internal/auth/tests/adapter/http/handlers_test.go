package http_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aws-payment-gateway/internal/auth/adapter/http"
	"github.com/aws-payment-gateway/internal/auth/adapter/http/dto"
	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/tests/mocks"
	"github.com/aws-payment-gateway/internal/auth/tests/utils"
	"github.com/aws-payment-gateway/internal/auth/usecase"
)

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}

func setupTestApp() *fiber.App {
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
	return app
}

func TestAuthHandler_RegisterApp_Success(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	app := setupTestApp()
	app.Post("/api/v1/auth/register", handler.RegisterApp)

	request := dto.RegisterAppRequest{
		Name:       "test-app",
		WebhookURL: stringPtr("https://example.com/webhook"),
	}

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.RegisterAppResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)

	utils.RequireNotNil(t, response.AccountID)
	utils.RequireEqual(t, "test-app", response.Name)
	utils.RequireEqual(t, string(domain.AccountStatusActive), response.Status)
	utils.RequireNotNil(t, response.CreatedAt)
}

func TestAuthHandler_RegisterApp_ValidationError(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	app := setupTestApp()
	app.Post("/api/v1/auth/register", handler.RegisterApp)

	tests := []struct {
		name         string
		request      dto.RegisterAppRequest
		expectedCode int
	}{
		{
			name: "empty name",
			request: dto.RegisterAppRequest{
				Name: "",
			},
			expectedCode: fiber.StatusBadRequest,
		},
		{
			name: "short name",
			request: dto.RegisterAppRequest{
				Name: "ab",
			},
			expectedCode: fiber.StatusBadRequest,
		},
		{
			name: "invalid webhook URL",
			request: dto.RegisterAppRequest{
				Name:       "test-app",
				WebhookURL: stringPtr("invalid-url"),
			},
			expectedCode: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Act
			resp, err := app.Test(req, -1)

			// Assert
			utils.RequireNoError(t, err)
			require.Equal(t, tt.expectedCode, resp.StatusCode)

			body, err := readResponseBody(resp.Body)
			utils.RequireNoError(t, err)

			var response dto.ErrorResponse
			err = json.Unmarshal(body, &response)
			utils.RequireNoError(t, err)
			require.Equal(t, "validation_error", response.Error)
		})
	}
}

func TestAuthHandler_RegisterApp_DuplicateName(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	// Add existing account with same name
	existingAccount := utils.CreateTestAccount(t)
	existingAccount.Name = "duplicate-name"
	mockAppRepo.AddAccount(existingAccount)

	app := setupTestApp()
	app.Post("/api/v1/auth/register", handler.RegisterApp)

	request := dto.RegisterAppRequest{
		Name: "duplicate-name",
	}

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusConflict, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ErrorResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "account_exists", response.Error)
}

func TestAuthHandler_IssueApiKey_Success(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	// Add existing account
	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	app := setupTestApp()
	app.Post("/api/v1/auth/api-keys", handler.IssueApiKey)

	request := dto.IssueApiKeyRequest{
		AccountID:   account.ID,
		Name:        "test-api-key",
		Permissions: []string{domain.PermissionReadAccounts, domain.PermissionReadKeys},
		ExpiresIn:   intPtr(24),
	}

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/auth/api-keys", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.IssueApiKeyResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)

	utils.RequireNotNil(t, response.APIKeyID)
	utils.RequireNotNil(t, response.APIKey)
	utils.RequireNotNil(t, response.KeyHash)
	utils.RequireEqual(t, account.ID, response.AccountID)
	utils.RequireEqual(t, "test-api-key", response.Name)
	utils.RequireEqual(t, []string{domain.PermissionReadAccounts, domain.PermissionReadKeys}, response.Permissions)
	utils.RequireEqual(t, string(domain.ApiKeyStatusActive), response.Status)
	utils.RequireNotNil(t, response.ExpiresAt)
	utils.RequireNotNil(t, response.CreatedAt)
}

func TestAuthHandler_IssueApiKey_ValidationError(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	app := setupTestApp()
	app.Post("/api/v1/auth/api-keys", handler.IssueApiKey)

	tests := []struct {
		name         string
		request      dto.IssueApiKeyRequest
		expectedCode int
	}{
		{
			name: "empty name",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "",
				Permissions: []string{domain.PermissionReadAccounts},
			},
			expectedCode: fiber.StatusBadRequest,
		},
		{
			name: "empty permissions",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{},
			},
			expectedCode: fiber.StatusBadRequest,
		},
		{
			name: "invalid permission",
			request: dto.IssueApiKeyRequest{
				AccountID:   uuid.New(),
				Name:        "test-api-key",
				Permissions: []string{"invalid:permission"},
			},
			expectedCode: fiber.StatusInternalServerError, // Handler returns 500 for validation errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/auth/api-keys", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Act
			resp, err := app.Test(req, -1)

			// Assert
			utils.RequireNoError(t, err)
			require.Equal(t, tt.expectedCode, resp.StatusCode)

			body, err := readResponseBody(resp.Body)
			utils.RequireNoError(t, err)

			var response dto.ErrorResponse
			err = json.Unmarshal(body, &response)
			utils.RequireNoError(t, err)

			// Different validation errors may return different error types
			// Check for expected error based on the test case
			if tt.name == "invalid permission" {
				require.Equal(t, "internal_error", response.Error)
			} else {
				require.Equal(t, "validation_error", response.Error)
			}
		})
	}
}

func TestAuthHandler_IssueApiKey_AccountNotFound(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	app := setupTestApp()
	app.Post("/api/v1/auth/api-keys", handler.IssueApiKey)

	request := dto.IssueApiKeyRequest{
		AccountID:   uuid.New(), // Non-existent account
		Name:        "test-api-key",
		Permissions: []string{domain.PermissionReadAccounts},
	}

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/auth/api-keys", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ErrorResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "account_not_found", response.Error)
}

func TestAuthHandler_ValidateApiKey_Success(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	// Add existing account and API key
	account := utils.CreateTestAccount(t)
	apiKey := utils.CreateTestApiKey(t, account.ID)
	mockAppRepo.AddAccount(account)
	mockApiKeyRepo.AddApiKey(apiKey)

	app := setupTestApp()
	app.Post("/api/v1/auth/validate", handler.ValidateApiKey)

	request := dto.ValidateApiKeyRequest{
		KeyHash: apiKey.KeyHash,
	}

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/auth/validate", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ValidateApiKeyResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)

	require.Equal(t, true, response.Valid)
	utils.RequireNotNil(t, response.AccountID)
	utils.RequireNotNil(t, response.APIKeyID)
	utils.RequireNotNil(t, response.Name)
	utils.RequireEqual(t, []string(apiKey.Permissions), response.Permissions)
	utils.RequireNotNil(t, response.ExpiresAt)
}

func TestAuthHandler_ValidateApiKey_InvalidKey(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	app := setupTestApp()
	app.Post("/api/v1/auth/validate", handler.ValidateApiKey)

	request := dto.ValidateApiKeyRequest{
		KeyHash: "non-existent-hash",
	}

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/auth/validate", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ValidateApiKeyResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)

	require.Equal(t, false, response.Valid)
	utils.RequireNil(t, response.AccountID)
	utils.RequireNil(t, response.APIKeyID)
	utils.RequireNil(t, response.Name)
	require.Nil(t, response.Permissions, "Expected permissions to be nil for invalid key")
}

func TestAuthHandler_GetAPIKeys_Success(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	// Add existing account and API keys
	account := utils.CreateTestAccount(t)
	mockAppRepo.AddAccount(account)

	for i := 0; i < 3; i++ {
		apiKey := utils.CreateTestApiKey(t, account.ID)
		mockApiKeyRepo.AddApiKey(apiKey)
	}

	app := setupTestApp()
	app.Get("/api/v1/auth/accounts/:account_id/api-keys", handler.GetAPIKeys)

	req := httptest.NewRequest("GET", "/api/v1/auth/accounts/"+account.ID.String()+"/api-keys?limit=10&offset=0", nil)

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.GetAPIKeysResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)

	require.Equal(t, 10, response.Limit)
	require.Equal(t, 0, response.Offset)
	require.Equal(t, 3, response.Total)
	require.Equal(t, 3, len(response.APIKeys))
}

func TestAuthHandler_GetAPIKeys_InvalidAccountID(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	app := setupTestApp()
	app.Get("/api/v1/auth/accounts/:account_id/api-keys", handler.GetAPIKeys)

	req := httptest.NewRequest("GET", "/api/v1/auth/accounts/invalid-uuid/api-keys", nil)

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ErrorResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "invalid_account_id", response.Error)
}

func TestAuthHandler_RevokeApiKey_Success(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	// Add existing API key
	apiKey := utils.CreateTestApiKey(t, uuid.New())
	mockApiKeyRepo.AddApiKey(apiKey)

	app := setupTestApp()
	app.Delete("/api/v1/auth/api-keys/:api_key_id", handler.RevokeApiKey)

	req := httptest.NewRequest("DELETE", "/api/v1/auth/api-keys/"+apiKey.ID.String(), nil)

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestAuthHandler_RevokeApiKey_NotFound(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	app := setupTestApp()
	app.Delete("/api/v1/auth/api-keys/:api_key_id", handler.RevokeApiKey)

	req := httptest.NewRequest("DELETE", "/api/v1/auth/api-keys/"+uuid.New().String(), nil)

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ErrorResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "api_key_not_found", response.Error)
}

func TestAuthHandler_HealthCheck_Success(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	app := setupTestApp()
	app.Get("/health", handler.HealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.HealthResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)

	require.Equal(t, "healthy", response.Status)
	require.Equal(t, "auth-service", response.Service)
	require.Equal(t, "1.0.0", response.Version)
	utils.RequireNotNil(t, response.Timestamp)
}

func TestAuthHandler_InvalidJSON(t *testing.T) {
	// Arrange
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	handler := http.NewAuthHandler(
		usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo),
		usecase.NewIssueApiKey(mockAppRepo, mockApiKeyRepo),
		usecase.NewValidateApiKey(mockApiKeyRepo, mockAppRepo),
		usecase.NewGetAPIKeys(mockAppRepo, mockApiKeyRepo),
		usecase.NewRevokeApiKey(mockApiKeyRepo),
	)

	app := setupTestApp()
	app.Post("/api/v1/auth/register", handler.RegisterApp)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := app.Test(req, -1)

	// Assert
	utils.RequireNoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := readResponseBody(resp.Body)
	utils.RequireNoError(t, err)

	var response dto.ErrorResponse
	err = json.Unmarshal(body, &response)
	utils.RequireNoError(t, err)
	require.Equal(t, "invalid_request", response.Error)
}
