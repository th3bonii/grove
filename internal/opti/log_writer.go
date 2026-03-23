package opti

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogWriter handles writing to the GROVE-OPTI-LOG.md file.
type LogWriter struct {
	logPath string
	mu      sync.Mutex
}

// LogData represents the complete log data structure.
type LogData struct {
	Invocations  []InvocationEntry `json:"invocations"`
	UserProfile  *UserProfileData  `json:"user_profile,omitempty"`
	EditPatterns []EditPatternData `json:"edit_patterns,omitempty"`
}

// InvocationEntry represents a single invocation log entry.
type InvocationEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Intent     string    `json:"intent"`
	Tokens     int       `json:"tokens"`
	Files      []string  `json:"files"`
	Layers     []int     `json:"layers"`
	UserAction string    `json:"user_action"` // send|edit|reject
	SkillsUsed []string  `json:"skills_used"`
}

// UserProfileData tracks user interaction patterns.
type UserProfileData struct {
	FileReference      CategoryStats `json:"file-reference"`
	ScopeBoundary      CategoryStats `json:"scope-boundary"`
	SkillInvocation    CategoryStats `json:"skill-invocation"`
	SuccessCriteria    CategoryStats `json:"success-criteria"`
	PlanMode           CategoryStats `json:"plan-mode"`
	OutOfScopeBoundary CategoryStats `json:"out-of-scope-boundary"`
}

// CategoryStats tracks interaction frequency for a category.
type CategoryStats struct {
	TimesSeen int    `json:"times_seen"`
	LastSeen  string `json:"last_seen"`
}

// EditPatternData represents a learned pattern from user edits.
type EditPatternData struct {
	Type      string `json:"pattern_type"` // added|removed|rewritten
	Category  string `json:"category"`
	Frequency int    `json:"frequency"`
	Before    string `json:"example_before"`
	After     string `json:"example_after"`
}

// NewLogWriter creates a new LogWriter for the specified project root.
func NewLogWriter(projectRoot string) *LogWriter {
	return &LogWriter{
		logPath: filepath.Join(projectRoot, "GROVE-OPTI-LOG.md"),
	}
}

// AppendInvocation adds a new invocation entry to the log.
func (w *LogWriter) AppendInvocation(entry InvocationEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := w.LoadOrCreate()
	if err != nil {
		return err
	}

	data.Invocations = append(data.Invocations, entry)

	return w.saveAll(data)
}

// UpdateUserProfile updates the user profile in the log.
func (w *LogWriter) UpdateUserProfile(profile UserProfileData) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := w.LoadOrCreate()
	if err != nil {
		return err
	}

	data.UserProfile = &profile

	return w.saveAll(data)
}

// AddEditPattern adds or updates an edit pattern in the log.
func (w *LogWriter) AddEditPattern(pattern EditPatternData) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := w.LoadOrCreate()
	if err != nil {
		return err
	}

	// Check if pattern already exists
	found := false
	for i, p := range data.EditPatterns {
		if p.Category == pattern.Category && p.Type == pattern.Type {
			data.EditPatterns[i].Frequency = p.Frequency + 1
			if pattern.Before != "" {
				data.EditPatterns[i].Before = pattern.Before
			}
			if pattern.After != "" {
				data.EditPatterns[i].After = pattern.After
			}
			found = true
			break
		}
	}

	if !found {
		data.EditPatterns = append(data.EditPatterns, pattern)
	}

	return w.saveAll(data)
}

// LoadOrCreate loads existing log data or creates a new log file.
func (w *LogWriter) LoadOrCreate() (*LogData, error) {
	content, err := os.ReadFile(w.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new log file with default structure
			newData := &LogData{
				Invocations: make([]InvocationEntry, 0),
				UserProfile: &UserProfileData{
					FileReference:      CategoryStats{TimesSeen: 0, LastSeen: ""},
					ScopeBoundary:      CategoryStats{TimesSeen: 0, LastSeen: ""},
					SkillInvocation:    CategoryStats{TimesSeen: 0, LastSeen: ""},
					SuccessCriteria:    CategoryStats{TimesSeen: 0, LastSeen: ""},
					PlanMode:           CategoryStats{TimesSeen: 0, LastSeen: ""},
					OutOfScopeBoundary: CategoryStats{TimesSeen: 0, LastSeen: ""},
				},
				EditPatterns: make([]EditPatternData, 0),
			}
			return newData, w.saveAll(newData)
		}
		return nil, err
	}

	// Parse existing log file
	data := &LogData{
		Invocations:  make([]InvocationEntry, 0),
		UserProfile:  &UserProfileData{},
		EditPatterns: make([]EditPatternData, 0),
	}

	// Parse Invocation Log section
	invocations := w.parseInvocationLog(string(content))
	data.Invocations = invocations

	// Parse User Profile section
	profile := w.parseUserProfile(string(content))
	if profile != nil {
		data.UserProfile = profile
	}

	// Parse Edit Patterns section
	patterns := w.parseEditPatterns(string(content))
	data.EditPatterns = patterns

	return data, nil
}

