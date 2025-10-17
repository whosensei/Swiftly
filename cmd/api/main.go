package main

import (
	"fmt"
	"github/whosensei/shortenn/internal/database"
	"github/whosensei/shortenn/internal/handlers"
	"github/whosensei/shortenn/internal/middleware"
	"github/whosensei/shortenn/internal/redis"
	"github/whosensei/shortenn/internal/utils"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()
	connStr := os.Getenv("DATABASE_URL")
	db := database.Connect(connStr)
	defer db.Close()

	if err := redis.InitRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connected")

	mux := http.NewServeMux()
	handlers.RegisterRoute(mux, db)

	c := utils.CorsInit()

	loggedMux := c.Handler(middleware.Logger(mux))

	fmt.Println("The server started")
	http.ListenAndServe(":8080", loggedMux)
}
