package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/vmilasin/chirpy/internal/database"
)

// Test for the helloHandler
func TestHandlerReadiness(t *testing.T) {
	// Lock the test to run function
	mutex.Lock()
	defer mutex.Unlock()

	// Create a mock API config
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
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusOK)
	}

	// Check the response body
	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'", rr.Body.String(), expected)
	}

	// Tear down the mock API config
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

	req, err := http.NewRequest("GET", "/api/reset", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(mockApiConfig.HandlerMetricsReset)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusOK)
	}

	expected := "Hits: " + strconv.Itoa(0)
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'", rr.Body.String(), expected)
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

	req, err := http.NewRequest("GET", "/api/chirps", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(mockApiConfig.HandlerGetChirps)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusOK)
	}

	mockedChirps := []database.Chirp{
		{ID: 1, Body: "One not simply walk into Mordor."},
		{ID: 2, Body: "All right, then. Keep your secrets."},
	}
	expected, _ := json.Marshal(mockedChirps)

	if rr.Body.String() != string(expected) {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'", rr.Body.String(), string(expected))
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestHandlerGetChirpsPOST(t *testing.T) {
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
	req, err := http.NewRequest("POST", "/api/chirps", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(mockApiConfig.HandlerGetChirps)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusMethodNotAllowed)
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
	req, err := http.NewRequest("GET", "/api/chirps/2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(mockApiConfig.HandlerGetChirp)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusOK)
	}

	mockedChirps := database.Chirp{
		ID: 2, Body: "All right, then. Keep your secrets.",
	}
	expected, _ := json.Marshal(mockedChirps)

	if rr.Body.String() != string(expected) {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'", rr.Body.String(), string(expected))
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestHandlerPostChirpSuccess(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockApiConfig := InitMockApiConfig()

	chirpBody := `{"body":"This is a test chirp."}`

	req, err := http.NewRequest("POST", "/api/chirps", strings.NewReader(chirpBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(mockApiConfig.HandlerPostChirp)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusCreated)
	}

	var createdChirp database.Chirp
	err = json.Unmarshal(rr.Body.Bytes(), &createdChirp)
	if err != nil {
		t.Fatalf("Could not parse response body: '%v'", err)
	}

	if createdChirp.Body != "This is a test chirp." {
		t.Errorf("handler returned wrong chirp body: got '%v' want '%v'", createdChirp.Body, "This is a test chirp.")
	}

	loadedChirp, err := mockApiConfig.AppDatabase.ChirpDB.GetChirp(createdChirp.ID)
	if err != nil || loadedChirp.Body != createdChirp.Body {
		t.Errorf("Chirp was not correctly added to the database: got '%v' want '%v'", loadedChirp.Body, createdChirp.Body)
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestHandlerPostUserSuccess(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockApiConfig := InitMockApiConfig()

	userBody := `{"email":"vader@darth.com", "password":"c0m3 t0 th3 D4RK S1D3, we have cookies!"}`

	req, err := http.NewRequest("POST", "/api/users", strings.NewReader(userBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(mockApiConfig.HandlerPostUser)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusCreated)
	}

	var createdUser database.ReturnUser
	err = json.Unmarshal(rr.Body.Bytes(), &createdUser)
	if err != nil {
		t.Fatalf("Could not parse response body: %v", err)
	}

	if createdUser.Email != "vader@darth.com" {
		t.Errorf("handler returned user email: got '%v' want '%v'", createdUser.Email, "vader@darth.com")
	}

	loadedID, exists, err := mockApiConfig.AppDatabase.UserDB.UserLookup(createdUser.Email)
	if err != nil || createdUser.ID != loadedID || !exists {
		t.Errorf("User was not correctly added to the database: got email: '%v', ID: '%d', want email: '%v', ID: '%d'", createdUser.Email, createdUser.ID, "vader@darth.com", 1)
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

/*
	func TestHandlerPostUserWrongEmailEmpty(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()

		mockApiConfig := InitMockApiConfig()

		userBody := `{"email":"", "password":"c0m3 t0 th3 D4RK S1D3, we have cookies!"}`

		req, err := http.NewRequest("POST", "/api/users", strings.NewReader(userBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(mockApiConfig.HandlerPostUser)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusBadRequest)
		}

		var returnError errorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &returnError)
		if err != nil {
			t.Fatalf("Could not parse response body: %v", err)
		}

		expectedError := "Invalid e-mail address."
		if returnError.Error != expectedError {
			t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
		}

		errList := TeardownMockApiConfig()
		if len(errList) != 0 {
			for _, err := range errList {
				t.Error(err)
			}
			t.Fatal("Full teardown unsuccessful")
		}
	}

	func TestHandlerPostUserWrongEmailNoDomain(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()

		mockApiConfig := InitMockApiConfig()

		userBody := `{"email":"vader@darth", "password":"c0m3 t0 th3 D4RK S1D3, we have cookies!"}`

		req, err := http.NewRequest("POST", "/api/users", strings.NewReader(userBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(mockApiConfig.HandlerPostUser)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusBadRequest)
		}

		var returnError errorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &returnError)
		if err != nil {
			t.Fatalf("Could not parse response body: %v", err)
		}

		expectedError := "Invalid e-mail address."
		if returnError.Error != expectedError {
			t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
		}

		errList := TeardownMockApiConfig()
		if len(errList) != 0 {
			for _, err := range errList {
				t.Error(err)
			}
			t.Fatal("Full teardown unsuccessful")
		}
	}

	func TestHandlerPostUserWrongEmailNoAt(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()

		mockApiConfig := InitMockApiConfig()

		userBody := `{"email":"vader-darth.com", "password":"c0m3 t0 th3 D4RK S1D3, we have cookies!"}`

		req, err := http.NewRequest("POST", "/api/users", strings.NewReader(userBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(mockApiConfig.HandlerPostUser)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusBadRequest)
		}

		var returnError errorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &returnError)
		if err != nil {
			t.Fatalf("Could not parse response body: %v", err)
		}

		expectedError := "Invalid e-mail address."
		if returnError.Error != expectedError {
			t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
		}

		errList := TeardownMockApiConfig()
		if len(errList) != 0 {
			for _, err := range errList {
				t.Error(err)
			}
			t.Fatal("Full teardown unsuccessful")
		}
	}

	func TestHandlerPostUserEmptyPW(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()

		mockApiConfig := InitMockApiConfig()

		userBody := `{"email":"vader@darth.com", "password":""}`

		req, err := http.NewRequest("POST", "/api/users", strings.NewReader(userBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(mockApiConfig.HandlerPostUser)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusBadRequest)
		}

		var returnError errorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &returnError)
		if err != nil {
			t.Fatalf("Could not parse response body: %v", err)
		}

		expectedError := "The password should be at least 6 characters long."
		if returnError.Error != expectedError {
			t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
		}

		errList := TeardownMockApiConfig()
		if len(errList) != 0 {
			for _, err := range errList {
				t.Error(err)
			}
			t.Fatal("Full teardown unsuccessful")
		}
	}

	func TestHandlerPostUserLessThanSixCharsPW(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()

		mockApiConfig := InitMockApiConfig()

		userBody := `{"email":"hodor@thrones.com", "password":"hodor"}`

		req, err := http.NewRequest("POST", "/api/users", strings.NewReader(userBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(mockApiConfig.HandlerPostUser)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusBadRequest)
		}

		var returnError errorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &returnError)
		if err != nil {
			t.Fatalf("Could not parse response body: %v", err)
		}

		expectedError := "The password should be at least 6 characters long."
		if returnError.Error != expectedError {
			t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
		}

		errList := TeardownMockApiConfig()
		if len(errList) != 0 {
			for _, err := range errList {
				t.Error(err)
			}
			t.Fatal("Full teardown unsuccessful")
		}
	}

	func TestHandlerPostUserNoSmallCapsPW(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()

		mockApiConfig := InitMockApiConfig()

		userBody := `{"email":"sidious@darth.com", "password":"UNLIMITED POOOOOOWEEEEEEEEEEEEERRRRRRRRRRRRRR!!!11!!1!"}`

		req, err := http.NewRequest("POST", "/api/users", strings.NewReader(userBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(mockApiConfig.HandlerPostUser)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusBadRequest)
		}

		var returnError errorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &returnError)
		if err != nil {
			t.Fatalf("Could not parse response body: %v", err)
		}

		expectedError := "The password should contain at least one lowercase letter."
		if returnError.Error != expectedError {
			t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
		}

		errList := TeardownMockApiConfig()
		if len(errList) != 0 {
			for _, err := range errList {
				t.Error(err)
			}
			t.Fatal("Full teardown unsuccessful")
		}
	}

	func TestHandlerPostUserNoUppercasePW(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()

		mockApiConfig := InitMockApiConfig()

		userBody := `{"email":"obi@wan.com", "password":"hello there!123"}`

		req, err := http.NewRequest("POST", "/api/users", strings.NewReader(userBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(mockApiConfig.HandlerPostUser)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusBadRequest)
		}

		var returnError errorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &returnError)
		if err != nil {
			t.Fatalf("Could not parse response body: %v", err)
		}

		expectedError := "The password should contain at least one uppercase letter."
		if returnError.Error != expectedError {
			t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
		}

		errList := TeardownMockApiConfig()
		if len(errList) != 0 {
			for _, err := range errList {
				t.Error(err)
			}
			t.Fatal("Full teardown unsuccessful")
		}
	}

	func TestHandlerPostUserNoNumericPW(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()

		mockApiConfig := InitMockApiConfig()

		userBody := `{"email":"obi@wan.com", "password":"Hello there!"}`

		req, err := http.NewRequest("POST", "/api/users", strings.NewReader(userBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(mockApiConfig.HandlerPostUser)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusBadRequest)
		}

		var returnError errorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &returnError)
		if err != nil {
			t.Fatalf("Could not parse response body: %v", err)
		}

		expectedError := "The password should contain at least one digit."
		if returnError.Error != expectedError {
			t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
		}

		errList := TeardownMockApiConfig()
		if len(errList) != 0 {
			for _, err := range errList {
				t.Error(err)
			}
			t.Fatal("Full teardown unsuccessful")
		}
	}

	func TestHandlerPostUserSpecialCharsPW(t *testing.T) {
		mutex.Lock()
		defer mutex.Unlock()

		mockApiConfig := InitMockApiConfig()

		userBody := `{"email":"obi@wan.com", "password":"Hello there 1234"}`

		req, err := http.NewRequest("POST", "/api/users", strings.NewReader(userBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(mockApiConfig.HandlerPostUser)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusBadRequest)
		}

		var returnError errorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &returnError)
		if err != nil {
			t.Fatalf("Could not parse response body: %v", err)
		}

		expectedError := "The password should contain at least one special character. (space character excluded)"
		if returnError.Error != expectedError {
			t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
		}

		errList := TeardownMockApiConfig()
		if len(errList) != 0 {
			for _, err := range errList {
				t.Error(err)
			}
			t.Fatal("Full teardown unsuccessful")
		}
	}
*/
func TestHandlerUserLoginSuccess(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockApiConfig := InitMockApiConfig()
	mockApiConfig.AppDatabase.UserDB.CreateUser("vader@darth.com", "c0m3 t0 th3 D4RK S1D3, we have cookies!")

	userBody := `{"email":"vader@darth.com", "password":"c0m3 t0 th3 D4RK S1D3, we have cookies!"}`

	req, err := http.NewRequest("POST", "/api/login", strings.NewReader(userBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(mockApiConfig.HandlerLogin)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusOK)
	}

	var loginUser loginResponse
	err = json.Unmarshal(rr.Body.Bytes(), &loginUser)
	if err != nil {
		t.Fatalf("Could not parse response body: %v", err)
	}

	if loginUser.Email != "vader@darth.com" {
		t.Errorf("handler returned user email: got '%v' want '%v'", loginUser.Email, "vader@darth.com")
	}

	if loginUser.Token == "" {
		t.Errorf("An empty token was returned. Token: %s", loginUser.Token)
	}

	loadedID, exists, err := mockApiConfig.AppDatabase.UserDB.UserLookup(loginUser.Email)
	if err != nil || loginUser.ID != loadedID || !exists {
		t.Errorf("User was not correctly logged in: got email: '%v', ID: '%d', want email: '%v', ID: '%d'", loginUser.Email, loginUser.ID, "vader@darth.com", 1)
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestHandlerUserLoginWrongPW(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockApiConfig := InitMockApiConfig()
	mockApiConfig.AppDatabase.UserDB.CreateUser("vader@darth.com", "c0m3 t0 th3 D4RK S1D3, we have cookies!")

	userBody := `{"email":"vader@darth.com", "password":"xxx.123"}`

	req, err := http.NewRequest("POST", "/api/login", strings.NewReader(userBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(mockApiConfig.HandlerLogin)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusUnauthorized)
	}

	var returnError errorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &returnError)
	if err != nil {
		t.Fatalf("Could not parse response body: %v", err)
	}

	expectedError := "Wrong password."
	if returnError.Error != expectedError {
		t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestHandlerUserLoginWrongUser(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	mockApiConfig := InitMockApiConfig()
	mockApiConfig.AppDatabase.UserDB.CreateUser("vader@darth.com", "c0m3 t0 th3 D4RK S1D3, we have cookies!")

	userBody := `{"email":"sidious@darth.com", "password":"xxx.123"}`

	req, err := http.NewRequest("POST", "/api/login", strings.NewReader(userBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(mockApiConfig.HandlerLogin)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusUnauthorized)
	}

	var returnError errorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &returnError)
	if err != nil {
		t.Fatalf("Could not parse response body: %v", err)
	}

	expectedError := "Wrong e-mail address. Please type a valid one."
	if returnError.Error != expectedError {
		t.Errorf("handler returned wrong error: got '%v' want '%v'", returnError.Error, expectedError)
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestUpdateUserSuccess(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	// Initiate DB & create user
	mockApiConfig := InitMockApiConfig()
	mockApiConfig.AppDatabase.UserDB.CreateUser("vader@darth.com", "c0m3 t0 th3 D4RK S1D3, we have cookies!")

	// Log into the user
	userBody := `{"email":"vader@darth.com", "password":"c0m3 t0 th3 D4RK S1D3, we have cookies!"}`

	req, err := http.NewRequest("POST", "/api/login", strings.NewReader(userBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(mockApiConfig.HandlerLogin)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusOK)
	}

	var loginUser loginResponse
	err = json.Unmarshal(rr.Body.Bytes(), &loginUser)
	if err != nil {
		t.Fatalf("Could not parse response body: %v", err)
	}

	if loginUser.Email != "vader@darth.com" {
		t.Errorf("handler returned user email: got '%v' want '%v'", loginUser.Email, "vader@darth.com")
	}

	if loginUser.Token == "" {
		t.Errorf("An empty token was returned. Token: %s", loginUser.Token)
	}

	loadedID, exists, err := mockApiConfig.AppDatabase.UserDB.UserLookup(loginUser.Email)
	if err != nil || loginUser.ID != loadedID || !exists {
		t.Errorf("User was not correctly logged in: got email: '%v', ID: '%d', want email: '%v', ID: '%d'", loginUser.Email, loginUser.ID, "vader@darth.com", 1)
	}

	// Update user info
	updateInfoBody := `{"email":"sidious@darth.com", "password":"UnL1m1T3d POOOOOOOOOOWEEEEEEEEEERRRRRRRR11!1!!!!!!11"}`
	if err != nil {
		t.Errorf("An error occured while composing update info: '%s'", err)
	}

	req, err = http.NewRequest("PUT", "/api/users", strings.NewReader(updateInfoBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", loginUser.Token))

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(mockApiConfig.HandlerUpdateUser)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusOK)
	}

	var updatedUser database.ReturnUser
	err = json.Unmarshal(rr.Body.Bytes(), &updatedUser)
	if err != nil {
		t.Fatalf("Could not parse response body: %v", err)
	}

	if updatedUser.Email != "sidious@darth.com" || updatedUser.ID != loadedID {
		t.Errorf("There was an error in returned values, expected ID: %d, email: %s, received ID: %d, email: %s", loadedID, "sidious@darth.com", updatedUser.ID, updatedUser.Email)
	}

	// Try to log in using new credentials
	userBodyNew := `{"email":"sidious@darth.com", "password":"UnL1m1T3d POOOOOOOOOOWEEEEEEEEEERRRRRRRR11!1!!!!!!11"}`

	req, err = http.NewRequest("POST", "/api/login", strings.NewReader(userBodyNew))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(mockApiConfig.HandlerLogin)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got '%v' want '%v'", status, http.StatusOK)
	}

	var loginUserNew loginResponse
	err = json.Unmarshal(rr.Body.Bytes(), &loginUserNew)
	if err != nil {
		t.Fatalf("Could not parse response body: %v", err)
	}

	if loginUserNew.Email != "sidious@darth.com" {
		t.Errorf("handler returned user email: got '%v' want '%v'", loginUserNew.Email, "sidious@darth.com")
	}

	if loginUserNew.Token == "" {
		t.Errorf("An empty token was returned. Token: %s", loginUserNew.Token)
	}

	loadedID, exists, err = mockApiConfig.AppDatabase.UserDB.UserLookup(loginUserNew.Email)
	if err != nil || loginUserNew.ID != loadedID || !exists {
		t.Errorf("User was not correctly logged in: got email: '%v', ID: '%d', want email: '%v', ID: '%d'", loginUserNew.Email, loginUserNew.ID, "sidious@darth.com", 1)
	}

	errList := TeardownMockApiConfig()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

// Add tests here to check update info passwords
