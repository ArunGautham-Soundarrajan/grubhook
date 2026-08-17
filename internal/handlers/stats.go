// internal/handlers/dashboard.go
package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

type DashboardResponse struct {
	MonthTotal         float64      `json:"month_total"`
	LastMonthTotal     float64      `json:"last_month_total"`
	MonthAverage       float64      `json:"month_average"`
	OrderCount         int64        `json:"order_count"`
	TopSpenders        []StoreSpend `json:"top_spenders"`
	DaysSinceLastOrder *int         `json:"days_since_last_order"` // nil if no orders yet
}

type StoreSpend struct {
	Store string  `json:"store"`
	Total float64 `json:"total"`
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	monthTotal, err := h.Queries.GetMonthTotal(ctx)
	if err != nil {
		log.Println("error fetching month total:", err)
		http.Error(w, "failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}

	lastMonthTotal, err := h.Queries.GetPrevMonthTotal(ctx)
	if err != nil {
		log.Println("error fetching last month total:", err)
		http.Error(w, "failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}

	monthAverage, err := h.Queries.GetMonthAverage(ctx)
	if err != nil {
		log.Println("error fetching month average:", err)
		http.Error(w, "failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}

	orderCount, err := h.Queries.GetMonthOrderCount(ctx)
	if err != nil {
		log.Println("error fetching order count:", err)
		http.Error(w, "failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}

	topSpendersRows, err := h.Queries.GetTopSpenders(ctx)
	if err != nil {
		log.Println("error fetching top spenders:", err)
		http.Error(w, "failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}

	var daysSince *int
	lastOrderDate, err := h.Queries.GetLastOrderDate(ctx)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Println("error fetching last order date:", err)
			http.Error(w, "failed to fetch dashboard stats", http.StatusInternalServerError)
			return
		}
		// no orders yet — daysSince stays nil
	} else {
		days := int(time.Since(lastOrderDate.Time).Hours() / 24)
		daysSince = &days
	}

	topSpenders := make([]StoreSpend, len(topSpendersRows))
	for i, row := range topSpendersRows {
		total, _ := row.Total.Float64Value()
		topSpenders[i] = StoreSpend{Store: row.Store, Total: total.Float64}
	}

	monthTotalF, _ := monthTotal.Float64Value()
	lastMonthTotalF, _ := lastMonthTotal.Float64Value()
	monthAverageF, _ := monthAverage.Float64Value()

	writeJSON(w, http.StatusOK, DashboardResponse{
		MonthTotal:         monthTotalF.Float64,
		LastMonthTotal:     lastMonthTotalF.Float64,
		MonthAverage:       monthAverageF.Float64,
		OrderCount:         orderCount,
		TopSpenders:        topSpenders,
		DaysSinceLastOrder: daysSince,
	})
}
