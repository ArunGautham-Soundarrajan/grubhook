package handlers

import "net/http"

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *StatsHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if err := h.Pool.Ping(r.Context()); err != nil { // or pool.Ping directly
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
