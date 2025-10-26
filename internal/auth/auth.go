package auth

import (
	"context"
	"github/whosensei/shortenn/internal/utils"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

var jwks *keyfunc.JWKS

type contextKey string

const UserIDKey contextKey = "userID"

func InitJWKS() error {

	utils.LoadENV()
	betterAuthURL := os.Getenv("BETTER_AUTH_URL")
	if betterAuthURL == "" {
		betterAuthURL = "http://localhost:3000"
	}

	var err error
	jwks, err = keyfunc.Get(betterAuthURL+"/api/auth/jwks", keyfunc.Options{
		RefreshInterval:   1 * time.Hour,
		RefreshUnknownKID: true,
	})

	if err != nil {
		log.Fatal("failed to initalise JWKS")
		return err
	}

	log.Println("JWKS initialized")
	return nil
}

func JWTCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// warn if JWKS hasn't been initialized
		if jwks == nil {
			log.Println("JWKS not initialized; call InitJWKS() before using JWTCheckMiddleware")
		}
		authHeader := r.Header.Get("Authorization")

		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenString, jwks.Keyfunc,
				jwt.WithIssuer("http://localhost:3000"),
				jwt.WithAudience("http://localhost:3000"),
			)

			if err != nil {
				log.Printf("jwt.Parse error: %v", err)
			}

			if token == nil {
				log.Println("jwt.Parse returned nil token")
			} else {
				log.Printf("parsed token valid=%v", token.Valid)
			}

			if err == nil && token != nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if sub, ok := claims["sub"].(string); ok {
						r = r.WithContext(context.WithValue(r.Context(), UserIDKey, sub))
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func RequiredAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey)
		if userID == "" || userID == nil {
			http.Error(w, "User unauthorised", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUserId(r *http.Request) string {
	if userID, ok := r.Context().Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}
