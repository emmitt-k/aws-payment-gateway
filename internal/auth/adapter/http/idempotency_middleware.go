package http

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/usecase"
	"github.com/gofiber/fiber/v2"
)

// IdempotencyMiddleware provides idempotency handling for HTTP requests
type IdempotencyMiddleware struct {
	checkIdempotency    *usecase.CheckIdempotency
	createIdempotency   *usecase.CreateIdempotency
	completeIdempotency *usecase.CompleteIdempotency
}

// NewIdempotencyMiddleware creates a new IdempotencyMiddleware
func NewIdempotencyMiddleware(
	checkIdempotency *usecase.CheckIdempotency,
	createIdempotency *usecase.CreateIdempotency,
	completeIdempotency *usecase.CompleteIdempotency,
) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{
		checkIdempotency:    checkIdempotency,
		createIdempotency:   createIdempotency,
		completeIdempotency: completeIdempotency,
	}
}

// generateRequestHash generates a hash for the request
func (m *IdempotencyMiddleware) generateRequestHash(c *fiber.Ctx) string {
	// Get request method and path
	method := c.Method()
	path := c.Path()

	// Get request body (if any)
	var body string
	if c.Body() != nil {
		body = string(c.Body())
	}

	// Get relevant headers
	headers := make(map[string]string)
	for key, values := range c.GetReqHeaders() {
		if len(values) > 0 {
			headers[key] = values[0] // Take first value
		}
	}

	// Create normalized request string for hashing
	requestData := fmt.Sprintf("%s:%s:%s:%s", method, path, body)

	// Add headers to request data
	for key, value := range headers {
		requestData += fmt.Sprintf(":%s:%s", strings.ToLower(key), value)
	}

	// Hash the request data
	hash := sha256.Sum256([]byte(requestData))
	return hex.EncodeToString(hash[:])
}

// extractIdempotencyKey extracts idempotency key from request
func (m *IdempotencyMiddleware) extractIdempotencyKey(c *fiber.Ctx) string {
	// Try different header names
	idempotencyKey := c.Get("Idempotency-Key")
	if idempotencyKey == "" {
		idempotencyKey = c.Get("X-Idempotency-Key")
	}
	if idempotencyKey == "" {
		idempotencyKey = c.Get("X-Idempotency-Key")
	}

	return idempotencyKey
}

// Check creates a middleware that checks for existing idempotency keys
func (m *IdempotencyMiddleware) Check() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract idempotency key from request
		idempotencyKey := m.extractIdempotencyKey(c)
		if idempotencyKey == "" {
			// No idempotency key provided, skip check
			return c.Next()
		}

		// Generate request hash
		requestHash := m.generateRequestHash(c)

		// Check if idempotency key exists
		output, err := m.checkIdempotency.Execute(c.Context(), usecase.CheckIdempotencyInput{
			IdempotencyKey: idempotencyKey,
			RequestHash:    requestHash,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "idempotency_check_failed",
				"message": "Failed to check idempotency key",
				"details": err.Error(),
			})
		}

		if output.Exists {
			// Key exists, check status
			if output.Status == string(domain.IdempotencyKeyStatusCompleted) {
				// Request already completed, return cached response
				if output.Response != "" {
					c.Set("Content-Type", "application/json")
					return c.Status(200).SendString(output.Response)
				}
				return c.Status(200).JSON(fiber.Map{
					"status":       "completed",
					"completed_at": output.CreatedAt,
				})
			} else if output.Status == string(domain.IdempotencyKeyStatusExpired) {
				// Key exists but expired, treat as new request
				return c.Next()
			} else {
				// Key exists and is pending, request is in progress
				return c.Status(409).JSON(fiber.Map{
					"error":   "idempotency_key_pending",
					"message": "Request with this idempotency key is already in progress",
				})
			}
		}

		// No existing key, continue with request
		return c.Next()
	}
}

// Create creates a middleware that creates new idempotency keys
func (m *IdempotencyMiddleware) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract idempotency key from request
		idempotencyKey := m.extractIdempotencyKey(c)
		if idempotencyKey == "" {
			// No idempotency key provided, skip idempotency handling
			return c.Next()
		}

		// Generate request hash
		requestHash := m.generateRequestHash(c)

		// Create new idempotency key
		output, err := m.createIdempotency.Execute(c.Context(), usecase.CreateIdempotencyInput{
			IdempotencyKey: idempotencyKey,
			RequestHash:    requestHash,
			Response:       "", // Will be set by the actual handler
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "idempotency_creation_failed",
				"message": "Failed to create idempotency key",
				"details": err.Error(),
			})
		}

		// Store the generated idempotency key in response header for client to use
		c.Set("X-Idempotency-Key", output.IdempotencyKey)

		// Continue with request processing
		return c.Next()
	}
}

// Complete creates a middleware that completes idempotency keys
func (m *IdempotencyMiddleware) Complete() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract idempotency key from request
		idempotencyKey := m.extractIdempotencyKey(c)
		if idempotencyKey == "" {
			// No idempotency key provided, skip completion
			return c.Next()
		}

		// Complete the idempotency key
		_, err := m.completeIdempotency.Execute(c.Context(), usecase.CompleteIdempotencyInput{
			IdempotencyKey: idempotencyKey,
			Response:       "", // Will be set by the actual handler response
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "idempotency_completion_failed",
				"message": "Failed to complete idempotency key",
				"details": err.Error(),
			})
		}

		// Store the completion status in response header
		c.Set("X-Idempotency-Key", idempotencyKey)

		// Continue with request processing
		return c.Next()
	}
}
