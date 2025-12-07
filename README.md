# Tron USDT Payment Gateway

A simple, unified payment gateway for Tron USDT (TRC20) built on AWS.

## 🚀 Quick Start

### Local Development
```bash
# Start core services
make up

# Start all microservices
make up-all

# Start with admin tools
make up-admin

# Stop everything
make down
```

### AWS Deployment
```bash
# Deploy to dev
make deploy-dev

# Deploy to production
make deploy-prod
```

## 📋 Commands

| Command | What it does |
|---------|-------------|
| `make up` | Start core services (postgres, dynamodb, auth) |
| `make up-all` | Start all microservices |
| `make up-admin` | Include admin tools |
| `make down` | Stop all services |
| `make logs` | Show service logs |
| `make shell` | Open shell in auth service |
| `make test` | Run all tests |
| `make build` | Build all services |
| `make deploy-dev` | Deploy to AWS dev |
| `make deploy-prod` | Deploy to AWS production |
| `make openspec-list` | List active changes and specs |
| `make openspec-validate` | Validate all specs and changes |

## 🏗️ Architecture

### Services
- **Auth Service** (port 8080): API key management
- **Payin Service** (port 8081): Deposit addresses
- **Observer Service** (port 8082): Blockchain scanning
- **Ledger Service** (port 8083): Double-entry accounting
- **Payout Service** (port 8084): Withdrawal processing

### Data Stores
- **PostgreSQL**: Account data, transactions, balances
- **DynamoDB**: API keys, audit logs, webhooks

### AWS Infrastructure
- **ECS Fargate**: Serverless containers
- **RDS**: Managed PostgreSQL
- **Application Load Balancer**: Traffic distribution
- **CloudWatch**: Logging and monitoring

## 🔧 Development

### Local URLs
| Service | URL |
|---------|-----|
| Auth | http://localhost:8080 |
| Payin | http://localhost:8081 |
| Observer | http://localhost:8082 |
| Ledger | http://localhost:8083 |
| Payout | http://localhost:8084 |
| DynamoDB Admin | http://localhost:8001 |
| PostgreSQL | localhost:5432 |
| DynamoDB | http://localhost:8000 |

### Testing
```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration
```

## ☁️ AWS Deployment

### Prerequisites
- AWS CLI configured
- Terraform installed
- Docker installed

### Environment Files
- `terraform/dev.tfvars` - Dev infrastructure
- `terraform/prod.tfvars` - Production infrastructure

## 🛠️ Project Structure

```
.
├── cmd/                    # Service entry points
│   ├── auth-svc/
│   ├── payin-svc/
│   ├── observer-svc/
│   ├── ledger-svc/
│   └── payout-svc/
├── internal/               # Business logic
│   ├── auth/              # Authentication service
│   ├── payin/             # Payin service
│   ├── observer/          # Observer service
│   ├── ledger/            # Ledger service
│   └── common/            # Shared utilities
├── migrations/             # Database migrations
├── openspec/              # Specifications and change proposals
│   ├── specs/             # Current specifications
│   ├── changes/           # Change proposals
│   │   └── archive/       # Archived changes
│   └── AGENTS.md          # OpenSpec instructions
├── terraform/             # AWS infrastructure
│   ├── main.tf
│   ├── dev.tfvars
│   └── prod.tfvars
├── docker-compose.yml      # Local development
├── Makefile               # Build commands
└── README.md              # This file
```

## 🔒 Security

- **API Keys**: Secure authentication
- **Rate Limiting**: Prevent abuse
- **Audit Logging**: Track all actions
- **HTTPS**: SSL in production
- **Secrets Manager**: AWS secure storage

## 📊 Monitoring

- **Health Checks**: `/health` endpoint
- **CloudWatch**: Metrics and logs
- **Auto Scaling**: CPU/memory based
- **Alarms**: Email notifications

## 🐛 Troubleshooting

### Common Issues

```bash
# Port conflicts
lsof -i :8080

# Service logs
docker logs auth

# Database issues
docker logs postgres

# AWS issues
cd terraform && terraform show
```

### Debug Mode
```bash
# Verbose logs
ENVIRONMENT=local LOG_LEVEL=debug make up

# Test with output
go test -v ./...
```

## 📚 Documentation

- **Architecture**: Clean Architecture pattern
- **Database**: PostgreSQL + DynamoDB hybrid
- **Security**: API key authentication
- **Scaling**: Auto-scaling policies
- **Specifications**: See `openspec/specs/` for detailed requirements
- **Change Management**: See `openspec/AGENTS.md` for development workflow

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Deploy to dev: `make deploy-dev`
6. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.
