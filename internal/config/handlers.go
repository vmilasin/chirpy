package config

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/vmilasin/chirpy/internal/auth"
	"github.com/vmilasin/chirpy/internal/database"
)

type CreateUserParamsInput struct {
	Email    string
	Password string
}

type CreateUserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type UpdateUserInfo struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type CreateChirpRequest struct {
	Body string    `json:"body"`
	ID   uuid.UUID `json:"user_id"`
}

type CreateChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
}

type RefreshTokenResponse struct {
	Token string `json:"token"`
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

// Drop all DB entries when PLATFORM="dev" in .env file
func (cfg *ApiConfig) HandlerDBReset(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if cfg.Platform == "dev" {
			if err := cfg.Queries.TruncateAllTables(r.Context()); err != nil {
				output := func() {
					log.Printf("An error occured while trying to truncate all tables: %s.", err)
				}
				cfg.AppLogs.LogToFile(cfg.AppLogs.HandlerLog, output)
				msg := fmt.Sprintf("An error occured while trying to truncate all tables: %s.", err)
				cfg.respondWithError(w, http.StatusForbidden, msg)
				return
			}
			cfg.respondWithJSON(w, http.StatusOK, "OK")
			return
		}
		cfg.respondWithError(w, http.StatusForbidden, "Forbidden")
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.")
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

// USER HANDLERS
// Register a new user
func (cfg *ApiConfig) HandlerUserRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid request body.")
			return
		}

		// Parse JSON from request body
		var newUserInput = CreateUserParamsInput{}
		if err := json.Unmarshal(body, &newUserInput); err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid JSON.")
			return
		}

		// Validate email address & password provided by the user
		httpStatus, err := cfg.EmailValidation(r.Context(), newUserInput.Email)
		if err != nil {
			cfg.respondWithError(w, httpStatus, err.Error())
		}
		/*httpStatus, err = cfg.PasswordValidation(newUserInput.Password)
		if err != nil {
			cfg.respondWithError(w, httpStatus, err.Error())
		}*/

		// Create a new password hash from the provided PW
		newPwHash, err := auth.CreatePasswordHash(newUserInput.Password)
		if err != nil {
			output := func() {
				log.Printf("Failed to create password hash for new user '%s': %s.", newUserInput.Email, err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to password hash for new user '%s': %s.", newUserInput.Email, err))
			return
		}

		// Create user in database
		newUser := database.CreateUserParams{
			Email:        newUserInput.Email,
			PasswordHash: newPwHash,
		}

		createdUser, err := cfg.Queries.CreateUser(r.Context(), newUser)
		if err != nil {
			output := func() {
				log.Printf("Failed to create user %s: %s.", newUser.Email, err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create user: '%s'", err))
			return
		}

		createdUserResponse := CreateUserResponse{
			ID:        createdUser.ID,
			CreatedAt: createdUser.CreatedAt,
			UpdatedAt: createdUser.UpdatedAt,
			Email:     createdUser.Email,
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusCreated, createdUserResponse)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.")
	}
}

// Log into a user
func (cfg *ApiConfig) HandlerUserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid request body.")
			return
		}

		// Parse JSON from request body
		var loginReq loginRequest
		if err := json.Unmarshal(body, &loginReq); err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid JSON.")
			return
		}

		// Log in to the desired user
		loginUser, httpStatus, err := cfg.UserAuth(r.Context(), loginReq.Email, loginReq.Password)
		if err != nil {
			cfg.respondWithError(w, httpStatus, err.Error())
			return
		}

		// Create a new access token
		accessTokenString, err := auth.CreateAccessToken(loginUser.ID, cfg.JWTSecret)
		if err != nil {
			output := func() {
				log.Printf("An error ocurred while creating a new access token: %v", err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error ocurred while creating a new access token: %s", err))
			return
		}

		// Create a new refresh token
		refreshTokenString, tokenExpiration, err := auth.CreateRefreshToken()
		if err != nil {
			output := func() {
				log.Printf("An error ocurred while creating a new refresh token: %v", err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error ocurred while creating a new refresh token: %v", err))
			return
		}

		newRefreshToken := database.CreateRefreshTokenParams{
			UserID:       loginUser.ID,
			RefreshToken: refreshTokenString,
			ExpiresAt:    tokenExpiration,
		}
		if _, err := cfg.Queries.CreateRefreshToken(r.Context(), newRefreshToken); err != nil {
			output := func() {
				log.Printf("Could not save refresh token.")
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, "Could not save refresh token.")
			return
		}

		returnResponse := loginResponse{
			ID:           loginUser.ID,
			Email:        loginUser.Email,
			Token:        accessTokenString,
			RefreshToken: refreshTokenString,
		}

		cfg.respondWithJSON(w, http.StatusOK, returnResponse)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.")
	}
}

func (cfg *ApiConfig) HandlerUserUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		userID, err := uuid.Parse(r.Context().Value("userID").(string))
		if err != nil {
			output := func() {
				log.Printf("An error occured while reading userID from context: %s.", err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.ChirpLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error occured while reading userID from context: %s.", err))
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

		if updateInfo.Email != nil {
			// Validate email
			httpStatus, err := cfg.EmailValidation(r.Context(), *updateInfo.Email)
			if err != nil {
				cfg.respondWithError(w, httpStatus, err.Error())
			}
		}

		var newPwHash []byte
		if updateInfo.Password != nil {
			// Validate password
			/*httpStatus, err := cfg.PasswordValidation(*updateInfo.Password)
			if err != nil {
				cfg.respondWithError(w, httpStatus, err.Error())
			}*/
			// Create a new PW hash
			newPwHash, err = auth.CreatePasswordHash(*updateInfo.Password)
			if err != nil {
				output := func() {
					log.Printf("Failed to create password hash for existing user '%s': %s.", userID, err)
				}
				cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
				cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create password hash for existing user '%s': %s.", userID, err))
				return
			}
		}

		newParameters := database.UpdateUserParams{
			Column1: updateInfo.Email,
			Column2: newPwHash,
			ID:      userID,
		}
		updatedUser, err := cfg.Queries.UpdateUser(r.Context(), newParameters)
		if err != nil {
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error occured during user info update '%s'", err))
			return
		}
		cfg.respondWithJSON(w, http.StatusOK, updatedUser)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.")
	}
}

// GET all chirps
func (cfg *ApiConfig) HandlerChirpsGetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Fetch all chirps from the DB
		loadedChirps, err := cfg.Queries.GetChirpAll(r.Context())
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
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.")
	}
}

// GET a chirp
func (cfg *ApiConfig) HandlerChirpsGetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Get the chirp ID from placeholder "chirpID" in request
		requestedId, err := uuid.Parse(r.PathValue("chirpID"))
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Failed to get chirpID from the URL.")
			return
		}

		// Fetch the requested chirp from the DB
		loadedChirp, err := cfg.Queries.GetChirpByID(r.Context(), requestedId)
		if err != nil {
			cfg.respondWithError(w, http.StatusNotFound, "Chirp not found.")
			return
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusOK, loadedChirp)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.")
	}
}

