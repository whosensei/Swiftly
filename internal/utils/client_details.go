package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type ClientLoc struct {
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
	RegionName  string `json:"regionName"`
	Continent   string `json:"continent"`
}

func GetClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0]) // return the first one (real client)
	}
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return r.RemoteAddr // fallback raw address
		}
		return ip
	}
	return ip
}

//using ipapi for location.

func GetClientLoc(ip string) (*ClientLoc, error) {
	// Request country, countryCode, city, regionName, continent
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=country,countryCode,city,regionName,continent", ip)
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d %s", res.StatusCode, res.Status)
	}

	var loc ClientLoc
	if err := json.NewDecoder(res.Body).Decode(&loc); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	fmt.Println(loc)

	return &loc, nil
}
