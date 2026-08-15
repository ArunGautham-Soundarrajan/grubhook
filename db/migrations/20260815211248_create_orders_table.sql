-- +goose Up
CREATE TABLE orders (
  message_id TEXT PRIMARY KEY,
  partner TEXT NOT NULL,
  store TEXT NOT NULL,
  total NUMERIC(6,2) NOT NULL,
  order_date TIMESTAMPTZ NOT NULL,
  synced_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose Down
DROP TABLE orders;
