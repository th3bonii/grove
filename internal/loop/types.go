// Package loop provides types for the Ralph Loop orchestrator.
package loop

import "time"

// GentleClient for gentle-ai integration.
// Provides direct access to gentle-ai API for advanced operations.
type GentleClient struct {
	endpoint string
	apiKey   string
}

// GGAClient (Ángel Guardián Caballero) manages provider switching
// for LLM error recovery.
type GGAClient struct {
	currentProvider string
	providers       []string
	currentIndex    int
}

// NewGentleClient creates a new GentleClient instance.
func NewGentleClient(endpoint, apiKey string) *GentleClient {
	return &GentleClient{
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}

// NewGGAClient creates a new GGAClient with the given providers.
func NewGGAClient(providers []string) *GGAClient {
	if len(providers) == 0 {
		// Default providers
		providers = []string{"openai", "anthropic", "google"}
	}
	return &GGAClient{
		currentProvider: providers[0],
		providers:       providers,
		currentIndex:    0,
	}
}

// CurrentProvider returns the current LLM provider.
func (g *GGAClient) CurrentProvider() string {
	return g.currentProvider
}

// SwitchProvider switches to the next available LLM provider.
// Returns error if no more providers are available.
func (g *GGAClient) SwitchProvider() error {
	if len(g.providers) <= 1 {
		return nil // No need to switch
	}

	g.currentIndex = (g.currentIndex + 1) % len(g.providers)
	g.currentProvider = g.providers[g.currentIndex]

	return nil
}

// ResetProvider resets to the first provider in the list.
func (g *GGAClient) ResetProvider() {
	g.currentIndex = 0
	if len(g.providers) > 0 {
		g.currentProvider = g.providers[0]
	}
}

// DelegateContext contains the context passed to sub-agents for task execution.
type DelegateContext struct {
	Task              interface{} `json:"task"`
	SpecSections      []string    `json:"spec_sections"`
	RelevantSkills    []string    `json:"relevant_skills"`
	AgentsFile        string      `json:"agents_file"`
	Constraints       []string    `json:"constraints"`
	SuccessCriteria   string      `json:"success_criteria"`
	ProjectPath       string      `json:"project_path"`
	ChangeName        string      `json:"change_name"`
	ArtifactStoreMode string      `json:"artifact_store_mode"`
	Timestamp         time.Time   `json:"timestamp"`
}

// TaskResult represents the result of a task execution.
type TaskResult struct {
	TaskID   string                 `json:"task_id"`
	Status   string                 `json:"status"` // success, failure, warning
	Summary  string                 `json:"summary"`
	Duration time.Duration          `json:"duration"`
	FilesMod []string               `json:"files_modified,omitempty"`
	Output   string                 `json:"output,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// LoopConfig contains configuration for the Ralph Loop.
type LoopConfig struct {
	// Project path
	ProjectPath string

	// Documentation path
	DocsPath string

	// State directory for persistence
	StateDir string

	// Enable checkpoint/resume
	CheckpointEnabled bool

	// Max retries for failed tasks
	MaxRetries int

	// Base backoff in milliseconds
	BackoffBaseMs int64

	// Providers for GGA (Ángel Guardián Caballero)
	LLMProviders []string

	// Gentle AI endpoint
	GentleEndpoint string

	// Gentle AI API key
	GentleAPIKey string

	// Skill directory override
	SkillsDir string

	// Timeout for task execution
	TaskTimeout time.Duration

	// OnPhaseChange callback
	OnPhaseChange func(from, to LoopPhase)

	// OnTaskComplete callback
	OnTaskComplete func(task *Task, err error)

	// OnError callback
	OnError func(err error)
}

// DefaultLoopConfig returns a default configuration.
func DefaultLoopConfig() *LoopConfig {
	return &LoopConfig{
		CheckpointEnabled: true,
		MaxRetries:        3,
		BackoffBaseMs:     1000,
		LLMProviders:      []string{"openai", "anthropic", "google"},
		TaskTimeout:       5 * time.Minute,
	}
}
