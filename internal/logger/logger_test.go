package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var mutex = &sync.Mutex{}

// Helper function to check if a file exists
func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

// Create mock DB configuration and DB files
func InitMockLogs() *AppLogs {
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	// Log file paths
	mockLogFiles := make((map[string]string), 5)
	mockLogFiles["systemLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "system_test.log")
	mockLogFiles["handlerLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "handler_test.log")
	mockLogFiles["databaseLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "database_test.log")
	mockLogFiles["chirpLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "chirp_test.log")
	mockLogFiles["userLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "user_test.log")

	mockDB := InitiateLogs(mockLogFiles)
	return mockDB
}

// Delete mock DB files
func TeardownMockLogs() []error {
	errorList := make([]error, 0, 2)

	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	// Log file paths
	mockLogFiles := make((map[string]string), 5)
	mockLogFiles["systemLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "system_test.log")
	mockLogFiles["handlerLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "handler_test.log")
	mockLogFiles["databaseLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "database_test.log")
	mockLogFiles["chirpLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "chirp_test.log")
	mockLogFiles["userLog"] = filepath.Join(baseDir, "..", "..", "test", "logs", "user_test.log")

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

	mockLogs := InitMockLogs()

	exists := FileExists(mockLogs.SystemLog)
	if !exists {
		t.Error("System log file creation failed")
	}
	exists = FileExists(mockLogs.HandlerLog)
	if !exists {
		t.Error("Handler log file creation failed")
	}
	exists = FileExists(mockLogs.DatabaseLog)
	if !exists {
		t.Error("DB log file creation failed")
	}
	exists = FileExists(mockLogs.ChirpLog)
	if !exists {
		t.Error("Chirp log file creation failed")
	}
	exists = FileExists(mockLogs.UserLog)
	if !exists {
		t.Error("User log file creation failed")
	}

	errList := TeardownMockLogs()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestLogToFile(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockLogs := InitMockLogs()

	output := func() {
		log.SetFlags(0)
		log.Printf("Testing %d.", 123)
		log.SetFlags(log.Ldate | log.Ltime)
	}
	expectedOutput := "Testing 123.\n"

	// SystemLog
	targetFile := mockLogs.SystemLog
	err := mockLogs.LogToFile(targetFile, output)
	if err != nil {
		t.Errorf("There was an error when logging to file '%s': '%s'", targetFile, err)
	}
	fileContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("There was an error when reading file '%s': '%s'", targetFile, err)
	}
	if string(fileContent) != expectedOutput {
		t.Errorf("Logs in '%s' invalid.\nExpected: '%s'\nGot: '%s'", targetFile, expectedOutput, fileContent)
	}
	// ChirpLog
	targetFile = mockLogs.ChirpLog
	err = mockLogs.LogToFile(targetFile, output)
	if err != nil {
		t.Errorf("There was an error when logging to file '%s': '%s'", targetFile, err)
	}
	fileContent, err = os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("There was an error when reading file '%s': '%s'", targetFile, err)
	}
	if string(fileContent) != expectedOutput {
		t.Errorf("Logs in '%s' invalid.\nExpected: '%s'\nGot: '%s'", targetFile, expectedOutput, fileContent)
	}
	// UserLog
	targetFile = mockLogs.UserLog
	err = mockLogs.LogToFile(targetFile, output)
	if err != nil {
		t.Errorf("There was an error when logging to file '%s': '%s'", targetFile, err)
	}
	fileContent, err = os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("There was an error when reading file '%s': '%s'", targetFile, err)
	}
	if string(fileContent) != expectedOutput {
		t.Errorf("Logs in '%s' invalid.\nExpected: '%s'\nGot: '%s'", targetFile, expectedOutput, fileContent)
	}
	// DatabaseLog
	targetFile = mockLogs.DatabaseLog
	err = mockLogs.LogToFile(targetFile, output)
	if err != nil {
		t.Errorf("There was an error when logging to file '%s': '%s'", targetFile, err)
	}
	fileContent, err = os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("There was an error when reading file '%s': '%s'", targetFile, err)
	}
	if string(fileContent) != expectedOutput {
		t.Errorf("Logs in '%s' invalid.\nExpected: '%s'\nGot: '%s'", targetFile, expectedOutput, fileContent)
	}
	// HandlerLog
	targetFile = mockLogs.HandlerLog
	err = mockLogs.LogToFile(targetFile, output)
	if err != nil {
		t.Errorf("There was an error when logging to file '%s': '%s'", targetFile, err)
	}
	fileContent, err = os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("There was an error when reading file '%s': '%s'", targetFile, err)
	}
	if string(fileContent) != expectedOutput {
		t.Errorf("Logs in '%s' invalid.\nExpected: '%s'\nGot: '%s'", targetFile, expectedOutput, fileContent)
	}

	errList := TeardownMockLogs()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}
