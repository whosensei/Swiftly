package model

import "time"

type User_request struct {
	Long_url string `json:"longurl"`
	User_id       int `json:"userId"`
}

type Api_response struct{
	Success bool `json:"success"`
	Message string `json:"message"`
	Data interface{} `json:"data"`
}

type Short_url struct {
	User_id int `json:"userId"`
	Short_url string `json:"shorturl"`
	Long_url string `json:"longurl"`
	Created_at time.Time  `json:"created_at"`
}
