ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS company_id BIGINT REFERENCES companys(id);

CREATE INDEX IF NOT EXISTS idx_categories_company_id ON categories(company_id);
