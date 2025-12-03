package http

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws-payment-gateway/internal/auth/adapter/http/dto"
	"github.com/aws-payment-gateway/internal/auth/audit"
	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	registerApp    *usecase.RegisterApp
	issueApiKey    *usecase.IssueApiKey
	validateApiKey *usecase.ValidateApiKey
	getAPIKeys     *usecase.GetAPIKeys
	revokeApiKey   *usecase.RevokeApiKey
	auditLogger    audit.AuditLoggerInterface
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	registerApp *usecase.RegisterApp,
	issueApiKey *usecase.IssueApiKey,
	validateApiKey *usecase.ValidateApiKey,
	getAPIKeys *usecase.GetAPIKeys,
	revokeApiKey *usecase.RevokeApiKey,
	auditLogger audit.AuditLoggerInterface,
) *AuthHandler {
	return &AuthHandler{
		registerApp:    registerApp,
		issueApiKey:    issueApiKey,
		validateApiKey: validateApiKey,
		getAPIKeys:     getAPIKeys,
		revokeApiKey:   revokeApiKey,
		auditLogger:    auditLogger,
	}
}

// RegisterApp handles account registration
// @Summary Register a new application
// @Description Register a new application account in the system
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterAppRequest true "Registration request"
// @Success 201 {object} dto.RegisterAppResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) RegisterApp(c *fiber.Ctx) error {
	ctx := context.Background()

	var req dto.RegisterAppRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
			Details: err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: err.Error(),
		})
	}

	// Convert to use case input
	input := usecase.RegisterAppInput{
		Name:       req.Name,
		WebhookURL: req.WebhookURL,
	}

	// Execute use case
	output, err := h.registerApp.Execute(ctx, input)
	if err != nil {
		// Log failed account creation attempt
		h.auditLogger.LogAccountCreation(
			ctx,
			nil,
			&req.Name,
			c.IP(), c.Get("User-Agent"),
			map[string]string{
				"error":   err.Error(),
				"name":    req.Name,
				"success": "false",
			},
		)

		if err.Error() == fmt.Sprintf("app with name '%s' already exists", req.Name) {
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "account_exists",
				Message: "Account with this name already exists",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to register account",
			Details: err.Error(),
		})
	}

	// Log successful account creation
	h.auditLogger.LogAccountCreation(
		ctx,
		&output.AccountID,
		&output.Name,
		c.IP(), c.Get("User-Agent"),
		map[string]string{"success": "true"},
	)

	// Convert to response
	response := dto.RegisterAppResponse{
		AccountID: output.AccountID,
		Name:      output.Name,
		Status:    output.Status,
		CreatedAt: output.CreatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// IssueApiKey handles API key issuance
// @Summary Issue a new API key
// @Description Issue a new API key for an existing account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.IssueApiKeyRequest true "API key issuance request"
// @Success 201 {object} dto.IssueApiKeyResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/api-keys [post]
func (h *AuthHandler) IssueApiKey(c *fiber.Ctx) error {
	ctx := context.Background()

	var req dto.IssueApiKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
			Details: err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: err.Error(),
		})
	}

	// Convert to use case input
	input := usecase.IssueApiKeyInput{
		AccountID:   req.AccountID,
		Name:        req.Name,
		Permissions: domain.ApiKeyPermissions(req.Permissions),
		ExpiresIn:   req.ExpiresIn,
	}

	// Execute use case
	output, err := h.issueApiKey.Execute(ctx, input)
	if err != nil {
		// Log failed API key creation attempt
		h.auditLogger.LogAPIKeyCreation(
			ctx,
			&req.AccountID,
			nil,
			&req.Name,
			c.IP(), c.Get("User-Agent"),
			map[string]string{
				"error":   err.Error(),
				"success": "false",
			},
		)

		if err.Error() == "account not found or inactive" {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "account_not_found",
				Message: "Account not found or inactive",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to issue API key",
			Details: err.Error(),
		})
	}

	// Log successful API key creation
	h.auditLogger.LogAPIKeyCreation(
		ctx,
		&output.AccountID,
		&output.APIKeyID,
		&output.Name,
		c.IP(), c.Get("User-Agent"),
		map[string]string{"success": "true"},
	)

	// Convert to response
	response := dto.IssueApiKeyResponse{
		APIKeyID:    output.APIKeyID,
		KeyHash:     output.KeyHash,
		AccountID:   output.AccountID,
		Name:        output.Name,
		Permissions: []string(output.Permissions),
		Status:      output.Status,
		ExpiresAt:   output.ExpiresAt,
		CreatedAt:   output.CreatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// ValidateApiKey handles API key validation
// @Summary Validate an API key
// @Description Validate an API key and return account information
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ValidateApiKeyRequest true "API key validation request"
// @Success 200 {object} dto.ValidateApiKeyResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/validate [post]
func (h *AuthHandler) ValidateApiKey(c *fiber.Ctx) error {
	ctx := context.Background()

	var req dto.ValidateApiKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
			Details: err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: err.Error(),
		})
	}

	// Convert to use case input
	input := usecase.ValidateApiKeyInput{
		KeyHash: req.KeyHash,
	}

	// Execute use case
	output, err := h.validateApiKey.Execute(ctx, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to validate API key",
			Details: err.Error(),
		})
	}

	// Convert to response
	response := dto.ValidateApiKeyResponse{
		Valid:       output.Valid,
		AccountID:   output.AccountID,
		APIKeyID:    output.APIKeyID,
		Name:        output.Name,
		Permissions: []string(output.Permissions),
		LastUsedAt:  output.LastUsedAt,
		ExpiresAt:   output.ExpiresAt,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetAPIKeys handles getting API keys for an account
// @Summary Get API keys for an account
// @Description Retrieve all API keys for a specific account with pagination
// @Tags auth
// @Produce json
// @Param account_id path string true "Account ID"
// @Param limit query int false "Limit number of results" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} dto.GetAPIKeysResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/accounts/{account_id}/api-keys [get]
func (h *AuthHandler) GetAPIKeys(c *fiber.Ctx) error {
	ctx := context.Background()

	// Parse account ID
	accountIDStr := c.Params("account_id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_account_id",
			Message: "Invalid account ID format",
		})
	}

	// Parse pagination parameters
	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10 // Default limit
	}

	offsetStr := c.Query("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 // Default offset
	}

	// Convert to use case input
	input := usecase.GetAPIKeysInput{
		AccountID: accountID,
		Limit:     limit,
		Offset:    offset,
	}

	// Execute use case
	output, err := h.getAPIKeys.Execute(ctx, input)
	if err != nil {
		if err.Error() == "account not found or inactive" {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "account_not_found",
				Message: "Account not found or inactive",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get API keys",
			Details: err.Error(),
		})
	}

	// Convert API keys to response format
	apiKeys := make([]dto.ApiKeyResponse, len(output.APIKeys))
	for i, apiKey := range output.APIKeys {
		apiKeys[i] = dto.ApiKeyResponse{
			APIKeyID:    apiKey.ID,
			Name:        apiKey.Name,
			Permissions: []string(apiKey.Permissions),
			Status:      string(apiKey.Status),
			LastUsedAt:  apiKey.LastUsedAt,
			ExpiresAt:   apiKey.ExpiresAt,
			CreatedAt:   apiKey.CreatedAt,
		}
	}

	// Create response
	response := dto.GetAPIKeysResponse{
		APIKeys: apiKeys,
		Limit:   output.Limit,
		Offset:  output.Offset,
		Total:   output.Total,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// RevokeApiKey handles API key revocation
// @Summary Revoke an API key
// @Description Revoke (delete) an API key
// @Tags auth
// @Param api_key_id path string true "API Key ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/api-keys/{api_key_id} [delete]
func (h *AuthHandler) RevokeApiKey(c *fiber.Ctx) error {
	ctx := context.Background()

	// Parse API key ID
	apiKeyIDStr := c.Params("api_key_id")
	apiKeyID, err := uuid.Parse(apiKeyIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_api_key_id",
			Message: "Invalid API key ID format",
		})
	}

	// Get account ID from context for audit logging
	accountID, err := GetAccountID(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get account context",
			Details: err.Error(),
		})
	}

	// Get API key name from context for audit logging
	apiKeyName, err := GetAPIKeyName(c)
	if err != nil {
		// Continue even if we can't get the name
		apiKeyName = ""
	}

	// Convert to use case input
	input := usecase.RevokeApiKeyInput{
		APIKeyID: apiKeyID,
	}

	// Execute use case
	_, err = h.revokeApiKey.Execute(ctx, input)
	if err != nil {
		// Log failed API key revocation attempt
		h.auditLogger.LogAPIKeyRevocation(
			ctx,
			&accountID,
			&apiKeyID,
			&apiKeyName,
			c.IP(), c.Get("User-Agent"),
			map[string]string{
				"error":   err.Error(),
				"success": "false",
			},
		)

		if err.Error() == "API key not found" {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "api_key_not_found",
				Message: "API key not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to revoke API key",
			Details: err.Error(),
		})
	}

	// Log successful API key revocation
	h.auditLogger.LogAPIKeyRevocation(
		ctx,
		&accountID,
		&apiKeyID,
		&apiKeyName,
		c.IP(), c.Get("User-Agent"),
		map[string]string{"success": "true"},
	)

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// HealthCheck handles health check requests
// @Summary Health check
// @Description Check if the auth service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Router /health [get]
func (h *AuthHandler) HealthCheck(c *fiber.Ctx) error {
	response := dto.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "auth-service",
		Version:   "1.0.0",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
