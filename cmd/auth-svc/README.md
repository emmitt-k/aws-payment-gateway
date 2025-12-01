# Auth Service

The Auth Service is responsible for managing application registration, API key issuance, and authentication for the AWS Payment Gateway system.

## Features

- **Account Registration**: Register new applications/companies in the system
- **API Key Management**: Issue, validate, and revoke API keys
- **Permission Management**: Granular permissions for API keys
- **Secure Storage**: Hashed API keys stored in DynamoDB
- **Audit Logging**: Comprehensive logging of authentication events

## API Endpoints

### Public Endpoints

#### Register Application
```
POST /api/v1/auth/register
```

Request Body:
```json
{
  "name": "My Application",
  "webhook_url": "https://example.com/webhook"
}
```

Response:
```json
{
  "account_id": "uuid",
  "name": "My Application",
  "status": "active",
  "created_at": "2023-01-01T00:00:00Z"
}
```

#### Issue API Key
```
POST /api/v1/auth/api-keys
```

Request Body:
```json
{
  "account_id": "uuid",
  "name": "Production Key",
  "permissions": ["read:accounts", "write:accounts"],
  "expires_in": 8760
}
```

Response:
```json
{
  "api_key_id": "uuid",
  "api_key": "generated-key-here",
  "key_hash": "hashed-key",
  "account_id": "uuid",
  "name": "Production Key",
  "permissions": ["read:accounts", "write:accounts"],
  "status": "active",
  "expires_at": "2024-01-01T00:00:00Z",
  "created_at": "2023-01-01T00:00:00Z"
}
```

#### Validate API Key
```
POST /api/v1/auth/validate
```

Request Body:
```json
{
  "key_hash": "hashed-key"
}
```

Response:
```json
{
  "valid": true,
  "account_id": "uuid",
  "api_key_id": "uuid",
  "name": "Production Key",
  "permissions": ["read:accounts", "write:accounts"],
  "last_used_at": "2023-01-01T00:00:00Z",
  "expires_at": "2024-01-01T00:00:00Z"
}
```

### Protected Endpoints

All protected endpoints require an `x-api-key` header or `Authorization: Bearer <key>` header.

#### Get API Keys
```
GET /api/v1/auth/accounts/{account_id}/api-keys?limit=10&offset=0
```

Requires permission: `read:keys`

Response:
```json
{
  "api_keys": [
    {
      "api_key_id": "uuid",
      "name": "Production Key",
      "permissions": ["read:accounts", "write:accounts"],
      "status": "active",
      "last_used_at": "2023-01-01T00:00:00Z",
      "expires_at": "2024-01-01T00:00:00Z",
      "created_at": "2023-01-01T00:00:00Z"
    }
  ],
  "limit": 10,
  "offset": 0,
  "total": 1
}
```

#### Revoke API Key
```
DELETE /api/v1/auth/api-keys/{api_key_id}
```

Requires permission: `write:keys`

Response: `204 No Content`

#### Health Check
```
GET /health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2023-01-01T00:00:00Z",
  "service": "auth-service",
  "version": "1.0.0"
}
```

## Permissions

The following permissions are available for API keys:

- `read:accounts` - Read account information
- `write:accounts` - Modify account information
- `read:keys` - List API keys
- `write:keys` - Create/revoke API keys
- `manage:webhooks` - Manage webhook URLs

## Configuration

The service is configured via environment variables:

| Variable | Default | Description |
|-----------|----------|-------------|
| `PORT` | 8080 | HTTP server port |
| `AWS_REGION` | us-west-2 | AWS region for DynamoDB |
| `DYNAMODB_TABLE` | auth-service | DynamoDB table name |

## Deployment

### Docker

Build the Docker image:
```bash
docker build -t auth-service .
```

Run the container:
```bash
docker run -p 8080:8080 \
  -e PORT=8080 \
  -e AWS_REGION=us-west-2 \
  -e DYNAMODB_TABLE=auth-service \
  auth-service
```

### Local Development

Prerequisites:
- Go 1.23+
- AWS CLI configured with appropriate permissions
- DynamoDB table created with required schema

Run the service:
```bash
cd cmd/auth-svc
go run main.go
```

## Testing

Run unit tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

Generate coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Architecture

The service follows Clean Architecture principles:

- **Domain Layer**: Core business entities and rules
- **Use Case Layer**: Application business logic
- **Repository Layer**: Data access abstraction
- **Adapter Layer**: HTTP handlers and external integrations

## Security

- API keys are hashed using bcrypt before storage
- All authentication events are logged for audit purposes
- Permissions are enforced at the middleware level
- API keys have configurable expiration times

## Monitoring

The service provides:
- Structured JSON logging
- Health check endpoint
- Audit logging for security events
- Request/response logging

## Troubleshooting

### Common Issues

1. **DynamoDB Connection Failed**
   - Verify AWS credentials are configured
   - Check that the table exists in the specified region
   - Ensure IAM permissions are correct

2. **API Key Validation Fails**
   - Check that the key hasn't expired
   - Verify the key status is 'active'
   - Ensure the key hash is correctly stored

3. **Permission Denied**
   - Verify the API key has the required permission
   - Check that the account is in 'active' status

### Logs

Check the application logs for detailed error information:
```bash
docker logs <container-id>
```

For local development, logs are printed to stdout with structured JSON format.