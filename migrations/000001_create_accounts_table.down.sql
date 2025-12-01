-- +migrate Down
DROP INDEX IF EXISTS idx_accounts_name;
DROP INDEX IF EXISTS idx_accounts_status;
DROP TABLE IF EXISTS accounts;