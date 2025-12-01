-- +migrate Down
DROP INDEX IF EXISTS idx_payin_addresses_kms_key_id;
DROP INDEX IF EXISTS idx_payin_addresses_status;
DROP INDEX IF EXISTS idx_payin_addresses_ledger_account_id;
DROP INDEX IF EXISTS idx_payin_addresses_account_id;
ALTER TABLE payin_addresses DROP CONSTRAINT IF EXISTS uq_payin_addresses_tron_address_base58;
ALTER TABLE payin_addresses DROP CONSTRAINT IF EXISTS uq_payin_addresses_tron_address_hex;
DROP TABLE IF EXISTS payin_addresses;