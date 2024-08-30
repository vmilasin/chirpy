package database

import (
	"encoding/json"
	"os"
	"sync"
)

type ChirpDB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	NextChirpID int           `json:"nextChirpId"`
	Chirps      map[int]Chirp `json:"chirps"`
	NextUserID  int           `json:"nextUserId"`
	Users       map[int]User  `json:"users"`
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
	db.mux.Lock()
	defer db.mux.Unlock()

	// Check if the file already exists
	_, err := os.Stat(db.path)
	if err == nil {
		// The file exists
		return nil
	}
	if !os.IsNotExist(err) {
		// There was an error other than "file does not exist"
		return err
	}

	// File does not exist, create it
	file, err := os.Create(db.path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Initialize empty database structure in the file
	emptyDB := DBStructure{
		NextChirpID: 0,
		Chirps:      make(map[int]Chirp),
		NextUserID:  0,
		Users:       make(map[int]User),
	}
	return db.writeDB(emptyDB)
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