// parseInvocationLog parses the invocation log section from markdown.
func (w *LogWriter) parseInvocationLog(content string) []InvocationEntry {
	var entries []InvocationEntry

	// Find ## Invocation Log section
	sections := strings.Split(content, "## ")
	for _, section := range sections {
		if !strings.HasPrefix(section, "Invocation Log") {
			continue
		}

		lines := strings.Split(section, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Skip table header and separator
			if strings.HasPrefix(line, "|") && !strings.HasPrefix(line, "| Timestamp") && !strings.Contains(line, "|-----------|") {
				entry := w.parseInvocationRow(line)
				if entry.Intent != "" {
					entries = append(entries, entry)
				}
			}
		}
	}

	return entries
}

// parseInvocationRow parses a single invocation log row.
func (w *LogWriter) parseInvocationRow(line string) InvocationEntry {
	// Remove leading | and split
	line = strings.TrimPrefix(line, "|")
	parts := strings.Split(line, "|")
	if len(parts) < 6 {
		return InvocationEntry{}
	}

	// Parse each field
	timestamp := strings.TrimSpace(parts[0])
	intent := strings.TrimSpace(parts[1])
	tokens := 0
	fmt.Sscanf(strings.TrimSpace(parts[2]), "%d", &tokens)

	// Parse files array [file1 (L1), file2 (L2)]
	filesStr := strings.TrimSpace(parts[3])
	files := w.parseFileList(filesStr)

	// Parse layers
	layers := w.parseLayers(filesStr)

	userAction := strings.TrimSpace(parts[4])
	skillsStr := strings.TrimSpace(parts[5])
	skills := w.parseSkills(skillsStr)

	// Parse timestamp
	ts, _ := time.Parse(time.RFC3339, timestamp)
	if ts.IsZero() {
		ts, _ = time.Parse("2006-01-02T15:04:05Z", timestamp)
	}

	return InvocationEntry{
		Timestamp:  ts,
		Intent:     intent,
		Tokens:     tokens,
		Files:      files,
		Layers:     layers,
		UserAction: userAction,
		SkillsUsed: skills,
	}
}

// parseFileList parses a file list string like "[file1 (L1), file2 (L2)]".
func (w *LogWriter) parseFileList(s string) []string {
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" || s == "[]" {
		return []string{}
	}

	var files []string
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		// Extract file path before (L\d)
		if idx := strings.Index(p, " (L"); idx > 0 {
			p = p[:idx]
		}
		if p != "" {
			files = append(files, p)
		}
	}
	return files
}

// parseLayers parses layer numbers from file list string.
func (w *LogWriter) parseLayers(s string) []int {
	var layers []int
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		var layer int
		n, err := fmt.Sscanf(p, "%*s (L%d)", &layer)
		if n > 0 && err == nil {
			layers = append(layers, layer)
		}
	}
	return layers
}

// parseSkills parses skills from a comma-separated string.
func (w *LogWriter) parseSkills(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" || s == "[]" {
		return []string{}
	}

	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")

	var skills []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			skills = append(skills, p)
		}
	}
	return skills
}

// parseUserProfile parses the user profile section from markdown.
func (w *LogWriter) parseUserProfile(content string) *UserProfileData {
	sections := strings.Split(content, "## ")
	for _, section := range sections {
		if !strings.HasPrefix(section, "User Profile") {
			continue
		}

		// Find JSON block
		jsonStart := strings.Index(section, "```json")
		if jsonStart < 0 {
			jsonStart = strings.Index(section, "{")
		}
		jsonEnd := strings.LastIndex(section, "```")

		if jsonStart >= 0 && jsonEnd > jsonStart {
			var profile UserProfileData
			jsonContent := section[jsonStart:jsonEnd]
			// Remove markdown code fence if present
			jsonContent = strings.TrimPrefix(jsonContent, "```json")
			jsonContent = strings.TrimSuffix(jsonContent, "```")
			jsonContent = strings.TrimSpace(jsonContent)

			if err := json.Unmarshal([]byte(jsonContent), &profile); err == nil {
				return &profile
			}
		}
	}

	return nil
}

