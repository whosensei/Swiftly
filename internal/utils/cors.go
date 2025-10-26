package utils

import (
	"net/http"
	"os"
	"github.com/rs/cors"
)

func CorsInit() *cors.Cors{
	
	var c *cors.Cors

	if os.Getenv("ENV") == "development" {
		c = cors.New(cors.Options{
			AllowedOrigins: []string{
				"http://localhost:3000",
			},
			AllowedMethods: []string{http.MethodGet,http.MethodPost},
			AllowCredentials: true,
			AllowedHeaders: []string{
				"Authorization",
				"Content-Type",
			},
		})
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

	return c
}

