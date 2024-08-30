package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	// Configure fileserver root path and port to serve the site on
	const filepathRoot = "."
	const port = "8080"
	const dbFileName = "database.json"

	// --debug flag drops the table at the start for development purposes
	// WARNING: THIS DROPS THE DATABASE FILE!!!
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if *dbg {
		log.Print("DEBUG MODE INITIATED")
		// Check if the file already exists
		_, err := os.Stat(dbFileName)
		if err == nil {
			// The file exists
			err := os.Remove(dbFileName)
			checkError(err)
			log.Printf("Old database file %s removed", dbFileName)
		} else {
			log.Print("No old database to remove")
		}
	}

	// Initialize API config
	cfg, err := newApiConfig(dbFileName)
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
