package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type AppDatabase struct {
	ChirpDB *ChirpDB
	UserDB  *UserDB
}

type DB interface {
	Path() string
	Mux() *sync.RWMutex
}

type ChirpDB struct {
	path string
	mux  *sync.RWMutex
}

func (db *ChirpDB) Path() string {
	return db.path
}

func (db *ChirpDB) Mux() *sync.RWMutex {
	return db.mux
}

type UserDB struct {
	path string
	mux  *sync.RWMutex
}

func (db *UserDB) Path() string {
	return db.path
}

func (db *UserDB) Mux() *sync.RWMutex {
	return db.mux
}

type ChirpDBStructure struct {
	NextChirpID int           `json:"nextChirpId"`
	Chirps      map[int]Chirp `json:"chirps"`
}

type UserDBStructure struct {
	NextUserID int            `json:"nextUserId"`
	Users      map[int]User   `json:"users"`
	UserLookup map[string]int `json:"userLookup"`
}

// Initiate a new database - consists of a chirpDB and a userDB file
func NewDB(chirpPath, userPath string) (*AppDatabase, error) {
	chDB := &ChirpDB{
		path: chirpPath,
		mux:  &sync.RWMutex{},
	}
	err := ensureDB(chDB)
	if err != nil {
		log.Print(err.Error())
		log.Fatal("error during chirp DB creation")
	}

	uDB := &UserDB{
		path: userPath,
		mux:  &sync.RWMutex{},
	}
	err = ensureDB(uDB)
	if err != nil {
		log.Print(err.Error())
		log.Fatal("error during user DB creation")
	}

	appDB := &AppDatabase{
		ChirpDB: chDB,
		UserDB:  uDB,
	}

	return appDB, nil
}

// ensureDB creates a new database file if it doesn't exist
func ensureDB(db DB) error {
	mux := db.Mux()
	path := db.Path()

	mux.Lock()
	defer mux.Unlock()

	// Check if the file already exists
	_, err := os.Stat(path)
	if err == nil {
		// The file exists
		return nil
	}
	if !os.IsNotExist(err) {
		// There was an error other than "file does not exist"
		return err
	}

	// File does not exist, create it
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = initializeDB(db)
	if err != nil {
		return err
	}

	return nil
}

// Initialize empty database structure in the file
func initializeDB(db DB) error {
	switch db.(type) {
	case *ChirpDB:
		emptyDB := ChirpDBStructure{
			NextChirpID: 0,
			Chirps:      make(map[int]Chirp),
		}
		return writeToDB(db, emptyDB)
	case *UserDB:
		emptyDB := UserDBStructure{
			NextUserID: 0,
			Users:      make(map[int]User),
			UserLookup: make(map[string]int),
		}
		return writeToDB(db, emptyDB)
	}
	return nil
}

// loadDB reads the database file into memory
func loadDB(db DB) (*ChirpDBStructure, *UserDBStructure, error) {
	dbPath := db.Path()
	dat, err := os.ReadFile(dbPath)
	if err != nil {
		return nil, nil, err
	}

	switch db.(type) {
	case *ChirpDB:
		var loadedData = ChirpDBStructure{}
		err = json.Unmarshal(dat, &loadedData)
		if err != nil {
			return nil, nil, err
		}

		return &loadedData, nil, nil
	case *UserDB:
		var loadedData = UserDBStructure{}
		err = json.Unmarshal(dat, &loadedData)
		if err != nil {
			return nil, nil, err
		}

		return nil, &loadedData, nil
	}

	return nil, nil, errors.New("something went wrong while loading the database")
}

// writeDB writes the database file to disk
func writeToDB(db DB, dbStructure interface{}) error {
	dbPath := db.Path()

	dbData, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(dbPath, dbData, 0644)
	if err != nil {
		return err
	}

	return nil

}
