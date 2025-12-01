-- +migrate Up
CREATE TABLE balances (
    ledger_account_id UUID NOT NULL REFERENCES ledger_accounts(id) ON DELETE CASCADE,
    currency TEXT NOT NULL,
    amount_minor BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (ledger_account_id, currency)
);

-- Create index for updated_at queries
CREATE INDEX idx_balances_updated_at ON balances(updated_at);

-- Add check constraint to prevent negative balances for certain account types
-- This will be enforced at application level as rules vary by account type