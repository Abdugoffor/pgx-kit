CREATE TABLE IF NOT EXISTS orders (
    id                  BIGSERIAL       PRIMARY KEY,
    company_id          BIGINT          NOT NULL REFERENCES companys(id),
    user_id             BIGINT          NOT NULL REFERENCES users(id), -- kim amalga oshirdi
    type                VARCHAR(20)     NOT NULL,                      -- harakat turi
    reference_order_id  BIGINT          REFERENCES orders(id),         -- vazvrat_bizga uchun original sotuv
    total_sum           NUMERIC(12, 2)  NOT NULL DEFAULT 0,            -- hujjatning umumiy summasi
    note                TEXT,                                          -- izoh (ixtiyoriy)
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
);

-- Kompaniya bo'yicha filtrlash (asosiy filtr)
CREATE INDEX IF NOT EXISTS idx_orders_company_id  ON orders(company_id);

-- Tur bo'yicha filtrlash (sotuvlar / qaytarishlar alohida ko'rish uchun)
CREATE INDEX IF NOT EXISTS idx_orders_type        ON orders(type);

-- Sana bo'yicha filtrlash (kunlik/oylik hisobot uchun)
CREATE INDEX IF NOT EXISTS idx_orders_created_at  ON orders(created_at);

-- Kompaniya + tur + sana — eng ko'p ishlatiladigan kombinatsiya
CREATE INDEX IF NOT EXISTS idx_orders_company_type_date
    ON orders(company_id, type, created_at DESC);

-- Qaysi sotuvdan qaytarilganini tez topish (vazvrat_bizga uchun)
CREATE INDEX IF NOT EXISTS idx_orders_reference_order_id
    ON orders(reference_order_id)
    WHERE reference_order_id IS NOT NULL;
