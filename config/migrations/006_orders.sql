CREATE TABLE IF NOT EXISTS orders (
    id           BIGSERIAL      PRIMARY KEY,
    user_id      BIGINT         NOT NULL REFERENCES users(id),
    status       VARCHAR(50)    NOT NULL DEFAULT 'pending',
    total_amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
    note         TEXT,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id    ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status     ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders(deleted_at);

CREATE TABLE IF NOT EXISTS order_items (
    id         BIGSERIAL      PRIMARY KEY,
    order_id   BIGINT         NOT NULL REFERENCES orders(id),
    product_id BIGINT         NOT NULL REFERENCES products(id),
    quantity   INT            NOT NULL CHECK (quantity > 0),
    price      NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id   ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
