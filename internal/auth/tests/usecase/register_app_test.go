package usecase_test

import (
	"errors"
	"testing"

	"github.com/aws-payment-gateway/internal/auth/domain"
	"github.com/aws-payment-gateway/internal/auth/tests/mocks"
	"github.com/aws-payment-gateway/internal/auth/tests/utils"
	"github.com/aws-payment-gateway/internal/auth/usecase"
)

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

func TestRegisterApp_Execute_Success(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

	input := usecase.RegisterAppInput{
		Name:       "test-app",
		WebhookURL: stringPtr("https://example.com/webhook"),
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
	utils.RequireEqual(t, "test-app", output.Name)
	utils.RequireEqual(t, string(domain.AccountStatusActive), output.Status)
	utils.RequireNotNil(t, output.AccountID)
	utils.RequireNotNil(t, output.CreatedAt)

	// Verify account was created
	account, err := mockAppRepo.GetByID(ctx, output.AccountID)
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, account)
	utils.RequireEqual(t, input.Name, account.Name)
	utils.RequireEqual(t, domain.AccountStatusActive, account.Status)
	utils.RequireEqual(t, input.WebhookURL, account.WebhookURL)
}

func TestRegisterApp_Execute_DuplicateName(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

	// Create existing account with same name
	existingAccount := utils.CreateTestAccount(t)
	existingAccount.Name = "duplicate-name"
	mockAppRepo.AddAccount(existingAccount)

	input := usecase.RegisterAppInput{
		Name: "duplicate-name", // Same name as existing
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "app with name 'duplicate-name' already exists", err.Error())
}

func TestRegisterApp_Execute_EmptyName(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

	input := usecase.RegisterAppInput{
		Name: "", // Empty name
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: name is required", err.Error())
}

func TestRegisterApp_Execute_ShortName(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

	input := usecase.RegisterAppInput{
		Name: "ab", // Less than 3 characters
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: name must be at least 3 characters", err.Error())
}

func TestRegisterApp_Execute_InvalidWebhookURL(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

	input := usecase.RegisterAppInput{
		Name:       "test-app",
		WebhookURL: stringPtr("invalid-url"), // Invalid URL
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "invalid input: invalid webhook URL format", err.Error())
}

func TestRegisterApp_Execute_ValidWebhookURL(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

	input := usecase.RegisterAppInput{
		Name:       "test-app",
		WebhookURL: stringPtr("https://example.com/webhook"),
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)

	// Verify webhook URL was saved
	account, err := mockAppRepo.GetByID(ctx, output.AccountID)
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, account)
	utils.RequireEqual(t, input.WebhookURL, account.WebhookURL)
}

func TestRegisterApp_Execute_NilWebhookURL(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

	input := usecase.RegisterAppInput{
		Name:       "test-app",
		WebhookURL: nil, // No webhook URL
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)

	// Verify webhook URL is nil
	account, err := mockAppRepo.GetByID(ctx, output.AccountID)
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, account)
	utils.RequireNil(t, account.WebhookURL)
}

func TestRegisterApp_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

	// Set up repository to return error on GetByName
	repoError := errors.New("database error")
	mockAppRepo.SetGetError(repoError)

	input := usecase.RegisterAppInput{
		Name: "test-app",
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to check existing app: database error", err.Error())
}

func TestRegisterApp_Execute_CreateError(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

	// Set up repository to return error on Create
	repoError := errors.New("create error")
	mockAppRepo.SetCreateError(repoError)

	input := usecase.RegisterAppInput{
		Name: "test-app",
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	utils.RequireError(t, err)
	utils.RequireNil(t, output)
	utils.RequireEqual(t, "failed to create account: create error", err.Error())
}

func TestRegisterApp_Execute_LongName(t *testing.T) {
	// Arrange
	ctx := utils.TestContext(t)
	mockAppRepo := mocks.NewMockAppRepository()
	mockApiKeyRepo := mocks.NewMockApiKeyRepository()
	uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

	// Create a name longer than 100 characters
	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}

	input := usecase.RegisterAppInput{
		Name: longName,
	}

	// Act
	output, err := uc.Execute(ctx, input)

	// Assert
	// Note: The current implementation doesn't validate max length, but this test documents expected behavior
	// If max length validation is added, this test should expect an error
	utils.RequireNoError(t, err)
	utils.RequireNotNil(t, output)
}

func TestRegisterApp_Execute_ValidURLFormats(t *testing.T) {
	tests := []struct {
		name       string
		webhookURL string
		valid      bool
	}{
		{
			name:       "valid https URL",
			webhookURL: "https://example.com/webhook",
			valid:      true,
		},
		{
			name:       "valid http URL",
			webhookURL: "http://example.com/webhook",
			valid:      true,
		},
		{
			name:       "invalid URL - no protocol",
			webhookURL: "example.com/webhook",
			valid:      false,
		},
		{
			name:       "invalid URL - too short",
			webhookURL: "http://",
			valid:      false,
		},
		{
			name:       "invalid URL - ftp protocol",
			webhookURL: "ftp://example.com/webhook",
			valid:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := utils.TestContext(t)
			mockAppRepo := mocks.NewMockAppRepository()
			mockApiKeyRepo := mocks.NewMockApiKeyRepository()
			uc := usecase.NewRegisterApp(mockAppRepo, mockApiKeyRepo)

			input := usecase.RegisterAppInput{
				Name:       "test-app",
				WebhookURL: stringPtr(tt.webhookURL),
			}

			// Act
			output, err := uc.Execute(ctx, input)

			// Assert
			if tt.valid {
				utils.RequireNoError(t, err)
				utils.RequireNotNil(t, output)
			} else {
				utils.RequireError(t, err)
				utils.RequireNil(t, output)
			}
		})
	}
}
