-- +migrate Down
DROP INDEX IF EXISTS idx_system_wallets_kms_key_id;
DROP INDEX IF EXISTS idx_system_wallets_status;
DROP INDEX IF EXISTS idx_system_wallets_role;
ALTER TABLE system_wallets DROP CONSTRAINT IF EXISTS uq_system_wallets_address;
DROP TABLE IF EXISTS system_wallets;