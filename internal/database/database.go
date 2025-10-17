package database

import (
	"database/sql"
	"log"
	_ "github.com/jackc/pgx/v5/stdlib"
)


func Connect(connStr string) *sql.DB {

	db,err := sql.Open("pgx",connStr)

	if err != nil {
		log.Fatal("Failed to connect to database",err)
	}

	if err := db.Ping() ;err != nil {
		log.Fatal("Failed to connect")
	}

	log.Println("Database connected")
	return db
}