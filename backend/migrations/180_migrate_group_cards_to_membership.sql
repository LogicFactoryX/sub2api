-- Keep this legacy column for binary-only rollback compatibility. The new
-- application no longer reads or exposes it.
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS show_to_all_users BOOLEAN NOT NULL DEFAULT FALSE;

-- Convert groups that were actually granted by the former group-card system.
-- The latest entitlement supplies the fallback configured for that target.
WITH entitlement_target_config AS (
    SELECT DISTINCT ON (ent.group_id)
        ent.group_id,
        ent.fallback_group_id
    FROM user_group_entitlements AS ent
    ORDER BY ent.group_id, ent.expires_at DESC, ent.id DESC
), latest_target_config AS (
    SELECT DISTINCT ON (source.group_id)
        source.group_id,
        source.fallback_group_id
    FROM (
        SELECT group_id, fallback_group_id, 1 AS priority
        FROM entitlement_target_config
        UNION ALL
        SELECT group_id, fallback_group_id, 2 AS priority
        FROM legacy_group_redeem_code_config
    ) AS source
    ORDER BY source.group_id, source.priority
), valid_target_config AS (
    SELECT cfg.group_id, cfg.fallback_group_id
    FROM latest_target_config AS cfg
    JOIN groups AS target ON target.id = cfg.group_id
    JOIN groups AS fallback ON fallback.id = cfg.fallback_group_id
    WHERE target.deleted_at IS NULL
      AND target.status = 'active'
      AND target.subscription_type = 'standard'
      AND fallback.deleted_at IS NULL
      AND fallback.status = 'active'
      AND fallback.subscription_type = 'standard'
      AND fallback.is_exclusive = FALSE
      AND fallback.is_member_group = FALSE
      AND fallback.platform = target.platform
      AND NOT EXISTS (
          SELECT 1
          FROM latest_target_config AS sold
          WHERE sold.group_id = fallback.id
      )
)
UPDATE groups AS target
SET is_member_group = TRUE,
    is_exclusive = FALSE,
    member_fallback_group_id = cfg.fallback_group_id,
    updated_at = NOW()
FROM valid_target_config AS cfg
WHERE target.id = cfg.group_id;

-- Restore the legacy code shape in storage so rolling back to the preceding
-- binary still recognizes unused group cards. New code maps this type to
-- membership at the repository boundary.
UPDATE redeem_codes AS code
SET type = 'group',
    group_id = legacy.group_id,
    fallback_group_id = legacy.fallback_group_id
FROM legacy_group_redeem_code_config AS legacy
WHERE code.id = legacy.redeem_code_id
  AND code.type = 'membership';

WITH redeemed_config AS (
    SELECT DISTINCT ON (redeem_code_id)
        redeem_code_id, group_id, fallback_group_id
    FROM user_group_entitlements
    WHERE redeem_code_id IS NOT NULL
    ORDER BY redeem_code_id, expires_at DESC, id DESC
)
UPDATE redeem_codes AS code
SET type = 'group',
    group_id = cfg.group_id,
    fallback_group_id = cfg.fallback_group_id
FROM redeemed_config AS cfg
WHERE code.id = cfg.redeem_code_id
  AND code.type = 'membership';
