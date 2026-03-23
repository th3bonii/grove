package opti

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Classifier Tests
// =============================================================================

func TestClassifier_Classify_FeatureAddition(t *testing.T) {
	classifier := NewClassifier()

	tests := []struct {
		name     string
		input    string
		expected Intent
	}{
		{
			name:     "Simple add feature",
			input:    "add dark mode toggle to settings page",
			expected: IntentFeatureAddition,
		},
		{
			name:     "Create new component",
			input:    "create a new user profile component",
			expected: IntentFeatureAddition,
		},
		{
			name:     "Implement feature",
			input:    "implement pagination for the dashboard",
			expected: IntentFeatureAddition,
		},
		{
			name:     "Build new module",
			input:    "build an export to PDF feature",
			expected: IntentFeatureAddition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.input)
			if result.Intent != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.input, result.Intent, tt.expected)
			}
		})
	}
}

func TestClassifier_Classify_BugFix(t *testing.T) {
	classifier := NewClassifier()

	tests := []struct {
		name     string
		input    string
		expected Intent
	}{
		{
			name:     "Fix bug",
			input:    "fix the login button not working on mobile",
			expected: IntentBugFix,
		},
		{
			name:     "Fix error",
			input:    "fix the null pointer exception in auth service",
			expected: IntentBugFix,
		},
		{
			name:     "Bug report",
			input:    "there's a bug in the dashboard where charts don't load",
			expected: IntentBugFix,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.input)
			if result.Intent != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.input, result.Intent, tt.expected)
			}
		})
	}
}

func TestClassifier_Classify_Refactor(t *testing.T) {
	classifier := NewClassifier()

	tests := []struct {
		name     string
		input    string
		expected Intent
	}{
		{
			name:     "Refactor code",
			input:    "refactor the user service to use dependency injection",
			expected: IntentRefactor,
		},
		{
			name:     "Restructure",
			input:    "restructure the components folder",
			expected: IntentRefactor,
		},
		{
			name:     "Clean up",
			input:    "clean up the utils module",
			expected: IntentRefactor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.input)
			if result.Intent != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.input, result.Intent, tt.expected)
			}
		})
	}
}

func TestClassifier_Classify_Documentation(t *testing.T) {
	classifier := NewClassifier()

	tests := []struct {
		name     string
		input    string
		expected Intent
	}{
		{
			name:     "Update docs",
			input:    "update the API documentation",
			expected: IntentDocumentationUpdate,
		},
		{
			name:     "Add comments",
			input:    "add comments to the auth module",
			expected: IntentDocumentationUpdate,
		},
		{
			name:     "Write readme",
			input:    "write documentation for the new feature",
			expected: IntentDocumentationUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.input)
			if result.Intent != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.input, result.Intent, tt.expected)
			}
		})
	}
}

func TestClassifier_Classify_Configuration(t *testing.T) {
	classifier := NewClassifier()

	tests := []struct {
		name     string
		input    string
		expected Intent
	}{
		{
			name:     "Update config",
			input:    "update the database connection settings",
			expected: IntentConfigurationChange,
		},
		{
			name:     "Change env",
			input:    "change the environment variables for production",
			expected: IntentConfigurationChange,
		},
		{
			name:     "Add setting",
			input:    "add a new configuration option for rate limiting",
			expected: IntentConfigurationChange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.input)
			if result.Intent != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.input, result.Intent, tt.expected)
			}
		})
	}
}

func TestClassifier_Classify_Other(t *testing.T) {
	classifier := NewClassifier()

	tests := []struct {
		name     string
		input    string
		expected Intent
	}{
		{
			name:     "Vague input",
			input:    "make it work better",
			expected: IntentOther,
		},
		{
			name:     "Unclear intent",
			input:    "something is wrong with the app",
			expected: IntentOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.input)
			if result.Intent != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.input, result.Intent, tt.expected)
			}
		})
	}
}

