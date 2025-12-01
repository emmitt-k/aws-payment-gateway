-- +migrate Down
DROP INDEX IF EXISTS idx_balances_updated_at;
DROP TABLE IF EXISTS balances;