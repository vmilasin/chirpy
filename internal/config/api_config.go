package config

import (
	"log"
	"net/http"

	"github.com/vmilasin/chirpy/internal/database"
	"github.com/vmilasin/chirpy/internal/logger"
)

type ApiConfig struct {
	FileserverHits int
	AppDatabase    *database.AppDatabase
	AppLogs        *logger.AppLogs
}

func NewApiConfig(dbFiles, logFiles map[string]string) *ApiConfig {
	// Initialize chirp DB
	internalDB := database.NewDB(dbFiles)
	internalLogs := logger.InitiateLogs(logFiles)

	cfg := &ApiConfig{
		FileserverHits: 0,
		AppDatabase:    internalDB,
		AppLogs:        internalLogs,
	}

	loggerOutput := func() {
		output := `(%s) (
		Database files initialized {
			chirp database: %s
			user database: %s
		}
		Log files initialized {
			system logs: %s
			database logs: %s 
			chirp logs: %s 
			user logs: %s
		}
		)`
		log.Printf(
			output,
			cfg.AppLogs.CurrentTimestamp(),
			cfg.AppDatabase.ChirpDB.Path(),
			cfg.AppDatabase.UserDB.Path(),
			cfg.AppLogs.SystemLog,
			cfg.AppLogs.DatabaseLog,
			cfg.AppLogs.ChirpLog,
			cfg.AppLogs.UserLog,
		)
	}
	err := cfg.AppLogs.LogToFile(cfg.AppLogs.SystemLog, loggerOutput)
	if err != nil {
		log.Printf("Error logging to file %s: %s", cfg.AppLogs.SystemLog, err)
	}

	return cfg
}

/* MIDDLEWARE: */
func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits += 1
		next.ServeHTTP(w, r)
	})
}
