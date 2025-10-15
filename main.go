package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github/whosensei/shortenn/internal/database"
	"github/whosensei/shortenn/internal/handlers"
	"github/whosensei/shortenn/internal/middleware"
	"net/http"
	"os"
)

func main() {

	_ = godotenv.Load()
	connStr := os.Getenv("DATABASE_URL")
	db := database.Connect(connStr)
	defer db.Close()

	mux := http.NewServeMux()
	handlers.RegisterRoute(mux, db)

	var c *cors.Cors

	if os.Getenv("ENV") == "development" {
		c = cors.AllowAll()
	} else {
		c = cors.New(cors.Options{
			AllowedOrigins: []string{
				"https://swftly.dev",
				"https://www.swftly.dev",
			},
			AllowedMethods:   []string{http.MethodGet, http.MethodPost},
			AllowCredentials: true,
			AllowedHeaders: []string{
				"Authorization",
				"Content-Type",
			},
		})
	}

	loggedMux := c.Handler(middleware.Logger(mux))

	fmt.Println("The server started")
	http.ListenAndServe(":8080", loggedMux)
}
