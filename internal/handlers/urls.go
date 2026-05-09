package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"regexp"
	"time"

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
}

var (
	Anonymous_TTL    = time.Duration(30) * time.Minute // url expiry time 30 min
	Anonymous_Window = time.Duration(60) * time.Minute // 60 min rate limit window
	Anonymous_Limit  = 5                               // 5 url limit

	// Valid custom slug pattern: alphanumeric, hyphens, underscores, 3-30 chars
	customSlugRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
)

// hashPassword creates a SHA-256 hash of a password
func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

// checkPassword verifies a password against a hash
func checkPassword(password string, hash string) bool {
	return hashPassword(password) == hash
}

// validateURL checks if a string is a valid URL
func validateURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func (h *UserHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {

	var u model.User_request

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Failed to get the body", http.StatusBadRequest)
		return
	}

	// Validate URL
	if u.Long_url == "" {
		sendJSONError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Auto-prepend https:// if no scheme
	if !regexp.MustCompile(`^https?://`).MatchString(u.Long_url) {
		u.Long_url = "https://" + u.Long_url
	}

	if !validateURL(u.Long_url) {
		sendJSONError(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	// Determine short code (custom alias or generated)
	var shortCode string
	if u.Custom_slug != "" {
		// Validate custom slug
		if !customSlugRegex.MatchString(u.Custom_slug) {
			sendJSONError(w, "Custom alias must be 3-30 characters, using only letters, numbers, hyphens, and underscores", http.StatusBadRequest)
			return
		}
		// Check if taken
		exists, err := database.CheckShortCodeExists(h.DB, u.Custom_slug)
		if err != nil {
			sendJSONError(w, "Failed to check alias availability", http.StatusInternalServerError)
			return
		}
		if exists {
			sendJSONError(w, "This custom alias is already taken", http.StatusConflict)
			return
		}
		shortCode = u.Custom_slug
	}

	id := utils.GenerateId()
	if shortCode == "" {
		shortCode = utils.Url_shorten(id, u.Long_url)
	}

	userID := auth.GetUserId(r)

	if userID == "" {
		log.Println("Anonymous called")
		h.AnonymousShorten(w, r, u.Long_url, shortCode)
	} else {
		log.Println("Authenticated called")
		h.AuthenticatedShorten(w, r, u, shortCode)
	}
}

func (h *UserHandler) AnonymousShorten(w http.ResponseWriter, r *http.Request, longurl string, shortCode string) {

	anonymous_token := r.Header.Get("X-Anonymous-Token")
	if anonymous_token == "" {
		anonymous_token = uuid.New().String()
	}

	allowed, remaining, err := redis.CheckRateLimit(anonymous_token, 5, Anonymous_Window)

	if err != nil {
		log.Println("Failed to check the ratelimits", err)
	}
	if !allowed {
		log.Println("Ratelimits exceeded")
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	expires_at := time.Now().Add(30 * time.Minute)

	err = database.Add_anon_url(h.DB, shortCode, longurl, anonymous_token, utils.GetClientIP(r), expires_at)
	if err != nil {
		log.Println("Failed to add to db", err)
		sendJSONError(w, "Failed to create the url", http.StatusInternalServerError)
		return
	}

	// Scrape metadata in background
	go func() {
		metadata := utils.GetMetadata(longurl)
		if metadata.Title != "" || metadata.Description != "" || metadata.ImageURL != "" {
			database.UpdateURLMetadata(h.DB, shortCode, metadata.Title, metadata.Description, metadata.ImageURL)
		}
	}()

	baseurl := os.Getenv("BACKEND_URL")
	if baseurl == "" {
		baseurl = "https://localhost:8080"
	}

	response := model.ShortenResponse{
		Data:            fmt.Sprintf("%s/%s", baseurl, shortCode),
		Shortcode:       shortCode,
		Expires_at:      expires_at,
		Anonymous_Token: anonymous_token,
		Remaining:       remaining - 1,
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) AuthenticatedShorten(w http.ResponseWriter, r *http.Request, req model.User_request, shortCode string) {

	userID := auth.GetUserId(r)
	email := auth.GetUserEmail(r)
	name := auth.GetUserName(r)

	uuid, err := redis.UUIDfromRedis(userID)
	if err != nil {
		uuid, err = database.EnsureUserExists(h.DB, userID, email, name)
		if err != nil {
			log.Printf("Failed to ensure user exists: %v", err)
			sendJSONError(w, "Failed to process request", http.StatusInternalServerError)
			return
		}
		go func() {
			if err := redis.CacheUserUUID(userID, uuid); err != nil {
				log.Printf("Failed to cache user UUID: %v", err)
			}
		}()
	}

	// Parse expiration
	var expiresAt *time.Time
	if req.Expires_at != "" {
		t, err := time.Parse(time.RFC3339, req.Expires_at)
		if err != nil {
			sendJSONError(w, "Invalid expiration date format. Use ISO 8601 (e.g., 2026-12-31T23:59:59Z)", http.StatusBadRequest)
			return
		}
		if t.Before(time.Now()) {
			sendJSONError(w, "Expiration date must be in the future", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}

	// Hash password if provided
	var passwordHash string
	if req.Password != "" {
		passwordHash = hashPassword(req.Password)
	}

	// Scrape metadata synchronously for authenticated (fast DB)
	metadata := utils.GetMetadata(req.Long_url)

	err = database.Add_authenticated_url(h.DB, shortCode, req.Long_url, uuid, utils.GetClientIP(r), expiresAt, passwordHash, metadata.Title, metadata.Description, metadata.ImageURL)
	if err != nil {
		log.Printf("Failed to add URL to database: %v", err)
		sendJSONError(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	// Handle tags
	if len(req.Tags) > 0 {
		urlID, err := database.GetURLIDByShortCode(h.DB, shortCode)
		if err == nil {
			for _, tagName := range req.Tags {
				tagID, err := database.GetOrCreateTag(h.DB, uuid, tagName)
				if err == nil {
					database.LinkTagToURL(h.DB, urlID, tagID)
				}
			}
		}
	}

	baseurl := os.Getenv("BACKEND_URL")
	if baseurl == "" {
		baseurl = "https://localhost:8080"
	}

	response := model.ShortenResponse{
		Data:      fmt.Sprintf("%s/%s", baseurl, shortCode),
		Permanent: expiresAt == nil,
		Shortcode: shortCode,
	}
	if expiresAt != nil {
		response.Expires_at = *expiresAt
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

}

func (h *UserHandler) Redirect_to_website(w http.ResponseWriter, r *http.Request) {

	short_code := r.PathValue("short_code")
	info, expires_at := database.Redirect(h.DB, short_code)

	if info.LongURL == "" {
		http.NotFound(w, r)
		return
	}

	if expires_at.Valid && time.Now().After(expires_at.Time) {
		frontendURL := os.Getenv("BETTER_AUTH_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}
		http.Redirect(w, r, fmt.Sprintf("%s/expired", frontendURL), http.StatusFound)
		return
	}

	// Password-protected link: redirect to password entry page
	if info.HasPassword {
		// Check if password provided in query param or header
		password := r.URL.Query().Get("password")
		if password == "" {
			password = r.Header.Get("X-Link-Password")
		}

		if password == "" {
			// Redirect to frontend password page
			frontendURL := os.Getenv("BETTER_AUTH_URL")
			if frontendURL == "" {
				frontendURL = "http://localhost:3000"
			}
			http.Redirect(w, r, fmt.Sprintf("%s/password-check/%s", frontendURL, short_code), http.StatusFound)
			return
		}

		if !checkPassword(password, info.PasswordHash) {
			// Wrong password — redirect back to password page with error
			frontendURL := os.Getenv("BETTER_AUTH_URL")
			if frontendURL == "" {
				frontendURL = "http://localhost:3000"
			}
			http.Redirect(w, r, fmt.Sprintf("%s/password-check/%s?error=incorrect", frontendURL, short_code), http.StatusFound)
			return
		}
	}

	go func() {
		redis.IncrementClicks(short_code)

		user_IP := utils.GetClientIP(r)
		loc, locErr := utils.GetClientLoc(user_IP)
		country := ""
		city := ""
		if locErr == nil && loc != nil {
			if loc.CountryCode != "" {
				country = loc.CountryCode
			} else {
				country = loc.Country
			}
			city = loc.City
		}

		ua := r.UserAgent()
		details := utils.ParseUserAgent(ua)

		_, err := h.DB.Exec(`
            INSERT INTO clicks (url_id, ip_address, country, city, user_agent, referer, device_type, browser, os)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        `, info.UrlID, user_IP, country, city, ua, r.Referer(), details.Device, details.Browser, details.Platform)

		if err != nil {
			log.Printf("Failed to record click: %v", err)
		}
	}()

	http.Redirect(w, r, info.LongURL, http.StatusFound)
}

// CheckPassword endpoint for password-protected links
func (h *UserHandler) CheckLinkPassword(w http.ResponseWriter, r *http.Request) {
	short_code := r.PathValue("short_code")
	if short_code == "" {
		sendJSONError(w, "short_code required", http.StatusBadRequest)
		return
	}

	var req model.PasswordCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	info, _ := database.Redirect(h.DB, short_code)
	if info.LongURL == "" {
		sendJSONError(w, "URL not found", http.StatusNotFound)
		return
	}

	if !info.HasPassword {
		// No password needed, return the URL
		w.Header().Set("content-type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"long_url": info.LongURL,
		})
		return
	}

	if !checkPassword(req.Password, info.PasswordHash) {
		sendJSONError(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	// Record the click
	go func() {
		redis.IncrementClicks(short_code)
	}()

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"long_url": info.LongURL,
	})
}

func (h *UserHandler) Get_anon_urls(w http.ResponseWriter, r *http.Request) {

	anonymous_token := r.Header.Get("X-Anonymous-Token")
	if anonymous_token == "" {
		w.Header().Set("Content-type", "application/json")
		json.NewEncoder(w).Encode([]model.URL{})
		return
	}

	anon_urls, err := database.Get_anon_urls(h.DB, anonymous_token)
	if err != nil {
		log.Printf("Failed to fetch anonymous urls: %v", err)
		sendJSONError(w, "Failed to fetch URLs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(anon_urls)
}

func (h *UserHandler) Get_auth_urls(w http.ResponseWriter, r *http.Request) {

	userID := auth.GetUserId(r)
	email := auth.GetUserEmail(r)
	name := auth.GetUserName(r)

	uuid, err := redis.UUIDfromRedis(userID)
	if err != nil {
		uuid, err = database.EnsureUserExists(h.DB, userID, email, name)
		if err != nil {
			log.Printf("Failed to ensure user exists: %v", err)
			sendJSONError(w, "Failed to process request", http.StatusInternalServerError)
			return
		}
		go func() {
			if err := redis.CacheUserUUID(userID, uuid); err != nil {
				log.Printf("Failed to cache user UUID: %v", err)
			}
		}()
	}

	auth_urls, err := database.Get_auth_urls(h.DB, uuid)
	if err != nil {
		log.Printf("Failed to fetch URLs: %v", err)
		sendJSONError(w, "Failed to fetch URLs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(auth_urls)

}

// UpdateURL handler - Full edit: URL, tags, password, expiry
func (h *UserHandler) UpdateURL(w http.ResponseWriter, r *http.Request) {
	short_code := r.PathValue("short_code")
	if short_code == "" {
		sendJSONError(w, "short_code required", http.StatusBadRequest)
		return
	}

	var req model.UpdateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate URL if provided
	if req.Long_url != "" {
		if !regexp.MustCompile(`^https?://`).MatchString(req.Long_url) {
			req.Long_url = "https://" + req.Long_url
		}
		if !validateURL(req.Long_url) {
			sendJSONError(w, "Invalid URL format", http.StatusBadRequest)
			return
		}
	}

	userID := auth.GetUserId(r)
	email := auth.GetUserEmail(r)
	name := auth.GetUserName(r)

	uuid, err := redis.UUIDfromRedis(userID)
	if err != nil {
		uuid, err = database.EnsureUserExists(h.DB, userID, email, name)
		if err != nil {
			sendJSONError(w, "Failed to process request", http.StatusInternalServerError)
			return
		}
	}

	// Build update params
	var longURL *string
	if req.Long_url != "" {
		longURL = &req.Long_url
	}

	// Validate and prepare new slug
	var newSlug *string
	if req.NewSlug != "" && req.NewSlug != short_code {
		if !customSlugRegex.MatchString(req.NewSlug) {
			sendJSONError(w, "Custom alias must be 3-30 characters (letters, numbers, hyphens, underscores)", http.StatusBadRequest)
			return
		}
		exists, err := database.CheckShortCodeExists(h.DB, req.NewSlug)
		if err != nil {
			sendJSONError(w, "Failed to check alias availability", http.StatusInternalServerError)
			return
		}
		if exists {
			sendJSONError(w, "This custom alias is already taken", http.StatusConflict)
			return
		}
		newSlug = &req.NewSlug
	}

	var passwordHash *string
	if req.Password != nil && *req.Password != "" {
		h := hashPassword(*req.Password)
		passwordHash = &h
	}

	var expiresAt *time.Time
	if req.Expires_at != nil && *req.Expires_at != "" {
		t, err := time.Parse(time.RFC3339, *req.Expires_at)
		if err != nil {
			sendJSONError(w, "Invalid expiration date format", http.StatusBadRequest)
			return
		}
		if t.Before(time.Now()) {
			sendJSONError(w, "Expiration date must be in the future", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}

	err = database.FullUpdateURL(h.DB, short_code, uuid, longURL, newSlug, passwordHash, req.RemovePassword, expiresAt, req.RemoveExpiry)
	if err != nil {
		log.Printf("Failed to update URL: %v", err)
		sendJSONError(w, "Failed to update URL or unauthorized", http.StatusForbidden)
		return
	}

	// Handle tags: clear existing and re-link
	if req.Tags != nil {
		urlID, err := database.GetURLIDByShortCode(h.DB, short_code)
		if err == nil {
			database.ClearTagsForURL(h.DB, urlID)
			for _, tagName := range req.Tags {
				tagID, err := database.GetOrCreateTag(h.DB, uuid, tagName)
				if err == nil {
					database.LinkTagToURL(h.DB, urlID, tagID)
				}
			}
		}
	}

	// Re-scrape metadata if URL changed
	if req.Long_url != "" {
		go func() {
			metadata := utils.GetMetadata(req.Long_url)
			if metadata.Title != "" || metadata.Description != "" || metadata.ImageURL != "" {
				database.UpdateURLMetadata(h.DB, short_code, metadata.Title, metadata.Description, metadata.ImageURL)
			}
		}()
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "URL updated successfully",
	})
}

// GetUserTags returns all tags for the authenticated user
func (h *UserHandler) GetUserTags(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserId(r)
	email := auth.GetUserEmail(r)
	name := auth.GetUserName(r)

	uuid, err := redis.UUIDfromRedis(userID)
	if err != nil {
		uuid, err = database.EnsureUserExists(h.DB, userID, email, name)
		if err != nil {
			sendJSONError(w, "Failed to process request", http.StatusInternalServerError)
			return
		}
	}

	tags, err := database.GetUserTags(h.DB, uuid)
	if err != nil {
		sendJSONError(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

// DeleteTag deletes a tag
func (h *UserHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	tagID := r.PathValue("tag_id")
	if tagID == "" {
		sendJSONError(w, "tag_id required", http.StatusBadRequest)
		return
	}

	userID := auth.GetUserId(r)
	email := auth.GetUserEmail(r)
	name := auth.GetUserName(r)

	uuid, err := redis.UUIDfromRedis(userID)
	if err != nil {
		uuid, err = database.EnsureUserExists(h.DB, userID, email, name)
		if err != nil {
			sendJSONError(w, "Failed to process request", http.StatusInternalServerError)
			return
		}
	}

	err = database.DeleteTag(h.DB, tagID, uuid)
	if err != nil {
		sendJSONError(w, "Failed to delete tag", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tag deleted successfully",
	})
}

func CleanupExpiredURLs(database *sql.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		count, err := database.Exec(`
            DELETE FROM urls WHERE expires_at IS NOT NULL AND expires_at < NOW()
        `)
		if err == nil {
			rows, _ := count.RowsAffected()
			if rows > 0 {
				log.Printf("Cleaned up %d expired URLs", rows)
			}
		}
	}
}

// SyncClickCounts syncs Redis click counts back to Postgres
func SyncClickCounts(db *sql.DB) {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		counts, err := redis.GetAllClickCounts()
		if err != nil {
			log.Printf("Failed to get click counts from Redis: %v", err)
			continue
		}

		for shortCode, clickCount := range counts {
			_, err := db.Exec(`
				UPDATE urls SET clicks = $1, last_clicked_at = NOW() WHERE short_code = $2
			`, clickCount, shortCode)
			if err != nil {
				log.Printf("Failed to sync clicks for %s: %v", shortCode, err)
			}
		}

		if len(counts) > 0 {
			log.Printf("Synced click counts for %d URLs", len(counts))
		}
	}
}

func (h *UserHandler) Delete_url(w http.ResponseWriter, r *http.Request) {

	short_code := r.PathValue("short_code")
	if short_code == "" {
		sendJSONError(w, "short_code required", http.StatusBadRequest)
		return
	}

	userID := auth.GetUserId(r)
	email := auth.GetUserEmail(r)
	name := auth.GetUserName(r)

	if userID != "" {
		uuid, err := redis.UUIDfromRedis(userID)
		if err != nil {
			uuid, err = database.EnsureUserExists(h.DB, userID, email, name)
			if err != nil {
				log.Printf("Failed to ensure user exists: %v", err)
				sendJSONError(w, "Failed to process request", http.StatusInternalServerError)
				return
			}
			go func() {
				if err := redis.CacheUserUUID(userID, uuid); err != nil {
					log.Printf("Failed to cache user UUID: %v", err)
				}
			}()
		}

		owned, err := database.Verify_auth_url_ownership(h.DB, short_code, uuid)
		if !owned || err != nil {
			sendJSONError(w, "URL not found or unauthorised", http.StatusForbidden)
			return
		}
	} else {
		anonymous_token := r.Header.Get("X-Anonymous-Token")
		if anonymous_token == "" {
			sendJSONError(w, "Unauthorised", http.StatusUnauthorized)
			return
		}

		owned, err := database.Verify_anon_url_ownership(h.DB, short_code, anonymous_token)
		if !owned || err != nil {
			sendJSONError(w, "URL not found or unauthorised", http.StatusForbidden)
			return
		}
	}

	err := database.Delete_url(h.DB, short_code)
	if err != nil {
		sendJSONError(w, "Failed to delete the URL", http.StatusBadRequest)
		return
	}

	redis.ResetClickCount(short_code)

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "URL deleted successfully",
	})
}

// sendJSONError sends a consistent JSON error response
func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
