package database

import (
	"errors"
	"os"
	"sync"
)

type chirpDB struct {
	path string
	mux  *sync.RWMutex
}

func newDB(path string) (*chirpDB, error) {
	f, err := os.Open(path)
	if err == nil {
		f.Close()
		return nil, errors.New("database file already exists")
	}

	f, err = os.Create(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	chDB := &chirpDB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	return chDB, nil
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

/*
// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error)

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error)

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error)

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error


    os.ReadFile
    os.ErrNotExist
    os.WriteFile
    sort.Slice

*/
