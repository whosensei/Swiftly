package handlers

import (
	"database/sql"
	"net/http"
)

func RegisterRoute(mux *http.ServeMux , db *sql.DB) {

	userHandler := &UserHandler{DB:db}

	mux.HandleFunc("POST /shorten", userHandler.ShortenURL)
	mux.HandleFunc("GET /{shorturl}", userHandler.Redirect_to_website)
	mux.HandleFunc("GET /getall" ,userHandler.Gettallmaps)
}
