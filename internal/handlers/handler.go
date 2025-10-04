package handlers

import (
	"encoding/json"
	"github/whosensei/shortenn/internal/model"
	"github/whosensei/shortenn/internal/shortner"
	"net/http"
)

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	var u model.User_request
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Failed to get the body", http.StatusBadRequest)
		return
	}
	Short_url := shortner.Url_shorten(u.User_id,u.Long_url)

	w.Header().Set("content-type","application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.Api_response{
		Success: true,
		Message: "task executed successfully",
		Data: Short_url,
	})
}
