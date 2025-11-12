package utils

import (
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"
)

func CorsInit() *cors.Cors {

	var c *cors.Cors
	env := os.Getenv("ENV")
	log.Printf("CORS initialized with ENV=%s", env)

	if env == "development" {
		c = cors.New(cors.Options{
			AllowedOrigins: []string{
				"http://localhost:3000",
			},
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodOptions,
			},
			AllowCredentials: true,
			AllowedHeaders: []string{
				"Authorization",
				"Content-Type",
				"X-Anonymous-token",
			},
		})
	} else {
		c = cors.New(cors.Options{
			AllowedOrigins: []string{
				"https://swftly.dev",
				"https://www.swftly.dev",
			},
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodOptions,
			},
			AllowCredentials: true,
			AllowedHeaders: []string{
				"Authorization",
				"Content-Type",
				"X-Anonymous-token",
			},
		})
	}

	return c
}
