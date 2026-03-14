package service

import (
	"strings"
)

var bannedWords = map[string]bool{
	"nigger":   true,
	"niggers":  true,
	"faggot":   true,
	"faggots":  true,
	"kike":     true,
	"kikes":    true,
	"spic":     true,
	"spics":    true,
	"chink":    true,
	"chinks":   true,
	"wetback":  true,
	"wetbacks": true,
	"raghead":  true,
	"ragheads": true,
	"tranny":   true,
	"trannies": true,
	"retard":   true,
	"retards":  true,
	"retarded": true,
}

// ContainsBannedContent checks if the given content contains any banned words.
func ContainsBannedContent(content string) bool {
	words := strings.Fields(strings.ToLower(content))
	for _, word := range words {
		// Strip common punctuation from edges
		cleaned := strings.Trim(word, ".,!?;:\"'()[]{}#@*~`")
		if bannedWords[cleaned] {
			return true
		}
	}
	return false
}
