# Design: Simple Database Migrations for AWS

## Approach
Use the existing golang-migrate tool with a simple script that connects to AWS RDS and applies pending migrations.

## Simple Makefile Commands
The goal is to make it easy for beginners with these commands:

```bash
# Migrate databases
make migrate-dev      # Migrate dev database
make migrate-prod     # Migrate prod database

# Deploy (auto-migrates first)
make deploy-dev       # Deploy to dev (includes migration)
make deploy-prod      # Deploy to prod (includes migration)

# Check migration status
make migrate-status-dev   # Check dev migration status
make migrate-status-prod  # Check prod migration status
```

## Implementation Details
- Migration script: `scripts/migrate.sh`
- Script reads environment (dev/prod) from command line
- Uses existing migration files in `migrations/` directory
- Database connection via AWS credentials and Terraform outputs
- Deploy commands automatically run migrations before deploying services

## Security
- Use existing AWS credentials for database access
- No additional IAM roles needed
- Simple connection string with password from tfvars