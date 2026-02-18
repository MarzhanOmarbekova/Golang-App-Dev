package middleware

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s",
			time.Now().Format(time.RFC3339),
			r.Method,
			r.RequestURI,
		)
		next.ServeHTTP(w, r)
	})
}
