package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/vmilasin/chirpy/internal/auth"
)

type contextKey string

const (
	ctxUserID                contextKey = "userID"
	ctxRefreshToken          contextKey = "refreshToken"
	ctxRefreshTokenRevokedAt contextKey = "refreshTokenRevokedAt"
)

/* MIDDLEWARE: */
// File access metrics
func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			err := errors.New("invalid or missing Authorization header")
			cfg.resolveAuthTokenError(w, err)
			return
		}

		userID, err := auth.AccessTokenAuth(tokenString, cfg.JWTSecret)
		if err != nil {
			cfg.resolveAuthTokenError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (cfg *ApiConfig) RefreshTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			cfg.respondWithError(w, http.StatusUnauthorized, "Invalid or missing refresh token.")
			return
		}

		var token string
		if strings.HasPrefix(tokenString, "Bearer ") {
			token = strings.TrimPrefix(tokenString, "Bearer ")
			token = strings.TrimSpace(token)
		} else {
			cfg.respondWithError(w, http.StatusUnauthorized, "Invalid or missing Authorization header.")
		}

		returnedToken, err := cfg.Queries.CheckRefreshTokenValidity(r.Context(), token)
		if err != nil {
			if err == sql.ErrNoRows {
				cfg.respondWithError(w, http.StatusUnauthorized, "No refresh token found.")
			} else {
				output := func() {
					log.Printf("An error occured during refresh token validation: %s.", err)
				}
				cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
				cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error occured during refresh token validation: %s.", err))
				return
			}
		}

		ctx := context.WithValue(r.Context(), ctxRefreshToken, token)
		ctx = context.WithValue(ctx, ctxUserID, returnedToken.UserID)
		ctx = context.WithValue(ctx, ctxRefreshTokenRevokedAt, returnedToken.RevokedAt)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (cfg *ApiConfig) PolkaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providedPolkaKey := r.Header.Get("Authorization")
		if providedPolkaKey == "" {
			cfg.respondWithError(w, http.StatusUnauthorized, "Invalid or missing Polka key.")
			return
		}

		var key string
		if strings.HasPrefix(providedPolkaKey, "ApiKey ") {
			key = strings.TrimPrefix(providedPolkaKey, "ApiKey ")
			key = strings.TrimSpace(key)
		} else {
			cfg.respondWithError(w, http.StatusUnauthorized, "Invalid Polka key.")
			return
		}

		if key != cfg.PolkaKey {
			cfg.respondWithError(w, http.StatusUnauthorized, "Invalid Polka key.")
			return
		}

		next.ServeHTTP(w, r)
	})
}
