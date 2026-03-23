package opti

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ExplanationLevel represents the depth of explanation to provide.
type ExplanationLevel int

const (
	ExplanationFull  ExplanationLevel = iota // Full 2-sentence explanation
	ExplanationShort                         // 1-sentence reminder
	ExplanationLabel                         // Label only, no explanation
)

// Note: LearnedPattern is defined in bidirectional.go and shared across the package

// UserProfile tracks the user's learning history for adaptive explanations.
type UserProfile struct {
	Categories map[string]CategoryProfile `json:"categories"`
}

// CategoryProfile tracks experience with a specific optimization category.
type CategoryProfile struct {
	TimesSeen int    `json:"times_seen"`
	LastSeen  string `json:"last_seen"` // ISO date string
}

// Explanation represents a single WHY explanation.
type Explanation struct {
	Category    PromptElementType `json:"category"`
	Level       ExplanationLevel  `json:"level"`
	Text        string            `json:"text"`
	IsNewLearn  bool              `json:"is_new_learn"` // True if this is a newly learned pattern
	PatternDesc string            `json:"pattern_desc"` // Description of learned pattern
}

// Explainer generates adaptive WHY explanations based on user profile.
type Explainer struct {
	logPath      string
	explainAll   bool
	noTeach      bool
	userProfile  *UserProfile
	editPatterns []LearnedPattern
}

// NewExplainer creates a new ExplanationGenerator.
func NewExplainer(projectRoot string, explainAll, noTeach bool) *Explainer {
	return &Explainer{
		logPath:      filepath.Join(projectRoot, "GROVE-OPTI-LOG.md"),
		explainAll:   explainAll,
		noTeach:      noTeach,
		userProfile:  loadUserProfile(filepath.Join(projectRoot, "GROVE-OPTI-LOG.md")),
		editPatterns: loadEditPatterns(filepath.Join(projectRoot, "GROVE-OPTI-LOG.md")),
	}
}

// GenerateExplanations generates WHY explanations for each prompt element.
// It adapts explanation depth based on user profile (times_seen).
func (e *Explainer) GenerateExplanations(prompt *OptimizedPrompt) []Explanation {
	if e.noTeach {
		return nil
	}

	var explanations []Explanation

	for i := range prompt.Elements {
		element := &prompt.Elements[i]
		profile := e.getCategoryProfile(string(element.Type))

		level := e.determineExplanationLevel(profile.TimesSeen)
		text := e.generateExplanationText(element, level)

		// Check for learned patterns
		isNewLearn, patternDesc := e.checkLearnedPattern(element.Type)

		explanations = append(explanations, Explanation{
			Category:    element.Type,
			Level:       level,
			Text:        text,
			IsNewLearn:  isNewLearn,
			PatternDesc: patternDesc,
		})

		// Apply learned pattern if available
		if isNewLearn && patternDesc != "" {
			element.Explanation = fmt.Sprintf("→ Learned: %s. %s", patternDesc, text)
		} else {
			element.Explanation = text
		}
	}

	return explanations
}

// determineExplanationLevel decides the explanation depth based on times_seen.
func (e *Explainer) determineExplanationLevel(timesSeen int) ExplanationLevel {
	if e.explainAll {
		return ExplanationFull
	}

	switch {
	case timesSeen <= 0:
		return ExplanationFull
	case timesSeen <= 3:
		return ExplanationFull
	case timesSeen <= 10:
		return ExplanationShort
	default:
		return ExplanationLabel
	}
}

