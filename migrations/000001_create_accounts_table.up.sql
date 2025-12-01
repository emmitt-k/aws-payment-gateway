-- +migrate Up
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'suspended', 'deleted')),
    webhook_url TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for status queries
CREATE INDEX idx_accounts_status ON accounts(status);

-- Create index for name searches
CREATE INDEX idx_accounts_name ON accounts(name);