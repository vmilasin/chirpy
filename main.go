package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/vmilasin/chirpy/internal/config"
	"github.com/vmilasin/chirpy/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Configure fileserver root path and port to serve the site on
	const filepathRoot = "."
	const port = "8080"
	// Log file paths
	logFiles := make((map[string]string), 5)
	logFiles["systemLog"] = filepath.Join(baseDir, "logs", "system.log")
	logFiles["handlerLog"] = filepath.Join(baseDir, "logs", "handler.log")
	logFiles["databaseLog"] = filepath.Join(baseDir, "logs", "database.log")
	logFiles["chirpLog"] = filepath.Join(baseDir, "logs", "chirp.log")
	logFiles["userLog"] = filepath.Join(baseDir, "logs", "user.log")

	// --debug flag drops the table at the start for development purposes
	// WARNING: THIS DROPS ALL DATABASE ENTRIES!!!
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		log.Print("DEBUG MODE INITIATED")
		dropFile(logFiles["systemLog"])
		dropFile(logFiles["handlerLog"])
		dropFile(logFiles["databaseLog"])
		dropFile(logFiles["chirpLog"])
		dropFile(logFiles["userLog"])
	}

	// Load env variables
	// Look for .env file in the current dir
	godotenv.Load()
	// Get the database connection URL
	dbURL := os.Getenv("DB_URL")
	// Open a connection to the DB
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Unable to open connection to the database: %v", err)
	}
	// Create an instance of Queries with the open db connection
	queries := database.New(db)
	// Get the JWT secret
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	// Get the PLATFORM value
	platform := os.Getenv("PLATFORM")
	// Get the key for Polka webhooks
	polkaKey := os.Getenv("POLKA_KEY")
	// Initialize API config
	cfg := config.NewApiConfig(db, queries, logFiles, jwtSecret, platform, polkaKey)

	if *dbg {
		cfg.Queries.TruncateAllTables(context.Background())
		log.Print("ALL DATABASE TABLES TRUNCATED")
	}

	// ServeMux is an HTTP request router
	mux := http.NewServeMux()

	// Configure root path for your fileserver
	fileserver := http.FileServer(http.Dir(filepathRoot))

	/* HANDLER REGISTRATION: */
	mux.Handle("/app/*", cfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileserver)))

	mux.HandleFunc("GET /api/healthz", cfg.HandlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.HandlerDBReset)
	mux.HandleFunc("GET /api/reset", cfg.HandlerMetricsReset)

	mux.HandleFunc("GET /api/chirps", cfg.HandlerChirpsGetAll)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.HandlerChirpsGetByID)

	mux.HandleFunc("POST /api/users", cfg.HandlerUserRegistration)
	mux.HandleFunc("POST /api/login", cfg.HandlerUserLogin)

	mux.Handle("POST /api/chirps", cfg.AuthTokenMiddleware(http.HandlerFunc(cfg.HandlerChirpsCreate)))
	mux.Handle("DELETE /api/chirps/{chirpID}", cfg.AuthTokenMiddleware((http.HandlerFunc(cfg.HandlerChirpsDelete))))
	mux.Handle("PUT /api/users", cfg.AuthTokenMiddleware(http.HandlerFunc(cfg.HandlerUserUpdate)))

	mux.Handle("POST /api/refresh", cfg.RefreshTokenMiddleware(http.HandlerFunc(cfg.HandlerRefreshTokenRefresh)))
	mux.Handle("POST /api/revoke", cfg.RefreshTokenMiddleware(http.HandlerFunc(cfg.HandlerRefreshTokenRevoke)))

	mux.Handle("POST /api/polka/webhooks", cfg.PolkaMiddleware(http.HandlerFunc(cfg.HandlerWebhooksPolkaEnableChirpyRed)))

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
