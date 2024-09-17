package database

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestCreateUser(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	modckDB := InitMockDB()

	email := "saul@bettercall.com"
	password := "Xxx.123.17!"
	newUserReturn, err := modckDB.UserDB.CreateUser(email, password)
	if err != nil {
		t.Errorf("Failed to create a new user: '%s'\nERROR: %s", email, err)
	}
	if newUserReturn.Email != email || newUserReturn.ID != 1 {
		t.Errorf("Failed to return new user. Expected ID: '%d', email: '%s', got ID: '%d', email: '%s'.", 1, email, newUserReturn.ID, newUserReturn.Email)
	}
	userDBContent, err := os.ReadFile(modckDB.UserDB.Path())
	if err != nil {
		t.Fatalf("Failed to read UserDB file: %v", err)
	}
	var userDB UserDBStructure
	if err := json.Unmarshal(userDBContent, &userDB); err != nil {
		t.Errorf("Failed to unmarshal UserDB: %v", err)
	}
	if user, exists := userDB.Users[1]; !exists {
		t.Fatalf("Expected user with ID '%d' to exist", 1)
	} else {
		if user.Email != email {
			t.Errorf("Expected user email to be '%s', got '%s'", email, user.Email)
		}
	}

	errList := TeardownMockDB()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestUserLookup(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	modckDB := InitMockDB()

	email := "saul@bettercall.com"
	password := "Xxx.123.17!"
	_, err := modckDB.UserDB.CreateUser(email, password)
	if err != nil {
		t.Errorf("Failed to create a new user: '%s'\nERROR: %s", email, err)
	}

	userLookupID, userLookupExists, err := modckDB.UserDB.UserLookup(email)
	if err != nil {
		t.Errorf("There was an error during lookup of email: '%s'\nERROR: %s", email, err)
	} else if !userLookupExists {
		t.Errorf("The user %s should have been returned, but wasn't", email)
	} else if userLookupID != 1 {
		t.Errorf("The returned user's ID was expected to be: '%d', got: '%d'", 1, userLookupID)
	}

	email = "heisenberg@betternotcall.com"
	expectedError := errors.New("user does not exist")
	userLookupID, userLookupExists, err = modckDB.UserDB.UserLookup(email)
	if err != nil && err.Error() != expectedError.Error() {
		t.Errorf("There was an error during lookup of email: '%s'\nERROR: %s", email, err)
	} else if userLookupExists {
		t.Errorf("The user %s should not have been returned, but was", email)
	} else if userLookupID != 0 {
		t.Errorf("The returned user's ID was expected to be: '%d', got: '%d'", 0, userLookupID)
	}

	errList := TeardownMockDB()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestUserLogin(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	modckDB := InitMockDB()

	email := "saul@bettercall.com"
	password := "Xxx.123.17!"
	_, err := modckDB.UserDB.CreateUser(email, password)
	if err != nil {
		t.Errorf("Failed to create a new chirp: '%s'\nERROR: %s", email, err)
	}

	loggedInUser, err := modckDB.UserDB.LoginUser(email, password)
	if err != nil {
		t.Errorf("There was an error during login to: '%s'\nERROR: %s", email, err)
	} else if loggedInUser.ID != 1 || loggedInUser.Email != email {
		t.Errorf("The returned value for user being logged in to are invalid, expected ID: '%d', email: '%s', got ID: '%d', email: '%s'", 1, email, loggedInUser.ID, loggedInUser.Email)
	}

	loggedInUser, err = modckDB.UserDB.LoginUser(email, "xxx.123")
	expectedError := errors.New("incorrect password")
	if err != nil && err.Error() != expectedError.Error() {
		t.Errorf("There was an error during login to: '%s'\nERROR: %s", email, err)
	}
	if err.Error() != expectedError.Error() {
		t.Errorf("An error should've occured, but didn't: '%s'", expectedError)
	}

	loggedInUser, err = modckDB.UserDB.LoginUser("heisenberg@betternotcall.com", "xxx.123")
	expectedError = errors.New("user does not exist")
	if err != nil && err.Error() != expectedError.Error() {
		t.Errorf("There was an error during login to: '%s'\nERROR: %s", email, err)
	}
	if err.Error() != expectedError.Error() {
		t.Errorf("An error should've occured, but didn't: '%s'", expectedError)
	}
}
