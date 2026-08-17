package handlers

import (
	"net/http"
)

func (h *Handler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	total, err := h.Queries.SumLastNDays(r.Context(), 30)
	if err != nil {
		http.Error(w, "failed to compute total", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, total)
}
