package logger

import (
	"log"
	"os"
)

type AppLogs struct {
	databaseLog      string
	databaseErrorLog string
	chirpLog         string
	chirpErrorLog    string
	userLog          string
	userErrorLog     string
}

// Open a file for writing logs
func openLogFile(path string) {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %s, ERROR: %s", path, err)
	}
	defer logFile.Close()
}

// Initiate all log files
func InitiateLogs(logFiles map[string]string) *AppLogs {
	openLogFile(logFiles["databaseLog"])
	openLogFile(logFiles["chirpLog"])
	openLogFile(logFiles["userLog"])

	appLogs := &AppLogs{
		databaseLog:      logFiles["databaseLog"],
		databaseErrorLog: logFiles["databaseErrorLog"],
		chirpLog:         logFiles["chirpLog"],
		chirpErrorLog:    logFiles["chirpErrorLog"],
		userLog:          logFiles["userLog"],
		userErrorLog:     logFiles["userErrorLog"],
	}

	return appLogs
}
