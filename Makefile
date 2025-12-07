.PHONY: help up down build test clean check-prereqs check-aws-prereqs plan-dev plan-prod deploy-dev deploy-prod destroy-dev destroy-prod aws-status aws-logs-dev aws-logs-prod logs shell init-terraform

help:
	@echo "Payment Gateway - Simple Commands"
	@echo ""
	@echo "Local Development:"
	@echo "  up              - Start all services locally"
	@echo "  down            - Stop all services"
	@echo "  logs            - Show service logs"
	@echo "  shell           - Open shell in auth service"
	@echo ""
	@echo "Testing:"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo ""
	@echo "AWS Deployment:"
	@echo "  plan-dev        - Plan dev deployment"
	@echo "  plan-prod       - Plan production deployment"
	@echo "  deploy-dev      - Deploy to dev environment"
	@echo "  deploy-prod     - Deploy to production"
	@echo "  destroy-dev     - Destroy dev environment"
	@echo "  destroy-prod    - Destroy production environment"
	@echo "  aws-status      - Check AWS deployment status"
	@echo "  aws-logs-dev    - Show dev logs"
	@echo "  aws-logs-prod   - Show production logs"
	@echo ""
	@echo "Prerequisites:"
	@echo "  check-prereqs   - Check local development prerequisites"
	@echo "  check-aws-prereqs - Check AWS deployment prerequisites"
	@echo ""
	@echo "Build:"
	@echo "  build           - Build all services"
	@echo "  clean           - Clean build artifacts"

# Local Development
up:
	@echo "Starting services..."
	docker-compose up -d postgres dynamodb auth
	@echo "Services running:"
	@echo "  Auth: http://localhost:8080"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  DynamoDB: http://localhost:8000"

up-all:
	@echo "Starting all services including microservices..."
	docker-compose --profile microservices up -d
	@echo "All services running:"
	@echo "  Auth: http://localhost:8080"
	@echo "  Payin: http://localhost:8081"
	@echo "  Observer: http://localhost:8082"
	@echo "  Ledger: http://localhost:8083"
	@echo "  Payout: http://localhost:8084"

up-admin:
	@echo "Starting services with admin tools..."
	docker-compose --profile admin up -d
	@echo "Admin tools available:"
	@echo "  DynamoDB Admin: http://localhost:8001"

down:
	@echo "Stopping all services..."
	docker-compose down -v --remove-orphans

logs:
	docker-compose logs -f

shell:
	docker exec -it auth /bin/sh

# Testing
test:
	@echo "Running all tests..."
	go test -v ./...

test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./internal/auth/tests/integration/...

# Build
build:
	@echo "Building all services..."
	go build -o bin/auth ./cmd/auth-svc
	go build -o bin/payin ./cmd/payin-svc
	go build -o bin/observer ./cmd/observer-svc
	go build -o bin/ledger ./cmd/ledger-svc
	go build -o bin/payout ./cmd/payout-svc

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage-*.out
	rm -f coverage.html

# Check prerequisites
check-prereqs:
	@echo "Checking prerequisites..."
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed."; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed."; exit 1; }
	@echo "✅ Local prerequisites OK"

check-aws-prereqs:
	@echo "Checking AWS prerequisites..."
	@command -v aws >/dev/null 2>&1 || { echo "❌ AWS CLI is required but not installed."; exit 1; }
	@command -v terraform >/dev/null 2>&1 || { echo "❌ Terraform is required but not installed."; exit 1; }
	@echo "✅ AWS prerequisites OK"

# AWS Deployment
init-terraform: check-aws-prereqs
	@echo "Initializing Terraform..."
	cd terraform && terraform init

plan-dev:
	@echo "Planning dev deployment..."
	cd terraform && terraform plan -var-file="dev.tfvars"

plan-prod:
	@echo "Planning production deployment..."
	cd terraform && terraform plan -var-file="prod.tfvars"

deploy-dev: init-terraform
	@echo "Deploying to dev environment..."
	cd terraform && terraform apply -var-file="dev.tfvars" -auto-approve
	@echo "✓ Dev deployment complete!"

deploy-prod: init-terraform
	@echo "Deploying to production environment..."
	cd terraform && terraform apply -var-file="prod.tfvars" -auto-approve
	@echo "✓ Production deployment complete!"

destroy-dev:
	@echo "Destroying dev environment..."
	@echo "This will destroy all dev resources. Are you sure? [y/N]"
	@read -r confirm && [ "$$confirm" = "y" ] || exit 1
	cd terraform && terraform destroy -var-file="dev.tfvars" -auto-approve
	@echo "✓ Dev infrastructure destroyed!"

destroy-prod:
	@echo "Destroying production environment..."
	@echo "⚠️  WARNING: This will destroy PRODUCTION! ⚠️"
	@echo "Type 'destroy-prod' to confirm:"
	@read -r confirm && [ "$$confirm" = "destroy-prod" ] || exit 1
	cd terraform && terraform destroy -var-file="prod.tfvars" -auto-approve
	@echo "✓ Production infrastructure destroyed!"

# AWS Utilities
aws-logs-dev:
	@echo "Showing dev logs..."
	cd terraform && terraform output cluster_name | xargs -I {} aws logs tail /ecs/{}-auth-service --follow --region us-west-2

aws-logs-prod:
	@echo "Showing production logs..."
	cd terraform && terraform output cluster_name | xargs -I {} aws logs tail /ecs/{}-auth-service --follow --region us-west-2

aws-status:
	@echo "Checking AWS deployment status..."
	cd terraform && terraform output