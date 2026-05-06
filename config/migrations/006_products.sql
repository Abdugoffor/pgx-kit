CREATE TABLE IF NOT EXISTS products (
    id          BIGSERIAL       PRIMARY KEY,
    company_id  BIGINT          NOT NULL REFERENCES companys(id),
    category_id BIGINT          NOT NULL REFERENCES categories(id),
    name        VARCHAR(255)    NOT NULL,
    description VARCHAR(1000),
    cost_price  NUMERIC(12, 2)  NOT NULL,
    sell_price  NUMERIC(12, 2)  NOT NULL,
    is_active   BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_company_id  ON products(company_id);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_is_active   ON products(is_active);
