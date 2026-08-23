CREATE TABLE IF NOT EXISTS urls (
    id          UUID PRIMARY KEY,
    original    TEXT NOT NULL UNIQUE,
    short_code  VARCHAR(12) NOT NULL UNIQUE,
    created_at  BIGINT NOT NULL,
    updated_at  BIGINT NOT NULL
);
