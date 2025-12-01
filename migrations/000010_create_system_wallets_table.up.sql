-- +migrate Up
CREATE TABLE system_wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role TEXT NOT NULL CHECK (role IN ('hot', 'cold', 'gas')),
    address TEXT NOT NULL,
    kms_key_id TEXT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create unique constraint for wallet addresses
ALTER TABLE system_wallets ADD CONSTRAINT uq_system_wallets_address UNIQUE (address);

-- Create indexes for performance
CREATE INDEX idx_system_wallets_role ON system_wallets(role);
CREATE INDEX idx_system_wallets_status ON system_wallets(status);
CREATE INDEX idx_system_wallets_kms_key_id ON system_wallets(kms_key_id) WHERE kms_key_id IS NOT NULL;