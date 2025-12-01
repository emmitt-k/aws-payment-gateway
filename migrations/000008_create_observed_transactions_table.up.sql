-- +migrate Up
CREATE TABLE observed_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    address_id UUID NOT NULL REFERENCES payin_addresses(id) ON DELETE CASCADE,
    tx_hash TEXT NOT NULL,
    block_number BIGINT NOT NULL,
    amount_minor BIGINT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('detected', 'confirmed')),
    raw_event JSONB NOT NULL,
    journal_entry_id UUID NULL REFERENCES journal_entries(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create unique constraint for transaction hash
ALTER TABLE observed_transactions ADD CONSTRAINT uq_observed_transactions_tx_hash UNIQUE (tx_hash);

-- Create indexes for performance
CREATE INDEX idx_observed_transactions_account_id ON observed_transactions(account_id);
CREATE INDEX idx_observed_transactions_address_id ON observed_transactions(address_id);
CREATE INDEX idx_observed_transactions_status ON observed_transactions(status);
CREATE INDEX idx_observed_transactions_block_number ON observed_transactions(block_number);
CREATE INDEX idx_observed_transactions_journal_entry_id ON observed_transactions(journal_entry_id) WHERE journal_entry_id IS NOT NULL;
CREATE INDEX idx_observed_transactions_created_at ON observed_transactions(created_at);