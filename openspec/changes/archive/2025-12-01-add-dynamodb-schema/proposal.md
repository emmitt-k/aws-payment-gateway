# Change: Add DynamoDB Schema Definition

## Why
We need to define the DynamoDB table schemas for high-volume operational data as specified in README.md. These tables will handle audit logs, webhook events, API keys, and idempotency with TTL support for cost-effective storage.

## What Changes
- Create Terraform configuration for DynamoDB tables
- Define table schemas with proper partition keys, sort keys, and GSIs
- Configure TTL settings for automatic data expiration
- Set up appropriate capacity modes and billing
- Add table-specific attributes and indexes

## Impact
- Affected specs: dynamodb (new capability)
- Affected code: deploy/terraform/dynamodb.tf
- Dependencies: AWS infrastructure setup required before table creation