// Package utils provides shared utility functions for string manipulation.
package utils

import (
	"regexp"
	"strings"
)

// ToLower returns a lowercase version of the string.
// Wrapper around strings.ToLower for consistency across the codebase.
func ToLower(s string) string {
	return strings.ToLower(s)
}

// Contains checks if substr is contained in s.
// Wrapper around strings.Contains for consistency.
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Truncate truncates a string to a maximum length.
// If the string is shorter than maxLen, it returns the original string.
// If maxLen is less than or equal to 0, it returns an empty string.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// ExtractKeywords extracts keywords from text by removing punctuation,
// converting to lowercase, and filtering out stop words.
// Returns a slice of meaningful words.
func ExtractKeywords(text string) []string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove punctuation and split into words
	re := regexp.MustCompile(`[^\p{L}\p{N}]+`)
	words := re.Split(text, -1)

	// Filter out stop words and empty strings
	var keywords []string
	for _, word := range words {
		if word != "" && !IsStopWord(word) {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// IsStopWord checks if a word is a common stop word.
// Stop words are common words that are usually filtered out in text analysis.
func IsStopWord(word string) bool {
	stopWords := map[string]bool{
		"a":       true,
		"an":      true,
		"and":     true,
		"are":     true,
		"as":      true,
		"at":      true,
		"be":      true,
		"been":    true,
		"but":     true,
		"by":      true,
		"can":     true,
		"could":   true,
		"did":     true,
		"do":      true,
		"does":    true,
		"doing":   true,
		"done":    true,
		"for":     true,
		"from":    true,
		"had":     true,
		"has":     true,
		"have":    true,
		"having":  true,
		"he":      true,
		"her":     true,
		"here":    true,
		"him":     true,
		"his":     true,
		"how":     true,
		"i":       true,
		"if":      true,
		"in":      true,
		"into":    true,
		"is":      true,
		"it":      true,
		"its":     true,
		"just":    true,
		"me":      true,
		"more":    true,
		"most":    true,
		"my":      true,
		"no":      true,
		"not":     true,
		"of":      true,
		"on":      true,
		"or":      true,
		"other":   true,
		"our":     true,
		"out":     true,
		"said":    true,
		"she":     true,
		"so":      true,
		"some":    true,
		"such":    true,
		"than":    true,
		"that":    true,
		"the":     true,
		"their":   true,
		"them":    true,
		"then":    true,
		"there":   true,
		"these":   true,
		"they":    true,
		"this":    true,
		"those":   true,
		"through": true,
		"to":      true,
		"too":     true,
		"under":   true,
		"up":      true,
		"very":    true,
		"was":     true,
		"we":      true,
		"were":    true,
		"what":    true,
		"when":    true,
		"where":   true,
		"which":   true,
		"while":   true,
		"who":     true,
		"will":    true,
		"with":    true,
		"would":   true,
		"you":     true,
		"your":    true,
	}

	return stopWords[word]
}
