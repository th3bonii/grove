// Package opti provides the GROVE Opti Prompt engine for transforming
// natural language user prompts into precise, project-aware OpenCode instructions.
package opti

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	groveerrors "github.com/Gentleman-Programming/grove/internal/errors"
)

// PromptDiff represents the difference between original and edited prompts.
type PromptDiff struct {
	Original  string
	Edited    string
	Added     []string
	Removed   []string
	Rewritten []string
}

// LearnedPattern represents a learned pattern from user edits.
type LearnedPattern struct {
	Key       string    `json:"key"`        // Unique pattern key (category + content hash)
	Type      string    `json:"type"`       // "added", "removed", "rewritten"
	Category  string    `json:"category"`   // Intent category (e.g., "feature-addition")
	Before    string    `json:"before"`     // What was before the edit
	After     string    `json:"after"`      // What was after the edit
	Frequency int       `json:"frequency"`  // How many times this pattern was observed
	AutoApply bool      `json:"auto_apply"` // Whether to auto-apply this pattern
	LastSeen  time.Time `json:"last_seen"`  // Last time this pattern was observed
	Timestamp time.Time `json:"timestamp"`  // When this pattern was first learned
}

// BidirectionalLearner learns from user edits to improve future optimizations.
type BidirectionalLearner struct {
	logPath   string
	patterns  map[string]*LearnedPattern
	autoApply map[string]bool // Patterns ready for auto-application
}

// NewBidirectionalLearner creates a new BidirectionalLearner.
func NewBidirectionalLearner(projectRoot string) *BidirectionalLearner {
	logPath := filepath.Join(projectRoot, "GROVE-OPTI-LOG.md")

	// Try to load existing patterns from log file
	patterns := make(map[string]*LearnedPattern)
	autoApply := make(map[string]bool)

	if _, err := os.Stat(logPath); err == nil {
		existingPatterns, err := loadPatternsFromLog(logPath)
		if err != nil {
			slog.Debug("failed to load existing patterns",
				slog.String("error", err.Error()))
		} else {
			for key, pattern := range existingPatterns {
				patterns[key] = pattern
				if pattern.Frequency >= 3 {
					autoApply[key] = true
				}
			}
			slog.Debug("loaded existing patterns",
				slog.Int("count", len(patterns)))
		}
	}

	return &BidirectionalLearner{
		logPath:   logPath,
		patterns:  patterns,
		autoApply: autoApply,
	}
}

// LearnFromEdit learns from a user edit and updates patterns.
func (l *BidirectionalLearner) LearnFromEdit(original, edited string, category string) error {
	if original == edited {
		slog.Debug("no changes to learn from (identical strings)")
		return nil
	}

	diff := l.computeDiff(original, edited)

	// Extract pattern from diff
	pattern := l.extractPattern(diff, category)

	// Generate a unique key for this pattern
	pattern.Key = l.generatePatternKey(pattern)

	// Check if pattern already exists
	if existing, ok := l.patterns[pattern.Key]; ok {
		existing.Frequency++
		existing.LastSeen = time.Now()

		// Auto-apply threshold
		if existing.Frequency >= 3 && !existing.AutoApply {
			existing.AutoApply = true
			l.autoApply[pattern.Key] = true
			slog.Info("pattern auto-apply enabled",
				slog.String("key", pattern.Key),
				slog.Int("frequency", existing.Frequency))
		}

		slog.Debug("updated existing pattern",
			slog.String("key", pattern.Key),
			slog.Int("frequency", existing.Frequency))
	} else {
		// New pattern
		pattern.Timestamp = time.Now()
		pattern.LastSeen = time.Now()
		l.patterns[pattern.Key] = &pattern

		slog.Debug("learned new pattern",
			slog.String("key", pattern.Key),
			slog.String("type", pattern.Type))
	}

	// Save to log file
	return l.savePatternToLog(&pattern)
}

// computeDiff computes the diff between original and edited prompts.
func (l *BidirectionalLearner) computeDiff(original, edited string) *PromptDiff {
	origLines := strings.Split(original, "\n")
	editLines := strings.Split(edited, "\n")

	diff := &PromptDiff{
		Original: original,
		Edited:   edited,
	}

	// Simple line-by-line diff algorithm
	origSet := make(map[string]bool)
	editSet := make(map[string]bool)

	for _, line := range origLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			origSet[trimmed] = true
		}
	}

	for _, line := range editLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			editSet[trimmed] = true
		}
	}

	// Find added lines (in edited but not in original)
	for line := range editSet {
		if !origSet[line] {
			diff.Added = append(diff.Added, line)
		}
	}

	// Find removed lines (in original but not in edited)
	for line := range origSet {
		if !editSet[line] {
			diff.Removed = append(diff.Removed, line)
		}
	}

	// Detect rewritten: lines that are similar but not identical
	// Use simple word overlap detection
	for _, removed := range diff.Removed {
		for _, added := range diff.Added {
			if l.wordsOverlap(removed, added) >= 0.5 {
				diff.Rewritten = append(diff.Rewritten, fmt.Sprintf("%s -> %s", removed, added))
				// Remove from added/removed to avoid duplication
				diff.Added = removeString(diff.Added, added)
				diff.Removed = removeString(diff.Removed, removed)
				break
			}
		}
	}

	return diff
}

