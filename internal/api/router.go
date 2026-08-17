package api

import (
	"net/http"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/handlers"
)

func NewRouter(h *handlers.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handlers.Healthz)
	mux.HandleFunc("GET /readyz", h.Readyz)
	mux.HandleFunc("GET /stats", h.GetMonthlyStats)
	mux.HandleFunc("GET /sync", h.Sync)

	return mux
}