func TestClassifier_ExtractDomain(t *testing.T) {
	classifier := NewClassifier()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Settings domain",
			input:    "add dark mode to settings page",
			expected: "Settings",
		},
		{
			name:     "Auth domain",
			input:    "fix the login flow",
			expected: "Auth",
		},
		{
			name:     "Dashboard domain",
			input:    "update the dashboard component",
			expected: "Dashboard",
		},
		{
			name:     "No domain",
			input:    "do something generic",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.extractDomain(tt.input)
			if result != tt.expected {
				t.Errorf("extractDomain(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestClassifier_ExtractKeywords(t *testing.T) {
	classifier := NewClassifier()

	tests := []struct {
		name     string
		input    string
		minCount int
	}{
		{
			name:     "Feature keywords",
			input:    "add dark mode toggle to settings page",
			minCount: 3, // dark, mode, settings, page, toggle
		},
		{
			name:     "CamelCase split",
			input:    "update the useAuth hook",
			minCount: 3, // update, useAuth, use, auth, hook
		},
		{
			name:     "Kebab case",
			input:    "fix the user-profile component",
			minCount: 2, // user, profile, component
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.extractKeywords(tt.input)
			if len(result) < tt.minCount {
				t.Errorf("extractKeywords(%q) returned %v (%d items), want at least %d",
					tt.input, result, len(result), tt.minCount)
			}
		})
	}
}

func TestClassifier_ExtractKeywords_StopsWords(t *testing.T) {
	classifier := NewClassifier()

	input := "the quick brown fox jumps over the lazy dog"
	result := classifier.extractKeywords(input)

	// Should not contain stop words
	for _, word := range result {
		lower := strings.ToLower(word)
		if lower == "the" || lower == "over" || lower == "the" {
			t.Errorf("extractKeywords(%q) contains stop word: %q", input, word)
		}
	}
}

func TestSplitCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "useAuth",
			expected: []string{"use", "Auth"},
		},
		{
			input:    "UserProfile",
			expected: []string{"User", "Profile"},
		},
		{
			input:    "getUserById",
			expected: []string{"get", "User", "By", "Id"},
		},
		{
			input:    "darkMode",
			expected: []string{"dark", "Mode"},
		},
		{
			input:    "simple",
			expected: []string{"simple"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitCamelCase(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitCamelCase(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestClassifier_Confidence(t *testing.T) {
	classifier := NewClassifier()

	// High confidence: clear bug fix language
	bugResult := classifier.Classify("fix the null pointer exception in user service")
	if bugResult.Confidence < 0.5 {
		t.Errorf("Bug fix should have high confidence, got %f", bugResult.Confidence)
	}

	// Lower confidence: vague input
	vagueResult := classifier.Classify("make things work")
	if vagueResult.Confidence > 0.6 {
		t.Errorf("Vague input should have lower confidence, got %f", vagueResult.Confidence)
	}
}

// =============================================================================
// UserProfile Tests
// =============================================================================

func TestUserProfile_CategoryTracking(t *testing.T) {
	profile := &UserProfile{
		Categories: map[string]CategoryProfile{
			"file-reference":   {TimesSeen: 5, LastSeen: "2024-01-15"},
			"scope-boundary":   {TimesSeen: 2, LastSeen: "2024-01-14"},
			"success-criteria": {TimesSeen: 0, LastSeen: ""},
		},
	}

	// Test getting existing category
	if cat, ok := profile.Categories["file-reference"]; !ok {
		t.Error("Expected file-reference category to exist")
	} else if cat.TimesSeen != 5 {
		t.Errorf("Expected TimesSeen=5, got %d", cat.TimesSeen)
	}

	// Test getting non-existing category (should get zero values)
	if cat, ok := profile.Categories["new-category"]; ok {
		t.Error("Expected new-category to NOT exist")
	} else if cat.TimesSeen != 0 {
		t.Errorf("Expected default TimesSeen=0, got %d", cat.TimesSeen)
	}
}

func TestUserProfile_JSONSerialization(t *testing.T) {
	profile := &UserProfile{
		Categories: map[string]CategoryProfile{
			"file-reference": {TimesSeen: 10, LastSeen: "2024-01-15"},
			"scope-boundary": {TimesSeen: 3, LastSeen: "2024-01-14"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("Failed to marshal UserProfile: %v", err)
	}

	// Unmarshal back
	var result UserProfile
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal UserProfile: %v", err)
	}

	// Verify
	if len(result.Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(result.Categories))
	}

	if result.Categories["file-reference"].TimesSeen != 10 {
		t.Errorf("Expected TimesSeen=10, got %d", result.Categories["file-reference"].TimesSeen)
	}
}

func TestExplainer_DetermineExplanationLevel(t *testing.T) {
	explainer := &Explainer{
		explainAll: false,
	}

	tests := []struct {
		name      string
		timesSeen int
		expected  ExplanationLevel
	}{
		{"Never seen", 0, ExplanationFull},
		{"Seen 1 time", 1, ExplanationFull},
		{"Seen 3 times", 3, ExplanationFull},
		{"Seen 4 times", 4, ExplanationShort},
		{"Seen 7 times", 7, ExplanationShort},
		{"Seen 10 times", 10, ExplanationShort},
		{"Seen 11 times", 11, ExplanationLabel},
		{"Seen 20 times", 20, ExplanationLabel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := explainer.determineExplanationLevel(tt.timesSeen)
			if result != tt.expected {
				t.Errorf("determineExplanationLevel(%d) = %v, want %v",
					tt.timesSeen, result, tt.expected)
			}
		})
	}
}

func TestExplainer_DetermineExplanationLevel_ExplainAllFlag(t *testing.T) {
	explainer := &Explainer{
		explainAll: true, // Force full explanations
	}

	// Even after 100 uses, should still give full explanation
	result := explainer.determineExplanationLevel(100)
	if result != ExplanationFull {
		t.Errorf("With explainAll=true, expected ExplanationFull, got %v", result)
	}
}

func TestExplainer_GenerateExplanationText(t *testing.T) {
	explainer := &Explainer{}

	tests := []struct {
		elementType PromptElementType
		level       ExplanationLevel
		shouldHave  string
	}{
		{ElementFileReference, ExplanationFull, "WHY:"},
		{ElementFileReference, ExplanationShort, "WHY:"},
		{ElementFileReference, ExplanationLabel, "[file-reference]"},
		{ElementScopeBoundary, ExplanationLabel, "[scope-boundary]"},
		{ElementSuccessCriteria, ExplanationFull, "WHY:"},
		{ElementPlanMode, ExplanationLabel, "[plan-mode]"},
	}

	for _, tt := range tests {
		t.Run(string(tt.elementType)+"_"+tt.level.String(), func(t *testing.T) {
			element := &PromptElement{
				Type:    tt.elementType,
				Content: "test content",
			}
			result := explainer.generateExplanationText(element, tt.level)
			if !strings.Contains(result, tt.shouldHave) {
				t.Errorf("Expected explanation to contain %q, got %q", tt.shouldHave, result)
			}
		})
	}
}

func TestExplainer_UpdateProfile(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "GROVE-OPTI-LOG.md")

	explainer := &Explainer{
		logPath:     logPath,
		noTeach:     false,
		userProfile: &UserProfile{Categories: make(map[string]CategoryProfile)},
	}

	elements := []PromptElement{
		{Type: ElementFileReference},
		{Type: ElementScopeBoundary},
		{Type: ElementSuccessCriteria},
	}

	// Test update on send
	err := explainer.UpdateProfile(elements, "send")
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}

	// Verify profile was updated
	if explainer.userProfile.Categories["file-reference"].TimesSeen != 1 {
		t.Errorf("Expected TimesSeen=1, got %d",
			explainer.userProfile.Categories["file-reference"].TimesSeen)
	}

	// Test update on reject (should not increment)
	err = explainer.UpdateProfile(elements, "reject")
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}

	// times_seen should still be 1
	if explainer.userProfile.Categories["file-reference"].TimesSeen != 1 {
		t.Errorf("Reject should not increment, expected 1, got %d",
			explainer.userProfile.Categories["file-reference"].TimesSeen)
	}
}

