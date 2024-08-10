package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	// Initialize API config
	cfg := &apiConfig{
		FileserverHits: 0,
	}

	// ServeMux is an HTTP request router
	mux := http.NewServeMux()

	// Configure root path for your fileserver
	fileserver := http.FileServer(http.Dir(filepathRoot))

	/* HANDLER REGISTRATION (check custom_handlers.go for custom handlers): */
	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app", fileserver)))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("GET /api/reset", cfg.handlerMetricsReset)
	mux.HandleFunc("POST /api/validate_chirp", cfg.handlerValidateChirp)

	// A Server defines parameters for running an HTTP server
	// We use a pointer to specify the same server instance, instead of working with multiple copies (for each request)
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
