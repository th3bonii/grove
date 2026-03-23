// Package opti provides the GROVE Opti Prompt engine for transforming
// natural language user prompts into precise, project-aware OpenCode instructions.
package opti

import (
	"regexp"
	"strings"
)

// Intent represents the classified user intent.
type Intent string

const (
	IntentFeatureAddition     Intent = "feature-addition"
	IntentBugFix              Intent = "bug-fix"
	IntentRefactor            Intent = "refactor"
	IntentDocumentationUpdate Intent = "documentation-update"
	IntentConfigurationChange Intent = "configuration-change"
	IntentOther               Intent = "other"
)

// IntentClassification contains the result of intent classification.
type IntentClassification struct {
	Intent     Intent   `json:"intent"`     // Classified intent type
	Domain     string   `json:"domain"`     // Extracted domain/module name
	Keywords   []string `json:"keywords"`   // Extracted keywords for file matching
	Confidence float64  `json:"confidence"` // Confidence score 0.0-1.0
	RawInput   string   `json:"raw_input"`  // Original input for reference
}

// Classifier handles intent classification of user prompts.
type Classifier struct {
	// Keyword patterns for each intent type
	intentPatterns map[Intent][]*regexp.Regexp
}

// NewClassifier creates a new IntentClassifier with default patterns.
func NewClassifier() *Classifier {
	return &Classifier{
		intentPatterns: map[Intent][]*regexp.Regexp{
			IntentFeatureAddition: {
				regexp.MustCompile(`(?i)\b(add|create|implement|build|introduce|new)\b`),
				regexp.MustCompile(`(?i)\b(feature|component|module|function|endpoint)\b`),
			},
			IntentBugFix: {
				regexp.MustCompile(`(?i)\b(fix|bug|issue|error|problem|broken|fail|crash)\b`),
				regexp.MustCompile(`(?i)\b(bug|defect|crash|exception|null|nil|undefined)\b`),
			},
			IntentRefactor: {
				regexp.MustCompile(`(?i)\b(refactor|restructure|reorganize|clean|optimize|improve)\b`),
				regexp.MustCompile(`(?i)\b(rename|extract|move|consolidate|decouple)\b`),
			},
			IntentDocumentationUpdate: {
				regexp.MustCompile(`(?i)\b(doc|document|comment|readme|guide|wiki)\b`),
				regexp.MustCompile(`(?i)\b(update|write|add|create|fix)\b.*\b(documentation|docs|comment)\b`),
			},
			IntentConfigurationChange: {
				regexp.MustCompile(`(?i)\b(config|setting|env|environment|config|parameter)\b`),
				regexp.MustCompile(`(?i)\b(change|update|modify|add|remove)\b.*\b(config|setting|env)\b`),
			},
		},
	}
}

// Classify classifies the user input into one of the defined intent types.
// Returns the IntentClassification with extracted domain and keywords.
func (c *Classifier) Classify(input string) IntentClassification {
	input = strings.TrimSpace(input)
	domain := c.extractDomain(input)
	keywords := c.extractKeywords(input)

	// Calculate confidence scores for each intent
	var bestIntent Intent = IntentOther
	var bestScore float64 = 0.0

	for intent, patterns := range c.intentPatterns {
		score := c.calculateIntentScore(input, patterns)
		if score > bestScore {
			bestScore = score
			bestIntent = intent
		}
	}

	// Boost confidence if keywords are strong
	if len(keywords) >= 3 {
		bestScore = min(bestScore*1.2, 1.0)
	}

	// Determine domain from keywords if not explicitly found
	if domain == "" && len(keywords) > 0 {
		domain = keywords[0]
	}

	return IntentClassification{
		Intent:     bestIntent,
		Domain:     domain,
		Keywords:   keywords,
		Confidence: bestScore,
		RawInput:   input,
	}
}

// calculateIntentScore calculates a score for an intent based on pattern matches.
func (c *Classifier) calculateIntentScore(input string, patterns []*regexp.Regexp) float64 {
	if len(patterns) == 0 {
		return 0.0
	}

	var score float64 = 0.0
	patternMatches := 0

	for _, pattern := range patterns {
		if pattern.MatchString(input) {
			patternMatches++
		}
	}

	// Score based on number of matching patterns
	if patternMatches == len(patterns) {
		score = 1.0
	} else if patternMatches == 1 && len(patterns) > 1 {
		score = 0.4
	} else if patternMatches > 0 {
		score = 0.7
	}

	return score
}

// extractDomain extracts the primary domain/module from the input.
// Looks for common domain patterns like "settings", "auth", "dashboard", etc.
func (c *Classifier) extractDomain(input string) string {
	// Common domain patterns to look for
	domainPatterns := []string{
		`\b(settings|config|configuration)\b`,
		`\b(auth|authentication|login|signup|user)\b`,
		`\b(dashboard|home|landing)\b`,
		`\b(profile|account|user-settings)\b`,
		`\b(api|endpoint|route)\b`,
		`\b(navigation|menu|sidebar)\b`,
		`\b(theme|dark-mode|light-mode)\b`,
		`\b(payment|billing|subscription)\b`,
		`\b(notification|alert|message)\b`,
		`\b(search|filter|sort)\b`,
	}

	lowerInput := strings.ToLower(input)

	for _, pattern := range domainPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindString(lowerInput); matches != "" {
			// Return the first match, cleaned up
			return strings.Title(strings.ReplaceAll(matches, "-", " "))
		}
	}

	// Try to extract domain from @file references
	fileRefPattern := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	matches := fileRefPattern.FindAllStringSubmatch(input, -1)
	if len(matches) > 0 {
		// Return the first file reference path component
		for _, match := range matches {
			if len(match) > 1 {
				return match[1]
			}
		}
	}

	return ""
}

// extractKeywords extracts meaningful keywords from the input for file matching.
// Handles camelCase, PascalCase, kebab-case, and snake_case.
func (c *Classifier) extractKeywords(input string) []string {
	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "can": true, "this": true,
		"that": true, "these": true, "those": true, "i": true, "you": true,
		"we": true, "they": true, "it": true, "my": true, "your": true,
		"our": true, "their": true, "please": true, "want": true, "need": true,
		"also": true, "just": true, "like": true, "when": true, "what": true,
	}

	var keywords []string
	seen := make(map[string]bool)

	// Split on whitespace and punctuation
	words := regexp.MustCompile(`[\s,.\-:;!?]+`).Split(input, -1)

	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		lower := strings.ToLower(word)

		// Skip stop words
		if stopWords[lower] {
			continue
		}

		// Skip very short words (less than 2 chars)
		if len(word) < 2 {
			continue
		}

		// Add the word if not seen
		if !seen[lower] {
			seen[lower] = true
			keywords = append(keywords, word)
		}

		// Also split camelCase and PascalCase into separate words
		camelParts := splitCamelCase(word)
		for _, part := range camelParts {
			lowerPart := strings.ToLower(part)
			if len(part) >= 2 && !stopWords[lowerPart] && !seen[lowerPart] {
				seen[lowerPart] = true
				keywords = append(keywords, part)
			}
		}
	}

	return keywords
}

// splitCamelCase splits camelCase and PascalCase words into separate parts.
func splitCamelCase(s string) []string {
	if s == "" {
		return nil
	}

	var parts []string
	var current strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Check if previous char was lowercase (true camelCase)
			prevRunes := []rune(s)
			prevR := prevRunes[i-1]
			if prevR >= 'a' && prevR <= 'z' {
				parts = append(parts, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
