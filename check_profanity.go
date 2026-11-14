package main

import "strings"

func CheckProfanity(value string) string {
	targets := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Fields(value)

	cleanedWords := make([]string, len(words))

	for _, word := range words {
		lowercaseWord := strings.ToLower(word)
		containsTargetWord := false
		for _, target := range targets {
			if strings.Contains(lowercaseWord, target) {
				cleanedWords = append(cleanedWords, "****")
				containsTargetWord = true
				break
			}
		}

		if !containsTargetWord {
			cleanedWords = append(cleanedWords, word)
		}
	}

	return strings.TrimSpace(strings.Join(cleanedWords, " "))
}
