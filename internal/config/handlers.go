package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vmilasin/chirpy/internal/database"
)

type errorResponse struct {
	Error string `json:"error"`
}

type loginRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
}

type loginResponse struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type UpdateUserInfo struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

// Response helper functions
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

func (cfg *ApiConfig) resolveAuthTokenError(w http.ResponseWriter, err error) {
	if err.Error() == "invalid or missing Authorization header" {
		cfg.respondWithError(w, http.StatusUnauthorized, "Invalid or missing Authorization header")
		return
	}
	if err.Error() == "failed to convert userID from subject to int" {
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

// Health check
func (cfg *ApiConfig) HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// Usage metrics
func (cfg *ApiConfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Template HTML
	metricsTemplate := "assets/html_templates/metrics.gohtml"

	t, err := template.ParseFiles(metricsTemplate)
	if err != nil {
		output := func() {
			log.Printf("There was an error while trying to parse template %s: %s.", metricsTemplate, err)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.HandlerLog, output)
		return
	}

	err = t.Execute(w, cfg)
	if err != nil {
		output := func() {
			log.Printf("There was an error while trying to execute template %s: %s.", metricsTemplate, err)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.HandlerLog, output)
	}
}

// Usage metrics reset
func (cfg *ApiConfig) HandlerMetricsReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	cfg.FileserverHits = 0
	output := "Hits: " + strconv.Itoa(cfg.FileserverHits)
	w.Write([]byte(output))
}

// GET all chirps
func (cfg *ApiConfig) HandlerGetChirps(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Fetch all chirps from the DB
		loadedChirps, err := cfg.AppDatabase.ChirpDB.GetChirps()
		if err != nil {
			output := func() {
				log.Printf("An error occured while fetching chirps: %s.", err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.ChirpLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error occured while fetching chirps: '%s'", err))
			return
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusOK, loadedChirps)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.") // HTTP requests should be GET
	}
}

// GET a chirp
func (cfg *ApiConfig) HandlerGetChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid chirp ID.")
			return
		}

		// Convert the chirp ID from string to int
		requestedId, err := strconv.Atoi(pathParts[3])
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid chirp ID.")
			return
		}

		// Fetch the requested chirp from the DB
		loadedChirp, err := cfg.AppDatabase.ChirpDB.GetChirp(requestedId)
		if err != nil {
			cfg.respondWithError(w, http.StatusNotFound, "Chirp not found.")
			return
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusOK, loadedChirp)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.") // HTTP requests should be GET
	}
}

// POST a chirp
func (cfg *ApiConfig) HandlerPostChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid request body.")
			return
		}

		// Parse JSON
		var chirp = database.Chirp{}
		if err := json.Unmarshal(body, &chirp); err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid JSON.")
			return
		}

		// Validate chirp
		if chirp.Body == "" || len(chirp.Body) > 140 {
			cfg.respondWithError(w, http.StatusBadRequest, "Chirp must be long 140 characters or less.")
			return
		}

		// Create chirp in database
		newChirp, err := cfg.AppDatabase.ChirpDB.CreateChirp(chirp.Body)
		if err != nil {
			output := func() {
				log.Printf("An error occured during chirp creation: %s.", err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.ChirpLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create chirp: '%s'", err))
			return
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusCreated, newChirp)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.") // HTTP requests should be GET - added for extra security
	}
}

// POST an User
func (cfg *ApiConfig) HandlerPostUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid request body.")
			return
		}

		// Parse JSON
		var user = database.User{}
		if err := json.Unmarshal(body, &user); err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid JSON.")
			return
		}

		// Validate if email already exists in the DB
		_, exists, err := cfg.AppDatabase.UserDB.UserLookup(user.Email)
		if exists {
			cfg.respondWithError(w, http.StatusBadRequest, "E-mail address already in use. Please try another one.")
			return
		}

		/*
			// Validate email complexity
			emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			re := regexp.MustCompile(emailPattern)
			if !re.MatchString(user.Email) {
				cfg.respondWithError(w, http.StatusBadRequest, "Invalid e-mail address.")
				return
			}

			// Validate password
			// Check password length
			if len(*user.Password) < 6 {
				cfg.respondWithError(w, http.StatusBadRequest, "The password should be at least 6 characters long.")
				return
			}

			// Check for at least one lowercase letter
			hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(*user.Password)
			if !hasLowercase {
				cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one lowercase letter.")
				return
			}

			// Check for at least one uppercase letter
			hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(*user.Password)
			if !hasUppercase {
				cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one uppercase letter.")
				return
			}

			// Check for at least one digit
			hasDigit := regexp.MustCompile(`\d`).MatchString(*user.Password)
			if !hasDigit {
				cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one digit.")
				return
			}

			// Check for at least one special character
			hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(*user.Password)
			if !hasSpecial {
				cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one special character. (space character excluded)")
				return
			}
		*/

		// Create user in database
		newUser, err := cfg.AppDatabase.UserDB.CreateUser(user.Email, *user.Password)
		if err != nil {
			output := func() {
				log.Printf("Failed to create user %s: %s.", newUser.Email, err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create user: '%s'", err))
			return
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusCreated, newUser)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.") // HTTP requests should be GET
	}
}

