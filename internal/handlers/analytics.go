package handlers

import (
	"encoding/json"
	"net/http"

	"github/whosensei/shortenn/internal/database"
)

func (h *UserHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		http.Error(w, "short_code required", http.StatusBadRequest)
		return
	}

	data, err := database.GetAnalyticsBreakdownByShortCode(h.DB, shortCode)
	if err != nil {
		http.Error(w, "failed to fetch analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(data)
}
