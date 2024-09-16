package config

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"text/template"

	"github.com/vmilasin/chirpy/internal/database"
)

// Response helper functions
func (cfg *ApiConfig) respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		output := func() {
			log.Printf("(%s) Responding with 5XX error: %s.\n", cfg.AppLogs.CurrentTimestamp(), msg)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.HandlerLog, output)
	}
	type errorResponse struct {
		Error string `json:"error"`
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
			log.Printf("(%s) There was an error while trying to marshal JSON: %s.\n", cfg.AppLogs.CurrentTimestamp(), err)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.HandlerLog, output)

		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

// Health check
func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
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
			log.Printf("(%s) There was an error while trying to parse template %s: %s.\n", cfg.AppLogs.CurrentTimestamp(), metricsTemplate, err)
		}
		cfg.AppLogs.LogToFile(cfg.AppLogs.HandlerLog, output)
	}

	err = t.Execute(w, cfg)
	if err != nil {
		output := func() {
			log.Printf("(%s) There was an error while trying to execute template %s: %s.\n", cfg.AppLogs.CurrentTimestamp(), metricsTemplate, err)
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
				log.Printf("(%s) An error occured while fetching chirps: %s.\n", cfg.AppLogs.CurrentTimestamp(), err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.ChirpLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, "An error occured while fetching chirps.\n")
			return
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusOK, loadedChirps)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.\n") // HTTP requests should be GET
	}
}

// GET a chirp
func (cfg *ApiConfig) HandlerGetChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Get the chirp ID from the request
		requestedId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid chirp ID.\n")
		}

		// Fetch the requested chirp from the DB
		loadedChirp, err := cfg.AppDatabase.ChirpDB.GetChirp(requestedId)
		if err != nil {
			cfg.respondWithError(w, http.StatusNotFound, "Chirp not found.\n")
			return
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusOK, loadedChirp)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.\n") // HTTP requests should be GET
	}
}

// POST a chirp
func (cfg *ApiConfig) HandlerPostChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid request body.\n")
			return
		}

		// Parse JSON
		var chirp = database.Chirp{}
		if err := json.Unmarshal(body, &chirp); err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid JSON.\n")
			return
		}

		// Validate chirp
		if chirp.Body == "" || len(chirp.Body) > 140 {
			cfg.respondWithError(w, http.StatusBadRequest, "Chirp must be long 140 characters or less.\n")
			return
		}

		// Create chirp in database
		newChirp, err := cfg.AppDatabase.ChirpDB.CreateChirp(chirp.Body)
		if err != nil {
			output := func() {
				log.Printf("(%s) An error occured during chirp creation: %s.\n", cfg.AppLogs.CurrentTimestamp(), err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.ChirpLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, "Failed to create chirp.\n")
			return
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusCreated, newChirp)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.\n") // HTTP requests should be GET
	}
}

// POST an User
func (cfg *ApiConfig) HandlerPostUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid request body.\n")
			return
		}

		// Parse JSON
		var user = database.User{}
		if err := json.Unmarshal(body, &user); err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid JSON.\n")
			return
		}

		// Validate if email already exists in the DB
		_, exists, err := cfg.AppDatabase.UserDB.UserLookup(user.Email)
		if err != nil {
			output := func() {
				log.Printf("(%s) Failed lookup during user creation: %s.\n", cfg.AppLogs.CurrentTimestamp(), err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, "An error occured during user authentication.\n")
		}
		if exists {
			cfg.respondWithError(w, http.StatusBadRequest, "E-mail address already in use. Please try another one.\n")
		}

		// Validate email complexity
		emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		re := regexp.MustCompile(emailPattern)
		if !re.MatchString(user.Email) {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid e-mail address.\n")
		}

		/* PASSWORD VALIDATION TEMPORARILY TURNED OFF
		// Validate password
		// Check password length
		if len(*user.Password) < 6 {
			cfg.respondWithError(w, http.StatusBadRequest, "The password should be at least 6 characters long.\n")
			return
		}

		// Check for at least one lowercase letter
		hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(*user.Password)
		if !hasLowercase {
			cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one lowercase letter.\n")
			return
		}

		// Check for at least one uppercase letter
		hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(*user.Password)
		if !hasUppercase {
			cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one uppercase letter.\n")
			return
		}

		// Check for at least one digit
		hasDigit := regexp.MustCompile(`\d`).MatchString(*user.Password)
		if !hasDigit {
			cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one digit.\n")
			return
		}

		// Check for at least one special character
		hasSpecial := regexp.MustCompile(`[\W_]`).MatchString(*user.Password)
		if !hasSpecial {
			cfg.respondWithError(w, http.StatusBadRequest, "The password should contain at least one special character.\n")
			return
		}
		*/

		// Create user in database
		newUser, err := cfg.AppDatabase.UserDB.CreateUser(user.Email, *user.Password)
		if err != nil {
			output := func() {
				log.Printf("(%s) Failed to create user %s: %s.\n", cfg.AppLogs.CurrentTimestamp(), newUser.Email, err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, "Failed to create user.\n")
			return
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusCreated, newUser)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.\n") // HTTP requests should be GET
	}
}

// User login
func (cfg *ApiConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid request body.\n")
			return
		}

		// Parse JSON
		var user = database.User{}
		if err := json.Unmarshal(body, &user); err != nil {
			cfg.respondWithError(w, http.StatusBadRequest, "Invalid JSON.\n")
			return
		}

		// Check if the provided user exists in the DB
		_, exists, err := cfg.AppDatabase.UserDB.UserLookup(user.Email)
		if err != nil {
			output := func() {
				log.Printf("(%s) Failed lookup during user login: %s.\n", cfg.AppLogs.CurrentTimestamp(), err)
			}
			cfg.AppLogs.LogToFile(cfg.AppLogs.UserLog, output)
			cfg.respondWithError(w, http.StatusInternalServerError, "An error occured during user authentication.\n")
			return
		}
		if !exists {
			cfg.respondWithError(w, http.StatusBadRequest, "Wrong e-mail address. Please type a valid one.\n")
			return
		}

		// Log in to the desired user
		currentUser, err := cfg.AppDatabase.UserDB.LoginUser(user.Email, *user.Password)
		if err != nil {
			cfg.respondWithError(w, http.StatusUnauthorized, "User authentication failed.\n")
			return
		}

		// Respond with JSON
		cfg.respondWithJSON(w, http.StatusOK, currentUser)
	} else {
		cfg.respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method.\n") // HTTP requests should be GET
	}
}
