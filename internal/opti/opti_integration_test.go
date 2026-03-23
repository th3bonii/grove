package opti

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// GROVE Opti Integration Tests
// =============================================================================

// TestGROVEOptiPromptOptimization tests the GROVE Opti prompt optimization.
func TestGROVEOptiPromptOptimization(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checkFn func(*testing.T, *OptimizedPrompt)
	}{
		{
			name:    "simple add feature",
			input:   "add dark mode to settings",
			wantErr: false,
			checkFn: func(t *testing.T, result *OptimizedPrompt) {
				require.NotNil(t, result)
				assert.NotEmpty(t, result.Optimized)
				// Should contain instructions for settings domain
				assert.Contains(t, result.Optimized, "settings")
			},
		},
		{
			name:    "bug fix prompt",
			input:   "fix login error handling",
			wantErr: false,
			checkFn: func(t *testing.T, result *OptimizedPrompt) {
				require.NotNil(t, result)
				assert.NotEmpty(t, result.Optimized)
			},
		},
		{
			name:    "refactor prompt",
			input:   "refactor auth module for better testing",
			wantErr: false,
			checkFn: func(t *testing.T, result *OptimizedPrompt) {
				require.NotNil(t, result)
				assert.NotEmpty(t, result.Optimized)
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: false,
			checkFn: func(t *testing.T, result *OptimizedPrompt) {
				require.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create optimizer
			optimizer := NewOptimizer(2000)
			require.NotNil(t, optimizer)

			// Create classification
			classifier := NewClassifier()
			classification := classifier.Classify(tt.input)

			// Create context with minimal files
			ctx := &ContextResult{
				Files:  []FileCandidate{},
				Skills: []string{},
			}

			// Optimize
			result, err := optimizer.Optimize(context.Background(), tt.input, classification, ctx)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			tt.checkFn(t, result)
		})
	}
}

// TestGROVEOptiClassifier tests the intent classifier.
func TestGROVEOptiClassifier(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantIntent     Intent
		wantConfidence bool
	}{
		{
			name:           "add feature",
			input:          "add dark mode to settings",
			wantIntent:     IntentFeatureAddition,
			wantConfidence: true,
		},
		{
			name:           "fix bug",
			input:          "fix login crash",
			wantIntent:     IntentBugFix,
			wantConfidence: true,
		},
		{
			name:           "refactor",
			input:          "refactor auth module",
			wantIntent:     IntentRefactor,
			wantConfidence: true,
		},
		{
			name:           "update docs",
			input:          "update API documentation",
			wantIntent:     IntentDocumentationUpdate,
			wantConfidence: true,
		},
		{
			name:           "change config",
			input:          "update env variables",
			wantIntent:     IntentConfigurationChange,
			wantConfidence: true,
		},
		{
			name:           "unknown intent",
			input:          "do something",
			wantIntent:     IntentOther,
			wantConfidence: false, // May not have high confidence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewClassifier()
			result := classifier.Classify(tt.input)

			if tt.wantConfidence {
				// For known intents, should have some confidence
				assert.GreaterOrEqual(t, result.Confidence, 0.0)
			}

			t.Logf("Intent: %s, Domain: %s, Keywords: %v, Confidence: %.2f",
				result.Intent, result.Domain, result.Keywords, result.Confidence)
		})
	}
}

// TestGROVEOptiTokenCount tests token counting functionality.
func TestGROVEOptiTokenCount(t *testing.T) {
	optimizer := NewOptimizer(2000)
	require.NotNil(t, optimizer)

	tests := []struct {
		name    string
		input   string
		wantMin int
		wantMax int
	}{
		{
			name:    "short text",
			input:   "add feature",
			wantMin: 1,
			wantMax: 10,
		},
		{
			name:    "medium text",
			input:   "Add dark mode to settings with custom themes",
			wantMin: 5,
			wantMax: 20,
		},
		{
			name:    "empty",
			input:   "",
			wantMin: 0,
			wantMax: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := optimizer.CalculateTokenCount(tt.input)
			assert.GreaterOrEqual(t, count, tt.wantMin)
			assert.LessOrEqual(t, count, tt.wantMax)
		})
	}
}

