# AWS Provider Configuration
provider "aws" {
  region = var.aws_region
  
  default_tags {
    Name        = "${var.environment}-payment-gateway"
    Environment = var.environment
    Project     = var.tags["Project"]
    ManagedBy   = var.tags["ManagedBy"]
  }
}

# Terraform Backend Configuration
terraform {
  backend "s3" {
    bucket = "tron-payment-gateway-terraform-state"
    key    = "dynamodb/${var.environment}/terraform.tfstate"
    region = var.aws_region
  }
}

# Outputs
output "dynamodb_tables" {
  description = "DynamoDB table names and ARNs"
  value = {
    api_keys = {
      name = aws_dynamodb_table.api_keys.name
      arn  = aws_dynamodb_table.api_keys.arn
    }
    idempotency_keys = {
      name = aws_dynamodb_table.idempotency_keys.name
      arn  = aws_dynamodb_table.idempotency_keys.arn
    }
    webhook_events = {
      name = aws_dynamodb_table.webhook_events.name
      arn  = aws_dynamodb_table.webhook_events.arn
    }
    audit_logs = {
      name = aws_dynamodb_table.audit_logs.name
      arn  = aws_dynamodb_table.audit_logs.arn
    }
  }
}