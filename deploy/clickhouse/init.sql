CREATE DATABASE IF NOT EXISTS nexuslink;

-- One row per click. ReplacingMergeTree(inserted_at) collapses duplicate
-- deliveries with the same (short_code, clicked_at, request_id), which makes
-- at-least-once consumption safe to query with FINAL (or uniq by request_id).
CREATE TABLE IF NOT EXISTS nexuslink.clicks
(
    short_code  LowCardinality(String),
    clicked_at  DateTime64(3, 'UTC'),
    ip          String,
    user_agent  String,
    referrer    String,
    request_id  String,
    inserted_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(inserted_at)
PARTITION BY toYYYYMM(clicked_at)
ORDER BY (short_code, clicked_at, request_id);
