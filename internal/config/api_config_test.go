package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/vmilasin/chirpy/internal/database"
)

var mutex = &sync.Mutex{}

// Helper function to check if a file exists
func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

// Create mock API configuration and DB and log files
func InitMockApiConfig() *ApiConfig {
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	// Database file paths
	mockDatabaseFiles := make(map[string]string, 2)
	mockDatabaseFiles["chirpDBFileName"] = filepath.Join(baseDir, "..", "..", "test", "db", "chirp_database_test.json")
	mockDatabaseFiles["userDBFileName"] = filepath.Join(baseDir, "..", "..", "test", "db", "user_database_test.json")
	// Log file paths
	mockLogFiles := make((map[string]string), 5)
	mockLogFiles["systemLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "system_test.log")
	mockLogFiles["handlerLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "handler_test.log")
	mockLogFiles["databaseLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "database_test.log")
	mockLogFiles["chirpLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "chirp_test.log")
	mockLogFiles["userLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "user_test.log")

	cfg := NewApiConfig(mockDatabaseFiles, mockLogFiles)
	return cfg
}

// Delete mock DB and log files
func TeardownMockApiConfig() []error {
	errorList := make([]error, 0, 7)

	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	// Database file paths
	mockDatabaseFiles := make(map[string]string, 2)
	mockDatabaseFiles["chirpDBFileName"] = filepath.Join(baseDir, "..", "..", "test", "db", "chirp_database_test.json")
	mockDatabaseFiles["userDBFileName"] = filepath.Join(baseDir, "..", "..", "test", "db", "user_database_test.json")
	// Log file paths
	mockLogFiles := make((map[string]string), 5)
	mockLogFiles["systemLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "system_test.log")
	mockLogFiles["handlerLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "handler_test.log")
	mockLogFiles["databaseLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "database_test.log")
	mockLogFiles["chirpLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "chirp_test.log")
	mockLogFiles["userLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "user_test.log")

	for index, file := range mockDatabaseFiles {
		err := os.Remove(file)
		if err != nil {
			output := fmt.Errorf("mock DB file %s was not removed from location %s: %s", index, file, err)
			errorList = append(errorList, output)
		}
	}

	for index, file := range mockLogFiles {
		err := os.Remove(file)
		if err != nil {
			output := fmt.Errorf("mock log file %s was not removed from location %s: %s", index, file, err)
			errorList = append(errorList, output)
		}
	}
	return errorList
}

func TestNewApiConfig(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockCFG := InitMockApiConfig()

	// DB files
	// Chirp DB
	exists := FileExists(mockCFG.AppDatabase.ChirpDB.Path())
	if !exists {
		t.Error("Chirp DB file creation failed")
	}
	chirpDBContent, err := os.ReadFile(mockCFG.AppDatabase.ChirpDB.Path())
	if err != nil {
		t.Fatalf("Failed to read ChirpDB file: %v", err)
	}
	var chirpDB database.ChirpDBStructure
	if err := json.Unmarshal(chirpDBContent, &chirpDB); err != nil {
		t.Errorf("Failed to unmarshal ChirpDB: %v", err)
	}
	if chirpDB.NextChirpID != 0 || len(chirpDB.Chirps) != 0 {
		t.Errorf("ChirpDB is not initialized correctly: %+v", chirpDB)
	}
	// User DB
	exists = FileExists(mockCFG.AppDatabase.UserDB.Path())
	if !exists {
		t.Error("User DB file creation failed")
	}
	userDBContent, err := os.ReadFile(mockCFG.AppDatabase.UserDB.Path())
	if err != nil {
		t.Fatalf("Failed to read UserDB file: %v", err)
	}
	var userDB database.UserDBStructure
	if err := json.Unmarshal(userDBContent, &userDB); err != nil {
		t.Errorf("Failed to unmarshal UserDB: %v", err)
	}
	if userDB.NextUserID != 0 || len(userDB.Users) != 0 || len(userDB.UserLookup) != 0 {
		t.Errorf("UserDB is not initialized correctly: %+v", userDB)
	}

	// Log files
	exists = FileExists(mockCFG.AppLogs.SystemLog)
	if !exists {
		t.Error("System log file creation failed")
	}
	exists = FileExists(mockCFG.AppLogs.HandlerLog)
	if !exists {
		t.Error("Handler log file creation failed")
	}
	exists = FileExists(mockCFG.AppLogs.DatabaseLog)
	if !exists {
		t.Error("DB log file creation failed")
	}
	exists = FileExists(mockCFG.AppLogs.ChirpLog)
	if !exists {
		t.Error("Chirp log file creation failed")
	}
	exists = FileExists(mockCFG.AppLogs.UserLog)
	if !exists {
		t.Error("User log file creation failed")
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}
