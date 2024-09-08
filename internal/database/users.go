package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int     `json:"id"`
	Email        string  `json:"email"`
	Password     *string `json:"password,omitempty"`
	PasswordHash []byte  `json:"passwordHash"`
}

// Remove password from returning to user on success
type ReturnUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

// CreateUser creates a new user and saves it to disk
func (db *ChirpDB) Createuser(userEmail, userPassword string) (ReturnUser, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	_, dbDat, err := loadDB(db)
	if err != nil {
		return ReturnUser{}, err
	}

	dbDat.NextUserID += 1
	newPwHash, err := CreatePasswordHash(userPassword)
	if err != nil {
		return ReturnUser{}, errors.New("failed to create a password hash")
	}

	newUser := User{
		ID:           dbDat.NextUserID,
		Email:        userEmail,
		PasswordHash: newPwHash,
	}

	dbDat.Users[dbDat.NextUserID] = newUser
	dbDat.UserLookup[newUser.Email] = newUser.ID
	err = writeToDB(db, dbDat)
	if err != nil {
		return ReturnUser{}, err
	}

	result := ReturnUser{
		ID:    newUser.ID,
		Email: newUser.Email,
	}

	return result, nil
}

// Create password hash
func CreatePasswordHash(password string) ([]byte, error) {
	pw := []byte(password)

	newHash, err := bcrypt.GenerateFromPassword(pw, 12)
	if err != nil {
		return []byte{}, err
	}

	return newHash, nil
}
