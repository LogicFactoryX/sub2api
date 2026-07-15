ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS show_to_all_users BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_groups_show_to_all_users
    ON groups (show_to_all_users)
    WHERE deleted_at IS NULL AND show_to_all_users = TRUE;
