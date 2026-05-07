CREATE TABLE IF NOT EXISTS product_history (
    id                BIGSERIAL       PRIMARY KEY,
    company_id        BIGINT          NOT NULL,
    user_id           BIGINT          NOT NULL,                               -- kim amalga oshirdi
    product_id        BIGINT          NOT NULL,
    product_value_id  BIGINT          NOT NULL,                               -- qaysi lot
    order_id          BIGINT,                                                 -- NULL = prihod
    order_type        VARCHAR(20)     NOT NULL,                               -- harakat turi
    quantity          NUMERIC(12, 3)  NOT NULL,                               -- o'zgarish miqdori
    quantity_before   NUMERIC(12, 3)  NOT NULL,                               -- harakatdan oldin
    quantity_after    NUMERIC(12, 3)  NOT NULL,                               -- harakatdan keyin
    price             NUMERIC(12, 2),                                         -- narx (yuqorida izoh)
    created_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Lot tarixi — "Rulon 3 ga nima bo'ldi?" (eng ko'p ishlatiladigan)
CREATE INDEX IF NOT EXISTS idx_product_history_product_value_id
    ON product_history(product_value_id);

-- Tovar bo'yicha barcha harakatlar
CREATE INDEX IF NOT EXISTS idx_product_history_product_id
    ON product_history(product_id);

-- Kompaniya bo'yicha filtrlash
CREATE INDEX IF NOT EXISTS idx_product_history_company_id
    ON product_history(company_id);

-- Order bo'yicha barcha harakatlar (1 sotuv bir nechta lot ochishi mumkin)
CREATE INDEX IF NOT EXISTS idx_product_history_order_id
    ON product_history(order_id)
    WHERE order_id IS NOT NULL;

-- Tur bo'yicha filtrlash (faqat kirimlar / faqat sotuvlar)
CREATE INDEX IF NOT EXISTS idx_product_history_order_type
    ON product_history(order_type);

-- Sana bo'yicha filtrlash (kunlik/oylik hisobot)
CREATE INDEX IF NOT EXISTS idx_product_history_created_at
    ON product_history(created_at DESC);

-- Kompaniya + tur + sana — hisobotlarda eng ko'p ishlatiladigan kombinatsiya
CREATE INDEX IF NOT EXISTS idx_product_history_company_type_date
    ON product_history(company_id, order_type, created_at DESC);
