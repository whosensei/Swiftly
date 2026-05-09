package handlers

import (
	"database/sql"
	"net/http"
)

func RegisterRoute(mux *http.ServeMux, db *sql.DB) {

	userHandler := &UserHandler{DB: db}

	// URL operations
	mux.HandleFunc("POST /shorten", userHandler.ShortenURL)
	mux.HandleFunc("GET /urls/anonymous", userHandler.Get_anon_urls)
	mux.HandleFunc("GET /urls/authenticated", userHandler.Get_auth_urls)
	mux.HandleFunc("PUT /urls/update/{short_code}", userHandler.UpdateURL)
	mux.HandleFunc("DELETE /urls/delete/{short_code}", userHandler.Delete_url)

	// Analytics
	mux.HandleFunc("GET /analytics/{short_code}", userHandler.GetAnalytics)

	// Tags
	mux.HandleFunc("GET /tags", userHandler.GetUserTags)
	mux.HandleFunc("DELETE /tags/{tag_id}", userHandler.DeleteTag)

	// Password-protected link check
	mux.HandleFunc("POST /check-password/{short_code}", userHandler.CheckLinkPassword)

	// Redirect (keep this last - catches all single-segment paths)
	mux.HandleFunc("GET /{short_code}", userHandler.Redirect_to_website)
}
