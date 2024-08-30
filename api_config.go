package main

import (
	"net/http"

	"github.com/vmilasin/chirpy/internal/database"
)

type apiConfig struct {
	FileserverHits int
	db             *database.ChirpDB
}

func newApiConfig(dbFileName string) (*apiConfig, error) {
	// Initialize chirp DB
	internalDB, err := database.NewDB(dbFileName)
	if err != nil {
		return &apiConfig{}, err
	}

	cfg := &apiConfig{
		FileserverHits: 0,
		db:             internalDB,
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
