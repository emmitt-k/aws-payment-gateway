-- +migrate Down
DROP INDEX IF EXISTS idx_chain_cursors_updated_at;
DROP TABLE IF EXISTS chain_cursors;