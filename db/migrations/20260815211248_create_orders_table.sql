-- +goose Up
CREATE TABLE orders (
  order_id TEXT PRIMARY KEY,
  restaurant_id TEXT NOT NULL,
  total NUMERIC(6,2) NOT NULL,
  restaurant_name TEXT NOT NULL,
  order_date TIMESTAMPTZ NOT NULL,
  synced_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose Down
DROP TABLE orders;