// wordsOverlap calculates the word overlap ratio between two strings.
func (l *BidirectionalLearner) wordsOverlap(a, b string) float64 {
	wordsA := strings.Fields(strings.ToLower(a))
	wordsB := strings.Fields(strings.ToLower(b))

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0
	}

	// Count common words
	common := 0
	seenB := make(map[string]bool)
	for _, word := range wordsB {
		seenB[word] = true
	}

	for _, word := range wordsA {
		if seenB[word] {
			common++
		}
	}

	// Use Jaccard-like similarity
	minLen := min(len(wordsA), len(wordsB))
	if minLen == 0 {
		return 0
	}

	return float64(common) / float64(minLen)
}

// removeString removes a string from a slice (for rewritten detection).
func removeString(slice []string, s string) []string {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// extractPattern extracts a LearnedPattern from a PromptDiff.
func (l *BidirectionalLearner) extractPattern(diff *PromptDiff, category string) LearnedPattern {
	pattern := LearnedPattern{
		Category: category,
	}

	// Determine pattern type based on diff content
	if len(diff.Added) > 0 && len(diff.Removed) == 0 {
		pattern.Type = "added"
		pattern.After = strings.Join(diff.Added, "; ")
	} else if len(diff.Removed) > 0 && len(diff.Added) == 0 {
		pattern.Type = "removed"
		pattern.Before = strings.Join(diff.Removed, "; ")
	} else if len(diff.Rewritten) > 0 {
		pattern.Type = "rewritten"
		pattern.Before = strings.Join(diff.Removed, "; ")
		pattern.After = strings.Join(diff.Added, "; ")
	} else {
		// Mixed changes
		pattern.Type = "mixed"
		pattern.Before = strings.Join(diff.Removed, "; ")
		pattern.After = strings.Join(diff.Added, "; ")
	}

	// Detect specific edit category based on content
	pattern.Category = l.detectEditCategory(diff)

	return pattern
}

// detectEditCategory detects a specific category for the edit pattern.
func (l *BidirectionalLearner) detectEditCategory(diff *PromptDiff) string {
	// Look for common patterns in the changes
	combined := strings.Join(append(append(diff.Added, diff.Removed...), diff.Rewritten...), " ")

	categories := map[string][]string{
		"scope-specification": {"scope", "focus", "only", "modify", "files"},
		"success-criteria":    {"done", "when", "complete", "success", "verified"},
		"skill-invocation":    {"skill", "invoke", "load", "skill("},
		"out-of-scope":        {"not", "do not", "don't", "exclude", "avoid"},
		"file-reference":      {"@", "file", "path", "import"},
		"context-addition":    {"context", "note", "consider", "remember"},
	}

	for category, keywords := range categories {
		count := 0
		lowerCombined := strings.ToLower(combined)
		for _, keyword := range keywords {
			if strings.Contains(lowerCombined, keyword) {
				count++
			}
		}
		if count >= 2 {
			return category
		}
	}

	return "general"
}

// generatePatternKey generates a unique key for a pattern.
func (l *BidirectionalLearner) generatePatternKey(pattern LearnedPattern) string {
	// Create a key from category + type + content hash
	content := pattern.Category + pattern.Type + pattern.Before + pattern.After
	hash := simpleHash(content)
	return fmt.Sprintf("%s_%s_%d", pattern.Category, pattern.Type, hash)
}

// simpleHash creates a simple hash from a string.
func simpleHash(s string) int {
	hash := 0
	for i, c := range s {
		hash += int(c) * (i + 1)
	}
	return hash
}

// ShouldAutoApply checks if a pattern should be auto-applied for a category.
func (l *BidirectionalLearner) ShouldAutoApply(category string, patternType string) bool {
	for key := range l.autoApply {
		if strings.HasPrefix(key, category+"_"+patternType) {
			return true
		}
	}
	return false
}

// GetAutoApplyPatterns returns all patterns ready for auto-application.
func (l *BidirectionalLearner) GetAutoApplyPatterns() map[string]*LearnedPattern {
	result := make(map[string]*LearnedPattern)
	for key, pattern := range l.patterns {
		if pattern.AutoApply {
			result[key] = pattern
		}
	}
	return result
}

// savePatternToLog saves a pattern to the GROVE-OPTI-LOG.md file.
func (l *BidirectionalLearner) savePatternToLog(pattern *LearnedPattern) error {
	// Ensure the log file exists with proper header
	if _, err := os.Stat(l.logPath); os.IsNotExist(err) {
		header := `# GROVE Opti Prompt - Edit Patterns Log

This file tracks learned patterns from user edits to improve prompt optimization.

## Auto-Apply Patterns

Patterns with frequency >= 3 are automatically applied to future optimizations.

## Learned Patterns

| Category | Type | Before | After | Frequency | Auto-Apply |
|----------|------|--------|-------|-----------|------------|
`
		if err := os.WriteFile(l.logPath, []byte(header), 0644); err != nil {
			return groveerrors.NewOptiError("save_pattern", "create log file", err)
		}
	}

	// Read existing content
	content, err := os.ReadFile(l.logPath)
	if err != nil {
		return groveerrors.NewOptiError("save_pattern", "read log file", err)
	}

	// Build the new row
	newRow := fmt.Sprintf("| %s | %s | %s | %s | %d | %s |\n",
		pattern.Category,
		pattern.Type,
		truncateForLog(pattern.Before),
		truncateForLog(pattern.After),
		pattern.Frequency,
		formatBool(pattern.AutoApply),
	)

	// Check if pattern already exists in table
	existingRow := fmt.Sprintf("| %s | %s |", pattern.Category, pattern.Type)
	if strings.Contains(string(content), existingRow) {
		// Update existing row - this is a simplified approach
		// In production, you'd want more sophisticated update logic
		slog.Debug("pattern already exists in log (row update not implemented)")
		return nil
	}

	// Append new pattern row
	updatedContent := string(content) + newRow

	return os.WriteFile(l.logPath, []byte(updatedContent), 0644)
}

// loadPatternsFromLog loads patterns from the GROVE-OPTI-LOG.md file.
func loadPatternsFromLog(logPath string) (map[string]*LearnedPattern, error) {
	patterns := make(map[string]*LearnedPattern)

	content, err := os.ReadFile(logPath)
	if err != nil {
		return patterns, err
	}

	lines := strings.Split(string(content), "\n")
	inTable := false

	for _, line := range lines {
		// Look for the table header
		if strings.Contains(line, "| Category | Type |") {
			inTable = true
			continue
		}

		if inTable && strings.HasPrefix(line, "| ") {
			// Parse table row
			parts := strings.Split(line, "|")
			if len(parts) >= 6 {
				category := strings.TrimSpace(parts[1])
				patternType := strings.TrimSpace(parts[2])
				before := strings.TrimSpace(parts[3])
				after := strings.TrimSpace(parts[4])
				freqStr := strings.TrimSpace(parts[5])

				var frequency int
				fmt.Sscanf(freqStr, "%d", &frequency)

				key := fmt.Sprintf("%s_%s", category, patternType)
				patterns[key] = &LearnedPattern{
					Key:       key,
					Category:  category,
					Type:      patternType,
					Before:    before,
					After:     after,
					Frequency: frequency,
					AutoApply: frequency >= 3,
				}
			}
		}
	}

	return patterns, nil
}

// truncateForLog truncates a string for table display.
func truncateForLog(s string) string {
	maxLen := 50
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// formatBool formats a boolean for table display.
func formatBool(b bool) string {
	if b {
		return "✓"
	}
	return ""
}

// ApplyLearnedPatterns applies learned patterns to a prompt.
func (l *BidirectionalLearner) ApplyLearnedPatterns(prompt string, category string) string {
	result := prompt

	for key, pattern := range l.patterns {
		if !pattern.AutoApply {
			continue
		}

		// Only apply patterns for the same category or general ones
		if pattern.Category != category && pattern.Category != "general" {
			continue
		}

		switch pattern.Type {
		case "added":
			// Add the content if not present
			if !strings.Contains(result, pattern.After) {
				result += "\n" + pattern.After
			}
		case "removed":
			// Remove the content if present
			result = strings.ReplaceAll(result, pattern.Before, "")
		case "rewritten":
			// Replace before with after
			result = strings.ReplaceAll(result, pattern.Before, pattern.After)
		}

		slog.Debug("applied learned pattern",
			slog.String("key", key),
			slog.String("type", pattern.Type))
	}

	return result
}

// GetPatternCount returns the number of learned patterns.
func (l *BidirectionalLearner) GetPatternCount() int {
	return len(l.patterns)
}

// GetAutoApplyCount returns the number of patterns ready for auto-application.
func (l *BidirectionalLearner) GetAutoApplyCount() int {
	return len(l.autoApply)
}
