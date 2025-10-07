package handlers

import (
	"encoding/json"
	"fmt"
	"github/whosensei/shortenn/internal/model"
	"github/whosensei/shortenn/internal/shortner"
	"github/whosensei/shortenn/internal/store"
	"net/http"
)

func ShortenURL(w http.ResponseWriter, r *http.Request) {

	var u model.User_request

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Failed to get the body", http.StatusBadRequest)
		return
	}

	Short_url := shortner.Url_shorten(u.Long_url)

	store.Add_mapping(Short_url, u.Long_url)

	redirect_link := fmt.Sprintf("%s/redirect/%s", "http://localhost:8080", Short_url)

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.Api_response{
		Success: true,
		Message: "task executed successfully",
		Data:    redirect_link,
	})
}

func Redirect_to_website(w http.ResponseWriter, r *http.Request) {

	shorturl := r.PathValue("shorturl")
	longurl := store.Redirect(shorturl)

	fmt.Println(longurl)
	if longurl == "" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, longurl, http.StatusFound)
}


func Gettallmaps(w http.ResponseWriter, r* http.Request){
	data := store.Getallmaps()
	w.Header().Set("content-type","application-json")
	json.NewEncoder(w).Encode(data)
}