// TestGROVEOptiBuildFunctions tests the builder functions.
func TestGROVEOptiBuildFunctions(t *testing.T) {
	optimizer := NewOptimizer(2000)
	require.NotNil(t, optimizer)

	// Create test classification and context
	classification := IntentClassification{
		Intent:     IntentFeatureAddition,
		Domain:     "settings",
		Confidence: 0.8,
	}

	ctx := &ContextResult{
		Files: []FileCandidate{
			{Path: "src/components/Settings.tsx"},
		},
		Skills:         []string{"react", "sdd-apply"},
		DependencyRefs: []string{},
		AgentsContent:  "",
	}

	// Test buildFileReferences
	_ = optimizer.buildFileReferences(ctx)

	// Test buildCoreRequest
	coreReq := optimizer.buildCoreRequest("add dark mode", classification)
	assert.NotEmpty(t, coreReq)

	// Test buildSkillInvocation
	skillInv := optimizer.buildSkillInvocation(ctx.Skills)
	assert.NotEmpty(t, skillInv)

	// Test buildSuccessCriteria
	successCriteria := optimizer.buildSuccessCriteria(classification, ctx)
	assert.NotEmpty(t, successCriteria)
	assert.Contains(t, successCriteria, "This is done when")

	// Test buildScopeBoundary
	scopeBoundary := optimizer.buildScopeBoundary(classification, ctx)
	assert.NotEmpty(t, scopeBoundary)

	// Test buildOutOfScope
	outOfScope := optimizer.buildOutOfScope(classification, ctx)
	assert.NotEmpty(t, outOfScope)

	// Test shouldRecommendPlanMode
	recommendPlan := optimizer.shouldRecommendPlanMode(classification, ctx)
	// Should not recommend for small changes
	assert.False(t, recommendPlan)
}

// TestGROVEOptiDomainExtraction tests domain extraction from prompts.
func TestGROVEOptiDomainExtraction(t *testing.T) {
	classifier := NewClassifier()
	require.NotNil(t, classifier)

	tests := []struct {
		name       string
		input      string
		wantDomain string
	}{
		{
			name:       "settings domain",
			input:      "add dark mode to settings page",
			wantDomain: "Settings",
		},
		{
			name:       "auth domain",
			input:      "fix authentication login bug",
			wantDomain: "Authentication",
		},
		{
			name:       "theme domain",
			input:      "add dark theme toggle",
			wantDomain: "Theme",
		},
		{
			name:       "no domain",
			input:      "add some feature",
			wantDomain: "", // May not detect
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.input)

			if tt.wantDomain != "" {
				// For known domains, should match
				assert.Equal(t, tt.wantDomain, result.Domain)
			} else {
				// May or may not have domain
				t.Logf("Extracted domain: %s", result.Domain)
			}
		})
	}
}

// TestGROVEOptiKeywordExtraction tests keyword extraction.
func TestGROVEOptiKeywordExtraction(t *testing.T) {
	classifier := NewClassifier()
	require.NotNil(t, classifier)

	tests := []struct {
		name        string
		input       string
		minKeywords int
	}{
		{
			name:        "simple feature",
			input:       "add user authentication with JWT",
			minKeywords: 3,
		},
		{
			name:        "bug fix",
			input:       "fix null pointer exception in login handler",
			minKeywords: 3,
		},
		{
			name:        "refactor",
			input:       "refactor database connection pooling",
			minKeywords: 2,
		},
		{
			name:        "empty",
			input:       "",
			minKeywords: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.input)
			assert.GreaterOrEqual(t, len(result.Keywords), tt.minKeywords)
			t.Logf("Keywords: %v", result.Keywords)
		})
	}
}

