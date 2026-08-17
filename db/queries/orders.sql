-- name: InsertOrder :exec
INSERT INTO orders (message_id, partner, store, total, order_date)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (message_id) DO NOTHING;

-- name: ExistingMessageIDs :many
SELECT message_id FROM orders;

-- name: GetLastOrderDate :one
SELECT order_date FROM orders
ORDER BY order_date DESC
LIMIT 1;

-- name: GetMonthTotal :one
SELECT COALESCE(SUM(total), 0)::numeric(10,2) as total
FROM orders
WHERE order_date >= date_trunc('month', now())
  AND order_date < date_trunc('month', now()) + interval '1 month';

-- name: GetPrevMonthTotal :one
SELECT COALESCE(SUM(total), 0)::numeric(10,2) as total
FROM orders
WHERE order_date >= date_trunc('month', now()) - interval '1 month'
  AND order_date < date_trunc('month', now());

-- name: GetMonthAverage :one
SELECT COALESCE(AVG(total), 0)::numeric(10,2) as average
FROM orders
WHERE order_date >= date_trunc('month', now())
  AND order_date < date_trunc('month', now()) + interval '1 month';

-- name: GetMonthOrderCount :one
SELECT COUNT(*) as count
FROM orders
WHERE order_date >= date_trunc('month', now())
  AND order_date < date_trunc('month', now()) + interval '1 month';

-- name: GetTopSpenders :many
SELECT store, SUM(total)::numeric(10,2) as total
FROM orders
WHERE order_date >= date_trunc('month', now())
  AND order_date < date_trunc('month', now()) + interval '1 month'
GROUP BY store
ORDER BY total DESC
LIMIT 3;
