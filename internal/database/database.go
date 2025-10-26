package database

import (
	"database/sql"
	"github/whosensei/shortenn/internal/utils"
	"log"
	"os"
	_ "github.com/jackc/pgx/v5/stdlib"
)


func InitDB() (*sql.DB,error) {

	utils.LoadENV()

	dbUrl := os.Getenv("DATABASE_URL");
	if dbUrl == "" {
		dbUrl = "postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable"
	}

	db,err := sql.Open("pgx",dbUrl)

	if err != nil {
		log.Fatal("Failed to connect to database",err)
		return nil,err
	}

	if err := db.Ping() ;err != nil {
		log.Fatal("Failed to connect")
		return nil,err
	}

	log.Println("Database connected")
	return db,nil
}