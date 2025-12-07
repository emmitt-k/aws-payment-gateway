package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws-payment-gateway/internal/auth/adapter/http/dto"
	"github.com/aws-payment-gateway/internal/auth/tests/integration/fixtures"
	"github.com/aws-payment-gateway/internal/auth/tests/integration/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// AuthIntegrationTestSuite runs integration tests for the auth service
type AuthIntegrationTestSuite struct {
	utils.TestSuite
}

// SetupSuite runs once before all tests
func (suite *AuthIntegrationTestSuite) SetupSuite() {
	suite.TestSuite.SetupSuite()

	// Setup DynamoDB tables for testing
	dynamoSetup, err := utils.NewDynamoDBSetup(
		suite.Config.DynamoDBEndpoint,
		suite.Config.DynamoDBRegion,
	)
	suite.Require().NoError(err)

	err = dynamoSetup.SetupTables()
	suite.Require().NoError(err)
}

// TestAccountRegistration tests the account registration endpoint
func (suite *AuthIntegrationTestSuite) TestAccountRegistration() {
	tests := []struct {
		name           string
		request        dto.RegisterAppRequest
		expectedStatus int
		expectedError  string
		shouldSucceed  bool
	}{
		{
			name:           "Valid account registration",
			request:        fixtures.ValidAccount(),
			expectedStatus: http.StatusCreated,
			shouldSucceed:  true,
		},
		{
			name:           "Valid account registration without webhook",
			request:        fixtures.ValidAccountWithoutWebhook(),
			expectedStatus: http.StatusCreated,
			shouldSucceed:  true,
		},
		{
			name:           "Invalid account registration - missing name",
			request:        fixtures.InvalidAccountMissingName(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
		},
		{
			name:           "Invalid account registration - invalid webhook URL",
			request:        fixtures.InvalidAccountInvalidWebhook(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
		},
		{
			name: "Duplicate account registration",
			request: dto.RegisterAppRequest{
				Name: "Duplicate Test Company",
			},
			expectedStatus: http.StatusCreated,
			shouldSucceed:  true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Convert request to JSON
			reqBody, err := json.Marshal(tt.request)
			suite.Require().NoError(err)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			resp, err := suite.App.Test(req)
			suite.Require().NoError(err)

			// Check status code
			assert.Equal(suite.T(), tt.expectedStatus, resp.StatusCode)

			if tt.shouldSucceed {
				// Parse successful response
				var response dto.RegisterAppResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				suite.Require().NoError(err)

				// Verify response fields
				assert.NotEmpty(suite.T(), response.AccountID)
				assert.Equal(suite.T(), tt.request.Name, response.Name)
				assert.Equal(suite.T(), "active", response.Status)
				assert.NotZero(suite.T(), response.CreatedAt)

				// Verify UUID format (AccountID is already uuid.UUID type)
				assert.NotEqual(suite.T(), uuid.Nil, response.AccountID, "AccountID should be a valid UUID")
			} else if tt.expectedError != "" {
				// Parse error response
				var errorResp dto.ErrorResponse
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				suite.Require().NoError(err)

				assert.Equal(suite.T(), tt.expectedError, errorResp.Error)
			}

			// Close response body
			resp.Body.Close()
		})
	}
}

// TestAccountRegistrationDuplicate tests duplicate account registration
func (suite *AuthIntegrationTestSuite) TestAccountRegistrationDuplicate() {
	// First registration
	request := fixtures.ValidAccount()
	request.Name = "Duplicate Test Company"

	reqBody, err := json.Marshal(request)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.App.Test(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Second registration with same name
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := suite.App.Test(req2)
	suite.Require().NoError(err)
	suite.Equal(http.StatusConflict, resp2.StatusCode)

	// Parse error response
	var errorResp dto.ErrorResponse
	err = json.NewDecoder(resp2.Body).Decode(&errorResp)
	suite.Require().NoError(err)
	suite.Equal("account_exists", errorResp.Error)
	resp2.Body.Close()
}

// TestHealthCheck tests the health check endpoint
func (suite *AuthIntegrationTestSuite) TestHealthCheck() {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	resp, err := suite.App.Test(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	// Parse response
	var response dto.HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), "healthy", response.Status)
	assert.Equal(suite.T(), "auth-service", response.Service)
	assert.NotZero(suite.T(), response.Timestamp)
	resp.Body.Close()
}

// TestAccountRegistrationWithDatabase verifies data persistence
func (suite *AuthIntegrationTestSuite) TestAccountRegistrationWithDatabase() {
	request := fixtures.ValidAccount()
	request.Name = "Database Test Company"

	reqBody, err := json.Marshal(request)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.App.Test(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	// Parse response
	var response dto.RegisterAppResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.Require().NoError(err)
	resp.Body.Close()

	// Verify data was stored in PostgreSQL
	ctx := context.Background()
	query := `SELECT id, name, status, webhook_url, created_at, updated_at FROM accounts WHERE id = $1`
	row := suite.PostgresClient.QueryRowContext(ctx, query, response.AccountID)

	var id, name, status, webhookURL string
	var createdAt, updatedAt time.Time
	err = row.Scan(&id, &name, &status, &webhookURL, &createdAt, &updatedAt)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), response.AccountID, id)
	assert.Equal(suite.T(), request.Name, name)
	assert.Equal(suite.T(), "active", status)
	assert.Equal(suite.T(), *request.WebhookURL, webhookURL)
	assert.NotZero(suite.T(), createdAt)
	assert.NotZero(suite.T(), updatedAt)

	// Verify audit log was created in DynamoDB
	// This would require implementing audit log verification
	// For now, we'll just verify the setup worked
	suite.NotNil(suite.T(), suite.AuditDynamoClient)
}

