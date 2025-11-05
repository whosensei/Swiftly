package database

import (
	"database/sql"
	"fmt"
	"github/whosensei/shortenn/internal/model"
	"github/whosensei/shortenn/internal/redis"
	"log"
	"time"
)

// func checkifId(db *sql.DB, id int){
//  var u int
// 	query := `SELECT id FROM url WHERE id=$1`
// 	_,err := db.QueryRow(query,id).Scan(&u)
// 	return err;
// }

// func URL_Add(db *sql.DB, u model.URL) error {
// 	query := `INSERT INTO url (id,shorturl,longurl,created_at) VALUES ($1,$2,$3,$4)`
// 	_, err := db.Exec(query, u.Id, u.Short_url, u.Long_url, time.Now())

// 	if err != nil {
// 		fmt.Println("failed to insert into database")
// 	}

// 	return err
// }

func Redirect(db *sql.DB, shorturl string) string {
	var u string
	query := `SELECT longurl FROM url WHERE shorturl = $1`
	if err := db.QueryRow(query, shorturl).Scan(&u); err != nil {
		fmt.Println("Failed to fetch", err)
	}
	return u
}


func Add_anon_url(db *sql.DB, short_code string, long_url string, anonymous_token string, ip_address string, expires_at time.Time) error {
	query := `INSERT INTO urls (short_code,long_url,anonymous_token,ip_address,expires_at) VALUES ($1,$2,$3,$4,$5)`
	_, err := db.Exec(query, short_code, long_url, anonymous_token, ip_address, expires_at)
	return err
}

func Find_uuid_from_UserID(db *sql.DB,userID string) string {
	
	var u string;
	query := `SELECT uuid FROM users WHERE auth_user_id = $1`
	if err := db.QueryRow(query,userID).Scan(&u); err != nil {
		log.Fatal("failed to find uuid")
	}
	
	return u;

}

func Add_authenticated_url(db *sql.DB, short_code string , long_url string, userID string, ip_address string) error {
	query := `INSERT INTO urls (short_code, long_url, user_id, ip_address) VALUES ($1,$2,$3,$4)`
	_,err := db.Exec(query,short_code,long_url,userID,ip_address)
	return err;
}

func Get_anon_urls(db *sql.DB, anonymous_token string)([]model.URL,error){

	var u []model.URL

	query := `SELECT id, short_code, long_url, created_at, clicks, expires_at FROM urls WHERE anonymous_token=$1`

	rows,err := db.Query(query,anonymous_token)
	if err != nil {
		log.Fatal("Failed to get urls")
	}

	for rows.Next() {
		var data model.URL
		var dbClicks int
		rows.Scan(&data.Id,&data.Short_code,&data.Long_url,&dbClicks,&data.Expires_at)

		redisCounts,err := redis.GetClickCount(data.Short_code)
		if err != nil {
			log.Println("failed to get the redis count")
		}
		data.Clicks = int64(dbClicks)+redisCounts
		u = append(u, data)
	}
	return u, nil

}