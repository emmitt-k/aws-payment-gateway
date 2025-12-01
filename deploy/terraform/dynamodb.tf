# DynamoDB Tables for Tron USDT Custodial Payment Gateway
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# API Keys Table (Auth Service)
resource "aws_dynamodb_table" "api_keys" {
  name           = "api_keys"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "api_key_hash"

  attribute {
    name = "api_key_hash"
    type = "S"
  }

  attribute {
    name = "account_id"
    type = "S"
  }

  attribute {
    name = "name"
    type = "S"
  }

  attribute {
    name = "permissions"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  attribute {
    name = "last_used_at"
    type = "S"
  }

  attribute {
    name = "expires_at"
    type = "N"
  }

  attribute {
    name = "ttl"
    type = "N"
  }

  global_secondary_index {
    name     = "gsi_account_id"
    hash_key = "account_id"
    projection_type = "ALL"
  }

  ttl {
    attribute_name = "ttl"
    enabled = true
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Service     = "payment-gateway"
    Environment = var.environment
    Table       = "api_keys"
  }
}

# Idempotency Keys Table (All Services)
resource "aws_dynamodb_table" "idempotency_keys" {
  name           = "idempotency_keys"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "key"

  attribute {
    name = "key"
    type = "S"
  }

  attribute {
    name = "account_id"
    type = "S"
  }

  attribute {
    name = "scope"
    type = "S"
  }

  attribute {
    name = "request_hash"
    type = "S"
  }

  attribute {
    name = "response"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  attribute {
    name = "created_at"
    type = "S"
  }

  attribute {
    name = "ttl"
    type = "N"
  }

  global_secondary_index {
    name     = "gsi_account_id"
    hash_key = "account_id"
    projection_type = "ALL"
  }

  ttl {
    attribute_name = "ttl"
    enabled = true
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Service     = "payment-gateway"
    Environment = var.environment
    Table       = "idempotency_keys"
  }
}

# Webhook Events Table (Notification Service)
resource "aws_dynamodb_table" "webhook_events" {
  name           = "webhook_events"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "account_id"
  range_key      = "event_id"

  attribute {
    name = "account_id"
    type = "S"
  }

  attribute {
    name = "event_id"
    type = "S"
  }

  attribute {
    name = "event_type"
    type = "S"
  }

  attribute {
    name = "payload"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  attribute {
    name = "attempts"
    type = "N"
  }

  attribute {
    name = "last_attempt_at"
    type = "S"
  }

  attribute {
    name = "last_error"
    type = "S"
  }

  attribute {
    name = "next_retry_at"
    type = "S"
  }

  attribute {
    name = "ttl"
    type = "N"
  }

  global_secondary_index {
    name     = "gsi_status_next_retry"
    hash_key = "status"
    range_key  = "next_retry_at"
    projection_type = "ALL"
  }

  ttl {
    attribute_name = "ttl"
    enabled = true
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Service     = "payment-gateway"
    Environment = var.environment
    Table       = "webhook_events"
  }
}

# Audit Logs Table (Compliance)
resource "aws_dynamodb_table" "audit_logs" {
  name           = "audit_logs"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "account_id"
  range_key      = "timestamp"

  attribute {
    name = "account_id"
    type = "S"
  }

  attribute {
    name = "timestamp"
    type = "S"
  }

  attribute {
    name = "action"
    type = "S"
  }

  attribute {
    name = "actor_id"
    type = "S"
  }

  attribute {
    name = "actor_type"
    type = "S"
  }

  attribute {
    name = "data"
    type = "S"
  }

  attribute {
    name = "ip_address"
    type = "S"
  }

  attribute {
    name = "ttl"
    type = "N"
  }

  ttl {
    attribute_name = "ttl"
    enabled = true
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Service     = "payment-gateway"
    Environment = var.environment
    Table       = "audit_logs"
  }
}