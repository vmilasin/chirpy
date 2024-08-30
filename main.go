package main

import (
	"log"
	"net/http"
)

func main() {
	// Configure fileserver root path and port to serve the site on
	const filepathRoot = "."
	const port = "8080"

	// Initialize API config
	cfg, err := newApiConfig()
	checkError(err)

	// ServeMux is an HTTP request router
	mux := http.NewServeMux()

	// Configure root path for your fileserver
	fileserver := http.FileServer(http.Dir(filepathRoot))

	/* HANDLER REGISTRATION: */
	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app", fileserver)))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("GET /api/reset", cfg.handlerMetricsReset)
	mux.HandleFunc("GET /api/chirps", cfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{id}", cfg.handlerGetChirp)
	mux.HandleFunc("POST /api/chirps", cfg.handlerPostChirp)
	mux.HandleFunc("POST /api/users", cfg.handlerPostUser)

	// Server parameters
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving at %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

// Log error checks
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
