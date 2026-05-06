CREATE TABLE IF NOT EXISTS product_value (
    id                BIGSERIAL       PRIMARY KEY,
    company_id        BIGINT          NOT NULL REFERENCES companys(id),
    product_id        BIGINT          NOT NULL REFERENCES products(id),
    price             NUMERIC(12, 2)  NOT NULL,                  -- kelish narxi
    quantity_before        NUMERIC(12, 3)  NOT NULL DEFAULT 0,   -- kelgan miqdor
    quantity_after         NUMERIC(12, 3)  NOT NULL,             -- qolgan miqdor
    unit              VARCHAR(20)     NOT NULL,                  -- o'lchov birligi
    created_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
);

CREATE INDEX IF NOT EXISTS idx_product_value_company_id   ON product_value(company_id);

CREATE INDEX IF NOT EXISTS idx_product_value_product_id   ON product_value(product_id);

CREATE INDEX IF NOT EXISTS idx_product_value_in_stock
    ON product_value(company_id, product_id)
    WHERE quantity_after > 0;

CREATE INDEX IF NOT EXISTS idx_product_value_company_product
    ON product_value(company_id, product_id);
