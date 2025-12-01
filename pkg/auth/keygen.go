package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// GenerateAPIKey generates a new secure API key
func GenerateAPIKey() (string, error) {
	// Generate 32 random bytes
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string
	apiKey := hex.EncodeToString(keyBytes)

	return apiKey, nil
}

// HashAPIKey creates a bcrypt hash of the API key for secure storage
func HashAPIKey(apiKey string) (string, error) {
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash API key: %w", err)
	}

	return string(hashedKey), nil
}

// ValidateAPIKey compares a raw API key with its hash
func ValidateAPIKey(apiKey, hashedKey string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(apiKey))
}

// GenerateAPIKeyWithHash generates a new API key and returns both the key and its hash
func GenerateAPIKeyWithHash() (apiKey string, keyHash string, err error) {
	// Generate API key
	apiKey, err = GenerateAPIKey()
	if err != nil {
		return "", "", err
	}

	// Hash the API key
	keyHash, err = HashAPIKey(apiKey)
	if err != nil {
		return "", "", err
	}

	return apiKey, keyHash, nil
}