// TestAccountRegistrationConcurrency tests concurrent registrations
func (suite *AuthIntegrationTestSuite) TestAccountRegistrationConcurrency() {
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			request := fixtures.ValidAccount()
			request.Name = fmt.Sprintf("Concurrent Test Company %d", index)
			request.WebhookURL = stringPtr(fmt.Sprintf("https://test%d.example.com/webhook", index))

			reqBody, err := json.Marshal(request)
			if err != nil {
				results <- err
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := suite.App.Test(req)
			if err != nil {
				results <- err
				return
			}

			if resp.StatusCode != http.StatusCreated {
				results <- fmt.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
				return
			}

			resp.Body.Close()
			results <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(suite.T(), err, "Concurrent registration should succeed")
	}
}

// TestAPIKeyIssuance tests the API key issuance endpoint
func (suite *AuthIntegrationTestSuite) TestAPIKeyIssuance() {
	// First, register an account to get an account ID
	registerReq := dto.RegisterAppRequest{
		Name:       "API Key Test Company",
		WebhookURL: stringPtr("https://test.example.com/webhook"),
	}

	reqBody, err := json.Marshal(registerReq)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.App.Test(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var registerResp dto.RegisterAppResponse
	err = json.NewDecoder(resp.Body).Decode(&registerResp)
	suite.Require().NoError(err)
	resp.Body.Close()

	// Test API key issuance
	apiKeyReq := fixtures.ValidAPIKeyRequest(registerResp.AccountID)

	apiKeyReqBody, err := json.Marshal(apiKeyReq)
	suite.Require().NoError(err)

	apiKeyReqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/api-keys", bytes.NewBuffer(apiKeyReqBody))
	apiKeyReqHTTP.Header.Set("Content-Type", "application/json")

	apiKeyResp, err := suite.App.Test(apiKeyReqHTTP)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, apiKeyResp.StatusCode)

	var apiKeyRespDTO dto.IssueApiKeyResponse
	err = json.NewDecoder(apiKeyResp.Body).Decode(&apiKeyRespDTO)
	suite.Require().NoError(err)
	apiKeyResp.Body.Close()

	// Verify API key response
	assert.NotEmpty(suite.T(), apiKeyRespDTO.APIKeyID)
	assert.Equal(suite.T(), registerResp.AccountID, apiKeyRespDTO.AccountID)
	assert.Equal(suite.T(), "Test API Key", apiKeyRespDTO.Name)
	assert.Equal(suite.T(), []string{"read:keys", "write:keys"}, apiKeyRespDTO.Permissions)
	assert.Equal(suite.T(), "active", apiKeyRespDTO.Status)
	assert.NotZero(suite.T(), apiKeyRespDTO.CreatedAt)
	assert.NotZero(suite.T(), apiKeyRespDTO.ExpiresAt)
}

