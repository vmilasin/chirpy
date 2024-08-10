package main

import (
	"net/http"
)

type apiConfig struct {
	FileserverHits int
}

/* MIDDLEWARE: */
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits += 1
		next.ServeHTTP(w, r)
	})
}
