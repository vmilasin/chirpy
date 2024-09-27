package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// Remove password from returning to user on success
type ReturnUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}*/

// Email validation for registration & update
func (cfg *ApiConfig) EmailValidation(email string, w http.ResponseWriter, r *http.Request) {
	// Validate if email already exists in the DB
	id, err := cfg.Queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error occured when trying to lookup email address in the database: %s", err))
		return
	}
	if id != uuid.Nil {
		cfg.respondWithError(w, http.StatusBadRequest, "E-mail address already in use. Please try another one.")
		return
	}

	// Validate email complexity
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailPattern)
	if !re.MatchString(email) {
		cfg.respondWithError(w, http.StatusBadRequest, "Invalid e-mail address.")
		return
	}
}

// Password validation for registration & update
func (cfg *ApiConfig) PasswordValidation(password string, w http.ResponseWriter, r *http.Request) {
	// Check password length
	if len(password) < 6 {
		cfg.respondWithError(w, http.StatusBadRequest, "The password should be at least 6 characters long.")
		return
	}

	// Check for at least one lowercase letter
	hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLowercase {
		cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one lowercase letter.")
		return
	}

	// Check for at least one uppercase letter
	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUppercase {
		cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one uppercase letter.")
		return
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	if !hasDigit {
		cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one digit.")
		return
	}

	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	if !hasSpecial {
		cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one special character. (space character excluded)")
		return
	}
}

// Create password hash for new users on sign-up
func CreatePasswordHash(password string) ([]byte, error) {
	pw := []byte(password)

	newHash, err := bcrypt.GenerateFromPassword(pw, 12)
	if err != nil {
		return []byte{}, err
	}

	return newHash, nil
}

// User login
func (cfg *ApiConfig) LoginUser(password string, w http.ResponseWriter, r *http.Request) (ReturnUser, error) {
	/*db.mux.RLock()
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
	providedPW := []byte(userPassword)*/
	user := d

	if err := bcrypt.CompareHashAndPassword(*hashedPW, providedPW); err != nil {
		return ReturnUser{}, errors.New("incorrect password")
	}

	result := ReturnUser{
		ID:    desiredUser.ID,
		Email: desiredUser.Email,
	}
	return result, nil
}

func (db *UserDB) AddAccessTokenToUser(jwtExpiration, userID int, JWTSecret []byte) (string, error) {
	now := time.Now().UTC()

	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(jwtExpiration) * time.Second)),
		Subject:   strconv.Itoa(userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", err
	}

	return signedString, nil
}

// Add a new refresh token to user (on login)
func (db *UserDB) AddRefreshTokenToUser(userID int) (string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	// Get user by ID
	user, err := db.UserLookupByID(userID)
	if err != nil {
		return "", err
	}

	// Create a refresh token string
	randBytes := make([]byte, 32)
	_, err = rand.Read(randBytes)
	if err != nil {
		return "", err
	}
	hexString := hex.EncodeToString(randBytes)

	// Define token expiration timestamp
	now := time.Now().UTC()
	expirationTimestamp := now.Add(time.Duration(60*24) * time.Hour)

	user.RefreshToken = hexString
	user.RefreshTokenExpiration = expirationTimestamp

	// Load the DB && write to it
	_, dbDat, err := loadDB(db)
	if err != nil {
		return "", err
	}
	dbDat.Users[user.ID] = user

	return hexString, nil
}

// Authorization request using an access token
func (db *UserDB) AccessTokenAuthorization(header string, JWTSecret []byte) (int, error) {
	var token string
	if strings.HasPrefix(header, "Bearer ") {
		token = strings.TrimPrefix(header, "Bearer ")
		token = strings.TrimSpace(token)
	} else {
		err := errors.New("invalid or missing Authorization header")
		return 0, err
	}

	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})
	if err != nil {
		return 0, err
	}

	userID := claims.Subject
	convertedUserID, err := strconv.Atoi(userID)
	if err != nil {
		return 0, errors.New("failed to convert userID from subject to int")
	}

	return convertedUserID, nil
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
