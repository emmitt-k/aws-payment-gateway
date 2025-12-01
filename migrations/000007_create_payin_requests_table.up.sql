-- +migrate Up
CREATE TABLE payin_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    ledger_account_id UUID NULL REFERENCES ledger_accounts(id) ON DELETE SET NULL,
    address_id UUID NOT NULL REFERENCES payin_addresses(id) ON DELETE CASCADE,
    client_reference TEXT NULL,
    state TEXT NOT NULL CHECK (state IN ('pending', 'active', 'disabled')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_payin_requests_account_id ON payin_requests(account_id);
CREATE INDEX idx_payin_requests_ledger_account_id ON payin_requests(ledger_account_id) WHERE ledger_account_id IS NOT NULL;
CREATE INDEX idx_payin_requests_address_id ON payin_requests(address_id);
CREATE INDEX idx_payin_requests_state ON payin_requests(state);
CREATE INDEX idx_payin_requests_client_reference ON payin_requests(client_reference) WHERE client_reference IS NOT NULL;