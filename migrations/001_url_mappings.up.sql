CREATE TABLE IF NOT EXISTS url_mappings (
    short_code   VARCHAR(16) PRIMARY KEY,
    long_url     TEXT NOT NULL,
    snowflake_id BIGINT UNIQUE,
    user_id      UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    http_status  SMALLINT NOT NULL DEFAULT 302
);

CREATE INDEX IF NOT EXISTS idx_url_mappings_user_id ON url_mappings (user_id);
CREATE INDEX IF NOT EXISTS idx_url_mappings_created_at ON url_mappings (created_at);
