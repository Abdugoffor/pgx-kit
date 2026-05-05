CREATE TABLE IF NOT EXISTS companys (
    id         BIGSERIAL      PRIMARY KEY,
    name       VARCHAR(255)   NOT NULL,
    is_active  BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
)