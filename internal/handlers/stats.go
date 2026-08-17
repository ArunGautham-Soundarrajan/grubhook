package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/db"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/parser"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/gmail/v1"
)

type StatsHandler struct {
	Queries  *db.Queries
	GmailSrv *gmail.Service
	Pool     *pgxpool.Pool
}

func (h *StatsHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	messages, err := parser.GetLast30dMessageIDs(r.Context(), h.GmailSrv, "takeaways")
	if err != nil {
		log.Fatalf("error fetching messages: %v", err)
	}
	for _, m := range messages {
		orderDetails, err := parser.GetOrderTotalFromMsg(r.Context(), h.GmailSrv, m)
		if err != nil {
			log.Println("error parsing details for message id:", m.Id, err)
		}
		fmt.Printf("%+v\n", orderDetails)
	}
}