func TestExplainer_LastSeenUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "GROVE-OPTI-LOG.md")

	explainer := &Explainer{
		logPath:     logPath,
		noTeach:     false,
		userProfile: &UserProfile{Categories: make(map[string]CategoryProfile)},
	}

	elements := []PromptElement{
		{Type: ElementFileReference, Content: "test"},
	}

	// Update profile
	explainer.UpdateProfile(elements, "send")

	// Check last_seen was updated
	lastSeen := explainer.userProfile.Categories["file-reference"].LastSeen
	today := time.Now().Format("2006-01-02")

	if lastSeen != today {
		t.Errorf("Expected lastSeen=%s, got %s", today, lastSeen)
	}
}

func TestExplainer_RecordEditPattern(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "GROVE-OPTI-LOG.md")

	explainer := &Explainer{
		logPath:      logPath,
		editPatterns: make([]EditPattern, 0),
	}

	// Record a new pattern
	original := "@src/components/Auth.tsx"
	final := "@src/components/Auth.tsx\nskill({ name: 'auth-skill' })"

	err := explainer.RecordEditPattern(original, final, ElementSkillInvocation)
	if err != nil {
		t.Fatalf("RecordEditPattern failed: %v", err)
	}

	if len(explainer.editPatterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(explainer.editPatterns))
	}

	if explainer.editPatterns[0].PatternType != "added" {
		t.Errorf("Expected patternType='added', got %s", explainer.editPatterns[0].PatternType)
	}

	if explainer.editPatterns[0].Frequency != 1 {
		t.Errorf("Expected frequency=1, got %d", explainer.editPatterns[0].Frequency)
	}
}

