package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int     `json:"id"`
	Email        string  `json:"email"`
	Password     *string `json:"password,omitempty"`
	PasswordHash *[]byte `json:"passwordHash"`
}

// Remove password from returning to user on success
type ReturnUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

// User lookup
func (db *UserDB) UserLookup(userEmail string) (int, bool, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	_, dbDat, err := loadDB(db)
	if err != nil {
		return 0, true, err
	}

	id, exists := dbDat.UserLookup[userEmail]

	if !exists {
		return 0, false, errors.New("user does not exist")
	}

	return id, true, nil
}

// User lookup
func (db *UserDB) UserLookupByID(id int) (User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	_, dbDat, err := loadDB(db)
	if err != nil {
		return User{}, errors.New("user does not exist")
	}

	user := dbDat.Users[id]
	return user, nil
}

// CreateUser creates a new user and saves it to disk
func (db *UserDB) CreateUser(userEmail, userPassword string) (ReturnUser, error) {
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
		PasswordHash: &newPwHash,
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

// User login
func (db *UserDB) LoginUser(userEmail, userPassword string) (ReturnUser, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	userID, _, err := db.UserLookup(userEmail)
	if err != nil {
		return ReturnUser{}, err
	}

	_, dbDat, err := loadDB(db)
	if err != nil {
		return ReturnUser{}, err
	}

	desiredUser := dbDat.Users[userID]
	hashedPW := desiredUser.PasswordHash
	providedPW := []byte(userPassword)

	if err := bcrypt.CompareHashAndPassword(*hashedPW, providedPW); err != nil {
		return ReturnUser{}, errors.New("incorrect password")
	}

	result := ReturnUser{
		ID:    desiredUser.ID,
		Email: desiredUser.Email,
	}
	return result, nil
}

func (db *UserDB) UpdateUser(userID int, email *string, password *string) (ReturnUser, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	user, err := db.UserLookupByID(userID)
	if err != nil {
		return ReturnUser{}, err
	}

	if email != nil {
		user.Email = *email
	}
	if password != nil {
		newPwHash, err := CreatePasswordHash(*password)
		if err != nil {
			return ReturnUser{}, errors.New("failed to create a password hash")
		}
		user.PasswordHash = &newPwHash
	}

	_, dbDat, err := loadDB(db)
	if err != nil {
		return ReturnUser{}, err
	}

	dbDat.Users[userID] = user
	if email != nil {
		dbDat.UserLookup[*email] = userID
	}

	err = writeToDB(db, dbDat)
	if err != nil {
		return ReturnUser{}, err
	}

	result := ReturnUser{
		ID:    user.ID,
		Email: user.Email,
	}
	return result, nil
}
