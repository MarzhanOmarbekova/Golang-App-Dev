package middleware

import (
	"encoding/json"
	"net/http"
)

const apiKey = "secret12345"

func APIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "unauthorized",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
