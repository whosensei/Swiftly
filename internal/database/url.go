package database

import (
	"database/sql"
	"fmt"
	"github/whosensei/shortenn/internal/model"
	"log"
	"time"
)

// func checkifId(db *sql.DB, id int){
//  var u int
// 	query := `SELECT id FROM url WHERE id=$1`
// 	_,err := db.QueryRow(query,id).Scan(&u)
// 	return err;
// }

func URL_Add(db *sql.DB, u model.URL) error {
	query := `INSERT INTO url (id,shorturl,longurl,created_at) VALUES ($1,$2,$3,$4)`
	_, err := db.Exec(query, u.Id, u.Short_url, u.Long_url, time.Now())

	if err != nil {
		fmt.Println("failed to insert into database")
	}

	return err
}

func Redirect(db *sql.DB, shorturl string) string {
	var u string
	query := `SELECT longurl FROM url WHERE shorturl = $1`
	if err := db.QueryRow(query, shorturl).Scan(&u); err != nil {
		fmt.Println("Failed to fetch", err)
	}
	return u
}

func Getallurls(db *sql.DB) []model.URL{
	query := `SELECT id, shorturl,longurl,created_at FROM url`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Failed to get all the mapping", err)
	}

	var u []model.URL

	for rows.Next() {
		var url model.URL
		if err := rows.Scan(&url.Id, &url.Short_url, &url.Long_url,&url.Created_At); err != nil {
			log.Println("error occured", err)
			continue
		}
		u = append(u, url)
	}
	return u
}