// User login
func (cfg *ApiConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid request body.")
			return
		}

		// Parse JSON
		var loginReq loginRequest
		if err := json.Unmarshal(body, &loginReq); err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid JSON.")
			return
		}

		// Check if the provided user exists in the DB
		_, exists, err := cfg.AppDatabase.UserDB.UserLookup(loginReq.Email)
		if !exists {
			cfg.respondWithError(w, http.StatusUnauthorized, "Wrong e-mail address. Please type a valid one.")
			return
		}
		if err != nil {
			output := func() {
				log.Printf("Failed lookup during user login: %s.\n", err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error occured during user authentication: %s", err))
			return
		}

		// Log in to the desired user
		currentUser, err := cfg.AppDatabase.UserDB.LoginUser(loginReq.Email, loginReq.Password)
		if err != nil {
			cfg.respondWithError(w, http.StatusUnauthorized, "Wrong password.")
			return
		}

		// Create a new access token
		var jwtExpiration int
		if loginReq.ExpiresInSeconds != nil && *loginReq.ExpiresInSeconds <= 1*60*60 {
			jwtExpiration = *loginReq.ExpiresInSeconds
		} else {
			jwtExpiration = 1 * 60 * 60
		}

		accessTokenString, err := cfg.AppDatabase.UserDB.AddAccessTokenToUser(jwtExpiration, currentUser.ID, cfg.JWTSecret)
		if err != nil {
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error ocurred while creating a new access token: %s", err.Error()))
		}

		// Create a new refresh token
		refreshTokenString, err := cfg.AppDatabase.UserDB.AddRefreshTokenToUser(currentUser.ID)
		if err != nil {
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error ocurred while creating a new refresh token: %s", err.Error()))
		}

		returnResponse := loginResponse{
			ID:           currentUser.ID,
			Email:        currentUser.Email,
			Token:        accessTokenString,
			RefreshToken: refreshTokenString,
		}

		cfg.respondWithJSON(w, http.StatusOK, returnResponse)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.") // HTTP requests should be GET
	}
}

func (cfg *ApiConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		//Get the authorization header and pass it to the access token auth function it should return the user's ID if successful
		authHeader := r.Header.Get("Authorization")
		userID, err := cfg.AppDatabase.UserDB.AccessTokenAuthorization(authHeader, cfg.JWTSecret)
		if err != nil {
			cfg.resolveAuthTokenError(w, err)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid request body.")
			return
		}
		var updateInfo UpdateUserInfo
		if err := json.Unmarshal(body, &updateInfo); err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid JSON.")
			return
		}

		/*  ADD PASSWORD VALIDATION FOR THE NEW PW
		// Validate email complexity
		emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		re := regexp.MustCompile(emailPattern)
		if !re.MatchString(user.Email) {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid e-mail address.")
			return
		}

		// Validate password
		// Check password length
		if len(*user.Password) < 6 {
			cfg.respondWithError(w, http.StatusBadRequest, "The password should be at least 6 characters long.")
			return
		}

		// Check for at least one lowercase letter
		hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(*user.Password)
		if !hasLowercase {
			cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one lowercase letter.")
			return
		}

		// Check for at least one uppercase letter
		hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(*user.Password)
		if !hasUppercase {
			cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one uppercase letter.")
			return
		}

		// Check for at least one digit
		hasDigit := regexp.MustCompile(`\d`).MatchString(*user.Password)
		if !hasDigit {
			cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one digit.")
			return
		}

		// Check for at least one special character
		hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(*user.Password)
		if !hasSpecial {
			cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one special character. (space character excluded)")
			return
		}
		*/

		updatedUser, err := cfg.AppDatabase.UserDB.UpdateUser(userID, updateInfo.Email, updateInfo.Password)
		if err != nil {
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error occured during user info update '%s'", err))
			return
		}
		cfg.respondWithJSON(w, http.StatusOK, updatedUser)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.") // HTTP requests should be PUT
	}
}

/*func (cfg *ApiConfig) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {

	}
}
*/
