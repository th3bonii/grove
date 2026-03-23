// Package spec provides original idea tracking and comparison.
//
// This module ensures the engine never loses sight of the original idea
// while iterating and improving. Every iteration compares against the
// original to ensure completeness.
package spec

import (
	"fmt"
	"strings"
)

// =============================================================================
// Original Idea Tracking
// =============================================================================

// OriginalIdea represents the user's original idea.
type OriginalIdea struct {
	Content     string   `json:"content"`
	Components  []string `json:"components"`  // Components mentioned by user
	Keywords    []string `json:"keywords"`    // Key terms from original
	Intent      string   `json:"intent"`      // What user wants to achieve
	Constraints []string `json:"constraints"` // Any constraints mentioned
	References  []string `json:"references"`  // Visual references
}

// IdeaTracker tracks the original idea and ensures completeness.
type IdeaTracker struct {
	original *OriginalIdea
	covered  map[string]bool // What has been covered
	missing  []string        // What's still missing
	added    []string        // What was added beyond original
	drifted  []string        // What drifted from original intent
}

// NewIdeaTracker creates a new tracker from the original idea.
func NewIdeaTracker(content string) *IdeaTracker {
	original := parseOriginalIdea(content)
	return &IdeaTracker{
		original: original,
		covered:  make(map[string]bool),
		missing:  make([]string, 0),
		added:    make([]string, 0),
		drifted:  make([]string, 0),
	}
}

// parseOriginalIdea extracts components and keywords from the original idea.
func parseOriginalIdea(content string) *OriginalIdea {
	idea := &OriginalIdea{
		Content:     content,
		Components:  make([]string, 0),
		Keywords:    make([]string, 0),
		Constraints: make([]string, 0),
		References:  make([]string, 0),
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect component names (UI elements, features)
		if isComponentMention(line) {
			idea.Components = append(idea.Components, extractComponentName(line))
		}

		// Detect keywords
		idea.Keywords = append(idea.Keywords, extractKeywords(line)...)

		// Detect constraints
		if isConstraint(line) {
			idea.Constraints = append(idea.Constraints, line)
		}
	}

	// Extract intent
	idea.Intent = extractIntent(content)

	return idea
}

// CheckCompleteness compares current state against original idea.
func (t *IdeaTracker) CheckCompleteness(components []Component) CompletenessReport {
	report := CompletenessReport{
		OriginalComponents: len(t.original.Components),
		CoveredComponents:  0,
		MissingComponents:  make([]string, 0),
		ExtraComponents:    make([]string, 0),
		DriftDetected:      false,
		DriftItems:         make([]string, 0),
	}

	// Check if all original components are covered
	covered := make(map[string]bool)
	for _, comp := range components {
		covered[strings.ToLower(comp.Name)] = true
	}

	for _, origComp := range t.original.Components {
		if covered[strings.ToLower(origComp)] {
			report.CoveredComponents++
			t.covered[origComp] = true
		} else {
			report.MissingComponents = append(report.MissingComponents, origComp)
			t.missing = append(t.missing, origComp)
		}
	}

	// Check for extra components (added beyond original)
	originalSet := make(map[string]bool)
	for _, comp := range t.original.Components {
		originalSet[strings.ToLower(comp)] = true
	}

	for _, comp := range components {
		if !originalSet[strings.ToLower(comp.Name)] {
			report.ExtraComponents = append(report.ExtraComponents, comp.Name)
			t.added = append(t.added, comp.Name)
		}
	}

	// Check for drift from original intent
	drift := t.checkDrift(components)
	if len(drift) > 0 {
		report.DriftDetected = true
		report.DriftItems = drift
		t.drifted = drift
	}

	return report
}

// checkDrift detects if components drifted from original intent.
func (t *IdeaTracker) checkDrift(components []Component) []string {
	drift := make([]string, 0)

	for _, comp := range components {
		// Check if component aligns with original keywords
		compLower := strings.ToLower(comp.Name)
		aligned := false

		for _, keyword := range t.original.Keywords {
			if strings.Contains(compLower, strings.ToLower(keyword)) {
				aligned = true
				break
			}
		}

		// Check if component relates to original intent
		intentLower := strings.ToLower(t.original.Intent)
		if strings.Contains(compLower, intentLower) {
			aligned = true
		}

		if !aligned && len(t.original.Keywords) > 0 {
			drift = append(drift, comp.Name)
		}
	}

	return drift
}

