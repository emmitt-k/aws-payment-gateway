# Tasks: Add Automated Database Migrations for AWS Deployments

## 1. Create Migration Script
- [x] Create simple migration script that uses golang-migrate tool
- [x] Script reads database connection from environment variables
- [x] Script handles dev and prod environments automatically

## 2. Update Makefile (Simple Commands)
- [x] Add `migrate-dev` command to migrate dev database
- [x] Add `migrate-prod` command to migrate prod database
- [x] Update `deploy-dev` to auto-migrate before deploying
- [x] Update `deploy-prod` to auto-migrate before deploying
- [x] Add `migrate-status-dev` and `migrate-status-prod` commands

## 3. Update Terraform
- [x] Add database connection outputs to main.tf
- [x] Ensure security groups allow migration access

## 4. Simple Documentation
- [x] Add simple migration section to README
- [x] Show example commands for dev and prod