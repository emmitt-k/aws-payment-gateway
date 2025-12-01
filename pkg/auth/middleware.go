package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/aws-payment-gateway/internal/auth/adapter/http/dto"
	"github.com/aws-payment-gateway/internal/auth/repository"
)

// SharedAuthMiddleware provides authentication middleware for API key validation that can be shared across services
type SharedAuthMiddleware struct {
	apiKeyRepo repository.ApiKeyRepository
}

// NewSharedAuthMiddleware creates a new SharedAuthMiddleware
func NewSharedAuthMiddleware(apiKeyRepo repository.ApiKeyRepository) *SharedAuthMiddleware {
	return &SharedAuthMiddleware{
		apiKeyRepo: apiKeyRepo,
	}
}

// RequireAuth creates a middleware that requires valid API key
func (m *SharedAuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get API key from header
		apiKey := c.Get("x-api-key")
		if apiKey == "" {
			apiKey = c.Get("Authorization")
			if apiKey != "" {
				// Remove "Bearer " prefix if present
				if strings.HasPrefix(apiKey, "Bearer ") {
					apiKey = strings.TrimPrefix(apiKey, "Bearer ")
				}
			}
		}

		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "missing_api_key",
				Message: "API key is required",
			})
		}

		// Validate API key using repository method
		ctx := context.Background()
		validatedKey, err := m.apiKeyRepo.ValidateByKey(ctx, apiKey)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "validation_error",
				Message: "Failed to validate API key",
				Details: err.Error(),
			})
		}

		if validatedKey == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "invalid_api_key",
				Message: "API key is invalid or expired",
			})
		}

		// Store account context
		c.Locals("account_id", validatedKey.AccountID)
		c.Locals("api_key_id", validatedKey.ID)
		c.Locals("api_key_name", validatedKey.Name)
		c.Locals("permissions", validatedKey.Permissions)

		// Continue to next handler
		return c.Next()
	}
}

// RequirePermission creates a middleware that requires specific permission
func (m *SharedAuthMiddleware) RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get permissions from context (set by RequireAuth)
		permissions := c.Locals("permissions")
		if permissions == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "not_authenticated",
				Message: "Authentication required",
			})
		}

		// Check if user has required permission
		userPermissions, ok := permissions.([]string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "internal_error",
				Message: "Invalid permissions format",
			})
		}

		for _, p := range userPermissions {
			if p == permission {
				// User has required permission, continue
				return c.Next()
			}
		}

		// User doesn't have required permission
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
			Error:   "insufficient_permissions",
			Message: fmt.Sprintf("Permission '%s' is required", permission),
		})
	}
}

// RequireAnyPermission creates a middleware that requires any of the specified permissions
func (m *SharedAuthMiddleware) RequireAnyPermission(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get permissions from context (set by RequireAuth)
		userPermissions := c.Locals("permissions")
		if userPermissions == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "not_authenticated",
				Message: "Authentication required",
			})
		}

		// Check if user has any of the required permissions
		userPermList, ok := userPermissions.([]string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "internal_error",
				Message: "Invalid permissions format",
			})
		}

		for _, requiredPerm := range permissions {
			for _, userPerm := range userPermList {
				if userPerm == requiredPerm {
					// User has required permission, continue
					return c.Next()
				}
			}
		}

		// User doesn't have any of the required permissions
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
			Error:   "insufficient_permissions",
			Message: fmt.Sprintf("One of these permissions is required: %v", permissions),
		})
	}
}

// GetAccountID gets the account ID from the context
func GetAccountID(c *fiber.Ctx) (uuid.UUID, error) {
	accountID := c.Locals("account_id")
	if accountID == nil {
		return uuid.Nil, fmt.Errorf("account_id not found in context")
	}

	id, ok := accountID.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid account_id format in context")
	}

	return id, nil
}

// GetAPIKeyID gets the API key ID from the context
func GetAPIKeyID(c *fiber.Ctx) (uuid.UUID, error) {
	apiKeyID := c.Locals("api_key_id")
	if apiKeyID == nil {
		return uuid.Nil, fmt.Errorf("api_key_id not found in context")
	}

	id, ok := apiKeyID.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid api_key_id format in context")
	}

	return id, nil
}

// GetAPIKeyName gets the API key name from the context
func GetAPIKeyName(c *fiber.Ctx) (string, error) {
	apiKeyName := c.Locals("api_key_name")
	if apiKeyName == nil {
		return "", fmt.Errorf("api_key_name not found in context")
	}

	name, ok := apiKeyName.(string)
	if !ok {
		return "", fmt.Errorf("invalid api_key_name format in context")
	}

	return name, nil
}

// GetPermissions gets the permissions from the context
func GetPermissions(c *fiber.Ctx) ([]string, error) {
	permissions := c.Locals("permissions")
	if permissions == nil {
		return nil, fmt.Errorf("permissions not found in context")
	}

	perms, ok := permissions.([]string)
	if !ok {
		return nil, fmt.Errorf("invalid permissions format in context")
	}

	return perms, nil
}

// HasPermission checks if the current context has a specific permission
func HasPermission(c *fiber.Ctx, permission string) bool {
	permissions, err := GetPermissions(c)
	if err != nil {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}
