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

	anonymous_token := r.Header.Get("X-Anonymous-Token")
	if anonymous_token == "" {
		anonymous_token = uuid.New().String()
	}

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

	userID := auth.GetUserId(r)
	email := auth.GetUserEmail(r)
	name := auth.GetUserName(r)

	uuid, err := database.EnsureUserExists(h.DB, userID, email, name)
	if err != nil {
		log.Printf("Failed to ensure user exists: %v", err)
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	short_code := utils.Url_shorten(id, longurl)

	err = database.Add_authenticated_url(h.DB, short_code, longurl, uuid, utils.GetClientIP(r))
	if err != nil {
		log.Printf("Failed to add URL to database: %v", err)
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	baseurl := os.Getenv("BACKEND_URL")
	if baseurl == "" {
		baseurl = "https://localhost:8080"
	}

	response := model.ShortenResponse{
		Data:      fmt.Sprintf("%s/%s", baseurl, short_code),
		Permanent: true,
		Shortcode: short_code,
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

}

func (h *UserHandler) Redirect_to_website(w http.ResponseWriter, r *http.Request) {

	short_code := r.PathValue("short_code")
	longurl, url_id, expires_at := database.Redirect(h.DB, short_code)

	metadata := utils.GetMetadata(longurl)
	fmt.Println(metadata)
	//add metadata to db

	fmt.Println(longurl)
	if longurl == "" {
		http.NotFound(w, r)
		return
	}

	if expires_at.Valid && time.Now().After(expires_at.Time) {
		http.Error(w, "The short url has expired", http.StatusGone)
		return
	}

	go func() {
		redis.IncrementClicks(short_code)

		user_IP := utils.GetClientIP(r)
		utils.GetClientLoc(user_IP)

		ua := r.UserAgent()
		details := utils.ParseUserAgent(ua)
		fmt.Println(details.Device,details.Browser,details.Platform)
		//get details for the clicks using Ip

		//country, city               done - device_type, browser, os

		_,err := h.DB.Exec(`
            INSERT INTO clicks (url_id, ip_address, user_agent, referer, device_type, browser, os)
            VALUES ($1, $2, $3, $4, $5,$6, $7)
        `, url_id, user_IP, ua, r.Referer(), details.Device, details.Browser, details.Platform)
		
		if err != nil {
			fmt.Println("Failed to add")
			fmt.Println(err)
		}
	}()

	http.Redirect(w, r, longurl, http.StatusFound)
}

func (h *UserHandler) Get_anon_urls(w http.ResponseWriter, r *http.Request) {

	anonymous_token := r.Header.Get("X-Anonymous-Token")
	if anonymous_token == "" {
		json.NewEncoder(w).Encode([]model.URL{})
	}

	anon_urls, err := database.Get_anon_urls(h.DB, anonymous_token)
	if err != nil {
		log.Fatal("failed to fetch urls from database")
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(anon_urls)
}

func (h *UserHandler) Get_auth_urls(w http.ResponseWriter, r *http.Request) {

	userID := auth.GetUserId(r)
	email := auth.GetUserEmail(r)
	name := auth.GetUserName(r)

	uuid, err := database.EnsureUserExists(h.DB, userID, email, name)
	if err != nil {
		log.Printf("Failed to ensure user exists: %v", err)
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	auth_urls, err := database.Get_auth_urls(h.DB, uuid)
	if err != nil {
		log.Printf("Failed to fetch URLs: %v", err)
		http.Error(w, "Failed to fetch URLs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(auth_urls)

}

func CleanupExpiredURLs(database *sql.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		result, err := database.Exec(`
            DELETE FROM urls WHERE expires_at IS NOT NULL AND expires_at < NOW()
        `)
		if err == nil {
			count, _ := result.RowsAffected()
			if count > 0 {
				log.Printf("Cleaned up %d expired URLs", count)
			}
		}
	}
}

func (h *UserHandler) Delete_url(w http.ResponseWriter, r *http.Request) {

	short_code := r.PathValue("short_code")
	if short_code == "" {
		http.Error(w, "short_code required", http.StatusBadRequest)
		return
	}

	userID := auth.GetUserId(r)
	email := auth.GetUserEmail(r)
	name := auth.GetUserName(r)

	if userID != "" {
		uuid, err := database.EnsureUserExists(h.DB, userID, email, name)
		if err != nil {
			http.Error(w, "Failed to verify the user", http.StatusInternalServerError)
			return
		}

		owned, err := database.Verify_auth_url_ownership(h.DB, short_code, uuid)
		if !owned || err != nil {
			http.Error(w, "URL not found or unauthorised", http.StatusForbidden)
			return
		}
	} else {
		anonymous_token := r.Header.Get("X-Anonymous-Token")
		if anonymous_token == "" {
			http.Error(w, "Unauthorised", http.StatusUnauthorized)
			return
		}

		owned, err := database.Verify_anon_url_ownership(h.DB, short_code, anonymous_token)
		if !owned || err != nil {
			http.Error(w, "URL not found or unauthorised", http.StatusForbidden)
			return
		}
	}

	err := database.Delete_url(h.DB, short_code)
	if err != nil {
		http.Error(w, "Failed to delete the URL", http.StatusBadRequest)
		fmt.Println("Failed to delete the URL")
		return
	}

	redis.ResetClickCount(short_code)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Url deleted successfully")
}
