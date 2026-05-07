CREATE TABLE IF NOT EXISTS order_items (
    id                    BIGSERIAL       PRIMARY KEY,
    company_id            BIGINT          NOT NULL,
    order_id              BIGINT          NOT NULL,
    product_id            BIGINT          NOT NULL,
    product_value_id      BIGINT          NOT NULL, -- qaysi lotdan olindi
    quantity              NUMERIC(12, 3)  NOT NULL,                              -- qancha sotildi/qaytarildi
    sale_price            NUMERIC(12, 2)  NOT NULL,                              -- standart narx (skidkasiz)
    discount              NUMERIC(12, 2)   NOT NULL,                              -- haqiqiy narx (skidka bo'lsa kam)

    CONSTRAINT order_items_quantity_positive  CHECK (quantity > 0),
    CONSTRAINT order_items_price_positive     CHECK (sale_price > 0),
    CONSTRAINT order_items_discount_positive  CHECK (discount >= 0)
);

-- Order bo'yicha qatorlarni olish (eng tez-tez ishlatiladigan)
CREATE INDEX IF NOT EXISTS idx_order_items_order_id        ON order_items(order_id);

-- Tovar bo'yicha sotuv tarixi
CREATE INDEX IF NOT EXISTS idx_order_items_product_id      ON order_items(product_id);

-- Lot bo'yicha qaysi sotuvlarda ishlatilganini ko'rish
CREATE INDEX IF NOT EXISTS idx_order_items_product_value_id ON order_items(product_value_id);

-- Kompaniya bo'yicha filtrlash (hisobotlar uchun)
CREATE INDEX IF NOT EXISTS idx_order_items_company_id      ON order_items(company_id);