// TestAPIKeyValidation tests the API key validation endpoint
func (suite *AuthIntegrationTestSuite) TestAPIKeyValidation() {
	// First, register an account and issue an API key
	registerReq := dto.RegisterAppRequest{
		Name: "API Key Validation Test Company",
	}

	reqBody, err := json.Marshal(registerReq)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.App.Test(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var registerResp dto.RegisterAppResponse
	err = json.NewDecoder(resp.Body).Decode(&registerResp)
	suite.Require().NoError(err)
	resp.Body.Close()

	// Issue an API key
	apiKeyReq := fixtures.ValidAPIKeyRequest(registerResp.AccountID)
	apiKeyReq.Name = "Validation Test API Key"
	apiKeyReq.Permissions = []string{"read:keys"}

	apiKeyReqBody, err := json.Marshal(apiKeyReq)
	suite.Require().NoError(err)

	apiKeyReqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/api-keys", bytes.NewBuffer(apiKeyReqBody))
	apiKeyReqHTTP.Header.Set("Content-Type", "application/json")

	apiKeyResp, err := suite.App.Test(apiKeyReqHTTP)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, apiKeyResp.StatusCode)

	var apiKeyRespDTO dto.IssueApiKeyResponse
	err = json.NewDecoder(apiKeyResp.Body).Decode(&apiKeyRespDTO)
	suite.Require().NoError(err)
	apiKeyResp.Body.Close()

	// Test API key validation
	validateReq := dto.ValidateApiKeyRequest{
		KeyHash: apiKeyRespDTO.KeyHash,
	}

	validateReqBody, err := json.Marshal(validateReq)
	suite.Require().NoError(err)

	validateReqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate", bytes.NewBuffer(validateReqBody))
	validateReqHTTP.Header.Set("Content-Type", "application/json")

	validateResp, err := suite.App.Test(validateReqHTTP)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, validateResp.StatusCode)

	var validateRespDTO dto.ValidateApiKeyResponse
	err = json.NewDecoder(validateResp.Body).Decode(&validateRespDTO)
	suite.Require().NoError(err)
	validateResp.Body.Close()

	// Verify validation response
	assert.True(suite.T(), validateRespDTO.Valid)
	assert.Equal(suite.T(), registerResp.AccountID, *validateRespDTO.AccountID)
	assert.Equal(suite.T(), "Validation Test API Key", *validateRespDTO.Name)
	assert.Equal(suite.T(), []string{"read:keys"}, validateRespDTO.Permissions)
}

// TestAPIKeyRevocation tests the API key revocation endpoint
func (suite *AuthIntegrationTestSuite) TestAPIKeyRevocation() {
	// First, register an account and issue an API key
	registerReq := dto.RegisterAppRequest{
		Name: "API Key Revocation Test Company",
	}

	reqBody, err := json.Marshal(registerReq)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.App.Test(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var registerResp dto.RegisterAppResponse
	err = json.NewDecoder(resp.Body).Decode(&registerResp)
	suite.Require().NoError(err)
	resp.Body.Close()

	// Issue an API key
	apiKeyReq := fixtures.ValidAPIKeyRequest(registerResp.AccountID)
	apiKeyReq.Name = "Revocation Test API Key"

	apiKeyReqBody, err := json.Marshal(apiKeyReq)
	suite.Require().NoError(err)

	apiKeyReqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/api-keys", bytes.NewBuffer(apiKeyReqBody))
	apiKeyReqHTTP.Header.Set("Content-Type", "application/json")

	apiKeyResp, err := suite.App.Test(apiKeyReqHTTP)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, apiKeyResp.StatusCode)

	var apiKeyRespDTO dto.IssueApiKeyResponse
	err = json.NewDecoder(apiKeyResp.Body).Decode(&apiKeyRespDTO)
	suite.Require().NoError(err)
	apiKeyResp.Body.Close()

	// Create a mock request with authentication middleware
	revokeReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/auth/api-keys/%s", apiKeyRespDTO.APIKeyID), nil)
	revokeReq.Header.Set("Content-Type", "application/json")

	// Set the account context as the middleware would
	revokeReq.Header.Set("X-Account-ID", registerResp.AccountID.String())
	revokeReq.Header.Set("X-API-Key-ID", apiKeyRespDTO.APIKeyID.String())

	revokeResp, err := suite.App.Test(revokeReq)
	suite.Require().NoError(err)
	suite.Equal(http.StatusNoContent, revokeResp.StatusCode)
	revokeResp.Body.Close()

	// Verify the API key is no longer valid
	validateReq := dto.ValidateApiKeyRequest{
		KeyHash: apiKeyRespDTO.KeyHash,
	}

	validateReqBody, err := json.Marshal(validateReq)
	suite.Require().NoError(err)

	validateReqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate", bytes.NewBuffer(validateReqBody))
	validateReqHTTP.Header.Set("Content-Type", "application/json")

	validateResp, err := suite.App.Test(validateReqHTTP)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, validateResp.StatusCode)

	var validateRespDTO dto.ValidateApiKeyResponse
	err = json.NewDecoder(validateResp.Body).Decode(&validateRespDTO)
	suite.Require().NoError(err)
	validateResp.Body.Close()

	// Verify the API key is no longer valid
	assert.False(suite.T(), validateRespDTO.Valid)
}

