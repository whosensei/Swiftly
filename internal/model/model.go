package model

type User_request struct {
	Id int `json:"id"`
	Long_url string `json:"longurl"`
}

type Api_response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type URL struct {
	Id         int    `json:"id"`
	Long_url   string `json:"longurl"`
	Short_url  string `json:"shorturl"`
	Created_At string `json:"created_at"`
}
