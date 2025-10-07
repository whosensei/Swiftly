package handlers

import (
	"net/http"
)

func RegisterRoute(mux *http.ServeMux) {
	mux.HandleFunc("POST /shorten", ShortenURL)
	mux.HandleFunc("GET /redirect/{shorturl}", Redirect_to_website)
	mux.HandleFunc("GET /getall" ,Gettallmaps)
}
