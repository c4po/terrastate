package handlers

import (
	"net/http"
)

var (
	tokens = make(map[string]bool)
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "No token provided", http.StatusUnauthorized)
			return
		}

		if !tokens[token] {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
