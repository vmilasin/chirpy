package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/vmilasin/chirpy/internal/config"
	"github.com/vmilasin/chirpy/internal/database"
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
	// WARNING: THIS DROPS THE DATABASE FILE!!!
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
	// Initialize API config
	cfg := config.NewApiConfig(db, queries, logFiles, jwtSecret)

	// ServeMux is an HTTP request router
	mux := http.NewServeMux()

	// Configure root path for your fileserver
	fileserver := http.FileServer(http.Dir(filepathRoot))

	/* HANDLER REGISTRATION: */
	//http.Handle("/user", authMiddleware(http.HandlerFunc(userHandler)))
	mux.Handle("/app/*", cfg.MiddlewareMetricsInc(http.StripPrefix("/app", fileserver)))
	mux.HandleFunc("GET /api/healthz", cfg.HandlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("GET /api/reset", cfg.HandlerMetricsReset)

	mux.HandleFunc("GET /api/chirps", cfg.HandlerChirpsGetAll)
	mux.HandleFunc("GET /api/chirps/{id}", cfg.HandlerChirpsGetByID)

	mux.HandleFunc("POST /api/users", cfg.HandlerUserRegistration)
	mux.HandleFunc("POST /api/login", cfg.HandlerUserLogin)

	mux.Handle("POST /api/chirps", cfg.AuthMiddleware(http.HandlerFunc(cfg.HandlerChirpsCreate)))
	//mux.HandleFunc("POST /api/refresh", cfg.HandlerRefreshToken)
	//mux.HandleFunc("POST /api/revoke", cfg.HandlerRevokeToken)
	mux.Handle("PUT /api/users", cfg.AuthMiddleware(http.HandlerFunc(cfg.HandlerUserUpdate)))

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
