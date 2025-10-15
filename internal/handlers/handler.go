package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github/whosensei/shortenn/internal/database"
	"github/whosensei/shortenn/internal/model"
	"github/whosensei/shortenn/internal/shortner"
	"net/http"
)
type UserHandler struct {
	DB *sql.DB
}

func(h *UserHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {

	var u model.User_request

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Failed to get the body", http.StatusBadRequest)
		return
	}
	id := shortner.GenerateId()

	Short_url := shortner.Url_shorten(id,u.Long_url)
	data := model.URL{Id:id,Long_url: u.Long_url,Short_url: Short_url}

	if err:= database.URL_Add(h.DB, data); err!=nil {
		fmt.Println("Failed to add to database")
	}
	redirect_link := fmt.Sprintf("%s/redirect/%s","https://swftly.dev",Short_url)

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.Api_response{
		Success: true,
		Message: "task executed successfully",
		Data:    redirect_link,
	})
}

func(h *UserHandler) Redirect_to_website(w http.ResponseWriter, r *http.Request) {

	shorturl := r.PathValue("shorturl")
	longurl := database.Redirect(h.DB,shorturl)

	fmt.Println(longurl)
	if longurl == "" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, longurl, http.StatusFound)
}


func(h * UserHandler) Gettallmaps(w http.ResponseWriter, r* http.Request){
	data := database.Getallurls(h.DB)
	w.Header().Set("content-type","application-json")
	json.NewEncoder(w).Encode(data)
}