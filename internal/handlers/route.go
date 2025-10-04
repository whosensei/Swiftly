package handlers

import (
	"net/http"
)

func RegisterRoute(mux *http.ServeMux){
	mux.HandleFunc("POST /api/v1/shorten",ShortenURL)
}