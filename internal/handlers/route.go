package handlers

import (
	"database/sql"
	"net/http"
)


func RegisterRoute(mux *http.ServeMux, db *sql.DB) {

	userHandler := &UserHandler{DB: db}

	mux.HandleFunc("POST /shorten", userHandler.ShortenURL)
	mux.HandleFunc("GET /urls/anonymous", userHandler.Get_anon_urls)
	mux.HandleFunc("GET /urls/authenticated", userHandler.Get_auth_urls)
	mux.HandleFunc("DELETE /urls/delete/{short_code}",userHandler.Delete_url)
	mux.HandleFunc("GET /{short_code}", userHandler.Redirect_to_website)  // Keep this last!
}
