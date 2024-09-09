package main

import (
	"net/http"

	"github.com/vmilasin/chirpy/internal/database"
	"github.com/vmilasin/chirpy/internal/logger"
)

type apiConfig struct {
	FileserverHits int
	AppDatabase    *database.AppDatabase
	AppLogs        map[string]string
}

func newApiConfig(dbFiles, logFiles map[string]string) (*apiConfig, error) {
	// Initialize chirp DB
	internalDB, err := database.NewDB(dbFiles)
	if err != nil {
		return &apiConfig{}, err
	}

	internalLogs, err := logger.InitiateLogs(logFiles)

	cfg := &apiConfig{
		FileserverHits: 0,
		AppDatabase:    internalDB,
	}
	return cfg, nil
}

/* MIDDLEWARE: */
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits += 1
		next.ServeHTTP(w, r)
	})
}