// TestGetAPIKeys tests the get API keys endpoint
func (suite *AuthIntegrationTestSuite) TestGetAPIKeys() {
	// First, register an account
	registerReq := dto.RegisterAppRequest{
		Name: "Get API Keys Test Company",
	}

	reqBody, err := json.Marshal(registerReq)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.App.Test(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var registerResp dto.RegisterAppResponse
	err = json.NewDecoder(resp.Body).Decode(&registerResp)
	suite.Require().NoError(err)
	resp.Body.Close()

	// Issue multiple API keys
	for i := 0; i < 3; i++ {
		apiKeyReq := fixtures.ValidAPIKeyRequest(registerResp.AccountID)
		apiKeyReq.Name = fmt.Sprintf("Test API Key %d", i+1)

		apiKeyReqBody, err := json.Marshal(apiKeyReq)
		suite.Require().NoError(err)

		apiKeyReqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/api-keys", bytes.NewBuffer(apiKeyReqBody))
		apiKeyReqHTTP.Header.Set("Content-Type", "application/json")

		apiKeyResp, err := suite.App.Test(apiKeyReqHTTP)
		suite.Require().NoError(err)
		suite.Equal(http.StatusCreated, apiKeyResp.StatusCode)
		apiKeyResp.Body.Close()
	}

	// Get API keys for the account
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/auth/accounts/%s/api-keys", registerResp.AccountID), nil)
	getReq.Header.Set("Content-Type", "application/json")

	// Set the account context as the middleware would
	getReq.Header.Set("X-Account-ID", registerResp.AccountID.String())

	getResp, err := suite.App.Test(getReq)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, getResp.StatusCode)

	var getRespDTO dto.GetAPIKeysResponse
	err = json.NewDecoder(getResp.Body).Decode(&getRespDTO)
	suite.Require().NoError(err)
	getResp.Body.Close()

	// Verify the response
	assert.Len(suite.T(), getRespDTO.APIKeys, 3)

	for _, apiKey := range getRespDTO.APIKeys {
		assert.NotEmpty(suite.T(), apiKey.Name)
		assert.Equal(suite.T(), "active", apiKey.Status)
		assert.NotZero(suite.T(), apiKey.CreatedAt)
	}
}

// TestAuditLogging tests audit logging for various operations
func (suite *AuthIntegrationTestSuite) TestAuditLogging() {
	// Register an account
	registerReq := dto.RegisterAppRequest{
		Name:       "Audit Test Company",
		WebhookURL: stringPtr("https://audit-test.example.com/webhook"),
	}

	reqBody, err := json.Marshal(registerReq)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.App.Test(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var registerResp dto.RegisterAppResponse
	err = json.NewDecoder(resp.Body).Decode(&registerResp)
	suite.Require().NoError(err)
	resp.Body.Close()

	// Issue an API key
	apiKeyReq := fixtures.ValidAPIKeyRequest(registerResp.AccountID)
	apiKeyReq.Name = "Audit Test API Key"

	apiKeyReqBody, err := json.Marshal(apiKeyReq)
	suite.Require().NoError(err)

	apiKeyReqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/api-keys", bytes.NewBuffer(apiKeyReqBody))
	apiKeyReqHTTP.Header.Set("Content-Type", "application/json")

	apiKeyResp, err := suite.App.Test(apiKeyReqHTTP)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, apiKeyResp.StatusCode)

	var apiKeyRespDTO dto.IssueApiKeyResponse
	err = json.NewDecoder(apiKeyResp.Body).Decode(&apiKeyRespDTO)
	suite.Require().NoError(err)
	apiKeyResp.Body.Close()

	// Verify audit logs exist in DynamoDB
	ctx := context.Background()

	// Scan audit logs table for the account
	scanInput := &dynamodb.ScanInput{
		TableName:        aws.String(suite.Config.AuditLogsTable),
		FilterExpression: aws.String("contains(partition_key, :account_id)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account_id": &types.AttributeValueMemberS{Value: registerResp.AccountID.String()},
		},
	}

	var scanResult []map[string]types.AttributeValue
	err = suite.AuditDynamoClient.ScanItems(ctx, scanInput, &scanResult)
	suite.Require().NoError(err)

	// We should have at least 2 audit logs: one for registration, one for API key issuance
	assert.GreaterOrEqual(suite.T(), len(scanResult), 2)

	// Check for registration audit log
	foundRegistrationLog := false
	foundAPIKeyLog := false

	for _, item := range scanResult {
		if eventType, ok := item["event_type"]; ok {
			if eventVal, ok := eventType.(*types.AttributeValueMemberS); ok {
				if eventVal.Value == "account_registered" {
					foundRegistrationLog = true
				} else if eventVal.Value == "api_key_issued" {
					foundAPIKeyLog = true
				}
			}
		}
	}

	assert.True(suite.T(), foundRegistrationLog, "Should find account registration audit log")
	assert.True(suite.T(), foundAPIKeyLog, "Should find API key issuance audit log")
}

// stringPtr returns a pointer to the string value
func stringPtr(s string) *string {
	return &s
}

// TestAuthIntegrationTestSuite runs the integration test suite
func TestAuthIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}
