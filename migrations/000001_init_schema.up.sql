CREATE TABLE restaurants (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE menu_items (
    id            SERIAL PRIMARY KEY,
    restaurant_id INT NOT NULL REFERENCES restaurants (id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    price_cents   BIGINT NOT NULL CHECK (price_cents >= 0),
    stock         INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_menu_items_restaurant_id ON menu_items (restaurant_id);

CREATE TABLE orders (
    id            BIGSERIAL PRIMARY KEY,
    restaurant_id INT NOT NULL REFERENCES restaurants (id),
    status        TEXT NOT NULL DEFAULT 'new',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_restaurant_id ON orders (restaurant_id);

CREATE TABLE order_items (
    id           SERIAL PRIMARY KEY,
    order_id     BIGINT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    menu_item_id INT NOT NULL REFERENCES menu_items (id),
    quantity     INTEGER NOT NULL CHECK (quantity > 0),
    price_cents  BIGINT NOT NULL CHECK (price_cents >= 0)
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id);
