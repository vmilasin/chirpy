package config

import (
	"net/http"

	"github.com/vmilasin/chirpy/internal/profanity"
)

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
