package main

import (
	"fmt"
	"github/whosensei/shortenn/internal/auth"
	"github/whosensei/shortenn/internal/database"
	"github/whosensei/shortenn/internal/handlers"
	"github/whosensei/shortenn/internal/middleware"
	"github/whosensei/shortenn/internal/redis"
	"github/whosensei/shortenn/internal/utils"
	"log"
	"net/http"
)

func main() {

	utils.LoadENV()

	db,err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to Database",err)
	}
	defer db.Close()

	go handlers.CleanupExpiredURLs(db)

    redisClient, err := redis.InitRedis()
    if err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }
    defer redisClient.Close()
	
    if err := auth.InitJWKS(); err != nil {
        log.Fatal("Failed to initialize JWKS:", err)
    }

	mux := http.NewServeMux()
	handlers.RegisterRoute(mux, db)

	c := utils.CorsInit()

	loggedMux := c.Handler(auth.JWTCheckMiddleware(middleware.Logger(mux)))

	fmt.Println("The server started")
	http.ListenAndServe(":8080", loggedMux)
}
