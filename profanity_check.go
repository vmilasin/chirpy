package main

import "strings"

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
