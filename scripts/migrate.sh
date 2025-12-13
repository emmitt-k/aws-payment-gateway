#!/bin/bash

# Simple migration script for AWS RDS PostgreSQL
# Usage: ./scripts/migrate.sh [dev|prod] [up|down|status|force VERSION]

set -e

# Check if environment is provided
if [ -z "$1" ]; then
    echo "Usage: $0 [dev|prod] [up|down|status|force VERSION]"
    echo "Examples:"
    echo "  $0 dev up      # Apply pending migrations to dev"
    echo "  $0 prod up     # Apply pending migrations to prod"
    echo "  $0 dev status  # Check migration status for dev"
    echo "  $0 dev down    # Rollback last migration in dev"
    echo "  $0 dev force 5 # Set dev database to version 5"
    exit 1
fi

ENVIRONMENT=$1
COMMAND=$2
FORCE_VERSION=$3

# Set environment-specific variables
if [ "$ENVIRONMENT" = "dev" ]; then
    TFVARS_FILE="terraform/dev.tfvars"
    REGION="ap-southeast-1"
elif [ "$ENVIRONMENT" = "prod" ]; then
    TFVARS_FILE="terraform/prod.tfvars"
    REGION="ap-southeast-1"
else
    echo "Error: Environment must be 'dev' or 'prod'"
    exit 1
fi

# Check if terraform tfvars file exists
if [ ! -f "$TFVARS_FILE" ]; then
    echo "Error: Terraform tfvars file not found: $TFVARS_FILE"
    exit 1
fi

# Get database connection info from Terraform outputs
echo "Getting database connection info from Terraform..."
cd terraform

# Check if Terraform is initialized
if [ ! -d ".terraform" ]; then
    echo "Initializing Terraform..."
    terraform init
fi

# Get database outputs
DB_HOST=$(terraform output -raw db_host 2>/dev/null || echo "")
DB_PORT=$(terraform output -raw db_port 2>/dev/null || echo "5432")
DB_NAME=$(terraform output -raw db_name 2>/dev/null || echo "payment_gateway")
DB_USER=$(terraform output -raw db_user 2>/dev/null || echo "postgres")

# If outputs don't exist, try to get them from state or use defaults
if [ -z "$DB_HOST" ]; then
    echo "Database outputs not found. Using default values..."
    # Extract values from tfvars file as fallback
    DB_HOST=$(grep -E "postgres_password|aws_region" "$TFVARS_FILE" | head -1 | cut -d'=' -f2 | tr -d ' "')
    if [ -z "$DB_HOST" ]; then
        echo "Error: Could not determine database host. Please run 'terraform apply' first."
        exit 1
    fi
fi

# Get password from tfvars file
DB_PASSWORD=$(grep postgres_password "$TFVARS_FILE" | cut -d'=' -f2 | tr -d ' "')

if [ -z "$DB_PASSWORD" ]; then
    echo "Error: Could not find postgres_password in $TFVARS_FILE"
    exit 1
fi

cd ..

# Construct database URL
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require"

echo "Environment: $ENVIRONMENT"
echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo "Command: $COMMAND"

# Check if migrate tool is installed
if ! command -v migrate &> /dev/null; then
    echo "Error: migrate tool is not installed."
    echo "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Run migration command
case $COMMAND in
    "up")
        echo "Applying pending migrations..."
        migrate -path migrations -database "$DATABASE_URL" up
        echo "✅ Migrations applied successfully!"
        ;;
    "down")
        echo "Rolling back last migration..."
        migrate -path migrations -database "$DATABASE_URL" down 1
        echo "✅ Migration rolled back successfully!"
        ;;
    "status")
        echo "Checking migration status..."
        migrate -path migrations -database "$DATABASE_URL" version
        ;;
    "force")
        if [ -z "$FORCE_VERSION" ]; then
            echo "Error: force command requires a version number"
            echo "Usage: $0 $ENVIRONMENT force VERSION"
            exit 1
        fi
        echo "Setting database to version $FORCE_VERSION..."
        migrate -path migrations -database "$DATABASE_URL" force "$FORCE_VERSION"
        echo "✅ Database version set to $FORCE_VERSION"
        ;;
    *)
        echo "Error: Unknown command '$COMMAND'"
        echo "Available commands: up, down, status, force"
        exit 1
        ;;
esac