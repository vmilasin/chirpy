package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/vmilasin/chirpy/internal/config"
)

func main() {
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Configure fileserver root path and port to serve the site on
	const filepathRoot = "."
	const port = "8080"
	// Database file paths
	databaseFiles := make(map[string]string, 2)
	databaseFiles["chirpDBFileName"] = "chirp_database.json"
	databaseFiles["userDBFileName"] = "user_database.json"
	// Log file paths
	logFiles := make((map[string]string), 5)
	logFiles["systemLog"] = filepath.Join(baseDir, "logs", "system.log")
	logFiles["handlerLog"] = filepath.Join(baseDir, "logs", "handler.log")
	logFiles["databaseLog"] = filepath.Join(baseDir, "logs", "database.log")
	logFiles["chirpLog"] = filepath.Join(baseDir, "logs", "chirp.log")
	logFiles["userLog"] = filepath.Join(baseDir, "logs", "user.log")

	// --debug flag drops the table at the start for development purposes
	// WARNING: THIS DROPS THE DATABASE FILE!!!
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if *dbg {
		log.Print("DEBUG MODE INITIATED")
		dropFile(databaseFiles["chirpDBFileName"])
		dropFile(databaseFiles["userDBFileName"])
		dropFile(logFiles["systemLog"])
		dropFile(logFiles["handlerLog"])
		dropFile(logFiles["databaseLog"])
		dropFile(logFiles["chirpLog"])
		dropFile(logFiles["userLog"])
	}

	// Initialize API config
	cfg := config.NewApiConfig(databaseFiles, logFiles)

	// ServeMux is an HTTP request router
	mux := http.NewServeMux()

	// Configure root path for your fileserver
	fileserver := http.FileServer(http.Dir(filepathRoot))

	/* HANDLER REGISTRATION: */
	mux.Handle("/app/*", cfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileserver)))
	mux.HandleFunc("GET /api/healthz", config.HandlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("GET /api/reset", cfg.HandlerMetricsReset)
	mux.HandleFunc("GET /api/chirps", cfg.HandlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{id}", cfg.HandlerGetChirp)
	mux.HandleFunc("POST /api/chirps", cfg.HandlerPostChirp)
	mux.HandleFunc("POST /api/users", cfg.HandlerPostUser)
	mux.HandleFunc("POST /api/login", cfg.HandlerLogin)

	// Server parameters
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving at %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

// Drop files in debug mode
func dropFile(path string) {
	// Check if the file already exists
	_, err := os.Stat(path)
	if err == nil {
		// The file exists
		err := os.Remove(path)
		if err != nil {
			log.Printf("DEBUG: There was an issue when trying to remove the old file: %s", path)
		}
		log.Printf("DEBUG: old file %s successfully removed.", path)
	} else {
		log.Printf("DEBUG: No old file with path %s to remove", path)
	}
}
