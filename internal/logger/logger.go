package logger

import (
	"log"
	"os"
	"sync"
)

type AppLogs struct {
	SystemLog   string
	HandlerLog  string
	DatabaseLog string
	ChirpLog    string
	UserLog     string
	mux         *sync.RWMutex
}

var logInitMutex sync.Mutex

// Open a file for writing logs
func initLog(path string) error {
	logInitMutex.Lock()
	defer logInitMutex.Unlock()

	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to initialize log file: %s, ERROR: %s", path, err)
		return err
	}
	defer logFile.Close()
	return nil
}

// Write logs to a log file
func (logs *AppLogs) LogToFile(path string, output func()) error {
	logs.mux.Lock()
	defer logs.mux.Unlock()

	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %s, ERROR: %s", path, err)
		return err
	}
	defer logFile.Close()
	defer log.SetOutput(os.Stdout)

	log.SetOutput(logFile)
	output()
	return nil
}

// Initiate all log files
func InitiateLogs(logFiles map[string]string) *AppLogs {
	initLog(logFiles["systemLog"])
	initLog(logFiles["handlerLog"])
	initLog(logFiles["databaseLog"])
	initLog(logFiles["chirpLog"])
	initLog(logFiles["userLog"])

	appLogs := &AppLogs{
		SystemLog:   logFiles["systemLog"],
		HandlerLog:  logFiles["handlerLog"],
		DatabaseLog: logFiles["databaseLog"],
		ChirpLog:    logFiles["chirpLog"],
		UserLog:     logFiles["userLog"],
		mux:         &sync.RWMutex{},
	}

	return appLogs
}

/* Unnecessary so far
func (logs *AppLogs) CurrentTimestamp() string {
	// Get the current timestamp
	now := time.Now()

	// Load the CET timezone
	location, err := time.LoadLocation("CET")
	if err != nil {
		log.Printf("Error loading location: %s", err)
		return ""
	}

	// Convert the current time to CET
	cetTime := now.In(location)

	// Format the time as needed (e.g., RFC3339 format)
	cetTimeString := cetTime.Format(time.RFC3339)

	return cetTimeString
}
*/
