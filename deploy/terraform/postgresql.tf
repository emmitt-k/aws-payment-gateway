# PostgreSQL Configuration for Account Storage

# RDS PostgreSQL instance for account data
resource "aws_db_instance" "accounts" {
  identifier = "${var.environment}-payment-gateway-accounts"

  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.t3.micro"

  allocated_storage     = 20
  max_allocated_storage = 100
  storage_type         = "gp2"
  storage_encrypted    = true

  db_name  = "payment_gateway"
  username = var.postgres_username
  password = var.postgres_password

  port = 5432

  vpc_security_group_ids = [aws_security_group.rds.id]
  db_subnet_group_name   = aws_db_subnet_group.main.name

  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"

  skip_final_snapshot       = true
  final_snapshot_identifier = "${var.environment}-payment-gateway-accounts-final"

  tags = merge(var.tags, {
    Name = "${var.environment}-payment-gateway-accounts"
  })

  depends_on = [aws_db_subnet_group.main]
}

# DB subnet group for RDS
resource "aws_db_subnet_group" "main" {
  name       = "${var.environment}-payment-gateway-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = merge(var.tags, {
    Name = "${var.environment}-payment-gateway-db-subnet-group"
  })
}

# Security group for RDS
resource "aws_security_group" "rds" {
  name_prefix = "${var.environment}-payment-gateway-rds-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.environment}-payment-gateway-rds"
  })
}

# VPC for the application
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(var.tags, {
    Name = "${var.environment}-payment-gateway-vpc"
  })
}

# Private subnets for RDS
resource "aws_subnet" "private" {
  count = 2

  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = merge(var.tags, {
    Name = "${var.environment}-payment-gateway-private-${count.index + 1}"
  })
}

# Internet gateway for public subnets
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.environment}-payment-gateway-igw"
  })
}

# Route table for private subnets
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.environment}-payment-gateway-private-rt"
  })
}

# Route table associations for private subnets
resource "aws_route_table_association" "private" {
  count          = 2
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}

# Data source for availability zones
data "aws_availability_zones" "available" {
  state = "available"
}

# Outputs for PostgreSQL
output "postgresql" {
  description = "PostgreSQL instance configuration"
  value = {
    endpoint = aws_db_instance.accounts.endpoint
    port     = aws_db_instance.accounts.port
    database = aws_db_instance.accounts.db_name
    username = aws_db_instance.accounts.username
  }
}