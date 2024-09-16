package database

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCreateChirp(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	modckDB := InitMockDB()

	// Basic chirp
	newChirp := "This is a new chirp."
	_, err := modckDB.ChirpDB.CreateChirp(newChirp)
	if err != nil {
		t.Errorf("Failed to create a new chirp: '%s'\nERROR: %s", newChirp, err)
	}
	chirpDBContent, err := os.ReadFile(modckDB.ChirpDB.Path())
	if err != nil {
		t.Fatalf("Failed to read ChirpDB file: %v", err)
	}
	var chirpDB ChirpDBStructure
	if err := json.Unmarshal(chirpDBContent, &chirpDB); err != nil {
		t.Errorf("Failed to unmarshal ChirpDB: %v", err)
	}
	if chirp, exists := chirpDB.Chirps[1]; !exists {
		t.Fatalf("Expected chirp with ID '%d' to exist", 1)
	} else {
		if chirp.Body != newChirp {
			t.Errorf("Expected chirp body to be '%s', got '%s'", newChirp, chirp.Body)
		}
	}

	// Chirp ID 2
	newChirp = "This is another chirp.'); DROP TABLE Chirps;--)"
	_, err = modckDB.ChirpDB.CreateChirp(newChirp)
	if err != nil {
		t.Errorf("Failed to create a new chirp: '%s'\nERROR: %s", newChirp, err)
	}
	chirpDBContent, err = os.ReadFile(modckDB.ChirpDB.Path())
	if err != nil {
		t.Fatalf("Failed to read ChirpDB file: %v", err)
	}
	if err := json.Unmarshal(chirpDBContent, &chirpDB); err != nil {
		t.Errorf("Failed to unmarshal ChirpDB: %v", err)
	}
	if chirp, exists := chirpDB.Chirps[2]; !exists {
		t.Fatalf("Expected chirp with ID '%d' to exist", 2)
	} else {
		if chirp.Body != newChirp {
			t.Errorf("Expected chirp body to be '%s', got '%s'", newChirp, chirp.Body) // "You have no power here! :)"
		}
	}

	// Profanity check
	newChirp = "This site creates kerfuffle!"
	cleanNewChirp := "This site creates ****!"
	_, err = modckDB.ChirpDB.CreateChirp(newChirp)
	if err != nil {
		t.Errorf("Failed to create a new chirp: '%s'\nERROR: %s", newChirp, err)
	}
	chirpDBContent, err = os.ReadFile(modckDB.ChirpDB.Path())
	if err != nil {
		t.Fatalf("Failed to read ChirpDB file: %v", err)
	}
	if err := json.Unmarshal(chirpDBContent, &chirpDB); err != nil {
		t.Errorf("Failed to unmarshal ChirpDB: %v", err)
	}
	if chirp, exists := chirpDB.Chirps[3]; !exists {
		t.Fatalf("Expected chirp with ID: '%d' to exist", 3)
	} else {
		if chirp.Body != cleanNewChirp {
			t.Errorf("Expected chirp body to be: '%s', got: '%s'", cleanNewChirp, chirp.Body)
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

func TestGetAllChirps(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	modckDB := InitMockDB()

	body1 := "This is a new chirp."
	_, err := modckDB.ChirpDB.CreateChirp(body1)
	if err != nil {
		t.Errorf("Failed to create a new chirp.\nERROR: %s", err)
	}
	body2 := "This is another chirp."
	_, err = modckDB.ChirpDB.CreateChirp(body2)
	if err != nil {
		t.Errorf("Failed to create a new chirp.\nERROR: %s", err)
	}
	body3 := "This is yet another chirp."
	_, err = modckDB.ChirpDB.CreateChirp(body3)
	if err != nil {
		t.Errorf("Failed to create a new chirp.\nERROR: %s", err)
	}

	chirpList, err := modckDB.ChirpDB.GetChirps()
	if err != nil {
		t.Fatalf("Failed to fetch all chirps.\nERROR: %s", err)
	} else if chirpList[0].ID != 1 || chirpList[0].Body != body1 {
		t.Errorf("First chirp invalid, expected ID: '%d', Body: '%s', got ID: '%d', Body: '%s'", 1, body1, chirpList[0].ID, chirpList[0].Body)
	} else if chirpList[1].ID != 2 || chirpList[1].Body != body2 {
		t.Errorf("First chirp invalid, expected ID: '%d', Body: '%s', got ID: '%d', Body: '%s'", 2, body2, chirpList[1].ID, chirpList[1].Body)
	} else if chirpList[2].ID != 3 || chirpList[2].Body != body3 {
		t.Errorf("First chirp invalid, expected ID: '%d', Body: '%s', got ID: '%d', Body: '%s'", 3, body3, chirpList[2].ID, chirpList[2].Body)
	}

	errList := TeardownMockDB()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}

func TestGetSpecificChirp(t *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	modckDB := InitMockDB()

	body1 := "This is a new chirp."
	_, err := modckDB.ChirpDB.CreateChirp(body1)
	if err != nil {
		t.Errorf("Failed to create a new chirp.\nERROR: %s", err)
	}
	body2 := "This is another chirp."
	_, err = modckDB.ChirpDB.CreateChirp(body2)
	if err != nil {
		t.Errorf("Failed to create a new chirp.\nERROR: %s", err)
	}
	body3 := "This is yet another chirp."
	_, err = modckDB.ChirpDB.CreateChirp(body3)
	if err != nil {
		t.Errorf("Failed to create a new chirp.\nERROR: %s", err)
	}

	targetChirp, err := modckDB.ChirpDB.GetChirp(2)
	if err != nil {
		t.Fatalf("Failed to fetch chirp with ID: %d.\nERROR: %s", 1, err)
	} else if targetChirp.ID != 2 || targetChirp.Body != body2 {
		t.Errorf("Desired chirp fetch failed, expected ID: '%d', Body: '%s', got ID: '%d', Body: '%s'", 2, body2, targetChirp.ID, targetChirp.Body)
	}
	targetChirp, err = modckDB.ChirpDB.GetChirp(5)
	if err == nil {
		t.Fatalf("Received (but shouldn't) chirp with ID: %d.\nERROR: %s", 5, err)
	}

	errList := TeardownMockDB()
	if len(errList) != 0 {
		for _, err := range errList {
			t.Error(err)
		}
		t.Fatal("Full teardown unsuccessful")
	}
}
