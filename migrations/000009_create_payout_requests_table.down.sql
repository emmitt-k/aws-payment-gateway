-- +migrate Down
DROP INDEX IF EXISTS idx_payout_requests_updated_at;
DROP INDEX IF EXISTS idx_payout_requests_created_at;
DROP INDEX IF EXISTS idx_payout_requests_tron_tx_hash;
DROP INDEX IF EXISTS idx_payout_requests_state;
DROP INDEX IF EXISTS idx_payout_requests_from_ledger_account_id;
DROP INDEX IF EXISTS idx_payout_requests_account_id;
DROP TABLE IF EXISTS payout_requests;