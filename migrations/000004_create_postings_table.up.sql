-- +migrate Up
CREATE TABLE postings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    ledger_account_id UUID NOT NULL REFERENCES ledger_accounts(id) ON DELETE RESTRICT,
    currency TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    side TEXT NOT NULL CHECK (side IN ('debit', 'credit')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_postings_journal_entry_id ON postings(journal_entry_id);
CREATE INDEX idx_postings_ledger_account_id ON postings(ledger_account_id);
CREATE INDEX idx_postings_currency ON postings(currency);
CREATE INDEX idx_postings_side ON postings(side);
CREATE INDEX idx_postings_created_at ON postings(created_at);

-- Add check constraint to ensure debits equal credits per journal entry
-- This will be enforced at the application level with triggers if needed