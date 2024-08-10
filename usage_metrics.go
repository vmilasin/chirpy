package main

import (
	"html/template"
	"net/http"
	"strconv"
)

// Usage metrics
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	t, err := template.ParseFiles("assets/html_templates/metrics.gohtml")
	checkError(err)

	err = t.Execute(w, cfg)
	checkError(err)
}

// Usage metrics reset
func (cfg *apiConfig) handlerMetricsReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	cfg.FileserverHits = 0
	output := "Hits: " + strconv.Itoa(cfg.FileserverHits)
	w.Write([]byte(output))
}
