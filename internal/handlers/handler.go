package handlers

import (
	"database/sql"
	"encoding/json"
	"time"

	// "github.com/redis/go-redis/v9"
	"fmt"
	"github/whosensei/shortenn/internal/auth"
	"github/whosensei/shortenn/internal/database"
	"github/whosensei/shortenn/internal/model"
	"github/whosensei/shortenn/internal/redis"
	"github/whosensei/shortenn/internal/utils"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type UserHandler struct {
	DB *sql.DB
	// redis *redis.Client
}

var (
	Anonymous_TTL    = time.Duration(30) * time.Minute // url expiry time 30 min
	Anonymous_Window = time.Duration(60) * time.Minute // 60 min rate limit window
	Anonymous_Limit  = 5                               // 5 url limit
)

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

	anonymous_token := r.Header.Get("Anonymous_Token")
	if anonymous_token == "" {
		anonymous_token = uuid.New().String()
	}

	//check ratelimits
	allowed, remaining, err := redis.CheckRateLimit(anonymous_token, 5, Anonymous_Window)

	if err != nil {
		log.Println("Failed to check the ratelimits", err)
	}
	if !allowed {
		log.Println("Ratelimits execeeded")
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	short_code := utils.Url_shorten(id, longurl)
	expires_at := time.Now().Add(30 * time.Minute)

	err = database.Add_anon_url(h.DB, short_code, longurl, anonymous_token, utils.GetClientIP(r), expires_at)
	if err != nil {
		log.Println("Failed to add to db", err)
		http.Error(w, "Failed to create the url", http.StatusInternalServerError)
		return
	}

	baseurl := os.Getenv("BACKEND_URL")
	if baseurl == "" {
		baseurl = "https://localhost:8080"
	}

	response := model.ShortenResponse{
		Data:            fmt.Sprintf("%s/%s", baseurl, short_code),
		Shortcode:       short_code,
		Expires_at:      expires_at,
		Anonymous_Token: anonymous_token,
		Remaining:       remaining - 1,
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) AuthenticatedShorten(w http.ResponseWriter, r *http.Request, longurl string, id string) {
	//using userID find uuid for user,
	// add links for that uuid in url table;

	// userID := auth.GetUserId(r);

	short_code := utils.Url_shorten(id, longurl)

	baseurl := os.Getenv("BACKEND_URL")
	if baseurl == "" {
		baseurl = "https://localhost:8080"
	}

	response := model.ShortenResponse{
		Data:      fmt.Sprintf("%s/%s", baseurl, short_code),
		Permanent: true,

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
