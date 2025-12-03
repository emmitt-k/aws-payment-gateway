package security

import (
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ConstantTimeCompare performs constant-time comparison to prevent timing attacks
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// GenerateSecureAPIKey generates a secure API key with proper entropy
func GenerateSecureAPIKey() string {
	// Generate UUID-based API key with sufficient entropy
	return uuid.New().String()
}

// HashAPIKey securely hashes an API key using bcrypt
func HashAPIKey(apiKey string) (string, error) {
	// Use bcrypt with recommended cost for security
	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ValidateAPIKeyFormat validates that an API key meets security requirements
func ValidateAPIKeyFormat(apiKey string) bool {
	// Basic validation - in production, you might want stricter requirements
	if len(apiKey) < 32 {
		return false // Too short
	}

	if len(apiKey) > 256 {
		return false // Too long
	}

	// Check for common weak patterns
	weakPatterns := []string{
		"password", "secret", "key", "token", "api",
		"test", "demo", "sample", "example",
		"123456", "abcdef", "qwerty",
	}

	apiKeyLower := string(apiKey)
	for _, pattern := range weakPatterns {
		if len(apiKeyLower) > len(pattern) {
			// Check if the weak pattern is contained in the API key
			for i := 0; i <= len(apiKeyLower)-len(pattern); i++ {
				if apiKeyLower[i:i+len(pattern)] == pattern {
					return false // Contains weak pattern
				}
			}
		}
	}

	return true
}

// IsRecentKey checks if a key was created recently (for rotation policies)
func IsRecentKey(createdAt time.Time, maxAge time.Duration) bool {
	return time.Since(createdAt) < maxAge
}

// ShouldRotateKey determines if a key should be rotated based on age and usage
func ShouldRotateKey(createdAt time.Time, lastUsed *time.Time, maxAge time.Duration, maxUnusedDuration time.Duration) bool {
	// Rotate if key is too old
	if time.Since(createdAt) > maxAge {
		return true
	}

	// Rotate if key hasn't been used in a long time
	if lastUsed != nil && time.Since(*lastUsed) > maxUnusedDuration {
		return true
	}

	return false
}

// GenerateKeyRotationWarning creates a warning message for key rotation
func GenerateKeyRotationWarning(keyName string, daysUntilRotation int) map[string]string {
	return map[string]string{
		"warning_type":        "key_rotation_required",
		"key_name":            keyName,
		"days_until_rotation": fmt.Sprintf("%d", daysUntilRotation),
		"message":             fmt.Sprintf("API key '%s' should be rotated in %d days", keyName, daysUntilRotation),
	}
}

// SanitizeForLogging removes sensitive information from strings for logging
func SanitizeForLogging(input string) string {
	// Remove potential sensitive data from logs
	if len(input) > 8 {
		// Show only first 4 and last 4 characters for long keys
		return input[:4] + "****" + input[len(input)-4:]
	}
	return input
}

// ValidatePasswordComplexity checks if a password meets complexity requirements
func ValidatePasswordComplexity(password string) map[string]bool {
	requirements := make(map[string]bool)

	// Length requirement
	requirements["length"] = len(password) >= 12

	// Uppercase requirement
	hasUpper := false
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
			break
		}
	}
	requirements["uppercase"] = hasUpper

	// Lowercase requirement
	hasLower := false
	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
			break
		}
	}
	requirements["lowercase"] = hasLower

	// Number requirement
	hasNumber := false
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasNumber = true
			break
		}
	}
	requirements["number"] = hasNumber

	// Special character requirement
	hasSpecial := false
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, char := range password {
		for _, special := range specialChars {
			if char == special {
				hasSpecial = true
				break
			}
		}
	}
	requirements["special"] = hasSpecial

	return requirements
}
