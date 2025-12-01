-- +migrate Up
CREATE TABLE chain_cursors (
    id TEXT PRIMARY KEY,
    last_block_number BIGINT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for updated_at queries
CREATE INDEX idx_chain_cursors_updated_at ON chain_cursors(updated_at);

-- Insert initial cursor for Tron blockchain scanning
INSERT INTO chain_cursors (id, last_block_number) VALUES ('tron_mainnet', 0);