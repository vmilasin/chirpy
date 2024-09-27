package config

import (
	"context"
	"net/http"
)

/* MIDDLEWARE: */
// File access metrics
func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

// Auth middleware - extract user ID from JWT and set it in context
func (cfg *ApiConfig) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate extracting user ID from JWT
		userID := "12345" // Replace with actual JWT parsing logic

		// Create a new context with the user ID
		ctx := context.WithValue(r.Context(), "userID", userID)

		// Pass the new context to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
