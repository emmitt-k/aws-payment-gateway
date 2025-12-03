package http

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws-payment-gateway/internal/auth/repository"
	"github.com/gofiber/fiber/v2"
)

// RateLimitConfig defines configuration for rate limiting
type RateLimitConfig struct {
	// Requests per window
	Requests int
	// Window duration
	Window time.Duration
	// Key generator function (e.g., by IP, API key, account)
	KeyGenerator func(*fiber.Ctx) string
}

// RateLimitMiddleware provides rate limiting functionality
type RateLimitMiddleware struct {
	repository repository.RateLimitRepository
	configs    map[string]*RateLimitConfig
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware
func NewRateLimitMiddleware(repository repository.RateLimitRepository) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		repository: repository,
		configs:    make(map[string]*RateLimitConfig),
	}
}

// AddConfig adds a rate limit configuration
func (m *RateLimitMiddleware) AddConfig(name string, config *RateLimitConfig) {
	m.configs[name] = config
}

// ByIP creates a rate limiter that limits by IP address
func (m *RateLimitMiddleware) ByIP(requests int, window time.Duration) fiber.Handler {
	config := &RateLimitConfig{
		Requests:     requests,
		Window:       window,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
	}
	return m.createHandler("ip", config)
}

// ByAPIKey creates a rate limiter that limits by API key
func (m *RateLimitMiddleware) ByAPIKey(requests int, window time.Duration) fiber.Handler {
	config := &RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyGenerator: func(c *fiber.Ctx) string {
			if accountID := c.Locals("account_id"); accountID != nil {
				return fmt.Sprintf("account:%v", accountID)
			}
			return c.IP() // Fallback to IP
		},
	}
	return m.createHandler("apikey", config)
}

// ByEndpoint creates a rate limiter that limits by endpoint
func (m *RateLimitMiddleware) ByEndpoint(requests int, window time.Duration) fiber.Handler {
	config := &RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return fmt.Sprintf("%s:%s", c.IP(), c.Path())
		},
	}
	return m.createHandler("endpoint", config)
}

// createHandler creates the actual rate limiting handler
func (m *RateLimitMiddleware) createHandler(name string, config *RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := config.KeyGenerator(c)
		if key == "" {
			// No key available, skip rate limiting
			return c.Next()
		}

		// Check rate limit
		allowed, remaining, resetTime, err := m.repository.CheckRateLimit(
			context.Background(),
			fmt.Sprintf("%s:%s", name, key),
			config.Requests,
			config.Window,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "rate_limit_check_failed",
				"message": "Failed to check rate limit",
				"details": err.Error(),
			})
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":          "rate_limit_exceeded",
				"message":        "Rate limit exceeded",
				"limit":          config.Requests,
				"window_seconds": int(config.Window.Seconds()),
				"reset_time":     resetTime,
				"retry_after":    int(config.Window.Seconds()),
			})
		}

		return c.Next()
	}
}
