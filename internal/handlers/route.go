package handlers

import (
	"net/http"
)

func RegisterRoute(mux *http.ServeMux){
	mux.HandleFunc("POST /api/v1/shorten",ShortenURL)
	// mux.HandleFunc("GET /api/v1/redirect/{id}/{shorturl}",Redirect_to_website)
}