func TestExplainer_LearnedPatternDetection(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "GROVE-OPTI-LOG.md")

	explainer := &Explainer{
		logPath: logPath,
		editPatterns: []EditPattern{
			{
				PatternType:   "added",
				Category:      "skill-invocation",
				Frequency:     3, // Threshold for auto-application
				ExampleBefore: "no skill",
				ExampleAfter:  "with skill",
			},
		},
	}

	// Check if pattern is detected as "learned"
	isNewLearn, patternDesc := explainer.checkLearnedPattern(ElementSkillInvocation)

	if !isNewLearn {
		t.Error("Expected learned pattern to be detected")
	}

	if !strings.Contains(patternDesc, "skill-invocation") {
		t.Errorf("Expected pattern description to contain 'skill-invocation', got %s", patternDesc)
	}
}

func TestExplainer_NotLearnedPatternDetection(t *testing.T) {
	explainer := &Explainer{
		editPatterns: []EditPattern{
			{
				PatternType:   "added",
				Category:      "skill-invocation",
				Frequency:     2, // Below threshold
				ExampleBefore: "no skill",
				ExampleAfter:  "with skill",
			},
		},
	}

	isNewLearn, _ := explainer.checkLearnedPattern(ElementSkillInvocation)

	if isNewLearn {
		t.Error("Expected pattern NOT to be detected as learned (frequency < 3)")
	}
}

func TestExplainer_LogInvocation(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "GROVE-OPTI-LOG.md")

	explainer := &Explainer{
		logPath: logPath,
	}

	entry := InvocationLogEntry{
		Timestamp:            "2024-01-15T10:30:00Z",
		IntentClassification: "bug-fix",
		TokensUsed:           1250,
		FilesSelected: []FileCandidate{
			{Path: "/src/auth/service.go", Layer: 1, Score: 1.0},
		},
		UserAction:       "send",
		SkillsReferenced: []string{"auth-skill"},
	}

	err := explainer.LogInvocation(
		IntentClassification{Intent: "bug-fix"},
		1250,
		[]FileCandidate{{Path: "/src/auth/service.go", Layer: 1}},
		"send",
		[]string{"auth-skill"},
	)
	if err != nil {
		t.Fatalf("LogInvocation failed: %v", err)
	}

	// Verify log file was created
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "## Invocation Log") {
		t.Error("Expected log file to contain Invocation Log section")
	}

	if !strings.Contains(string(content), "bug-fix") {
		t.Error("Expected log to contain intent classification")
	}

	if !strings.Contains(string(content), "1250") {
		t.Error("Expected log to contain token count")
	}
}

func TestExplainer_ExtractPlaceholder(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"@src/components/Settings.tsx", "Settings.tsx"},
		{"skill({ name: 'auth-skill' })", "auth-skill"},
		{"This is a plain text response", "This is a plain text response"},
		{"@utils/helpers.ts", "helpers.ts"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractPlaceholder(tt.input)
			if result != tt.expected {
				t.Errorf("extractPlaceholder(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is a ..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestClassifier_Integration(t *testing.T) {
	classifier := NewClassifier()

	// Real-world scenario: bug report with file reference
	input := "fix the memory leak in useAuth hook"

	result := classifier.Classify(input)

	// Verify intent
	if result.Intent != IntentBugFix {
		t.Errorf("Expected bug-fix intent, got %s", result.Intent)
	}

	// Verify keywords were extracted
	if len(result.Keywords) < 2 {
		t.Errorf("Expected at least 2 keywords, got %d: %v", len(result.Keywords), result.Keywords)
	}

	// Verify keywords include relevant terms
	hasAuth := false
	hasMemory := false
	for _, kw := range result.Keywords {
		if strings.Contains(strings.ToLower(kw), "auth") {
			hasAuth = true
		}
		if strings.Contains(strings.ToLower(kw), "memory") || strings.Contains(strings.ToLower(kw), "leak") {
			hasMemory = true
		}
	}

	if !hasAuth {
		t.Error("Expected 'auth' to be in keywords")
	}
	if !hasMemory {
		t.Error("Expected 'memory' or 'leak' to be in keywords")
	}
}

func TestUserProfile_ExperienceProgression(t *testing.T) {
	profile := &UserProfile{
		Categories: make(map[string]CategoryProfile),
	}

	explainer := &Explainer{
		userProfile: profile,
		explainAll:  false,
	}

	// Simulate user seeing file-reference 15 times
	for i := 0; i < 15; i++ {
		profile.Categories["file-reference"] = CategoryProfile{
			TimesSeen: i + 1,
			LastSeen:  time.Now().Format("2006-01-02"),
		}

		level := explainer.determineExplanationLevel(i + 1)

		// First 3: full
		// 4-10: short
		// 11+: label
		var expected ExplanationLevel
		if i < 3 {
			expected = ExplanationFull
		} else if i < 10 {
			expected = ExplanationShort
		} else {
			expected = ExplanationLabel
		}

		if level != expected {
			t.Errorf("At times_seen=%d, expected %v, got %v", i+1, expected, level)
		}
	}
}
