# Change: Add Database Migration Files

## Why
We need to establish the database schema foundation for the Tron USDT Custodial Payment Gateway. The migration files will define all tables required for the double-entry bookkeeping system, user management, and transaction processing.

## What Changes
- Create SQL migration files for all database tables as defined in README.md
- Set up proper table relationships and constraints
- Include indexes for performance optimization
- Add initial data seeding where required

## Impact
- Affected specs: database (new capability)
- Affected code: migrations/ directory
- Dependencies: PostgreSQL database setup required before running migrations