// generateExplanationText generates the actual explanation text.
func (e *Explainer) generateExplanationText(element *PromptElement, level ExplanationLevel) string {
	// Templates for each category and level
	templates := map[PromptElementType]map[ExplanationLevel][3]string{
		ElementFileReference: {
			ExplanationFull: {
				"WHY: Adding %s ensures the agent edits the correct file instead of searching for it.",
				"Without this reference, the agent may modify the wrong component or introduce changes to unintended files.",
				"", // Combined full explanation
			},
			ExplanationShort: {
				"WHY: File reference added to scope the agent to the correct component.",
				"",
				"",
			},
			ExplanationLabel: {
				"[file-reference]",
				"",
				"",
			},
		},
		ElementScopeBoundary: {
			ExplanationFull: {
				"WHY: Scope boundaries help the agent understand exactly which files and components to focus on.",
				"Without clear scope, agents often make changes that are too broad or too narrow, leading to incomplete fixes or unintended modifications.",
				"",
			},
			ExplanationShort: {
				"WHY: Scope boundaries keep the agent focused on the right area.",
				"",
				"",
			},
			ExplanationLabel: {
				"[scope-boundary]",
				"",
				"",
			},
		},
		ElementSkillInvocation: {
			ExplanationFull: {
				"WHY: Invoking %s ensures the agent uses the correct workflow and conventions for this task.",
				"Skills provide specific instructions and patterns that improve consistency and reduce errors.",
				"",
			},
			ExplanationShort: {
				"WHY: Skill invocation ensures the agent follows the correct approach.",
				"",
				"",
			},
			ExplanationLabel: {
				"[skill-invocation]",
				"",
				"",
			},
		},
		ElementSuccessCriteria: {
			ExplanationFull: {
				"WHY: Success criteria define when the task is complete, preventing scope creep.",
				"Without clear criteria, agents may consider partial implementations as complete, leading to rework.",
				"",
			},
			ExplanationShort: {
				"WHY: Clear success criteria prevent incomplete implementations.",
				"",
				"",
			},
			ExplanationLabel: {
				"[success-criteria]",
				"",
				"",
			},
		},
		ElementPlanMode: {
			ExplanationFull: {
				"WHY: Plan mode is recommended for complex changes to ensure all implications are considered.",
				"This prevents costly mistakes that would require rollback or extensive rework.",
				"",
			},
			ExplanationShort: {
				"WHY: Plan mode helps avoid costly mistakes on complex changes.",
				"",
				"",
			},
			ExplanationLabel: {
				"[plan-mode]",
				"",
				"",
			},
		},
		ElementOutOfScope: {
			ExplanationFull: {
				"WHY: Explicitly stating what NOT to change reduces the risk of accidental modifications.",
				"This is especially important in projects with interdependent components.",
				"",
			},
			ExplanationShort: {
				"WHY: Out-of-scope boundaries prevent unintended changes.",
				"",
				"",
			},
			ExplanationLabel: {
				"[out-of-scope-boundary]",
				"",
				"",
			},
		},
	}

	// Get template for this element type
	typeTemplates, ok := templates[element.Type]
	if !ok {
		return "WHY: This optimization improves prompt clarity."
	}

	// Get text based on level
	var text string
	switch level {
	case ExplanationFull:
		if typeTemplates[ExplanationFull][2] != "" {
			text = typeTemplates[ExplanationFull][0] + " " + typeTemplates[ExplanationFull][1]
		} else {
			text = typeTemplates[ExplanationFull][0]
		}
	case ExplanationShort:
		text = typeTemplates[ExplanationShort][0]
	case ExplanationLabel:
		text = typeTemplates[ExplanationLabel][0]
	}

	// Fill in placeholders
	text = fmt.Sprintf(text, extractPlaceholder(element.Content))

	return text
}

// GenerateAdapted generates an explanation adapted to the user's experience level.
// It provides full explanation for new patterns, short reminders for familiar ones,
// and labels only for expert users.
func (e *Explainer) GenerateAdapted(category string, profile *CategoryProfile) string {
	if profile == nil {
		return e.getFullExplanation(category)
	}

	timesSeen := profile.TimesSeen

	if timesSeen <= 3 {
		// Full explanation (2 sentences) for new patterns
		return e.getFullExplanation(category)
	} else if timesSeen <= 10 {
		// Short reminder (1 sentence) for familiar patterns
		return e.getShortExplanation(category)
	} else {
		// Label only for expert users
		return fmt.Sprintf("[%s]", category)
	}
}

