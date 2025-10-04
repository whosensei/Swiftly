package main

import (
	"fmt"
	"github/whosensei/shortenn/internal/handlers"
	"net/http"
)

func main(){

	mux := http.NewServeMux()
	handlers.RegisterRoute(mux)
	fmt.Println("The server started")
	http.ListenAndServe(":8080",mux)
}