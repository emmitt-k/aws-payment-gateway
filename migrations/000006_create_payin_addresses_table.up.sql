-- +migrate Up
CREATE TABLE payin_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    ledger_account_id UUID NULL REFERENCES ledger_accounts(id) ON DELETE SET NULL,
    tron_address_base58 TEXT NOT NULL,
    tron_address_hex TEXT NOT NULL,
    kms_key_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'disabled')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create unique constraint for Tron addresses
ALTER TABLE payin_addresses ADD CONSTRAINT uq_payin_addresses_tron_address_hex UNIQUE (tron_address_hex);
ALTER TABLE payin_addresses ADD CONSTRAINT uq_payin_addresses_tron_address_base58 UNIQUE (tron_address_base58);

-- Create indexes for performance
CREATE INDEX idx_payin_addresses_account_id ON payin_addresses(account_id);
CREATE INDEX idx_payin_addresses_ledger_account_id ON payin_addresses(ledger_account_id) WHERE ledger_account_id IS NOT NULL;
CREATE INDEX idx_payin_addresses_status ON payin_addresses(status);
CREATE INDEX idx_payin_addresses_kms_key_id ON payin_addresses(kms_key_id);