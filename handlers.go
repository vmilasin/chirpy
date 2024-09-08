package main

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
func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

// Health check
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// Usage metrics
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	t, err := template.ParseFiles("assets/html_templates/metrics.gohtml")
	checkError(err)

	err = t.Execute(w, cfg)
	checkError(err)
}

// Usage metrics reset
func (cfg *apiConfig) handlerMetricsReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	cfg.FileserverHits = 0
	output := "Hits: " + strconv.Itoa(cfg.FileserverHits)
	w.Write([]byte(output))
}

// GET all chirps
func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Fetch all chirps from the DB
		loadedChirps, err := cfg.AppDatabase.ChirpDB.GetChirps()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error fetching chirps")
			return
		}

		// Respond with JSON
		respondWithJSON(w, http.StatusOK, loadedChirps)
	} else {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) // HTTP requests should be GET
	}
}

// GET a chirp
func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Get the chirp ID from the request
		requestedId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		}

		// Fetch the requested chirp from the DB
		loadedChirp, err := cfg.AppDatabase.ChirpDB.GetChirp(requestedId)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}

		// Respond with JSON
		respondWithJSON(w, http.StatusOK, loadedChirp)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) // HTTP requests should be GET
	}
}

// POST a chirp
func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Parse JSON
		var chirp = database.Chirp{}
		if err := json.Unmarshal(body, &chirp); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Validate chirp
		if chirp.Body == "" || len(chirp.Body) > 140 {
			respondWithError(w, http.StatusBadRequest, "Invalid chirp")
			return
		}

		// Create chirp in database
		newChirp, err := cfg.AppDatabase.ChirpDB.CreateChirp(chirp.Body)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
			return
		}

		// Respond with JSON
		respondWithJSON(w, http.StatusCreated, newChirp)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) //HTTP requests should be POST
	}
}

// POST an User
func (cfg *apiConfig) handlerPostUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Parse JSON
		var user = database.User{}
		if err := json.Unmarshal(body, &user); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Validate if email already exists in the DB
		_, exists := cfg.AppDatabase.UserDB.UserLookup[user.Email]
		if exists {
			respondWithError(w, http.StatusBadRequest, "Invalid e-mail address")
		}

		// Validate email complexity
		emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		re := regexp.MustCompile(emailPattern)
		if !re.MatchString(user.Email) {
			respondWithError(w, http.StatusBadRequest, "Invalid e-mail address")
		}

		// Validate password
		passwordPattern := `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[\W_]).{6,}$`
		re = regexp.MustCompile(passwordPattern)
		if !re.MatchString(*user.Password) {
			respondWithError(w, http.StatusBadRequest, "The password should contain at least 6 characters and include a lowercase, an uppercase letter, a number and a special character.")
		}

		// Create user in database
		newUser, err := cfg.AppDatabase.UserDB.Createuser(user.Email, *user.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Respond with JSON
		respondWithJSON(w, http.StatusCreated, newUser)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) //HTTP requests should be POST
	}
}

// User login
func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Parse JSON
		var user = database.User{}
		if err := json.Unmarshal(body, &user); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Create user in database
		newUser, err := cfg.AppDatabase.UserDB.Createuser(user.Email, user.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Respond with JSON
		respondWithJSON(w, http.StatusCreated, newUser)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) //HTTP requests should be POST
	}
}
