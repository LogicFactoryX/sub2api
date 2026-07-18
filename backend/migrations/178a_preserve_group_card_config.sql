CREATE TABLE IF NOT EXISTS legacy_group_redeem_code_config (
    redeem_code_id BIGINT PRIMARY KEY REFERENCES redeem_codes(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL,
    fallback_group_id BIGINT NOT NULL
);

INSERT INTO legacy_group_redeem_code_config (redeem_code_id, group_id, fallback_group_id)
SELECT id, group_id, fallback_group_id
FROM redeem_codes
WHERE type = 'group'
  AND group_id IS NOT NULL
  AND fallback_group_id IS NOT NULL
ON CONFLICT (redeem_code_id) DO NOTHING;
