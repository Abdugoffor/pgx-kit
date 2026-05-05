-- ============================================================
-- Seeder: 1000 languages, 1000 categories, 1000 products
-- Run: psql -d <dbname> -f config/seeds/seed.sql
-- ============================================================

BEGIN;

-- -------------------------------------------------------
-- LANGUAGES  (name <= 20 chars)
-- -------------------------------------------------------
INSERT INTO languages (name, description, is_active, created_at, updated_at)
SELECT
    CASE (i % 20)
        WHEN  0 THEN 'Uzbek'
        WHEN  1 THEN 'Russian'
        WHEN  2 THEN 'English'
        WHEN  3 THEN 'Chinese'
        WHEN  4 THEN 'Arabic'
        WHEN  5 THEN 'French'
        WHEN  6 THEN 'German'
        WHEN  7 THEN 'Spanish'
        WHEN  8 THEN 'Japanese'
        WHEN  9 THEN 'Korean'
        WHEN 10 THEN 'Turkish'
        WHEN 11 THEN 'Italian'
        WHEN 12 THEN 'Portuguese'
        WHEN 13 THEN 'Hindi'
        WHEN 14 THEN 'Persian'
        WHEN 15 THEN 'Polish'
        WHEN 16 THEN 'Dutch'
        WHEN 17 THEN 'Swedish'
        WHEN 18 THEN 'Greek'
        WHEN 19 THEN 'Czech'
    END || ' ' || CEIL(i::float / 20)::int                        AS name,
    'This is the description for language #' || i                  AS description,
    (i % 10 <> 0)                                                  AS is_active,
    NOW() - ((i % 365) || ' days')::interval                      AS created_at,
    NOW() - ((i % 30)  || ' days')::interval                      AS updated_at
FROM generate_series(1, 1000) AS s(i);

-- -------------------------------------------------------
-- CATEGORIES  (name <= 100 chars)
-- -------------------------------------------------------
INSERT INTO categories (name, is_active, created_at, updated_at)
SELECT
    CASE (i % 10)
        WHEN 1 THEN 'Electronics'
        WHEN 2 THEN 'Clothing & Apparel'
        WHEN 3 THEN 'Books & Stationery'
        WHEN 4 THEN 'Sports & Outdoors'
        WHEN 5 THEN 'Home & Garden'
        WHEN 6 THEN 'Toys & Games'
        WHEN 7 THEN 'Food & Beverage'
        WHEN 8 THEN 'Health & Beauty'
        WHEN 9 THEN 'Automotive'
        WHEN 0 THEN 'Office Supplies'
    END || ' #' || i                                               AS name,
    (i % 10 <> 0)                                                  AS is_active,
    NOW() - ((i % 365) || ' days')::interval                      AS created_at,
    NOW() - ((i % 30)  || ' days')::interval                      AS updated_at
FROM generate_series(1, 1000) AS s(i);

-- -------------------------------------------------------
-- PRODUCTS  (depends on categories being seeded above)
-- category_id cycles through 1..1000
-- price range: 5.00 – 9999.99
-- -------------------------------------------------------
INSERT INTO products (name, description, price, category_id, is_active, created_at, updated_at)
SELECT
    CASE (i % 10)
        WHEN 1 THEN 'Laptop'
        WHEN 2 THEN 'T-Shirt'
        WHEN 3 THEN 'Novel'
        WHEN 4 THEN 'Running Shoes'
        WHEN 5 THEN 'Garden Hose'
        WHEN 6 THEN 'LEGO Set'
        WHEN 7 THEN 'Green Tea'
        WHEN 8 THEN 'Face Cream'
        WHEN 9 THEN 'Car Wax'
        WHEN 0 THEN 'Stapler'
    END || ' Model-' || i                                          AS name,
    'High-quality product #' || i ||
        '. Perfect for everyday use. Manufactured with premium materials.' AS description,
    ROUND((5 + (i % 9995) + (i * 37 % 100) * 0.01)::numeric, 2)  AS price,
    ((i - 1) % 1000) + 1                                          AS category_id,
    (i % 10 <> 0)                                                  AS is_active,
    NOW() - ((i % 365) || ' days')::interval                      AS created_at,
    NOW() - ((i % 30)  || ' days')::interval                      AS updated_at
FROM generate_series(1, 1000) AS s(i);

COMMIT;