// parseEditPatterns parses the edit patterns section from markdown.
func (w *LogWriter) parseEditPatterns(content string) []EditPatternData {
	sections := strings.Split(content, "## ")
	for _, section := range sections {
		if !strings.HasPrefix(section, "Edit Patterns") {
			continue
		}

		// Try to find JSON array
		jsonStart := strings.Index(section, "[")
		jsonEnd := strings.LastIndex(section, "]")

		if jsonStart >= 0 && jsonEnd > jsonStart {
			var patterns []EditPatternData
			jsonContent := section[jsonStart : jsonEnd+1]

			if err := json.Unmarshal([]byte(jsonContent), &patterns); err == nil {
				return patterns
			}
		}

		// Fallback: parse table format
		return w.parseEditPatternsTable(section)
	}

	return nil
}

// parseEditPatternsTable parses edit patterns from table format.
func (w *LogWriter) parseEditPatternsTable(section string) []EditPatternData {
	var patterns []EditPatternData

	lines := strings.Split(section, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "|") && !strings.HasPrefix(line, "| Pattern") && !strings.Contains(line, "|---------|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 5 {
				pattern := EditPatternData{
					Type:     strings.TrimSpace(parts[1]),
					Category: strings.TrimSpace(parts[2]),
					// Frequency would need to be tracked separately
				}
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns
}

// saveAll saves the complete log data to markdown format.
func (w *LogWriter) saveAll(data *LogData) error {
	var sb strings.Builder

	// Write header
	sb.WriteString("# GROVE Opti Prompt - Invocation Log\n\n")

	// Write Invocation Log section
	sb.WriteString("## Invocation Log\n\n")
	sb.WriteString("| Timestamp | Intent | Tokens | Files | Action | Skills |\n")
	sb.WriteString("|-----------|--------|--------|-------|--------|--------|\n")

	for _, inv := range data.Invocations {
		filesStr := formatFiles(inv.Files, inv.Layers)
		skillsStr := formatSkills(inv.SkillsUsed)
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s | %s | %s |\n",
			inv.Timestamp.Format(time.RFC3339),
			inv.Intent,
			inv.Tokens,
			filesStr,
			inv.UserAction,
			skillsStr,
		))
	}

	sb.WriteString("\n")

	// Write User Profile section
	sb.WriteString("## User Profile\n\n")
	sb.WriteString("```json\n")
	profileJSON, err := json.MarshalIndent(data.UserProfile, "", "  ")
	if err != nil {
		return err
	}
	sb.Write(profileJSON)
	sb.WriteString("\n```\n\n")

	// Write Edit Patterns section
	sb.WriteString("## Edit Patterns\n\n")
	sb.WriteString("| Pattern | Category | Frequency | Example |\n")
	sb.WriteString("|---------|----------|-----------|---------|\n")

	for _, p := range data.EditPatterns {
		example := p.Before
		if example == "" {
			example = p.After
		}
		if len(example) > 50 {
			example = example[:50] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n",
			p.Type,
			p.Category,
			p.Frequency,
			example,
		))
	}

	// Also write JSON for Edit Patterns
	sb.WriteString("\n```json\n")
	patternsJSON, err := json.MarshalIndent(data.EditPatterns, "", "  ")
	if err != nil {
		return err
	}
	sb.Write(patternsJSON)
	sb.WriteString("\n```\n")

	return os.WriteFile(w.logPath, []byte(sb.String()), 0644)
}

// Helper functions

func formatFiles(files []string, layers []int) string {
	if len(files) == 0 {
		return "[]"
	}

	var parts []string
	for i, f := range files {
		layer := 0
		if i < len(layers) {
			layer = layers[i]
		}
		parts = append(parts, fmt.Sprintf("%s (L%d)", f, layer))
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

func formatSkills(skills []string) string {
	if len(skills) == 0 {
		return "[]"
	}
	return "[" + strings.Join(skills, ", ") + "]"
}
