package config

import (
	"context"
	"database/sql"
	"log"

	"github.com/vmilasin/chirpy/internal/database"
	"github.com/vmilasin/chirpy/internal/logger"
)

type ApiConfig struct {
	FileserverHits int
	DB             *sql.DB
	Queries        *database.Queries
	AppLogs        *logger.AppLogs
	JWTSecret      []byte
}

func NewApiConfig(db *sql.DB, queries *database.Queries, logFiles map[string]string, jwtSecret []byte) *ApiConfig {
	internalLogs := logger.InitiateLogs(logFiles)

	cfg := &ApiConfig{
		FileserverHits: 0,
		DB:             db,
		Queries:        queries,
		AppLogs:        internalLogs,
		JWTSecret:      jwtSecret,
	}

	loggerOutput := func() {
		output := `(
		Postgresql DB initialized,
		Log files initialized {
			system logs: %s
			database logs: %s 
			chirp logs: %s 
			user logs: %s
		}
		)`
		log.Printf(
			output,
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

func (cfg *ApiConfig) TransactionalQuery(ctx context.Context, txFunc func(tx *database.Queries) error) error {
	// Create a new transaction
	tx, err := cfg.DB.Begin()
	if err != nil {
		return err
	}

	// Create a new Qeuries instance with the transaction
	txQueries := cfg.Queries.WithTx(tx)

	// Ensure the transaction is rolled back if there’s an error
	defer func() {
		if err != nil {
			tx.Rollback() // Roll back if there was an error
		}
	}()

	// Execute the transaction function
	if err := txFunc(txQueries); err != nil {
		return err // Return the error to trigger the rollback
	}

	return tx.Commit()
}
