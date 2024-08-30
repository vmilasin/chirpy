package database

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

// CreateUser creates a new user and saves it to disk
func (db *ChirpDB) Createuser(userEmail string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbDat, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	// In a real app, I'd be using an SQL database, and would check if the user (email) is already in the DB

	dbDat.NextUserID += 1
	newUser := User{
		ID:    dbDat.NextUserID,
		Email: userEmail,
	}

	dbDat.Users[dbDat.NextUserID] = newUser
	err = db.writeDB(dbDat)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}
