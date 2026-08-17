package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/db"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/parser"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/gmail/v1"
)

type StatsHandler struct {
	Queries  *db.Queries
	GmailSrv *gmail.Service
	Pool     *pgxpool.Pool
}

func (h *StatsHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	existingMessageIds, err := h.Queries.ExistingMessageIDs(r.Context())
	if err != nil {
		http.Error(w, "failed to query db", http.StatusInternalServerError)
		return
	}
	messages, err := parser.GetLast30dMessageIDs(r.Context(), h.GmailSrv, "takeaways")
	if err != nil {
		log.Fatalf("error fetching messages: %v", err)
	}

	existing := make(map[string]bool, len(existingMessageIds))
	for _, id := range existingMessageIds {
		existing[id] = true
	}

	var newIDs []string
	for _, msg := range messages {
		if !existing[msg.Id] {
			newIDs = append(newIDs, msg.Id)
		}
	}

	for _, id := range newIDs {
		orderDetails, err := parser.GetOrderTotalFromMsg(r.Context(), h.GmailSrv, id)
		if err != nil {
			log.Println("error parsing details for message id:", id, err)
		}
		fmt.Printf("%+v\n", orderDetails)

		var total pgtype.Numeric
		if err := total.Scan(fmt.Sprintf("%.2f", orderDetails.Total)); err != nil {
			log.Println("error converting total for message id:", id, err)
			continue
		}

		err = h.Queries.InsertOrder(r.Context(), db.InsertOrderParams{
			MessageID: id,
			Partner:   orderDetails.Partner,
			Store:     orderDetails.Store,
			Total:     total,
			OrderDate: pgtype.Timestamptz{Time: orderDetails.Date, Valid: true},
		})
		if err != nil {
			log.Println("error inserting order for message id:", id, err)
		}
	}

	total, err := h.Queries.SumLastNDays(r.Context(), 30)
	if err != nil {
		http.Error(w, "failed to compute total", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, total)
}
