ALTER TABLE products
    ADD COLUMN IF NOT EXISTS company_id BIGINT REFERENCES companys(id);

CREATE INDEX IF NOT EXISTS idx_products_company_id ON products(company_id);
