-- +migrate Down
DROP INDEX IF EXISTS idx_payin_requests_client_reference;
DROP INDEX IF EXISTS idx_payin_requests_state;
DROP INDEX IF EXISTS idx_payin_requests_address_id;
DROP INDEX IF EXISTS idx_payin_requests_ledger_account_id;
DROP INDEX IF EXISTS idx_payin_requests_account_id;
DROP TABLE IF EXISTS payin_requests;