package fixtures

import (
	"github.com/aws-payment-gateway/internal/auth/adapter/http/dto"
	"github.com/google/uuid"
)

// ValidAccount returns a valid account registration request
func ValidAccount() dto.RegisterAppRequest {
	return dto.RegisterAppRequest{
		Name:       "Test Company",
		WebhookURL: stringPtr("https://test.example.com/webhook"),
	}
}

// ValidAccountWithoutWebhook returns a valid account registration request without webhook
func ValidAccountWithoutWebhook() dto.RegisterAppRequest {
	return dto.RegisterAppRequest{
		Name: "Test Company No Webhook",
	}
}

// InvalidAccountMissingName returns an invalid account registration request (missing name)
func InvalidAccountMissingName() dto.RegisterAppRequest {
	return dto.RegisterAppRequest{
		WebhookURL: stringPtr("https://test.example.com/webhook"),
	}
}

// InvalidAccountInvalidWebhook returns an invalid account registration request (invalid webhook URL)
func InvalidAccountInvalidWebhook() dto.RegisterAppRequest {
	return dto.RegisterAppRequest{
		Name:       "Test Company",
		WebhookURL: stringPtr("invalid-url"),
	}
}

// ValidAccountForAPIKey returns a valid account for API key testing
func ValidAccountForAPIKey() dto.RegisterAppRequest {
	return dto.RegisterAppRequest{
		Name: "API Key Test Company",
	}
}

// ValidAPIKeyRequest returns a valid API key issuance request
func ValidAPIKeyRequest(accountID uuid.UUID) dto.IssueApiKeyRequest {
	return dto.IssueApiKeyRequest{
		AccountID:   accountID,
		Name:        "Test API Key",
		Permissions: []string{"read:keys", "write:keys"},
	}
}

// ValidAPIKeyRequestWithExpiration returns a valid API key issuance request with expiration
func ValidAPIKeyRequestWithExpiration(accountID uuid.UUID) dto.IssueApiKeyRequest {
	expiresIn := 24 // 24 hours
	return dto.IssueApiKeyRequest{
		AccountID:   accountID,
		Name:        "Test API Key with Expiration",
		Permissions: []string{"read:keys"},
		ExpiresIn:   &expiresIn,
	}
}

// InvalidAPIKeyRequestMissingAccount returns an invalid API key issuance request (missing account ID)
func InvalidAPIKeyRequestMissingAccount() dto.IssueApiKeyRequest {
	return dto.IssueApiKeyRequest{
		Name:        "Test API Key",
		Permissions: []string{"read:keys"},
	}
}

// InvalidAPIKeyRequestMissingName returns an invalid API key issuance request (missing name)
func InvalidAPIKeyRequestMissingName(accountID uuid.UUID) dto.IssueApiKeyRequest {
	return dto.IssueApiKeyRequest{
		AccountID:   accountID,
		Permissions: []string{"read:keys"},
	}
}

// InvalidAPIKeyRequestMissingPermissions returns an invalid API key issuance request (missing permissions)
func InvalidAPIKeyRequestMissingPermissions(accountID uuid.UUID) dto.IssueApiKeyRequest {
	return dto.IssueApiKeyRequest{
		AccountID: accountID,
		Name:      "Test API Key",
	}
}

// stringPtr returns a pointer to the string value
func stringPtr(s string) *string {
	return &s
}