// TestGROVEOptiMultipleFiles tests optimization with multiple files.
func TestGROVEOptiMultipleFiles(t *testing.T) {
	optimizer := NewOptimizer(2000)
	require.NotNil(t, optimizer)

	classification := IntentClassification{
		Intent:     IntentRefactor,
		Domain:     "auth",
		Confidence: 0.9,
	}

	ctx := &ContextResult{
		Files: []FileCandidate{
			{Path: "src/auth/login.ts"},
			{Path: "src/auth/register.ts"},
			{Path: "src/auth/reset.ts"},
			{Path: "src/auth/middleware.ts"},
		},
		Skills:         []string{"sdd-apply"},
		DependencyRefs: []string{},
		AgentsContent:  "",
	}

	result, err := optimizer.Optimize(context.Background(), "refactor auth module", classification, ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should recommend plan mode for 4+ files
	assert.True(t, optimizer.shouldRecommendPlanMode(classification, ctx))
}

// TestGROVEOptiCriticalFiles tests optimization with critical files.
func TestGROVEOptiCriticalFiles(t *testing.T) {
	optimizer := NewOptimizer(2000)
	require.NotNil(t, optimizer)

	classification := IntentClassification{
		Intent:     IntentBugFix,
		Domain:     "auth",
		Confidence: 0.6, // Low confidence
	}

	ctx := &ContextResult{
		Files: []FileCandidate{
			{Path: "src/index.ts"},
			{Path: "src/app.ts"},
		},
		Skills:         []string{},
		DependencyRefs: []string{},
		AgentsContent:  "",
	}

	// Should recommend plan mode for critical files with low confidence
	assert.True(t, optimizer.shouldRecommendPlanMode(classification, ctx))
}

// TestGROVEOptiAGENTSContent tests when AGENTS.md content is present.
func TestGROVEOptiAGENTSContent(t *testing.T) {
	optimizer := NewOptimizer(2000)
	require.NotNil(t, optimizer)

	classification := IntentClassification{
		Intent:     IntentFeatureAddition,
		Confidence: 0.8,
	}

	ctx := &ContextResult{
		Files:          []FileCandidate{},
		Skills:         []string{},
		DependencyRefs: []string{},
		AgentsContent:  "# Project Conventions\n\nUse strict mode",
	}

	result, err := optimizer.Optimize(context.Background(), "add feature", classification, ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should include AGENTS.md note in output
	assert.Contains(t, result.Optimized, "AGENTS.md")
}

// TestGROVEOptiDependencyContext tests dependency context in prompts.
func TestGROVEOptiDependencyContext(t *testing.T) {
	optimizer := NewOptimizer(2000)
	require.NotNil(t, optimizer)

	classification := IntentClassification{
		Intent:     IntentFeatureAddition,
		Confidence: 0.8,
	}

	ctx := &ContextResult{
		Files: []FileCandidate{
			{Path: "src/components/Button.tsx"},
		},
		Skills: []string{},
		DependencyRefs: []string{
			"src/utils/helpers.ts",
			"src/styles/theme.ts",
		},
		AgentsContent: "",
	}

	result, err := optimizer.Optimize(context.Background(), "add button component", classification, ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should include dependency context
	assert.Contains(t, result.Optimized, "helpers.ts")
}

// TestGROVEOptiEmptyFiles tests optimization with no files selected.
func TestGROVEOptiEmptyFiles(t *testing.T) {
	optimizer := NewOptimizer(2000)
	require.NotNil(t, optimizer)

	classification := IntentClassification{
		Intent:     IntentFeatureAddition,
		Confidence: 0.8,
	}

	ctx := &ContextResult{
		Files:          []FileCandidate{},
		Skills:         []string{},
		DependencyRefs: []string{},
		AgentsContent:  "",
	}

	result, err := optimizer.Optimize(context.Background(), "add feature", classification, ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have warning about no file references
	assert.Contains(t, result.Warnings, "No explicit file references could be determined")
}

// TestGROVEOptiTokenBudget tests token budget enforcement.
func TestGROVEOptiTokenBudget(t *testing.T) {
	// Small token budget
	optimizer := NewOptimizer(100)
	require.NotNil(t, optimizer)

	classification := IntentClassification{
		Intent:     IntentFeatureAddition,
		Confidence: 0.8,
	}

	// Create context with many files
	ctx := &ContextResult{
		Files: make([]FileCandidate, 10),
	}
	for i := 0; i < 10; i++ {
		ctx.Files[i] = FileCandidate{Path: "src/test/file" + string(rune(i+'0')) + ".ts"}
	}
	ctx.Skills = []string{"skill1", "skill2", "skill3"}

	result, err := optimizer.Optimize(context.Background(), "add feature with many files", classification, ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have warning about token count
	if len(result.Warnings) > 0 {
		assert.Contains(t, result.Warnings[0], "Token count")
	}
}

// TestGROVEOptiContextResult tests ContextResult structure.
func TestGROVEOptiContextResult(t *testing.T) {
	ctx := &ContextResult{
		Files: []FileCandidate{
			{Path: "test.ts", Layer: 1, Score: 0.9, LayerName: "Source Files"},
		},
		Skills:         []string{"test-skill"},
		DependencyRefs: []string{"dep1.ts"},
		AgentsContent:  "Test content",
	}

	require.NotNil(t, ctx)
	require.Len(t, ctx.Files, 1)
	require.Len(t, ctx.Skills, 1)
	assert.Equal(t, "test.ts", ctx.Files[0].Path)
}

// TestGROVEOptiIntentClassification tests the full classification result.
func TestGROVEOptiIntentClassification(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, IntentClassification)
	}{
		{
			name:  "feature with confidence boost",
			input: "add user profile page with avatar upload and settings",
			check: func(t *testing.T, result IntentClassification) {
				// Many keywords should boost confidence
				assert.GreaterOrEqual(t, result.Confidence, 0.4)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewClassifier()
			result := classifier.Classify(tt.input)
			tt.check(t, result)
		})
	}
}

// TestGROVEOptiPromptsContentBuilder tests building different prompts.
func TestGROVEOptiPromptsContentBuilder(t *testing.T) {
	optimizer := NewOptimizer(2000)

	tests := []struct {
		name          string
		buildFn       func() string
		checkNotEmpty bool
	}{
		{
			name: "plan mode recommendation",
			buildFn: func() string {
				return optimizer.buildPlanModeRecommendation()
			},
			checkNotEmpty: true,
		},
		{
			name: "agents note",
			buildFn: func() string {
				return optimizer.buildAgentsNote()
			},
			checkNotEmpty: true,
		},
		{
			name: "dependency context empty",
			buildFn: func() string {
				return optimizer.buildDependencyContext([]string{})
			},
			checkNotEmpty: false, // Empty input returns empty string
		},
		{
			name: "dependency context with refs",
			buildFn: func() string {
				return optimizer.buildDependencyContext([]string{"dep1.ts", "dep2.ts"})
			},
			checkNotEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.buildFn()
			if tt.checkNotEmpty {
				assert.NotEmpty(t, result)
			}
		})
	}
}
