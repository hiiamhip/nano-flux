CREATE TABLE IF NOT EXISTS clicks (
    id          UUID PRIMARY KEY,
    short_code  VARCHAR(12) NOT NULL,
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    clicked_at  BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_clicks_short_code ON clicks(short_code);
