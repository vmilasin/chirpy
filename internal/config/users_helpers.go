package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

/*type User struct {
	ID                     int       `json:"id"`
	Email                  string    `json:"email"`
	Password               *string   `json:"password,omitempty"`
	PasswordHash           *[]byte   `json:"passwordHash"`
	RefreshToken           string    `json:"refresh_token,omitempty"`
	RefreshTokenExpiration time.Time `json:"refresh_token_expiration,omitempty"`
}
*/

// Remove password from returning to user on success
type LoginReturnUser struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// Email validation for registration & update
func (cfg *ApiConfig) EmailValidation(context context.Context, email string) (int, error) {
	// Validate if email already exists in the DB
	id, err := cfg.Queries.GetUserByEmail(context, email)
	if err != nil {
		output := func() {
			log.Printf("An error occured when trying to lookup email address in the database: %s", err)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
		returnError := errors.New(fmt.Sprintf("An error occured when trying to lookup email address in the database: %s", err))
		return http.StatusInternalServerError, returnError
	}
	if id != uuid.Nil {
		returnError := errors.New("E-mail address already in use. Please try another one.")
		return http.StatusBadRequest, returnError
	}

	// Validate email complexity
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailPattern)
	if !re.MatchString(email) {
		returnError := errors.New("Invalid e-mail address.")
		return http.StatusBadRequest, returnError
	}

	return 0, nil
}

// Password validation for registration & update
func (cfg *ApiConfig) PasswordValidation(password string) (int, error) {
	// Check password length
	if len(password) < 6 {
		returnError := errors.New("The password should be at least 6 characters long.")
		return http.StatusBadRequest, returnError
	}

	// Check for at least one lowercase letter
	hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLowercase {
		returnError := errors.New("The password should contain at least one lowercase letter.")
		return http.StatusBadRequest, returnError
	}

	// Check for at least one uppercase letter
	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUppercase {
		returnError := errors.New("The password should contain at least one uppercase letter.")
		return http.StatusBadRequest, returnError
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	if !hasDigit {
		returnError := errors.New("The password should contain at least one digit.")
		return http.StatusBadRequest, returnError
	}

	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	if !hasSpecial {
		returnError := errors.New("The password should contain at least one special character. (space character excluded)")
		return http.StatusBadRequest, returnError
	}

	return 0, nil
}

// User login
func (cfg *ApiConfig) UserAuth(context context.Context, email, password string) (LoginReturnUser, int, error) {
	// Check if the provided user exists in the DB
	userID, err := cfg.Queries.GetUserByEmail(context, email)
	if err == sql.ErrNoRows {
		returnError := errors.New("Wrong e-mail address. Please type a valid one.")
		return LoginReturnUser{}, http.StatusUnauthorized, returnError
	}
	if err != nil {
		output := func() {
			log.Printf("Failed lookup during user login: %s.", err)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
		returnError := errors.New(fmt.Sprintf("An error occured during user authentication: %s", err))
		return LoginReturnUser{}, http.StatusInternalServerError, returnError
	}

	hashedPW, err := cfg.Queries.GetPWHash(context, userID)
	if err != nil {
		output := func() {
			log.Printf("Failed lookup during user login: %s.", err)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
		returnError := errors.New(fmt.Sprintf("An error occured during user authentication: %s", err))
		return LoginReturnUser{}, http.StatusInternalServerError, returnError
	}

	if err := bcrypt.CompareHashAndPassword(hashedPW, []byte(password)); err != nil {
		returnError := errors.New("incorrect password")
		return LoginReturnUser{}, http.StatusUnauthorized, returnError
	}

	result := LoginReturnUser{
		ID:    userID,
		Email: email,
	}
	return result, 0, nil
}

// Update user info
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
