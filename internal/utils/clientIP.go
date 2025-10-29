package utils

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
        parts := strings.Split(ip, ",")
        return strings.TrimSpace(parts[0]) // return the first one (real client)
    }
	if ip == ""{
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == ""{
		 ip, _, err := net.SplitHostPort(r.RemoteAddr);
		 if err != nil {
			return r.RemoteAddr // fallback raw address
		}
		return ip
	}
	return ip
}
