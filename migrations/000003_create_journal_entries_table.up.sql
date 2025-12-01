-- +migrate Up
CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (event_type IN ('payin', 'payout', 'adjustment')),
    external_ref TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_journal_entries_account_id ON journal_entries(account_id);
CREATE INDEX idx_journal_entries_event_type ON journal_entries(event_type);
CREATE INDEX idx_journal_entries_created_at ON journal_entries(created_at);
CREATE INDEX idx_journal_entries_external_ref ON journal_entries(external_ref) WHERE external_ref IS NOT NULL;