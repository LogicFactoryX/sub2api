ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS is_member_group BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS member_fallback_group_id BIGINT NULL REFERENCES groups(id) ON DELETE RESTRICT;

DROP INDEX IF EXISTS idx_groups_show_to_all_users;
ALTER TABLE groups DROP COLUMN IF EXISTS show_to_all_users;

CREATE INDEX IF NOT EXISTS idx_groups_is_member_group
    ON groups (is_member_group)
    WHERE deleted_at IS NULL AND is_member_group = TRUE;

CREATE TABLE IF NOT EXISTS user_memberships (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_redeem_code_id BIGINT NULL REFERENCES redeem_codes(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_memberships_active_expiry
    ON user_memberships (expires_at)
    WHERE status = 'active';

-- Preserve users who already redeemed the former per-group cards by converting
-- their latest active entitlement into one site-wide membership expiry.
INSERT INTO user_memberships (user_id, expires_at, status, created_at, updated_at)
SELECT user_id, MAX(expires_at),
       CASE WHEN MAX(expires_at) > NOW() THEN 'active' ELSE 'expired' END,
       NOW(), NOW()
FROM user_group_entitlements
GROUP BY user_id
ON CONFLICT (user_id) DO UPDATE SET
    expires_at = GREATEST(user_memberships.expires_at, EXCLUDED.expires_at),
    status = CASE
        WHEN GREATEST(user_memberships.expires_at, EXCLUDED.expires_at) > NOW() THEN 'active'
        ELSE 'expired'
    END,
    updated_at = NOW();

UPDATE redeem_codes
SET type = 'membership', group_id = NULL, fallback_group_id = NULL
WHERE type = 'group';
