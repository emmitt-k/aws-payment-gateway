# Database Migrations

This directory contains SQL migration files for the Tron USDT Custodial Payment Gateway database schema.

## Migration Tool

We use `golang-migrate` for database migrations. Install it with:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Running Migrations

### Up Migrations
To apply all pending migrations:
```bash
migrate -path migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" up
```

### Down Migrations
To rollback the last migration:
```bash
migrate -path migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" down 1
```

### Force Version
To set the database to a specific version:
```bash
migrate -path migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" force 000005
```

## Migration Files

Each migration consists of two files:
- `XXXXXX_description.up.sql` - Applied when migrating up
- `XXXXXX_description.down.sql` - Applied when migrating down

Files are executed in numerical order.

## Schema Overview

The migrations create the following tables:

1. **accounts** - Company account registry
2. **ledger_accounts** - Double-entry accounts scoped by account_id
3. **journal_entries** - Immutable business events
4. **postings** - Line items for journal entries (double-entry)
5. **balances** - Materialized balances for fast reads
6. **payin_addresses** - Payin addresses and KMS bindings
7. **payin_requests** - Provisioned payin intents
8. **observed_transactions** - Raw chain transactions mapped to addresses
9. **payout_requests** - Outbound transfer lifecycle
10. **system_wallets** - Internal wallet configuration
11. **chain_cursors** - Blockchain scanning checkpoints

## Important Notes

- All amounts are stored as `BIGINT` in minor units (e.g., for USDT, 1 USDT = 1,000,000 minor units)
- All tables use UUID primary keys with `gen_random_uuid()` as default
- Foreign key constraints ensure referential integrity
- Check constraints enforce valid enum values
- Indexes are created for performance optimization
- Migration files use `+migrate` comments for proper tool recognition

## Creating New Migrations

To create a new migration:
```bash
migrate create -ext sql -dir migrations -seq add_new_feature
```

This will create both up and down migration files with the next sequence number.