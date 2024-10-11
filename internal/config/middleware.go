package config

import (
	"context"
	"errors"
	"net/http"

	"github.com/vmilasin/chirpy/internal/auth"
)

type contextKey string

const (
	ctxUserID contextKey = "userID"
)

/* MIDDLEWARE: */
// File access metrics
func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			err := errors.New("invalid or missing Authorization header")
			cfg.resolveAuthTokenError(w, err)
			return
		}

		userID, err := auth.AccessTokenAuthorization(tokenString, cfg.JWTSecret)
		if err != nil {
			cfg.resolveAuthTokenError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
