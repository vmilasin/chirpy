package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to check if a file exists
func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

// Create mock DB configuration and DB files
func InitMockDB() *AppDatabase {
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	// Database file paths
	mockDatabaseFiles := make(map[string]string, 2)
	mockDatabaseFiles["chirpDBFileName"] = filepath.Join(baseDir, "..", "..", "test", "db", "chirp_database_test.json")
	mockDatabaseFiles["userDBFileName"] = filepath.Join(baseDir, "..", "..", "test", "db", "user_database_test.json")

	mockDB := NewDB(mockDatabaseFiles)
	return mockDB
}

// Delete mock DB files
func TeardownMockDB() []error {
	errorList := make([]error, 0, 2)

	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	// Database file paths
	mockDatabaseFiles := make(map[string]string, 2)
	mockDatabaseFiles["chirpDBFileName"] = filepath.Join(baseDir, "..", "..", "test", "db", "chirp_database_test.json")
	mockDatabaseFiles["userDBFileName"] = filepath.Join(baseDir, "..", "..", "test", "db", "user_database_test.json")

	for index, file := range mockDatabaseFiles {
		err := os.Remove(file)
		if err != nil {
			output := fmt.Errorf("mock DB file %s was not removed from location %s: %s", index, file, err)
			errorList = append(errorList, output)
		}
	}
	return errorList
}

func TestNewDB(t *testing.T) {
	mockDB := InitMockDB()

	// Chirp DB
	exists := FileExists(mockDB.ChirpDB.Path())
	if !exists {
		t.Error("Chirp DB file creation failed")
	}
	chirpDBContent, err := os.ReadFile(mockDB.ChirpDB.Path())
	if err != nil {
		t.Fatalf("Failed to read ChirpDB file: %v", err)
	}
	var chirpDB ChirpDBStructure
	if err := json.Unmarshal(chirpDBContent, &chirpDB); err != nil {
		t.Errorf("Failed to unmarshal ChirpDB: %v", err)
	}
	if chirpDB.NextChirpID != 0 || len(chirpDB.Chirps) != 0 {
		t.Errorf("ChirpDB is not initialized correctly: %+v", chirpDB)
	}
	// User DB
	exists = FileExists(mockDB.UserDB.Path())
	if !exists {
		t.Error("User DB file creation failed")
	}
	userDBContent, err := os.ReadFile(mockDB.UserDB.Path())
	if err != nil {
		t.Fatalf("Failed to read UserDB file: %v", err)
	}
	var userDB UserDBStructure
	if err := json.Unmarshal(userDBContent, &userDB); err != nil {
		t.Errorf("Failed to unmarshal UserDB: %v", err)
	}
	if userDB.NextUserID != 0 || len(userDB.Users) != 0 || len(userDB.UserLookup) != 0 {
		t.Errorf("UserDB is not initialized correctly: %+v", userDB)
	}

	errList := TeardownMockDB()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}
