package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/db"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/deliveroo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SaveOrders inserts orders in a single transaction, skipping any already
// stored. It returns the number of newly inserted orders.
func SaveOrders(ctx context.Context, pool *pgxpool.Pool, orders []deliveroo.Order) (int64, error) {
	if len(orders) == 0 {
		return 0, nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) // no-op after Commit

	q := db.New(tx)
	var inserted int64
	for _, o := range orders {
		var total pgtype.Numeric
		if err := total.Scan(strings.TrimPrefix(o.Balance, "£")); err != nil {
			return 0, fmt.Errorf("order %s: parsing balance %q: %w", o.ID, o.Balance, err)
		}

		n, err := q.InsertOrder(ctx, db.InsertOrderParams{
			OrderID:        o.ID,
			RestaurantID:   o.RestaurantID,
			RestaurantName: o.Restaurant,
			Total:          total,
			OrderDate:      pgtype.Timestamptz{Time: o.Submitted, Valid: true},
		})
		if err != nil {
			return 0, fmt.Errorf("inserting order %s: %w", o.ID, err)
		}
		inserted += n
	}

	return inserted, tx.Commit(ctx)
}
