-- +migrate Down
DROP INDEX IF EXISTS idx_journal_entries_external_ref;
DROP INDEX IF EXISTS idx_journal_entries_created_at;
DROP INDEX IF EXISTS idx_journal_entries_event_type;
DROP INDEX IF EXISTS idx_journal_entries_account_id;
DROP TABLE IF EXISTS journal_entries;