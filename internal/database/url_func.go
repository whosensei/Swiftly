package database

import (
	"database/sql"
	"fmt"
	"github/whosensei/shortenn/internal/model"
	"github/whosensei/shortenn/internal/redis"
	"log"
	"time"
)


func Redirect(db *sql.DB, shorturl string) (string, string, sql.NullTime) {
	var u string
	var url_id string
	var expires_at sql.NullTime
	query := `SELECT long_url,id,expires_at FROM urls WHERE short_code = $1`
	if err := db.QueryRow(query, shorturl).Scan(&u, &url_id,&expires_at); err != nil {
		fmt.Println("Failed to fetch", err)
	}
	return u, url_id,expires_at
}

func Add_anon_url(db *sql.DB, short_code string, long_url string, anonymous_token string, ip_address string, expires_at time.Time) error {
	query := `INSERT INTO urls (short_code,long_url,anonymous_token,ip_address,expires_at) VALUES ($1,$2,$3,$4,$5)`
	_, err := db.Exec(query, short_code, long_url, anonymous_token, ip_address, expires_at)
	return err
}

func Find_uuid_from_UserID(db *sql.DB, userID string) string {

	var u string
	query := `SELECT id FROM users WHERE auth_user_id = $1`
	if err := db.QueryRow(query, userID).Scan(&u); err != nil {
		log.Fatal("failed to find uuid")
	}

	return u

}

func EnsureUserExists(db *sql.DB, authUserID string, email string, name string) (string, error) {
	var userUUID string

	query := `
		INSERT INTO users (auth_user_id, email, name, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (auth_user_id)
		DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	err := db.QueryRow(query, authUserID, email, name).Scan(&userUUID)
	if err != nil {
		return "", fmt.Errorf("failed to ensure user exists: %w", err)
	}

	return userUUID, nil
}

func Add_authenticated_url(db *sql.DB, short_code string, long_url string, userID string, ip_address string) error {
	query := `INSERT INTO urls (short_code, long_url, user_id, ip_address) VALUES ($1,$2,$3,$4)`
	_, err := db.Exec(query, short_code, long_url, userID, ip_address)
	return err
}

func Get_anon_urls(db *sql.DB, anonymous_token string) ([]model.URL,error) {

	var u []model.URL

	query := `SELECT id, short_code, long_url, created_at, clicks, expires_at FROM urls WHERE anonymous_token=$1 AND user_id is NULL AND (expires_at is NULL OR expires_at > NOW()) ORDER BY created_at DESC`

	rows, err := db.Query(query, anonymous_token)
	if err != nil {
		log.Fatal("Failed to get urls")
	}

	for rows.Next() {
		var data model.URL
		var dbClicks int
		rows.Scan(&data.Id, &data.Short_code, &data.Long_url,&data.Created_At, &dbClicks, &data.Expires_at)

		redisCounts, err := redis.GetClickCount(data.Short_code)
		if err != nil {
			log.Println("failed to get the redis count")
		}
		data.Clicks = int64(dbClicks) + redisCounts
		u = append(u, data)
	}
	return u,nil

}

func Get_auth_urls(db *sql.DB, user_id string)([]model.URL,error){

	query := "SELECT id,short_code,long_url,created_at,clicks FROM urls WHERE user_id = $1"

	rows, err := db.Query(query,user_id);
	if err != nil {
		log.Fatal("Failed to get urls")
	}
	var u []model.URL

	for rows.Next(){
		var data model.URL;
		var dbClicks int
		rows.Scan(&data.Id,&data.Short_code,&data.Long_url,&data.Created_At,&dbClicks)

		redisCount,err := redis.GetClickCount(data.Short_code)
		if err != nil {
			log.Println("Failed to get redis count")
		}
		data.Clicks = int64(dbClicks)+redisCount;
		u = append(u, data)
	}
	return u,nil
}

func Delete_url(db *sql.DB, short_code string) error{

	query := `DELETE * FROM urls WHERE short_code = $1`
	_,err := db.Exec(query,short_code)
	if err != nil {
		log.Println("Failed to delete")
		return err
	}
	return nil

}