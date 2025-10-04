package middleware

import (
	"fmt"
	"net/http"
	"time"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	Status_code int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w,http.StatusOK}
}

func (lrw * LoggingResponseWriter) WriteHeader(code int) {
	lrw.Status_code = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r* http.Request){
		start := time.Now()
		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw,r)

		fmt.Printf("[%s] %s %d %s \n",r.Method,r.URL.Path,lrw.Status_code,time.Since(start))

	})
}