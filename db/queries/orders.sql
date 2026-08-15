-- name: InsertOrder :exec
INSERT INTO orders (message_id, partner, store, total, order_date)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (message_id) DO NOTHING;

-- name: ExistingMessageIDs :many
SELECT message_id FROM orders;

-- name: SumLastNDays :one
SELECT COALESCE(SUM(total), 0)::float8 AS total
FROM orders
WHERE order_date >= now() - make_interval(days => sqlc.arg(days)::int);
