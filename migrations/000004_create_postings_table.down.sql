-- +migrate Down
DROP INDEX IF EXISTS idx_postings_created_at;
DROP INDEX IF EXISTS idx_postings_side;
DROP INDEX IF EXISTS idx_postings_currency;
DROP INDEX IF EXISTS idx_postings_ledger_account_id;
DROP INDEX IF EXISTS idx_postings_journal_entry_id;
DROP TABLE IF EXISTS postings;