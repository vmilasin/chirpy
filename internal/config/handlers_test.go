package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/vmilasin/chirpy/internal/database"
)

// Test for the helloHandler
func TestHandlerReadiness(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockApiConfig := InitMockApiConfig()

	// Create a request to pass to the handler
	req, err := http.NewRequest("GET", "/api/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler with the ResponseRecorder and request
	handler := http.HandlerFunc(mockApiConfig.HandlerReadiness)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestHandlerMetricsReset(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockApiConfig := InitMockApiConfig()

	// Create a request to pass to the handler
	req, err := http.NewRequest("GET", "/api/reset", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler with the ResponseRecorder and request
	handler := http.HandlerFunc(mockApiConfig.HandlerMetricsReset)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := "Hits: " + strconv.Itoa(0)
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestHandlerGetChirps(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockApiConfig := InitMockApiConfig()

	chirps := make([]string, 0, 2)
	chirps = append(chirps,
		"One not simply walk into Mordor.",
		"All right, then. Keep your secrets.")
	for _, chirp := range chirps {
		_, err := mockApiConfig.AppDatabase.ChirpDB.CreateChirp(chirp)
		if err != nil {
			t.Fatal("Error adding mock chirp.")
		}
	}
	// Create a request to pass to the handler
	req, err := http.NewRequest("GET", "/api/chirps", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler with the ResponseRecorder and request
	handler := http.HandlerFunc(mockApiConfig.HandlerGetChirps)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	mockedChirps := []database.Chirp{
		{ID: 1, Body: "One not simply walk into Mordor."},
		{ID: 2, Body: "All right, then. Keep your secrets."},
	}
	// Check the response body
	expected, _ := json.Marshal(mockedChirps)

	if rr.Body.String() != string(expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), string(expected))
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestHandlerGetChirp(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockApiConfig := InitMockApiConfig()

	chirps := make([]string, 0, 2)
	chirps = append(chirps,
		"One not simply walk into Mordor.",
		"All right, then. Keep your secrets.")
	for _, chirp := range chirps {
		_, err := mockApiConfig.AppDatabase.ChirpDB.CreateChirp(chirp)
		if err != nil {
			t.Fatalf("Error adding mock chirp: %s", err)
		}
	}
	// Create a request to pass to the handler
	req, err := http.NewRequest("GET", "/api/chirps/2", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler with the ResponseRecorder and request
	handler := http.HandlerFunc(mockApiConfig.HandlerGetChirp)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	mockedChirps := database.Chirp{
		ID: 2, Body: "All right, then. Keep your secrets.",
	}
	// Check the response body
	expected, _ := json.Marshal(mockedChirps)

	if rr.Body.String() != string(expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), string(expected))
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}
