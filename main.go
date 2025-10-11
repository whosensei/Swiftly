package main

import (
	"fmt"
	"github/whosensei/shortenn/internal/database"
	"github/whosensei/shortenn/internal/handlers"
	"github/whosensei/shortenn/internal/middleware"
	"net/http"
	"os"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main(){

	_ = godotenv.Load()
	connStr := os.Getenv("DATABASE_URL")
	db := database.Connect(connStr)
	defer db.Close()
	
	mux := http.NewServeMux()
	handlers.RegisterRoute(mux,db)

	c := cors.AllowAll()
	loggedMux := c.Handler(middleware.Logger(mux))

	fmt.Println("The server started")
	http.ListenAndServe(":8080",loggedMux)
}