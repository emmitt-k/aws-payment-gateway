-- +migrate Down
DROP INDEX IF EXISTS idx_observed_transactions_created_at;
DROP INDEX IF EXISTS idx_observed_transactions_journal_entry_id;
DROP INDEX IF EXISTS idx_observed_transactions_block_number;
DROP INDEX IF EXISTS idx_observed_transactions_status;
DROP INDEX IF EXISTS idx_observed_transactions_address_id;
DROP INDEX IF EXISTS idx_observed_transactions_account_id;
ALTER TABLE observed_transactions DROP CONSTRAINT IF EXISTS uq_observed_transactions_tx_hash;
DROP TABLE IF EXISTS observed_transactions;