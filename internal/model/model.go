package model

import "time"

type User_request struct {
	Long_url string `json:"longurl"`
}

type Api_response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type URL struct {
	Id              string    `json:"id"`
	Long_url        string    `json:"long_url"`
	Short_code       string   `json:"short_code"`
	Created_At      time.Time `json:"created_at"`
	Expires_at      time.Time `json:"expires_at"`
	Clicks          int64     `json:"clicks"`
	Last_clicked_at time.Time `json:"last_clicked_at"`
}

type ShortenResponse struct {
	Data            string    `json:"data"` //the shorturl complete
	Shortcode       string    `json:"short_code"`
	Created_at      time.Time `json:"created_at"`
	Expires_at      time.Time `json:"expires_at,omitempty"`
	Anonymous_Token string    `json:"anonymous_token,omitempty"`
	Remaining       int       `json:"remaining,omitempty"`
	Permanent       bool      `json:"permanent"`
}
