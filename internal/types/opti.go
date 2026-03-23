package types

import (
	"time"
)

// =============================================================================
// GROVE Opti Prompt Types
// =============================================================================

// IntentType categorizes the type of user intent.
type IntentType string

const (
	IntentFeatureAddition     IntentType = "feature-addition"
	IntentBugFix              IntentType = "bug-fix"
	IntentRefactor            IntentType = "refactor"
	IntentDocumentationUpdate IntentType = "documentation-update"
	IntentConfigurationChange IntentType = "configuration-change"
	IntentOther               IntentType = "other"
)

// String returns the intent type in a readable format.
func (it IntentType) String() string {
	switch it {
	case IntentFeatureAddition:
		return "Feature Addition"
	case IntentBugFix:
		return "Bug Fix"
	case IntentRefactor:
		return "Refactor"
	case IntentDocumentationUpdate:
		return "Documentation Update"
	case IntentConfigurationChange:
		return "Configuration Change"
	case IntentOther:
		return "Other"
	default:
		return "Unknown"
	}
}

// Intent represents the classified user intent from a natural language prompt.
type Intent struct {
	Type               IntentType `json:"type"`
	Keywords           []string   `json:"keywords"`
	PrimaryNouns       []string   `json:"primary_nouns"`
	Confidence         float64    `json:"confidence"` // 0.0 - 1.0
	Ambiguous          bool       `json:"ambiguous"`
	ClarifyingQuestion string     `json:"clarifying_question,omitempty"`
}

// PromptContext represents the collected project context for prompt optimization.
type PromptContext struct {
	Files          []FileSelection `json:"files"`
	SpecSections   []SpecSection   `json:"spec_sections"`
	AgentsFile     string          `json:"agents_file,omitempty"`
	RelevantSkills []string        `json:"relevant_skills"`
	Dependencies   []DependencyRef `json:"dependencies,omitempty"` // For cross-module intents
	TokenCount     int             `json:"token_count"`
	TokenBudget    int             `json:"token_budget"` // Default: 2000
}

// FileSelection represents a selected source file with its selection layer.
type FileSelection struct {
	Path       string         `json:"path"`
	Layer      SelectionLayer `json:"layer"`
	Reason     string         `json:"reason,omitempty"`
	TokenCount int            `json:"token_count,omitempty"`
}

// SelectionLayer represents the priority layer used for file selection.
type SelectionLayer int

const (
	LayerAgentsMD       SelectionLayer = iota + 1 // Layer 1: AGENTS.md explicit references
	LayerGitCommits                               // Layer 2: Recent git commits + keywords
	LayerPathMatch                                // Layer 3: Intent keyword path match
	LayerSpecComponents                           // Layer 4: SPEC.md component references
)

// SelectionLayerPriority provides string representation of layers.
var SelectionLayerPriority = map[SelectionLayer]string{
	LayerAgentsMD:       "AGENTS.md explicit references",
	LayerGitCommits:     "Recent git commits + intent keywords",
	LayerPathMatch:      "Intent keyword path match",
	LayerSpecComponents: "SPEC.md component references",
}

// SpecSection represents a relevant section from the specification.
type SpecSection struct {
	Title      string `json:"title"`
	Path       string `json:"path"`
	Content    string `json:"content"`
	Truncated  bool   `json:"truncated,omitempty"`
	TokenCount int    `json:"token_count,omitempty"`
}

// DependencyRef represents a dependency relationship for context inclusion.
type DependencyRef struct {
	SourceFile string `json:"source_file"`
	TargetFile string `json:"target_file"`
	Note       string `json:"note,omitempty"`
}

// OptimizedPrompt represents the final optimized prompt ready for execution.
type OptimizedPrompt struct {
	Original        string           `json:"original"`
	Optimized       string           `json:"optimized"`
	Intent          Intent           `json:"intent"`
	WhyExplanations []WhyExplanation `json:"why_explanations"`
	PlanMode        bool             `json:"plan_mode"`
	TokenCount      int              `json:"token_count"`
	TokenBudget     int              `json:"token_budget"`
	Warnings        []string         `json:"warnings,omitempty"`
}

