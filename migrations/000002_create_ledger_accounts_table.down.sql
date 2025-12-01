-- +migrate Down
DROP INDEX IF EXISTS idx_ledger_accounts_status;
DROP INDEX IF EXISTS idx_ledger_accounts_type;
DROP INDEX IF EXISTS idx_ledger_accounts_account_id;
ALTER TABLE ledger_accounts DROP CONSTRAINT IF EXISTS uq_ledger_accounts_account_code;
DROP TABLE IF EXISTS ledger_accounts;