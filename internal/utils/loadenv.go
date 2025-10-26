package utils

import (
	"log"
	"github.com/joho/godotenv"
)

func LoadENV(){
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, loading from system environment")
	}
}