-- UZS narxlari katta bo'lgani uchun NUMERIC(12,2) yetmaydi
-- total_sum: price * quantity * count jami juda katta son berishi mumkin

ALTER TABLE orders        ALTER COLUMN total_sum       TYPE NUMERIC(18, 2);
ALTER TABLE order_items   ALTER COLUMN sale_price      TYPE NUMERIC(18, 2);
ALTER TABLE order_items   ALTER COLUMN discount        TYPE NUMERIC(18, 2);
ALTER TABLE product_value ALTER COLUMN price           TYPE NUMERIC(18, 2);
ALTER TABLE product_history ALTER COLUMN price         TYPE NUMERIC(18, 2);
ALTER TABLE products      ALTER COLUMN price           TYPE NUMERIC(18, 2);
ALTER TABLE products      ALTER COLUMN sell_price      TYPE NUMERIC(18, 2);
