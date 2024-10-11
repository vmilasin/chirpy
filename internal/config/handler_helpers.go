package config

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type errorResponse struct {
	Error string `json:"error"`
}

// RESPONSE HELPER FUNCTIONS
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
