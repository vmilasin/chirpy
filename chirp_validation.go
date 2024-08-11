package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Validate chirp length
func (cfg apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Id          string  `json:"id"`
		Valid       bool    `json:"valid"`
		CleanedBody *string `json:"cleaned_body"`
	}

	// Unmarshal JSON payload from the received request body
	decoder := json.NewDecoder(r.Body)
	ch := reqBody{}
	err := decoder.Decode(&ch)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	// Check if the chirp is over the size limit
	const maxChirpLength = 140
	if len(ch.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	// Run the profanity check against the chirp
	cb := profanityCheck(ch.Body)

	// Get all the chirps saved in the database
	DBData := cfg.readDB()

	// Print the saved payload
	respondWithJSON(w, http.StatusCreated, returnVals{
		Id:          chID,
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

/*
// Read from database.json file
func (cfg apiConfig) readDB() chirpDatabase {
	cfg.MutexRW.RLock()
	defer cfg.MutexRW.RUnlock()

	dbData, err := os.ReadFile("database.json")
	checkError(err)

	dat := chirpDatabase{}
	err = json.Unmarshal(dbData, &dat)
	checkError(err)

	return dat
}

// Write to database.json file
func (cfg apiConfig) writeDB(DBData chirpDatabase, chirp string) {
	cfg.MutexRW.Lock()
	defer cfg.MutexRW.Unlock()

	nextIndex := reflect.ValueOf(DBData).Len() + 1

	DBData = append(DBData, payload)

	err = json.Marshal()
}
*/
