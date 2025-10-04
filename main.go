package main

import (
	"fmt"
	"github/whosensei/shortenn/internal/handlers"
	"github/whosensei/shortenn/internal/middleware"
	"net/http"
)

func main(){

	mux := http.NewServeMux()

	handlers.RegisterRoute(mux)

	loggedMux := middleware.Logger(mux)

	fmt.Println("The server started")
	http.ListenAndServe(":8080",loggedMux)
}