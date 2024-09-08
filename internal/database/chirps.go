package database

import (
	"errors"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

// GetChirps returns all chirps in the database
func (db *ChirpDB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dat, _, err := loadDB(db)
	if err != nil {
		return []Chirp{}, err
	}

	var datLen int = len(dat.Chirps)
	result := make([]Chirp, 0, datLen)
	for _, item := range dat.Chirps {
		result = append(result, item)
	}

	return result, nil
}

// GetChirp returns a specific chip from the db
func (db *ChirpDB) GetChirp(requestedId int) (Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dat, _, err := loadDB(db)
	if err != nil {
		return Chirp{}, err
	}

	result, exists := dat.Chirps[requestedId]
	if exists {
		return result, nil
	} else {
		return Chirp{}, errors.New("Chirp not found")
	}
}

// CreateChirp creates a new chirp and saves it to disk
func (db *ChirpDB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbDat, _, err := loadDB(db)
	if err != nil {
		return Chirp{}, err
	}

	cleanChirp := profanityCheck(body)

	dbDat.NextChirpID += 1
	newChirp := Chirp{
		ID:   dbDat.NextChirpID,
		Body: cleanChirp,
	}

	dbDat.Chirps[dbDat.NextChirpID] = newChirp
	err = writeToDB(db, dbDat)
	if err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}
