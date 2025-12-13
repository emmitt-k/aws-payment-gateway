# Database Specification

## Overview
PostgreSQL database for payment gateway with simple automated schema migrations.

## Schema Management
- Migration files stored in `migrations/` directory
- Use golang-migrate tool for applying migrations
- Migration tracking via schema_migrations table
- Sequential migration execution with version control

## Migration Requirements
- All migrations must be reversible (up/down files)
- Migrations must be idempotent
- No destructive changes to existing data
- Backward compatibility maintained during deployment

## ADDED Requirements

### Requirement: Simple Makefile Migration Commands
The system SHALL provide simple Makefile commands for database migrations in dev and prod environments.

#### Scenario:
- `make migrate-dev` SHALL apply pending migrations to dev database
- `make migrate-prod` SHALL apply pending migrations to prod database
- `make migrate-status-dev` SHALL show current migration status for dev
- `make migrate-status-prod` SHALL show current migration status for prod

### Requirement: Auto-Migrate on Deploy
The deploy commands SHALL automatically run migrations before deploying services.

#### Scenario:
- `make deploy-dev` SHALL automatically migrate dev database before deploying
- `make deploy-prod` SHALL automatically migrate prod database before deploying
- Migration failures SHALL prevent service deployment
- Clear migration status SHALL be shown during deployment

### Requirement: Simple Migration Script
The migration script SHALL be simple and easy for beginners to understand.

#### Scenario:
- Script SHALL read environment (dev/prod) from command line
- Script SHALL use existing golang-migrate tool
- Script SHALL provide clear success/failure messages
- Script SHALL handle database connection automatically