// GetMissing returns components from original that are not covered.
func (t *IdeaTracker) GetMissing() []string {
	return t.missing
}

// GetAdded returns components added beyond original.
func (t *IdeaTracker) GetAdded() []string {
	return t.added
}

// GetDrifted returns components that drifted from original intent.
func (t *IdeaTracker) GetDrifted() []string {
	return t.drifted
}

// GetOriginalComponents returns components mentioned in original idea.
func (t *IdeaTracker) GetOriginalComponents() []string {
	return t.original.Components
}

// GetOriginalKeywords returns keywords from original idea.
func (t *IdeaTracker) GetOriginalKeywords() []string {
	return t.original.Keywords
}

// CompletenessReport represents the completeness check results.
type CompletenessReport struct {
	OriginalComponents int      `json:"original_components"`
	CoveredComponents  int      `json:"covered_components"`
	MissingComponents  []string `json:"missing_components"`
	ExtraComponents    []string `json:"extra_components"`
	DriftDetected      bool     `json:"drift_detected"`
	DriftItems         []string `json:"drift_items"`
}

// IsComplete returns true if all original components are covered.
func (r *CompletenessReport) IsComplete() bool {
	return len(r.MissingComponents) == 0
}

// HasDrift returns true if components drifted from original intent.
func (r *CompletenessReport) HasDrift() bool {
	return r.DriftDetected
}

// Summary returns a human-readable summary.
func (r *CompletenessReport) Summary() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("Original components: %d\n", r.OriginalComponents))
	sb.WriteString(fmt.Sprintf("Covered: %d\n", r.CoveredComponents))
	sb.WriteString(fmt.Sprintf("Missing: %d\n", len(r.MissingComponents)))
	sb.WriteString(fmt.Sprintf("Extra: %d\n", len(r.ExtraComponents)))
	sb.WriteString(fmt.Sprintf("Drift: %v\n", r.DriftDetected))

	if len(r.MissingComponents) > 0 {
		sb.WriteString("\nMissing components:\n")
		for _, comp := range r.MissingComponents {
			sb.WriteString(fmt.Sprintf("  - %s\n", comp))
		}
	}

	if len(r.DriftItems) > 0 {
		sb.WriteString("\nDrifted from original:\n")
		for _, item := range r.DriftItems {
			sb.WriteString(fmt.Sprintf("  - %s\n", item))
		}
	}

	return sb.String()
}

// Helper functions

func isComponentMention(line string) bool {
	lower := strings.ToLower(line)
	componentTerms := []string{
		"button", "input", "form", "modal", "nav", "navigation",
		"header", "footer", "sidebar", "menu", "tab", "panel",
		"card", "list", "table", "chart", "graph", "icon",
		"image", "video", "audio", "slider", "dropdown", "toggle",
		"checkbox", "radio", "progress", "loading", "spinner",
		"alert", "notification", "tooltip", "popover", "dialog",
	}

	for _, term := range componentTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}

func extractComponentName(line string) string {
	// Simple extraction: take the first noun-like word
	words := strings.Fields(line)
	for _, word := range words {
		if len(word) > 2 && !isStopWord(word) {
			return word
		}
	}
	return line
}

func extractKeywords(line string) []string {
	words := strings.Fields(strings.ToLower(line))
	keywords := make([]string, 0)

	for _, word := range words {
		if len(word) > 2 && !isStopWord(word) {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

func isConstraint(line string) bool {
	lower := strings.ToLower(line)
	constraintTerms := []string{
		"must", "should", "required", "constraint", "limit",
		"only", "never", "always", "cannot", "must not",
	}

	for _, term := range constraintTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}

func extractIntent(content string) string {
	// Extract the main intent from content
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "Build the specified application"
}

func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "to": true,
		"in": true, "on": true, "at": true, "for": true, "and": true,
		"or": true, "but": true, "not": true, "with": true, "this": true,
		"that": true, "it": true, "of": true, "as": true, "be": true,
		"by": true, "from": true, "has": true, "have": true, "was": true,
		"que": true, "un": true, "una": true, "el": true, "la": true,
		"los": true, "las": true, "del": true, "al": true, "en": true,
		"con": true, "por": true, "para": true, "como": true, "más": true,
	}
	return stopWords[word]
}
