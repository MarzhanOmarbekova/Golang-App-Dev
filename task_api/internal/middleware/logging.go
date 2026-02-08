package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s request recieved",
			time.Now(),
			r.Method,
			r.URL.Path,
		)
		next.ServeHTTP(w, r)
	})
}
