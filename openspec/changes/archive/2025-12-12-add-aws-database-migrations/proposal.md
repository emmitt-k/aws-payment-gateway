# Change: Add Simple Database Migrations for AWS Deployments

## Why
The current AWS deployment process requires manual database schema migration using psql commands. This is error-prone and creates operational overhead during deployments, especially for beginners.

## What Changes
- Add simple migration script to execute schema migrations on AWS RDS
- Add simple Makefile commands for dev and prod environments
- Update deploy commands to auto-migrate before deploying services
- Add migration status commands to check current state

## Impact
- Affected specs: database
- Affected code: Makefile, terraform/main.tf
- Dependencies: AWS RDS PostgreSQL, golang-migrate tool

## Simple Usage for Beginners
```bash
# Migrate databases
make migrate-dev      # Migrate dev database
make migrate-prod     # Migrate prod database

# Deploy (auto-migrates first)
make deploy-dev       # Deploy to dev (includes migration)
make deploy-prod      # Deploy to prod (includes migration)