// POST a chirp
func (cfg *ApiConfig) HandlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		userID := r.Context().Value(ctxUserID).(uuid.UUID)

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid request body.")
			return
		}

		// Parse JSON
		var chirp = CreateChirpRequest{}
		if err := json.Unmarshal(body, &chirp); err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid JSON.")
			return
		}

		// Validate chirp
		cleanChirp := cfg.ChirpValidation(chirp.Body, w, r)

		newChirpData := database.CreateChirpParams{
			UserID: userID,
			Body:   cleanChirp,
		}

		// Create chirp in database
		newChirp, err := cfg.Queries.CreateChirp(r.Context(), newChirpData)
		if err != nil {
			output := func() {
				log.Printf("An error occured during chirp creation: %s.", err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.ChirpLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create chirp: '%s'", err))
			return
		}

		newChirpResponse := CreateChirpResponse{
			ID:        newChirp.ID,
			Body:      newChirp.Body,
			CreatedAt: newChirp.CreatedAt,
			UpdatedAt: newChirp.UpdatedAt,
			UserID:    newChirp.UserID,
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusCreated, newChirpResponse)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.")
	}
}

// Refresh - return a new access token if a user provides a valid refresh token
func (cfg *ApiConfig) HandlerRefreshTokenRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		userID := r.Context().Value(ctxUserID).(uuid.UUID)
		revokedAt := r.Context().Value(ctxRefreshTokenRevokedAt).(sql.NullTime)

		if !revokedAt.Valid {
			newAuthToken, err := auth.CreateAccessToken(userID, cfg.JWTSecret)
			if err != nil {
				output := func() {
					log.Printf("An error ocurred while creating a new access token: %v", err)
				}
				cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
				cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error ocurred while creating a new access token: %s", err))
				return
			}
			response := RefreshTokenResponse{
				Token: newAuthToken,
			}
			cfg.respondWithJSON(w, http.StatusOK, response)
			return
		}
		cfg.respondWithError(w, http.StatusUnauthorized, "The provided refresh token was revoked.")

	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.")
	}
}

// Revoke a refresh token
func (cfg *ApiConfig) HandlerRefreshTokenRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		refreshToken := r.Context().Value(ctxRefreshToken).(string)

		err := cfg.Queries.RevokeRefreshToken(r.Context(), refreshToken)
		if err != nil {
			output := func() {
				log.Printf("An error ocurred while revoking the refresh token in the database: %v", err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error ocurred while revoking the refresh token in the database: %s", err))
			return
		}

		cfg.respondWithJSON(w, http.StatusNoContent, nil)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.")
	}
}
