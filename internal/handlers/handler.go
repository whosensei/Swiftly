package handlers

import (
	"database/sql"
	"encoding/json"
	// "github.com/redis/go-redis/v9"
	"fmt"
	"github/whosensei/shortenn/internal/auth"
	"github/whosensei/shortenn/internal/database"
	"github/whosensei/shortenn/internal/model"
	"github/whosensei/shortenn/internal/utils"
	"log"
	"net/http"
	"os"
)

type UserHandler struct {
	DB *sql.DB
	// redis *redis.Client
}

func (h *UserHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {

	var u model.User_request

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Failed to get the body", http.StatusBadRequest)
		return
	}

	id := utils.GenerateId()

	// determine whether the request is authenticated
	userID := auth.GetUserId(r)

	if userID == "" {
		log.Println("Anonymous called")
		h.AnonymousShorten(w, r, u.Long_url, id)
	} else {
		log.Println("Authenticated called")
		h.AuthenticatedShorten(w, r, u.Long_url, id)
	}
}

func (h *UserHandler) AnonymousShorten(w http.ResponseWriter, r *http.Request, longurl string, id string) {

	token := r.Header.Get("Anonymous_Token")
	//check ratelimits
	Short_url := utils.Url_shorten(id, longurl)
	//add to database, Short_url,long_url,token,expires_at,created_at
	baseurl := os.Getenv("BACKEND_URL")
	if baseurl == "" {
		baseurl = "https://localhost:8080"
	}

	response := model.ShortenResponse{
		Data: fmt.Sprintf("%s/%s", baseurl, Short_url),
		Anonymous_Token: token,
		//add rest
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) AuthenticatedShorten(w http.ResponseWriter, r *http.Request, longurl string, id string) {
	//using userID find uuid for user,
	// add links for that uuid in url table;

	Short_url := utils.Url_shorten(id, longurl)

	baseurl := os.Getenv("BACKEND_URL")
	if baseurl == "" {
		baseurl = "https://localhost:8080"
	}

	response := model.ShortenResponse{
		Data: fmt.Sprintf("%s/%s", baseurl, Short_url),
		//add rest
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

}

func (h *UserHandler) Redirect_to_website(w http.ResponseWriter, r *http.Request) {

	shorturl := r.PathValue("shorturl")
	longurl := database.Redirect(h.DB, shorturl)

	fmt.Println(longurl)
	if longurl == "" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, longurl, http.StatusFound)
}

func (h *UserHandler) Gettallmaps(w http.ResponseWriter, r *http.Request) {
	data := database.Getallurls(h.DB)
	w.Header().Set("content-type", "application-json")
	json.NewEncoder(w).Encode(data)
}
