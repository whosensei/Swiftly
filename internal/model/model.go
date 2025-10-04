package model

import "time"

type User_request struct {
	Long_url string `json:"longurl"`
	Id  string `json:"userId"`
}

type Api_response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type Short_url struct {
	Long_url   string    `json:"longurl"`
	Short_url  string    `json:"shorturl"`
	Created_at time.Time `json:"created_at"`
}
