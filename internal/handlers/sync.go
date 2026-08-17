package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/db"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/parser"
	"github.com/jackc/pgx/v5/pgtype"
)

type SyncResult struct {
	NewOrders int `json:"new_order"`
	Failed    int `json:"failed"`
}

func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	existingMessageIds, err := h.Queries.ExistingMessageIDs(ctx)
	if err != nil {
		http.Error(w, "failed to query db", http.StatusInternalServerError)
		return
	}

	messages, err := parser.GetLast30dMessageIDs(ctx, h.GmailSrv, "takeaways")
	if err != nil {
		http.Error(w, "failed to fetch messages", http.StatusInternalServerError)
		return
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

	var synced, failed int
	for _, id := range newIDs {
		orderDetails, err := parser.GetOrderTotalFromMsg(ctx, h.GmailSrv, id)
		if err != nil {
			log.Println("error parsing details for message id:", id, err)
			failed++
			continue // was missing — prevented garbage inserts
		}

		var total pgtype.Numeric
		if err := total.Scan(fmt.Sprintf("%.2f", orderDetails.Total)); err != nil {
			log.Println("error converting total for message id:", id, err)
			failed++
			continue
		}

		err = h.Queries.InsertOrder(ctx, db.InsertOrderParams{
			MessageID: id,
			Partner:   orderDetails.Partner,
			Store:     orderDetails.Store,
			Total:     total,
			OrderDate: pgtype.Timestamptz{Time: orderDetails.Date, Valid: true},
		})
		if err != nil {
			log.Println("error inserting order for message id:", id, err)
			failed++
			continue
		}
		synced++
	}

	writeJSON(w, http.StatusOK, SyncResult{NewOrders: synced, Failed: failed})
}