// WhyExplanation provides the teaching explanation for a prompt optimization.
type WhyExplanation struct {
	Category WhyCategory `json:"category"`
	Full     string      `json:"full"`    // Full explanation (2 sentences)
	Brief    string      `json:"brief"`   // Brief reminder (1 sentence)
	Label    string      `json:"label"`   // Category label only
	Applied  WhyLevel    `json:"applied"` // Which explanation level was used
}

// WhyCategory categorizes the type of optimization explanation.
type WhyCategory string

const (
	WhyCategoryFileReference   WhyCategory = "file-reference"
	WhyCategoryScopeBoundary   WhyCategory = "scope-boundary"
	WhyCategorySkillInvocation WhyCategory = "skill-invocation"
	WhyCategorySuccessCriteria WhyCategory = "success-criteria"
	WhyCategoryPlanMode        WhyCategory = "plan-mode"
	WhyCategoryOutOfScope      WhyCategory = "out-of-scope-boundary"
)

// WhyLevel determines how much explanation to provide based on user profile.
type WhyLevel int

const (
	WhyLevelFull  WhyLevel = iota // Full 2-sentence explanation
	WhyLevelBrief                 // 1-sentence reminder
	WhyLevelLabel                 // Category label only
	WhyLevelNone                  // No explanation (--no-teach flag)
)

// UserProfile tracks user interaction patterns for adaptive explanations.
type UserProfile struct {
	Categories   map[WhyCategory]CategoryStats `json:"categories"`
	EditPatterns []EditPattern                 `json:"edit_patterns"`
	UpdatedAt    time.Time                     `json:"updated_at"`
}

