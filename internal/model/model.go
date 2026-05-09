package model

import "time"

type User_request struct {
	Long_url    string   `json:"longurl"`
	Custom_slug string   `json:"custom_slug,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Expires_at  string   `json:"expires_at,omitempty"`
	Password    string   `json:"password,omitempty"`
}

type UpdateURLRequest struct {
	Long_url       string   `json:"longurl,omitempty"`
	NewSlug        string   `json:"new_slug,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Password       *string  `json:"password"`
	Expires_at     *string  `json:"expires_at"`
	RemovePassword bool     `json:"remove_password"`
	RemoveExpiry   bool     `json:"remove_expiry"`
}

type PasswordCheckRequest struct {
	Password string `json:"password"`
}

type Api_response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type URL struct {
	Id              string    `json:"id"`
	Long_url        string    `json:"long_url"`
	Short_code      string    `json:"short_code"`
	Created_At      time.Time `json:"created_at"`
	Expires_at      time.Time `json:"expires_at"`
	Clicks          int64     `json:"clicks"`
	Last_clicked_at time.Time `json:"last_clicked_at"`
	Title           string    `json:"title,omitempty"`
	Description     string    `json:"description,omitempty"`
	Og_image        string    `json:"og_image,omitempty"`
	Has_password    bool      `json:"has_password"`
	Tags            []Tag     `json:"tags,omitempty"`
}

type Tag struct {
	Id   string `json:"id"`
	Name string `json:"name"`
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

type AnalyticsGroupCount struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type AnalyticsTimeSeries struct {
	Date   string `json:"date"`
	Clicks int64  `json:"clicks"`
}

type AnalyticsBreakdown struct {
	Countries     []AnalyticsGroupCount `json:"countries"`
	Cities        []AnalyticsGroupCount `json:"cities"`
	Referrers     []AnalyticsGroupCount `json:"referrers"`
	Devices       []AnalyticsGroupCount `json:"devices"`
	Browsers      []AnalyticsGroupCount `json:"browsers"`
	OS            []AnalyticsGroupCount `json:"os"`
	TimeSeries    []AnalyticsTimeSeries `json:"timeseries"`
	TotalClicks   int64                 `json:"total_clicks"`
	LastClickedAt *time.Time            `json:"last_clicked_at,omitempty"`
}

type RedirectInfo struct {
	LongURL      string
	UrlID        string
	HasPassword  bool
	PasswordHash string
}
