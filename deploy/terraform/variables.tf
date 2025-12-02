# Terraform Variables for DynamoDB Tables

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-1"
}

variable "tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default = {
    Project     = "tron-usdt-payment-gateway"
    ManagedBy   = "terraform"
  }
}

variable "dynamodb_tables" {
  description = "Configuration for DynamoDB tables"
  type = object({
    api_keys = object({
      name           = string
      billing_mode   = string
      hash_key       = string
      gsis = list(object({
        name           = string
        hash_key       = string
        projection_type = string
      }))
      ttl_enabled = bool
      point_in_time_recovery = bool
    })
    idempotency_keys = object({
      name           = string
      billing_mode   = string
      hash_key       = string
      gsis = list(object({
        name           = string
        hash_key       = string
        projection_type = string
      }))
      ttl_enabled = bool
      point_in_time_recovery = bool
    })
    webhook_events = object({
      name           = string
      billing_mode   = string
      hash_key       = string
      range_key      = string
      gsis = list(object({
        name           = string
        hash_key       = string
        range_key      = string
        projection_type = string
      }))
      ttl_enabled = bool
      point_in_time_recovery = bool
    })
    audit_logs = object({
      name           = string
      billing_mode   = string
      hash_key       = string
      range_key      = string
      ttl_enabled = bool
      point_in_time_recovery = bool
    })
  })
  default = {
    api_keys = {
      name           = "api_keys"
      billing_mode   = "PAY_PER_REQUEST"
      hash_key       = "api_key_hash"
      gsis = [
        {
          name           = "gsi_account_id"
          hash_key       = "account_id"
          projection_type = "ALL"
        }
      ]
      ttl_enabled = true
      point_in_time_recovery = true
    }
    idempotency_keys = {
      name           = "idempotency_keys"
      billing_mode   = "PAY_PER_REQUEST"
      hash_key       = "key"
      gsis = [
        {
          name           = "gsi_account_id"
          hash_key       = "account_id"
          projection_type = "ALL"
        }
      ]
      ttl_enabled = true
      point_in_time_recovery = true
    }
    webhook_events = {
      name           = "webhook_events"
      billing_mode   = "PAY_PER_REQUEST"
      hash_key       = "account_id"
      range_key      = "event_id"
      gsis = [
        {
          name           = "gsi_status_next_retry"
          hash_key       = "status"
          range_key      = "next_retry_at"
          projection_type = "ALL"
        }
      ]
      ttl_enabled = true
      point_in_time_recovery = true
    }
    audit_logs = {
      name           = "audit_logs"
      billing_mode   = "PAY_PER_REQUEST"
      hash_key       = "account_id"
      range_key      = "timestamp"
      ttl_enabled = true
      point_in_time_recovery = true
    }
  }
}

variable "postgres_username" {
  description = "PostgreSQL master username"
  type        = string
  default     = "postgres"
  sensitive   = true
}

variable "postgres_password" {
  description = "PostgreSQL master password"
  type        = string
  sensitive   = true
}