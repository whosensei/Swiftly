package handlers

import (
	"encoding/json"
	"fmt"
	"github/whosensei/shortenn/internal/model"
	"github/whosensei/shortenn/internal/shortner"
	"github/whosensei/shortenn/internal/store"
	"net/http"
	"time"
)

func ShortenURL(w http.ResponseWriter, r *http.Request) {

	var u model.User_request

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Failed to get the body", http.StatusBadRequest)
		return
	}

	Short_url := shortner.Url_shorten(u.Id,u.Long_url)

	var data = model.Short_url{
		Long_url: u.Long_url,
		Short_url: Short_url,
		Created_at: time.Now(),
	}
	res := store.Add_mapping(u.Id,data)
	
	if !res {
		http.Error(w,"Failed to save to DB",http.StatusBadGateway)
		return
	}
	redirect_link := fmt.Sprintf("%s/%s","http://localhost:8080",Short_url)

	w.Header().Set("content-type","application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.Api_response{
		Success: true,
		Message: "task executed successfully",
		Data: redirect_link,
	})
}

// func Redirect_to_website(w http.ResponseWriter, r* http.Request){
// 	id := r.PathValue("id")
// 	shorturl := r.PathValue("shorturl")
// 	shortlink := store.Redirect(id,shorturl)
// 	redirect_link := fmt.Sprintf("%s/%s","http://localhost:8080",shortlink)
// 	w.WriteHeader(http.StatusFound)
// 	http.Redirect(w,r,redirect_link,http.StatusFound)
// }