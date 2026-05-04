CREATE TABLE IF NOT EXISTS languages(
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(20) NOT NULL,
    description VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_languages_is_active ON languages(is_active);
