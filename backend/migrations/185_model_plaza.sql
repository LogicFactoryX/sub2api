CREATE TABLE IF NOT EXISTS model_plaza_entries (
    id BIGSERIAL PRIMARY KEY,
    model_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    tags TEXT[] NOT NULL DEFAULT '{}'::TEXT[],
    sort_order INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT model_plaza_entries_model_name_not_blank CHECK (BTRIM(model_name) <> ''),
    CONSTRAINT model_plaza_entries_model_group_unique UNIQUE (model_name, group_id)
);

CREATE INDEX IF NOT EXISTS idx_model_plaza_entries_visible
    ON model_plaza_entries (enabled, sort_order, id);

CREATE INDEX IF NOT EXISTS idx_model_plaza_entries_group_id
    ON model_plaza_entries (group_id);
