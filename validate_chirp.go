package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Validate chirp length
func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Valid       bool    `json:"valid"`
		CleanedBody *string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	ch := chirp{}
	err := decoder.Decode(&ch)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	const maxChirpLength = 140
	if len(ch.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cb := profanityCheck(ch.Body)
	respondWithJSON(w, http.StatusOK, returnVals{
		Valid:       true,
		CleanedBody: &cb,
	})
}

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

// Profanity checking
func profanityCheck(chBody string) (cleanBody string) {
	profaneWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	punctuationMarks := map[string]bool{
		".":  true,
		"?":  true,
		"!":  true,
		",":  true,
		";":  true,
		":":  true,
		"'":  true,
		"\"": true,
		"-":  true,
		"(":  true,
		")":  true,
		"[":  true,
		"]":  true,
		"/":  true,
		"\\": true,
		"_":  true,
	}

	// Split the chirp body string into separate words
	words := strings.Split(chBody, " ")
	for i, word := range words {
		word = strings.ToLower(word)
		chars := strings.Split(word, "")
		targetWord := word
		suffix := []string{}

		// Remove trailing punctionation marks from the check
		for i := 0; i < len(chars); i++ {
			lastChar := chars[len(chars)-(i+1)]
			if punctuationMarks[lastChar] {
				suffix = append(suffix, lastChar)
				targetWord = targetWord[:len(targetWord)-1]
			} else {
				break
			}
		}

		// Check the word for profanity
		if profaneWords[targetWord] {
			words[i] = "****" + strings.Join(suffix, "")
		}
	}
	cleanBody = strings.Join(words, " ")
	return cleanBody
}
