CREATE TABLE IF NOT EXISTS company_users (
    id         BIGSERIAL      PRIMARY KEY,
    company_id BIGINT         NOT NULL REFERENCES companys(id),
    user_id    BIGINT         NOT NULL REFERENCES users(id),
    is_active  BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
)