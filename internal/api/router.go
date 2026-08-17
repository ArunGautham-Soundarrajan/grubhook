package api

import (
	"net/http"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/handlers"
)

func NewRouter(stats *handlers.StatsHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handlers.Healthz)
	mux.HandleFunc("GET /readyz", stats.Readyz)
	mux.HandleFunc("GET /stats", stats.GetMonthlyStats)

	return mux
}