// CategoryStats tracks interaction frequency for a WHY category.
type CategoryStats struct {
	TimesSeen int       `json:"times_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// EditPattern represents a learned pattern from user edits.
type EditPattern struct {
	PatternType   PatternType `json:"pattern_type"` // "added", "removed", "rewritten"
	Category      string      `json:"category"`
	Frequency     int         `json:"frequency"`
	ExampleBefore string      `json:"example_before"`
	ExampleAfter  string      `json:"example_after"`
	LearnedAt     time.Time   `json:"learned_at"`
}

// PatternType categorizes the type of edit pattern.
type PatternType string

const (
	PatternTypeAdded     PatternType = "added"
	PatternTypeRemoved   PatternType = "removed"
	PatternTypeRewritten PatternType = "rewritten"
)

// InvocationLog represents a single optimization invocation record.
type InvocationLog struct {
	Timestamp            time.Time          `json:"timestamp"`
	IntentClassification IntentType         `json:"intent_classification"`
	TokensUsed           int                `json:"tokens_used"`
	FilesSelected        []FileSelectionLog `json:"files_selected"`
	UserAction           UserAction         `json:"user_action"`
	SkillsReferenced     []string           `json:"skills_referenced"`
}

// FileSelectionLog represents a logged file selection with layer info.
type FileSelectionLog struct {
	Path  string `json:"path"`
	Layer int    `json:"layer"`
}

// UserAction represents the user's response to the optimized prompt.
type UserAction string

const (
	UserActionSend   UserAction = "send"
	UserActionEdit   UserAction = "edit"
	UserActionReject UserAction = "reject"
)

// String returns the user action as a readable string.
func (ua UserAction) String() string {
	switch ua {
	case UserActionSend:
		return "Send"
	case UserActionEdit:
		return "Edit"
	case UserActionReject:
		return "Reject"
	default:
		return "Unknown"
	}
}

// OptiLog represents the complete GROVE Opti Prompt log file structure.
type OptiLog struct {
	InvocationLog []InvocationLog `json:"invocation_log"`
	UserProfile   *UserProfile    `json:"user_profile,omitempty"`
	EditPatterns  []EditPattern   `json:"edit_patterns,omitempty"`
}

// BatchResult represents the result of processing a single prompt in batch mode.
type BatchResult struct {
	LineNumber   int              `json:"line_number"`
	Original     string           `json:"original"`
	Optimized    *OptimizedPrompt `json:"optimized,omitempty"`
	Unclassified bool             `json:"unclassified"`
	Error        string           `json:"error,omitempty"`
	TokensUsed   int              `json:"tokens_used"`
}

// BatchReport represents the complete batch processing report.
type BatchReport struct {
	Filename          string        `json:"filename"`
	ProcessedAt       time.Time     `json:"processed_at"`
	TotalPrompts      int           `json:"total_prompts"`
	SuccessCount      int           `json:"success_count"`
	UnclassifiedCount int           `json:"unclassified_count"`
	ErrorCount        int           `json:"error_count"`
	TotalTokens       int           `json:"total_tokens"`
	Results           []BatchResult `json:"results"`
}

// OptiConfig represents the GROVE Opti Prompt configuration.
type OptiConfig struct {
	TokenBudget            int  `json:"token_budget"`             // Default: 2000
	MaxFiles               int  `json:"max_files"`                // Default: 3
	ExplainAll             bool `json:"explain_all"`              // --explain-all flag
	NoTeach                bool `json:"no_teach"`                 // --no-teach flag
	BatchMode              bool `json:"batch_mode"`               // --batch flag
	AutoClassify           bool `json:"auto_classify"`            // Auto-invoke on vague prompts
	LearnedPatternsEnabled bool `json:"learned_patterns_enabled"` // Enable bidirectional learning
	MinPatternFrequency    int  `json:"min_pattern_frequency"`    // Patterns applied after N instances
}

// WebCache represents a cached web search or MCP query result.
type WebCache struct {
	Query     string    `json:"query"`
	Source    string    `json:"source"` // "web" or "context7"
	Result    string    `json:"result"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PromptDiff represents the difference between optimized and edited prompt.
type PromptDiff struct {
	Original  string        `json:"original"`
	Optimized string        `json:"optimized"`
	Final     string        `json:"final"`
	Diff      []DiffSegment `json:"diff"`
}

// DiffSegment represents a segment of the diff.
type DiffSegment struct {
	Type    DiffType `json:"type"` // "added", "removed", "unchanged"
	Content string   `json:"content"`
}

// DiffType categorizes diff segment types.
type DiffType string

const (
	DiffTypeAdded     DiffType = "added"
	DiffTypeRemoved   DiffType = "removed"
	DiffTypeUnchanged DiffType = "unchanged"
)

// =============================================================================
// Helper Functions
// =============================================================================

// NewUserProfile creates a new user profile with default category stats.
func NewUserProfile() *UserProfile {
	categories := make(map[WhyCategory]CategoryStats)
	for _, cat := range []WhyCategory{
		WhyCategoryFileReference,
		WhyCategoryScopeBoundary,
		WhyCategorySkillInvocation,
		WhyCategorySuccessCriteria,
		WhyCategoryPlanMode,
		WhyCategoryOutOfScope,
	} {
		categories[cat] = CategoryStats{}
	}
	return &UserProfile{
		Categories: categories,
		UpdatedAt:  time.Now(),
	}
}

// GetWhyLevel determines the appropriate explanation level based on times_seen.
func GetWhyLevel(timesSeen int, forceFull bool) WhyLevel {
	if forceFull {
		return WhyLevelFull
	}
	switch {
	case timesSeen <= 0:
		return WhyLevelFull
	case timesSeen <= 3:
		return WhyLevelFull
	case timesSeen <= 10:
		return WhyLevelBrief
	default:
		return WhyLevelLabel
	}
}

// ShouldApplyLearnedPattern determines if a pattern should be auto-applied.
func ShouldApplyLearnedPattern(pattern EditPattern, minFrequency int) bool {
	return pattern.Frequency >= minFrequency
}