// getFullExplanation returns the full 2-sentence explanation for a category.
func (e *Explainer) getFullExplanation(category string) string {
	templates := map[string][2]string{
		"file-reference": {
			"WHY: Adding file references ensures the agent edits the correct file instead of searching for it.",
			"Without this reference, the agent may modify the wrong component or introduce changes to unintended files.",
		},
		"scope-boundary": {
			"WHY: Scope boundaries help the agent understand exactly which files and components to focus on.",
			"Without clear scope, agents often make changes that are too broad or too narrow.",
		},
		"skill-invocation": {
			"WHY: Invoking skills ensures the agent uses the correct workflow and conventions for this task.",
			"Skills provide specific instructions and patterns that improve consistency and reduce errors.",
		},
		"success-criteria": {
			"WHY: Success criteria define when the task is complete, preventing scope creep.",
			"Without clear criteria, agents may consider partial implementations as complete.",
		},
		"plan-mode": {
			"WHY: Plan mode is recommended for complex changes to ensure all implications are considered.",
			"This prevents costly mistakes that would require rollback or extensive rework.",
		},
		"out-of-scope-boundary": {
			"WHY: Explicitly stating what NOT to change reduces the risk of accidental modifications.",
			"This is especially important in projects with interdependent components.",
		},
	}

	if template, ok := templates[category]; ok {
		return template[0] + " " + template[1]
	}

	return fmt.Sprintf("WHY: This optimization improves prompt clarity for [%s].", category)
}

// getShortExplanation returns a single-sentence reminder for a category.
func (e *Explainer) getShortExplanation(category string) string {
	templates := map[string]string{
		"file-reference":        "WHY: File reference added to scope the agent to the correct component.",
		"scope-boundary":        "WHY: Scope boundaries keep the agent focused on the right area.",
		"skill-invocation":      "WHY: Skill invocation ensures the agent follows the correct approach.",
		"success-criteria":      "WHY: Clear success criteria prevent incomplete implementations.",
		"plan-mode":             "WHY: Plan mode helps avoid costly mistakes on complex changes.",
		"out-of-scope-boundary": "WHY: Out-of-scope boundaries prevent unintended changes.",
	}

	if template, ok := templates[category]; ok {
		return template
	}

	return fmt.Sprintf("WHY: [%s] optimization applied.", category)
}

