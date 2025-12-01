-- +migrate Up
CREATE TABLE ledger_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('asset', 'liability', 'equity', 'revenue', 'expense')),
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'archived')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create unique constraint for account_id + code combination
ALTER TABLE ledger_accounts ADD CONSTRAINT uq_ledger_accounts_account_code UNIQUE (account_id, code);

-- Create indexes for performance
CREATE INDEX idx_ledger_accounts_account_id ON ledger_accounts(account_id);
CREATE INDEX idx_ledger_accounts_type ON ledger_accounts(type);
CREATE INDEX idx_ledger_accounts_status ON ledger_accounts(status);