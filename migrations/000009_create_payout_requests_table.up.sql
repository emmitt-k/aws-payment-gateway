-- +migrate Up
CREATE TABLE payout_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    from_ledger_account_id UUID NOT NULL REFERENCES ledger_accounts(id) ON DELETE RESTRICT,
    to_address_base58 TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('requested', 'pending_sign', 'broadcasted', 'confirmed', 'failed')),
    tron_tx_hash TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_payout_requests_account_id ON payout_requests(account_id);
CREATE INDEX idx_payout_requests_from_ledger_account_id ON payout_requests(from_ledger_account_id);
CREATE INDEX idx_payout_requests_state ON payout_requests(state);
CREATE INDEX idx_payout_requests_tron_tx_hash ON payout_requests(tron_tx_hash) WHERE tron_tx_hash IS NOT NULL;
CREATE INDEX idx_payout_requests_created_at ON payout_requests(created_at);
CREATE INDEX idx_payout_requests_updated_at ON payout_requests(updated_at);