// extractPlaceholder extracts a meaningful placeholder from content.
func extractPlaceholder(content string) string {
	// For file references, extract the file name
	if strings.HasPrefix(content, "@") {
		parts := strings.Split(content, "/")
		return strings.TrimPrefix(parts[len(parts)-1], "@")
	}

	// For skill invocations, extract skill name
	if strings.HasPrefix(content, "skill(") {
		re := regexp.MustCompile(`name:\s*['"]([^'"]+)['"]`)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Return first 30 chars as fallback
	if len(content) > 30 {
		return content[:30] + "..."
	}
	return content
}

// getCategoryProfile retrieves or initializes a category profile.
func (e *Explainer) getCategoryProfile(category string) *CategoryProfile {
	if e.userProfile == nil {
		e.userProfile = &UserProfile{
			Categories: make(map[string]CategoryProfile),
		}
	}

	profile, exists := e.userProfile.Categories[category]
	if !exists {
		profile = CategoryProfile{
			TimesSeen: 0,
			LastSeen:  "",
		}
	}

	return &profile
}

// checkLearnedPattern checks if a learned pattern applies to this element.
func (e *Explainer) checkLearnedPattern(elementType PromptElementType) (bool, string) {
	category := string(elementType)

	for _, pattern := range e.editPatterns {
		if pattern.Category == category && pattern.Frequency >= 3 {
			switch pattern.Type {
			case "added":
				return true, fmt.Sprintf("you prefer to add %s", category)
			case "removed":
				return true, fmt.Sprintf("you prefer to simplify %s", category)
			case "rewritten":
				return true, fmt.Sprintf("you prefer to rewrite %s", category)
			}
		}
	}

	return false, ""
}

// UpdateProfile updates the user profile when user sends or edits.
func (e *Explainer) UpdateProfile(elements []PromptElement, action string) error {
	if action == "reject" {
		// Do not update on reject
		return nil
	}

	if e.userProfile == nil {
		e.userProfile = &UserProfile{
			Categories: make(map[string]CategoryProfile),
		}
	}

	now := time.Now().Format("2006-01-02")

	for _, element := range elements {
		category := string(element.Type)
		profile := e.userProfile.Categories[category]

		// Only increment times_seen on send or edit
		if action == "send" || action == "edit" {
			profile.TimesSeen++
		}

		// Always update last_seen when category appears
		profile.LastSeen = now
		e.userProfile.Categories[category] = profile
	}

	// Save updated profile
	return e.saveUserProfile()
}

// RecordEditPattern records a learned pattern from user edits.
func (e *Explainer) RecordEditPattern(original, final string, category PromptElementType) error {
	// Determine pattern type
	var patternType string
	if len(final) > len(original) {
		patternType = "added"
	} else if len(final) < len(original) {
		patternType = "removed"
	} else {
		patternType = "rewritten"
	}

	// Find existing pattern
	for i := range e.editPatterns {
		if e.editPatterns[i].Category == string(category) && e.editPatterns[i].Type == patternType {
			e.editPatterns[i].Frequency++
			if e.editPatterns[i].Before == "" {
				e.editPatterns[i].Before = truncate(original, 100)
			}
			e.editPatterns[i].After = truncate(final, 100)
			return e.saveEditPatterns()
		}
	}

	// Create new pattern
	e.editPatterns = append(e.editPatterns, LearnedPattern{
		Type:      patternType,
		Category:  string(category),
		Frequency: 1,
		Before:    truncate(original, 100),
		After:     truncate(final, 100),
	})

	return e.saveEditPatterns()
}

// LogInvocation records an invocation in the log file.
func (e *Explainer) LogInvocation(classification IntentClassification, tokens int, files []FileCandidate, action string, skills []string) error {
	entry := InvocationLogEntry{
		Timestamp:            time.Now().Format(time.RFC3339),
		IntentClassification: string(classification.Intent),
		TokensUsed:           tokens,
		FilesSelected:        files,
		UserAction:           action,
		SkillsReferenced:     skills,
	}

	return e.appendInvocationLog(entry)
}

// Helper functions

func truncate(s string, maxLen int) string {
	// Handle edge cases
	if maxLen <= 0 {
		return "..."
	}
	// Use runes to handle Unicode correctly (emojis, accented chars, etc.)
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func loadUserProfile(logPath string) *UserProfile {
	content, err := os.ReadFile(logPath)
	if err != nil {
		return nil
	}

	// Find ## User Profile section
	sections := strings.Split(string(content), "## ")
	for _, section := range sections {
		if strings.HasPrefix(section, "User Profile") {
			// Extract JSON content (skip the header lines)
			lines := strings.Split(section, "\n")
			var jsonLines []string
			inJson := false
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "```json") || strings.HasPrefix(strings.TrimSpace(line), "{") {
					inJson = true
				}
				if inJson && !strings.HasPrefix(strings.TrimSpace(line), "```") {
					jsonLines = append(jsonLines, line)
				}
				if inJson && strings.HasPrefix(strings.TrimSpace(line), "```") {
					break
				}
			}

			if len(jsonLines) > 0 {
				var profile UserProfile
				jsonContent := strings.Join(jsonLines, "\n")
				if err := json.Unmarshal([]byte(jsonContent), &profile); err == nil {
					return &profile
				}
			}
		}
	}

	return nil
}

func loadEditPatterns(logPath string) []LearnedPattern {
	content, err := os.ReadFile(logPath)
	if err != nil {
		return nil
	}

	// Find ## Edit Patterns section
	sections := strings.Split(string(content), "## ")
	for _, section := range sections {
		if strings.HasPrefix(section, "Edit Patterns") {
			// Try to parse as JSON array
			var patterns []LearnedPattern
			jsonStart := strings.Index(section, "[")
			jsonEnd := strings.LastIndex(section, "]")
			if jsonStart >= 0 && jsonEnd > jsonStart {
				jsonContent := section[jsonStart : jsonEnd+1]
				if err := json.Unmarshal([]byte(jsonContent), &patterns); err == nil {
					return patterns
				}
			}
		}
	}

	return nil
}

