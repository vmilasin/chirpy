package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type ChirpDB struct {
	path string
	mux  *sync.RWMutex
}

// Initiate a new database
func NewDB(path string) (*ChirpDB, error) {
	chDB := &ChirpDB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := chDB.ensureDB()
	return chDB, err
}

// ensureDB creates a new database file if it doesn't exist
func (db *ChirpDB) ensureDB() error {
	db.mux.RLock()
	defer db.mux.RUnlock()
	f, err := os.Open(db.path)
	if err == nil {
		f.Close()
		return errors.New("database file already exists")
	}

	db.mux.Lock()
	defer db.mux.Unlock()
	f, err = os.Create(db.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

// loadDB reads the database file into memory
func (db *ChirpDB) loadDB() (DBStructure, error) {
	dat, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	var loadedData = DBStructure{}
	err = json.Unmarshal(dat, &loadedData)
	if err != nil {
		return DBStructure{}, err
	}

	return loadedData, nil
}

// writeDB writes the database file to disk
func (db *ChirpDB) writeDB(dbStructure DBStructure) error {
	dbData, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dbData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetChirps returns all chirps in the database
func (db *ChirpDB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dat, err := db.loadDB()
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

// CreateChirp creates a new chirp and saves it to disk
func (db *ChirpDB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbDat, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	nextIndex := len(dbDat.Chirps) + 1
	newChirp := Chirp{
		ID:   nextIndex,
		Body: body,
	}

	dbDat.Chirps[nextIndex] = newChirp
	err = db.writeDB(dbDat)
	if err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}
