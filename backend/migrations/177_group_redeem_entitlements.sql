ALTER TABLE redeem_codes
    ADD COLUMN IF NOT EXISTS fallback_group_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_redeem_codes_fallback_group_id
    ON redeem_codes (fallback_group_id);

CREATE TABLE IF NOT EXISTS user_group_entitlements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    fallback_group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    redeem_code_id BIGINT NULL UNIQUE REFERENCES redeem_codes(id) ON DELETE SET NULL,
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    expired_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_group_entitlements_groups_differ CHECK (group_id <> fallback_group_id),
    CONSTRAINT user_group_entitlements_expiry_valid CHECK (expires_at > starts_at),
    CONSTRAINT user_group_entitlements_status_valid CHECK (status IN ('active', 'expired', 'revoked'))
);

CREATE INDEX IF NOT EXISTS idx_user_group_entitlements_active_user_group
    ON user_group_entitlements (user_id, group_id, expires_at)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_user_group_entitlements_expiry
    ON user_group_entitlements (expires_at, id)
    WHERE status = 'active';