func (e *Explainer) saveUserProfile() error {
	if e.userProfile == nil {
		return nil
	}

	content, err := os.ReadFile(e.logPath)
	if err != nil {
		// File doesn't exist, create it
		content = []byte{}
	}

	// Generate JSON
	jsonContent, err := json.MarshalIndent(e.userProfile, "", "  ")
	if err != nil {
		return err
	}

	// Find and replace User Profile section
	var newContent strings.Builder
	inUserProfile := false
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "## User Profile" {
			inUserProfile = true
			newContent.WriteString("## User Profile\n\n```json\n")
			newContent.Write(jsonContent)
			newContent.WriteString("\n```\n\n")
			continue
		}

		if inUserProfile {
			// Skip until we hit next section or end
			if strings.HasPrefix(line, "## ") || line == "" {
				inUserProfile = false
				newContent.WriteString(line + "\n")
			}
			continue
		}

		newContent.WriteString(line + "\n")
	}

	return os.WriteFile(e.logPath, []byte(newContent.String()), 0644)
}

func (e *Explainer) saveEditPatterns() error {
	content, err := os.ReadFile(e.logPath)
	if err != nil {
		content = []byte{}
	}

	// Generate JSON
	jsonContent, err := json.MarshalIndent(e.editPatterns, "", "  ")
	if err != nil {
		return err
	}

	// Find and replace Edit Patterns section
	var newContent strings.Builder
	inEditPatterns := false
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "## Edit Patterns" {
			inEditPatterns = true
			newContent.WriteString("## Edit Patterns\n\n```json\n")
			newContent.Write(jsonContent)
			newContent.WriteString("\n```\n\n")
			continue
		}

		if inEditPatterns {
			if strings.HasPrefix(line, "## ") || line == "" {
				inEditPatterns = false
				newContent.WriteString(line + "\n")
			}
			continue
		}

		newContent.WriteString(line + "\n")
	}

	return os.WriteFile(e.logPath, []byte(newContent.String()), 0644)
}

// InvocationLogEntry represents a single invocation log entry.
type InvocationLogEntry struct {
	Timestamp            string          `json:"timestamp"`
	IntentClassification string          `json:"intent_classification"`
	TokensUsed           int             `json:"tokens_used"`
	FilesSelected        []FileCandidate `json:"files_selected"`
	UserAction           string          `json:"user_action"`
	SkillsReferenced     []string        `json:"skills_referenced"`
}

func (e *Explainer) appendInvocationLog(entry InvocationLogEntry) error {
	// Format as markdown table row
	row := fmt.Sprintf("| %s | %s | %d | %s | %s | %s |",
		entry.Timestamp,
		entry.IntentClassification,
		entry.TokensUsed,
		formatFileList(entry.FilesSelected),
		entry.UserAction,
		strings.Join(entry.SkillsReferenced, ", "),
	)

	// Check if file exists and has headers
	content, err := os.ReadFile(e.logPath)
	if err != nil {
		// Create new log file
		header := `## Invocation Log

| Timestamp | Intent | Tokens | Files | Action | Skills |
|-----------|--------|--------|-------|--------|--------|
`
		content = []byte(header + row + "\n")
		return os.WriteFile(e.logPath, content, 0644)
	}

	// Append to existing log
	contentStr := string(content)
	if !strings.Contains(contentStr, "## Invocation Log") {
		contentStr = "## Invocation Log\n\n| Timestamp | Intent | Tokens | Files | Action | Skills |\n|-----------|--------|--------|-------|--------|--------|\n" + contentStr
	}

	return os.WriteFile(e.logPath, []byte(contentStr+row+"\n"), 0644)
}

func formatFileList(files []FileCandidate) string {
	if len(files) == 0 {
		return "[]"
	}

	var parts []string
	for _, f := range files {
		parts = append(parts, fmt.Sprintf("%s (L%d)", f.Path, f.Layer))
	}

	return "[" + strings.Join(parts, ", ") + "]"
}
