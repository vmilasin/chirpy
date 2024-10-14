package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vmilasin/chirpy/internal/profanity"
	"golang.org/x/crypto/bcrypt"
)

type errorResponse struct {
	Error string `json:"error"`
}

type AuthResponse struct {
	ID    uuid.UUID `json:"user_id"`
	Email string    `json:"email"`
}

// RESPONSE HELPER FUNCTIONS

// HTTP response when an error occurs
func (cfg *ApiConfig) respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		output := func() {
			log.Printf("Responding with 5XX error: %s.", msg)
		}
		err := cfg.AppLogs.LogToFile(cfg.AppLogs.HandlerLog, output)
		if err != nil {
			log.Printf("Error writing logs to file '%s': %s", cfg.AppLogs.HandlerLog, err)
		}
	}
	cfg.respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

// HTTP response
func (cfg *ApiConfig) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		output := func() {
			log.Printf("There was an error while trying to marshal JSON: %s.", err)
		}
		err := cfg.AppLogs.LogToFile(cfg.AppLogs.HandlerLog, output)
		if err != nil {
			log.Printf("Error writing logs to file '%s': %s", cfg.AppLogs.HandlerLog, err)
		}

		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

// USER HELPERS

// Email validation for registration & update
func (cfg *ApiConfig) EmailValidation(context context.Context, email string) (int, error) {
	// Validate if email already exists in the DB
	id, err := cfg.Queries.GetUserByEmail(context, email)
	if err != nil && err != sql.ErrNoRows {
		output := func() {
			log.Printf("An error occured when trying to lookup email address in the database: %s", err)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
		returnError := fmt.Errorf("an error occured when trying to lookup email address in the database: %s", err)
		return http.StatusInternalServerError, returnError
	}
	if id != uuid.Nil {
		returnError := errors.New("e-mail address already in use. Please try another one")
		return http.StatusBadRequest, returnError
	}

	// Validate email complexity
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailPattern)
	if !re.MatchString(email) {
		returnError := errors.New("invalid e-mail address")
		return http.StatusBadRequest, returnError
	}

	return 0, nil
}

// Password validation for registration & update
func (cfg *ApiConfig) PasswordValidation(password string) (int, error) {
	// Check password length
	if len(password) < 6 {
		returnError := errors.New("the password should be at least 6 characters long")
		return http.StatusBadRequest, returnError
	}

	// Check for at least one lowercase letter
	hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLowercase {
		returnError := errors.New("the password should contain at least one lowercase letter")
		return http.StatusBadRequest, returnError
	}

	// Check for at least one uppercase letter
	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUppercase {
		returnError := errors.New("the password should contain at least one uppercase letter")
		return http.StatusBadRequest, returnError
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	if !hasDigit {
		returnError := errors.New("the password should contain at least one digit")
		return http.StatusBadRequest, returnError
	}

	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	if !hasSpecial {
		returnError := errors.New("the password should contain at least one special character. (space character excluded)")
		return http.StatusBadRequest, returnError
	}

	return 0, nil
}

// User login
func (cfg *ApiConfig) UserAuth(context context.Context, email, password string) (AuthResponse, int, error) {
	// Check if the provided user exists in the DB
	userID, err := cfg.Queries.GetUserByEmail(context, email)
	if err == sql.ErrNoRows {
		returnError := errors.New("wrong e-mail address. Please type a valid one")
		return AuthResponse{}, http.StatusUnauthorized, returnError
	}
	if err != nil {
		output := func() {
			log.Printf("Failed lookup during user login: %s.", err)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
		returnError := fmt.Errorf("an error occured during user authentication: %s", err)
		return AuthResponse{}, http.StatusInternalServerError, returnError
	}

	hashedPW, err := cfg.Queries.GetPWHash(context, userID)
	if err != nil {
		output := func() {
			log.Printf("Failed lookup during user login: %s.", err)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
		returnError := fmt.Errorf("an error occured during user authentication: %s", err)
		return AuthResponse{}, http.StatusInternalServerError, returnError
	}

	if err := bcrypt.CompareHashAndPassword(hashedPW, []byte(password)); err != nil {
		returnError := errors.New("incorrect password")
		return AuthResponse{}, http.StatusUnauthorized, returnError
	}

	result := AuthResponse{
		ID:    userID,
		Email: email,
	}
	return result, 0, nil
}

// AUTH HELPERS

// Get the access token that was stored in the context after passing through authentication middleware
func GetAccessTokenFromContext(r *http.Request) string {
	if value := r.Context().Value("accessToken"); value != nil {
		if token, ok := value.(string); ok {
			return token
		}
	}
	return ""
}

// Return response based on the result of a failed authentication
func (cfg *ApiConfig) resolveAuthTokenError(w http.ResponseWriter, err error) {
	if err.Error() == "invalid or missing Authorization header" {
		cfg.respondWithError(w, http.StatusUnauthorized, "Invalid or missing Authorization header")
		return
	}
	if err.Error() == "failed to convert userID from subject to int" {
		output := func() {
			log.Printf("Failed to convert userID from subject to int")
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
		cfg.respondWithError(w, http.StatusInternalServerError, "Failed to convert userID from subject to int")
		return
	}
	if err == jwt.ErrSignatureInvalid {
		cfg.respondWithError(w, http.StatusUnauthorized, "Invalid token signature")
		return
	}
	if strings.Contains(err.Error(), "expired") {
		cfg.respondWithError(w, http.StatusUnauthorized, "Token has expired")
		return
	}
	cfg.respondWithError(w, http.StatusUnauthorized, "Invalid token")
}

// CHIRP HELPERS

// Check if the chirp passes the length and profanity requirements
func (cfg *ApiConfig) ChirpValidation(body string, w http.ResponseWriter, r *http.Request) string {
	// Check for maximum Chirp length
	if body == "" || len(body) > 140 {
		cfg.respondWithError(w, http.StatusBadRequest, "Chirp must be long 140 characters or less.")
		return ""
	}

	// Run the profanity check against the chirp
	cleanChirp := profanity.ProfanityCheck(body)
	return cleanChirp
}
