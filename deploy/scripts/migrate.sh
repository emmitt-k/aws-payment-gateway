#!/bin/bash

# Database Migration Script for Auth Service
# This script runs database migrations using golang-migrate

set -e

# Default values
MIGRATIONS_PATH="${MIGRATIONS_PATH:-./migrations}"
DATABASE_URL="${DATABASE_URL:-postgres://postgres:password@localhost:5432/payment_gateway?sslmode=disable}"
STEPS="${STEPS:-up}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting database migration...${NC}"

# Check if migrate tool is installed
if ! command -v migrate &> /dev/null; then
    echo -e "${RED}Error: migrate tool not found. Please install it with:${NC}"
    echo "go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Check if migrations directory exists
if [ ! -d "$MIGRATIONS_PATH" ]; then
    echo -e "${RED}Error: Migrations directory not found: $MIGRATIONS_PATH${NC}"
    exit 1
fi

# Test database connection
echo -e "${YELLOW}Testing database connection...${NC}"
if ! migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version 1>/dev/null 2>&1; then
    echo -e "${RED}Error: Cannot connect to database: $DATABASE_URL${NC}"
    exit 1
fi

echo -e "${GREEN}Database connection successful.${NC}"

# Get current database version
CURRENT_VERSION=$(migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version 2>/dev/null || echo "0")
echo -e "${YELLOW}Current database version: $CURRENT_VERSION${NC}"

# Run migrations based on steps
case "$STEPS" in
    "up")
        echo -e "${YELLOW}Running all pending migrations...${NC}"
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
        ;;
    "down")
        echo -e "${YELLOW}Rolling back last migration...${NC}"
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" down 1
        ;;
    "force")
        if [ -z "$VERSION" ]; then
            echo -e "${RED}Error: VERSION must be specified when using force${NC}"
            exit 1
        fi
        echo -e "${YELLOW}Forcing database to version $VERSION...${NC}"
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" force "$VERSION"
        ;;
    "version")
        echo -e "${YELLOW}Getting current database version...${NC}"
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version
        ;;
    *)
        echo -e "${RED}Error: Invalid STEPS value. Use 'up', 'down', 'force', or 'version'${NC}"
        exit 1
        ;;
esac

if [ "$STEPS" = "up" ]; then
    NEW_VERSION=$(migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version 2>/dev/null || echo "0")
    echo -e "${GREEN}Migration completed successfully. New version: $NEW_VERSION${NC}"
elif [ "$STEPS" = "down" ]; then
    NEW_VERSION=$(migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version 2>/dev/null || echo "0")
    echo -e "${GREEN}Rollback completed successfully. New version: $NEW_VERSION${NC}"
fi

echo -e "${GREEN}Database migration process completed.${NC}"