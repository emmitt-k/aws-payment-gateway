## 1. Remove File Logger Implementation
- [x] 1.1 Delete internal/auth/audit/logger.go file
- [x] 1.2 Remove AuditLogger struct and all its methods
- [x] 1.3 Remove NewAuditLogger function

## 2. Update DynamoDB Logger
- [x] 2.1 Remove file logger backup from DynamoDBAuditLogger struct
- [x] 2.2 Update NewDynamoDBAuditLogger to remove file logger initialization
- [x] 2.3 Remove all file logger calls from DynamoDB audit methods
- [x] 2.4 Remove error logging fallbacks that reference file logger
- [x] 2.5 Update error handling to use standard logging only

## 3. Update Main Application
- [x] 3.1 Remove any references to file logger in cmd/auth-svc/main.go
- [x] 3.2 Ensure only DynamoDBAuditLogger is instantiated and used

## 4. Update Interfaces and Types
- [x] 4.1 Move AuditEvent struct from logger.go to dynamodb_logger.go
- [x] 4.2 Move GetEventDescription function from logger.go to dynamodb_logger.go
- [x] 4.3 Ensure AuditLoggerInterface remains intact in dynamodb_logger.go

## 5. Update Imports
- [x] 5.1 Remove log package import from dynamodb_logger.go where no longer needed
- [x] 5.2 Update any other import statements that reference the removed file logger
