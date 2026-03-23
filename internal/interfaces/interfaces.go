// Package interfaces defines the core interfaces for the GROVE SDD framework.
// These interfaces enable dependency injection and testing throughout the system.
package interfaces

import (
	"context"
	"time"
)

// ============================================================================
// Core Types
// ============================================================================

// SpecPhase represents a phase in the SDD workflow.
type SpecPhase string

const (
	PhaseExplore  SpecPhase = "explore"
	PhaseProposal SpecPhase = "proposal"
	PhaseSpec     SpecPhase = "spec"
	PhaseDesign   SpecPhase = "design"
	PhaseTasks    SpecPhase = "tasks"
	PhaseApply    SpecPhase = "apply"
	PhaseVerify   SpecPhase = "verify"
	PhaseArchive  SpecPhase = "archive"
)

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusDeferred  TaskStatus = "deferred"
	TaskStatusSkipped   TaskStatus = "skipped"
)

// LoopState represents the current state of the Ralph Loop.
type LoopState struct {
	LoopNumber           int            `json:"loop_number"`
	CurrentTaskID        string         `json:"current_task_id"`
	CompletedTasks       []string       `json:"completed_tasks"`
	FailedTasks          []string       `json:"failed_tasks"`
	DeferredTasks        []string       `json:"deferred_tasks"`
	QualityScore         float64        `json:"quality_score"`
	DocumentationQuality *QualityScores `json:"documentation_quality,omitempty"`
	LastCheckpoint       time.Time      `json:"last_checkpoint"`
	Paused               bool           `json:"paused"`
	ErrorLog             []LoopError    `json:"error_log,omitempty"`
}

// QualityScores represents the 7-dimension quality scoring.
type QualityScores struct {
	FlowCoverage          float64 `json:"flow_coverage"`
	ComponentDepth        float64 `json:"component_depth"`
	LogicalConsistency    float64 `json:"logical_consistency"`
	Connectivity          float64 `json:"connectivity"`
	EdgeCases             float64 `json:"edge_cases"`
	DecisionJustification float64 `json:"decision_justification"`
	AgentConsumability    float64 `json:"agent_consumability"`
}

// CompositeScore calculates the weighted composite score.
func (q QualityScores) CompositeScore() float64 {
	return (q.FlowCoverage * 0.20) +
		(q.ComponentDepth * 0.20) +
		(q.LogicalConsistency * 0.15) +
		(q.Connectivity * 0.15) +
		(q.EdgeCases * 0.15) +
		(q.DecisionJustification * 0.10) +
		(q.AgentConsumability * 0.05)
}

// LoopError represents an error encountered during loop execution.
type LoopError struct {
	TaskID     string    `json:"task_id"`
	Error      string    `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
	Retriable  bool      `json:"retriable"`
	RetryCount int       `json:"retry_count"`
}

// SpecArtifact represents a specification artifact.
type SpecArtifact struct {
	Phase      SpecPhase `json:"phase"`
	Content    string    `json:"content"`
	TopicKey   string    `json:"topic_key"`
	Project    string    `json:"project"`
	ChangeName string    `json:"change_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Hash       string    `json:"hash,omitempty"`
}

// Task represents an implementation task.
type Task struct {
	ID           string        `json:"id"`
	Description  string        `json:"description"`
	Phase        SpecPhase     `json:"phase"`
	Dependencies []string      `json:"dependencies,omitempty"`
	Status       TaskStatus    `json:"status"`
	Attempts     int           `json:"attempts"`
	Duration     time.Duration `json:"duration,omitempty"`
	Notes        string        `json:"notes,omitempty"`
}

// ============================================================================
// Spec Engine Interfaces
// ============================================================================

// SpecEngine defines the core specification engine interface.
type SpecEngine interface {
	// GenerateSpec creates a new specification from input.
	GenerateSpec(ctx context.Context, input SpecInput) (*SpecOutput, error)

	// UpdateSpec updates an existing specification.
	UpdateSpec(ctx context.Context, update UpdateInput) (*SpecOutput, error)

	// ValidateSpec validates a specification for completeness.
	ValidateSpec(ctx context.Context, spec SpecArtifact) (*ValidationResult, error)

	// Decompose breaks down requirements into components.
	Decompose(ctx context.Context, req DecomposeInput) ([]Component, error)
}

// SpecInput represents input for specification generation.
type SpecInput struct {
	ProjectPath string   `json:"project_path"`
	InputFiles  []string `json:"input_files"`
	Stack       []string `json:"stack,omitempty"`
}

// SpecOutput represents the output of specification generation.
type SpecOutput struct {
	Spec    *SpecArtifact `json:"spec"`
	Design  *SpecArtifact `json:"design"`
	Tasks   []Task        `json:"tasks"`
	Changes []ChangeEntry `json:"changes,omitempty"`
}

// UpdateInput represents input for specification update.
type UpdateInput struct {
	BaselineArtifacts []SpecArtifact   `json:"baseline_artifacts"`
	DiffFiles         []string         `json:"diff_files,omitempty"`
	FeedbackPayload   *FeedbackPayload `json:"feedback_payload,omitempty"`
}

// FeedbackPayload represents structured feedback from Ralph Loop.
type FeedbackPayload struct {
	Trigger       string        `json:"trigger"`
	LoopNumber    int           `json:"loop_number"`
	QualityScore  float64       `json:"quality_score"`
	Missing       []string      `json:"missing"`
	QualityScores QualityScores `json:"quality_scores"`
	FailedTasks   []string      `json:"failed_tasks"`
	Observations  []string      `json:"observations"`
}

// DecomposeInput represents input for decomposition.
type DecomposeInput struct {
	ComponentName string `json:"component_name"`
	Context       string `json:"context"`
	Scope         string `json:"scope,omitempty"`
}

// Component represents a decomposed component.
type Component struct {
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	SubComponents []Component `json:"sub_components,omitempty"`
	States        []string    `json:"states,omitempty"`
	Behaviors     []string    `json:"behaviors,omitempty"`
	EdgeCases     []string    `json:"edge_cases,omitempty"`
	Justification string      `json:"justification"`
}

// ChangeEntry represents a change made during specification update.
type ChangeEntry struct {
	ChangeType string `json:"change_type"` // "added", "modified", "removed"
	Location   string `json:"location"`
	Before     string `json:"before,omitempty"`
	After      string `json:"after,omitempty"`
	Reason     string `json:"reason"`
}

// ValidationResult represents the result of specification validation.
type ValidationResult struct {
	Valid         bool           `json:"valid"`
	Errors        []string       `json:"errors,omitempty"`
	Warnings      []string       `json:"warnings,omitempty"`
	QualityScores *QualityScores `json:"quality_scores,omitempty"`
	PassedChecks  []string       `json:"passed_checks"`
}

// ============================================================================
// Spec Scorer Interface
// ============================================================================

// SpecScorer evaluates the quality of specifications.
type SpecScorer interface {
	// ScoreSpec evaluates a specification against 7 dimensions.
	ScoreSpec(ctx context.Context, spec SpecArtifact) (*QualityScores, error)

	// IsComplete checks if the specification meets the completion threshold.
	IsComplete(ctx context.Context, scores QualityScores) (bool, string, error)

	// CalculateDelta calculates the change percentage between two specs.
	CalculateDelta(ctx context.Context, oldSpec, newSpec SpecArtifact) (float64, error)
}

// ============================================================================
// Loop Validator Interface
// ============================================================================

// LoopValidator validates pre-conditions before starting the build loop.
type LoopValidator interface {
	// ValidateAgents validates ROOT and scoped AGENTS.md files.
	ValidateAgents(ctx context.Context, projectPath string) (*AgentsValidationResult, error)

	// ValidateSkills validates SKILL.md files referenced in AGENTS.md.
	ValidateSkills(ctx context.Context, projectPath string) (*SkillsValidationResult, error)

	// ValidateSpecs validates SPEC.md, DESIGN.md, and TASKS.md.
	ValidateSpecs(ctx context.Context, projectPath string) (*SpecsValidationResult, error)

	// ValidateStack validates tech stack coherence.
	ValidateStack(ctx context.Context, design SpecArtifact) (*StackValidationResult, error)

	// ValidateDependencies validates task dependencies in TASKS.md.
	ValidateDependencies(ctx context.Context, tasks []Task) (*DependenciesValidationResult, error)

	// ValidateAll runs all validations and returns a combined result.
	ValidateAll(ctx context.Context, projectPath string) (*AllValidationsResult, error)
}

// AgentsValidationResult represents validation of AGENTS.md files.
type AgentsValidationResult struct {
	RootValid          bool     `json:"root_valid"`
	ScopedValid        bool     `json:"scoped_valid"`
	SkillRegistryValid bool     `json:"skill_registry_valid"`
	MissingFiles       []string `json:"missing_files,omitempty"`
	InvalidFiles       []string `json:"invalid_files,omitempty"`
	Errors             []string `json:"errors,omitempty"`
}

// SkillsValidationResult represents validation of SKILL.md files.
type SkillsValidationResult struct {
	Valid         bool     `json:"valid"`
	MissingSkills []string `json:"missing_skills,omitempty"`
	InvalidSkills []string `json:"invalid_skills,omitempty"`
	Errors        []string `json:"errors,omitempty"`
}

// SpecsValidationResult represents validation of spec files.
type SpecsValidationResult struct {
	SpecValid    bool     `json:"spec_valid"`
	DesignValid  bool     `json:"design_valid"`
	TasksValid   bool     `json:"tasks_valid"`
	MissingFiles []string `json:"missing_files,omitempty"`
	EmptyFiles   []string `json:"empty_files,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

// StackValidationResult represents validation of tech stack.
type StackValidationResult struct {
	Valid    bool     `json:"valid"`
	Declared []string `json:"declared_stack"`
	Detected []string `json:"detected_stack"`
	Coherent bool     `json:"coherent"`
	Warnings []string `json:"warnings,omitempty"`
}

// DependenciesValidationResult represents validation of task dependencies.
type DependenciesValidationResult struct {
	Valid                 bool     `json:"valid"`
	CircularDependencies  []string `json:"circular_dependencies,omitempty"`
	UndefinedDependencies []string `json:"undefined_dependencies,omitempty"`
	Errors                []string `json:"errors,omitempty"`
}

// AllValidationsResult represents combined validation results.
type AllValidationsResult struct {
	Agents     *AgentsValidationResult       `json:"agents"`
	Skills     *SkillsValidationResult       `json:"skills"`
	Specs      *SpecsValidationResult        `json:"specs"`
	Stack      *StackValidationResult        `json:"stack"`
	Depends    *DependenciesValidationResult `json:"dependencies"`
	Ready      bool                          `json:"ready"`
	Errors     []string                      `json:"errors,omitempty"`
	CanProceed bool                          `json:"can_proceed"`
}

// ============================================================================
// Loop Orchestrator Interface
// ============================================================================

// LoopOrchestrator manages the Ralph Loop execution.
type LoopOrchestrator interface {
	// Start starts the loop execution.
	Start(ctx context.Context, config LoopConfig) error

	// Pause pauses the loop execution.
	Pause(ctx context.Context) error

	// Resume resumes a paused or interrupted loop.
	Resume(ctx context.Context) error

	// Stop stops the loop execution.
	Stop(ctx context.Context) error

	// GetState returns the current loop state.
	GetState(ctx context.Context) (*LoopState, error)

	// ExecuteTask executes a single task.
	ExecuteTask(ctx context.Context, task Task) (*TaskResult, error)

	// ExecutePhase executes all tasks in a phase.
	ExecutePhase(ctx context.Context, phase SpecPhase) (*PhaseResult, error)
}

// LoopConfig represents configuration for the loop orchestrator.
type LoopConfig struct {
	ProjectPath    string        `json:"project_path"`
	MaxRetries     int           `json:"max_retries"`
	RetryBackoff   time.Duration `json:"retry_backoff"`
	PauseAfterTask string        `json:"pause_after_task,omitempty"`
	ForceRestart   bool          `json:"force_restart"`
	ValidateOnly   bool          `json:"validate_only"`
}

// TaskResult represents the result of a task execution.
type TaskResult struct {
	TaskID       string        `json:"task_id"`
	Success      bool          `json:"success"`
	Duration     time.Duration `json:"duration"`
	Error        string        `json:"error,omitempty"`
	Attempts     int           `json:"attempts"`
	VerifyReport *VerifyReport `json:"verify_report,omitempty"`
}

// PhaseResult represents the result of a phase execution.
type PhaseResult struct {
	Phase      SpecPhase     `json:"phase"`
	TotalTasks int           `json:"total_tasks"`
	Completed  int           `json:"completed"`
	Failed     int           `json:"failed"`
	Deferred   int           `json:"deferred"`
	Duration   time.Duration `json:"duration"`
	Results    []TaskResult  `json:"results"`
}

// ============================================================================
// Intent Classifier Interface
// ============================================================================

// IntentClassifier classifies user intent from input.
type IntentClassifier interface {
	// ClassifyIntent determines the user's intent from input.
	ClassifyIntent(ctx context.Context, input string) (*IntentClassification, error)

	// ExtractEntities extracts entities from input.
	ExtractEntities(ctx context.Context, input string) ([]Entity, error)

	// DetectAmbiguity detects ambiguous input.
	DetectAmbiguity(ctx context.Context, input string) (*AmbiguityReport, error)
}

// IntentClassification represents the classified intent.
type IntentClassification struct {
	PrimaryIntent    IntentType             `json:"primary_intent"`
	SecondaryIntents []IntentType           `json:"secondary_intents,omitempty"`
	Confidence       float64                `json:"confidence"`
	Parameters       map[string]interface{} `json:"parameters,omitempty"`
}

// IntentType represents types of user intent.
type IntentType string

const (
	IntentCreate   IntentType = "create"
	IntentUpdate   IntentType = "update"
	IntentDelete   IntentType = "delete"
	IntentQuery    IntentType = "query"
	IntentClarify  IntentType = "clarify"
	IntentContinue IntentType = "continue"
	IntentStop     IntentType = "stop"
	IntentHelp     IntentType = "help"
)

// Entity represents an extracted entity.
type Entity struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Span  [2]int `json:"span"`
}

// AmbiguityReport represents ambiguity detection results.
type AmbiguityReport struct {
	Ambiguous      bool     `json:"ambiguous"`
	Ambiguities    []string `json:"ambiguities,omitempty"`
	Suggestions    []string `json:"suggestions,omitempty"`
	Clarifications []string `json:"clarifications,omitempty"`
}

// ============================================================================
// Context Collector Interface
// ============================================================================

// ContextCollector collects and manages context for agents.
type ContextCollector interface {
	// CollectTaskContext collects context for a specific task.
	CollectTaskContext(ctx context.Context, task Task) (*TaskContext, error)

	// CollectSpecContext collects relevant spec sections.
	CollectSpecContext(ctx context.Context, scope string) (*SpecContext, error)

	// CollectAgentsContext collects AGENTS.md context.
	CollectAgentsContext(ctx context.Context, module string) (*AgentsContext, error)

	// CollectSkillsContext collects skill context for triggers.
	CollectSkillsContext(ctx context.Context, triggers []string) ([]SkillContext, error)

	// BoundContext ensures context stays within size limits.
	BoundContext(ctx context.Context, context Context, maxTokens int) (*BoundedContext, error)
}

// TaskContext represents context for task execution.
type TaskContext struct {
	Task        Task           `json:"task"`
	Spec        *SpecContext   `json:"spec,omitempty"`
	Skills      []SkillContext `json:"skills,omitempty"`
	Constraints []string       `json:"constraints,omitempty"`
}

// SpecContext represents relevant specification sections.
type SpecContext struct {
	RelevantSections []SpecSection `json:"relevant_sections"`
	FlowCoverage     []string      `json:"flow_coverage,omitempty"`
	EdgeCases        []string      `json:"edge_cases,omitempty"`
}

// SpecSection represents a section of the specification.
type SpecSection struct {
	Title    string    `json:"title"`
	Phase    SpecPhase `json:"phase"`
	Content  string    `json:"content"`
	Priority int       `json:"priority"`
}

// AgentsContext represents AGENTS.md context.
type AgentsContext struct {
	RootAgents   *RootAgentsData   `json:"root_agents"`
	ScopedAgents *ScopedAgentsData `json:"scoped_agents,omitempty"`
}

// RootAgentsData represents ROOT AGENTS.md data.
type RootAgentsData struct {
	ProjectContext  string            `json:"project_context"`
	SkillRegistry   map[string]string `json:"skill_registry"`
	AutoInvocations map[string]string `json:"auto_invocations"`
	FileTree        string            `json:"file_tree,omitempty"`
}

// ScopedAgentsData represents scoped AGENTS.md data.
type ScopedAgentsData struct {
	Module      string   `json:"module"`
	Scope       string   `json:"scope"`
	Skills      []string `json:"skills"`
	Triggers    []string `json:"triggers"`
	Constraints []string `json:"constraints"`
	TargetFiles []string `json:"target_files,omitempty"`
}

// SkillContext represents context from a skill.
type SkillContext struct {
	SkillName    string `json:"skill_name"`
	SkillPath    string `json:"skill_path"`
	Instructions string `json:"instructions"`
}

// Context represents a generic context container.
type Context interface{}

// BoundedContext represents context bounded by token limit.
type BoundedContext struct {
	Content   string `json:"content"`
	Tokens    int    `json:"tokens"`
	Truncated bool   `json:"truncated"`
	Cutoff    string `json:"cutoff,omitempty"`
}

// ============================================================================
// State Manager Interface
// ============================================================================

// StateManager manages loop and spec state persistence.
type StateManager interface {
	// SaveLoopState saves the current loop state.
	SaveLoopState(ctx context.Context, state *LoopState) error

	// LoadLoopState loads the last saved loop state.
	LoadLoopState(ctx context.Context, projectPath string) (*LoopState, error)

	// SaveSpecState saves spec loop state.
	SaveSpecState(ctx context.Context, state *SpecLoopState) error

	// LoadSpecState loads spec loop state.
	LoadSpecState(ctx context.Context, projectPath string) (*SpecLoopState, error)

	// ClearState clears all state for a project.
	ClearState(ctx context.Context, projectPath string) error

	// AtomicWrite writes data atomically to prevent corruption.
	AtomicWrite(ctx context.Context, path string, data interface{}) error
}

// SpecLoopState represents state for the spec loop.
type SpecLoopState struct {
	LoopNumber               int           `json:"loop_number"`
	DimensionScores          QualityScores `json:"dimension_scores"`
	CompositeScore           float64       `json:"composite_score"`
	ContentDeltaPct          float64       `json:"content_delta_pct"`
	ConsecutiveLowDeltaCount int           `json:"consecutive_low_delta_count"`
	ExitCondition            ExitCondition `json:"exit_condition"`
	LastRun                  time.Time     `json:"last_run"`
}

// ExitCondition represents loop exit conditions.
type ExitCondition string

const (
	ExitNormal    ExitCondition = "normal"
	ExitSafetyNet ExitCondition = "safety_net"
	ExitManual    ExitCondition = "manual"
	ExitError     ExitCondition = "error"
)

// ============================================================================
// Audit Logger Interface
// ============================================================================

// AuditLogger provides comprehensive audit logging.
type AuditLogger interface {
	// LogTask logs a task execution event.
	LogTask(ctx context.Context, event TaskLogEvent) error

	// LogLoop logs a loop event.
	LogLoop(ctx context.Context, event LoopLogEvent) error

	// LogError logs an error event.
	LogError(ctx context.Context, event ErrorLogEvent) error

	// LogSpecChange logs a specification change.
	LogSpecChange(ctx context.Context, event SpecChangeEvent) error

	// LogDecision logs a decision with rationale.
	LogDecision(ctx context.Context, event DecisionLogEvent) error

	// GetAuditTrail returns the audit trail for a period.
	GetAuditTrail(ctx context.Context, projectPath string, since time.Time) ([]AuditEntry, error)

	// ExportMetrics exports build performance metrics.
	ExportMetrics(ctx context.Context, projectPath string) (*BuildMetrics, error)
}

// TaskLogEvent represents a task log event.
type TaskLogEvent struct {
	TaskID       string        `json:"task_id"`
	LoopNumber   int           `json:"loop_number"`
	Phase        SpecPhase     `json:"phase"`
	Action       string        `json:"action"` // started, completed, failed, retried
	Duration     time.Duration `json:"duration,omitempty"`
	VerifyResult string        `json:"verify_result,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// LoopLogEvent represents a loop log event.
type LoopLogEvent struct {
	LoopNumber   int       `json:"loop_number"`
	Action       string    `json:"action"` // started, paused, resumed, completed
	QualityScore float64   `json:"quality_score,omitempty"`
	TasksTotal   int       `json:"tasks_total,omitempty"`
	TasksDone    int       `json:"tasks_done,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// ErrorLogEvent represents an error log event.
type ErrorLogEvent struct {
	TaskID     string    `json:"task_id,omitempty"`
	LoopNumber int       `json:"loop_number,omitempty"`
	Error      string    `json:"error"`
	Type       string    `json:"type"` // llm_failure, network_error, validation_error
	Retriable  bool      `json:"retriable"`
	Recovered  bool      `json:"recovered,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Status     string    `json:"status"` // flagged, warning, resolved
}

// SpecChangeEvent represents a specification change event.
type SpecChangeEvent struct {
	LoopNumber  int       `json:"loop_number"`
	ChangeType  string    `json:"change_type"` // added, modified, removed
	Location    string    `json:"location"`
	BeforeHash  string    `json:"before_hash,omitempty"`
	AfterHash   string    `json:"after_hash,omitempty"`
	DeltaTokens int       `json:"delta_tokens,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// DecisionLogEvent represents a decision log event.
type DecisionLogEvent struct {
	DecisionID   string    `json:"decision_id"`
	Category     string    `json:"category"`
	Decision     string    `json:"decision"`
	Rationale    string    `json:"rationale"`
	Alternatives []string  `json:"alternatives,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// AuditEntry represents a generic audit entry.
type AuditEntry struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      string    `json:"data"`
}

// BuildMetrics represents build performance metrics.
type BuildMetrics struct {
	ProjectID      string                  `json:"project_id"`
	TechStack      string                  `json:"tech_stack"`
	LoopNumber     int                     `json:"loop_number"`
	PhaseDurations map[string]float64      `json:"phase_durations"`
	TaskDurations  map[string]TaskDuration `json:"task_durations"`
	Bottlenecks    []string                `json:"bottleneck_phases"`
	Timestamp      time.Time               `json:"timestamp"`
}

// TaskDuration represents duration metrics for a task type.
type TaskDuration struct {
	AvgSeconds  float64 `json:"avg_seconds"`
	Count       int     `json:"count"`
	FailureRate float64 `json:"failure_rate"`
}

// ============================================================================
// Verify Reporter Interface
// ============================================================================

// VerifyReporter generates verification reports.
type VerifyReporter interface {
	// GenerateReport generates a verification report.
	GenerateReport(ctx context.Context, task Task, spec *SpecContext, result VerifyInput) (*VerifyReport, error)

	// CompareResults compares two verification results.
	CompareResults(ctx context.Context, old, new VerifyReport) (*VerifyComparison, error)
}

// VerifyInput represents input for verification.
type VerifyInput struct {
	TaskDescription    string
	SpecRequirements   []string
	ImplementationPath string
	Changes            []string
}

// VerifyReport represents a verification report.
type VerifyReport struct {
	TaskID      string    `json:"task_id"`
	Verdict     Verdict   `json:"verdict"` // PASS, FAIL
	Checks      []Check   `json:"checks"`
	PassedCount int       `json:"passed_count"`
	FailedCount int       `json:"failed_count"`
	Issues      []Issue   `json:"issues,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// Verdict represents verification verdict.
type Verdict string

const (
	VerdictPass Verdict = "PASS"
	VerdictFail Verdict = "FAIL"
)

// Check represents a single verification check.
type Check struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details,omitempty"`
}

// Issue represents a verification issue.
type Issue struct {
	Severity   string `json:"severity"` // critical, major, minor
	Location   string `json:"location"`
	Problem    string `json:"problem"`
	Suggestion string `json:"suggestion,omitempty"`
}

// VerifyComparison represents comparison of verification results.
type VerifyComparison struct {
	Improved   bool        `json:"improved"`
	Regressed  bool        `json:"regressed"`
	DiffChecks []CheckDiff `json:"diff_checks"`
}

// CheckDiff represents a difference in checks between verifications.
type CheckDiff struct {
	CheckName string `json:"check_name"`
	Before    bool   `json:"before"`
	After     bool   `json:"after"`
}

// ============================================================================
// Web Search Cache Interface
// ============================================================================

// WebSearchCache caches web search and MCP query results.
type WebSearchCache interface {
	// Get retrieves a cached result.
	Get(ctx context.Context, query string) (*CachedResult, error)

	// Set stores a result in the cache.
	Set(ctx context.Context, query string, result *CachedResult) error

	// IsExpired checks if a cached result has expired.
	IsExpired(ctx context.Context, query string, maxAge time.Duration) (bool, error)

	// Clear clears the cache.
	Clear(ctx context.Context) error

	// Save persists the cache to disk.
	Save(ctx context.Context) error

	// Load loads the cache from disk.
	Load(ctx context.Context) error
}

// CachedResult represents a cached search result.
type CachedResult struct {
	Query     string    `json:"query"`
	Source    string    `json:"source"` // "web" or "context7"
	Summary   string    `json:"summary"`
	Timestamp time.Time `json:"timestamp"`
	RawData   string    `json:"raw_data,omitempty"`
}

// ============================================================================
// Agent Spawner Interface
// ============================================================================

// AgentSpawner spawns sub-agents for task execution.
type AgentSpawner interface {
	// SpawnImplementation spawns an implementation sub-agent.
	SpawnImplementation(ctx context.Context, config AgentConfig) (*AgentSession, error)

	// SpawnVerification spawns a verification sub-agent.
	SpawnVerification(ctx context.Context, config AgentConfig) (*AgentSession, error)

	// Terminate terminates an agent session.
	Terminate(ctx context.Context, sessionID string) error

	// GetSessionStatus gets the status of a session.
	GetSessionStatus(ctx context.Context, sessionID string) (*AgentSessionStatus, error)
}

// AgentConfig represents configuration for agent spawning.
type AgentConfig struct {
	TaskID       string          `json:"task_id"`
	TaskType     string          `json:"task_type"` // "implementation", "verification"
	Context      *BoundedContext `json:"context"`
	Skills       []string        `json:"skills"`
	AgentsMDPath string          `json:"agents_md_path"`
	Constraints  []string        `json:"constraints,omitempty"`
}

// AgentSession represents an active agent session.
type AgentSession struct {
	SessionID   string    `json:"session_id"`
	TaskID      string    `json:"task_id"`
	StartedAt   time.Time `json:"started_at"`
	ContextUsed int       `json:"context_used"`
}

// AgentSessionStatus represents the status of an agent session.
type AgentSessionStatus struct {
	SessionID string  `json:"session_id"`
	Status    string  `json:"status"` // "running", "completed", "failed"
	Progress  float64 `json:"progress"`
	Output    string  `json:"output,omitempty"`
	Error     string  `json:"error,omitempty"`
	Completed bool    `json:"completed"`
	ExitCode  int     `json:"exit_code,omitempty"`
}

// ============================================================================
// Document Generator Interface
// ============================================================================

// DocumentGenerator generates SDD documentation files.
type DocumentGenerator interface {
	// GenerateSpecDoc generates SPEC.md.
	GenerateSpecDoc(ctx context.Context, spec *SpecOutput) (*GeneratedDoc, error)

	// GenerateDesignDoc generates DESIGN.md.
	GenerateDesignDoc(ctx context.Context, spec *SpecOutput) (*GeneratedDoc, error)

	// GenerateTasksDoc generates TASKS.md.
	GenerateTasksDoc(ctx context.Context, tasks []Task) (*GeneratedDoc, error)

	// GenerateAgentsDoc generates AGENTS.md.
	GenerateAgentsDoc(ctx context.Context, config AgentsDocConfig) (*GeneratedDoc, error)

	// GenerateSkillDoc generates a SKILL.md file.
	GenerateSkillDoc(ctx context.Context, config SkillDocConfig) (*GeneratedDoc, error)
}

// GeneratedDoc represents a generated document.
type GeneratedDoc struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Hash    string `json:"hash"`
}

// AgentsDocConfig represents configuration for AGENTS.md generation.
type AgentsDocConfig struct {
	ProjectPath string   `json:"project_path"`
	Scope       string   `json:"scope"` // "root" or module name
	Modules     []string `json:"modules,omitempty"`
	Skills      []string `json:"skills,omitempty"`
}

// SkillDocConfig represents configuration for SKILL.md generation.
type SkillDocConfig struct {
	SkillName    string   `json:"skill_name"`
	SkillPath    string   `json:"skill_path"`
	Stack        string   `json:"stack"`
	Triggers     []string `json:"triggers"`
	Description  string   `json:"description"`
	Instructions string   `json:"instructions"`
}

// ============================================================================
// Quality Gate Interface
// ============================================================================

// QualityGate evaluates documentation quality and determines if loop can proceed.
type QualityGate interface {
	// Evaluate evaluates documentation quality.
	Evaluate(ctx context.Context, specs []SpecArtifact) (*QualityGateResult, error)

	// ShouldInvokeSpec determines if GROVE Spec should be re-invoked.
	ShouldInvokeSpec(ctx context.Context, result *QualityGateResult) (bool, *FeedbackPayload, error)

	// GetThresholds returns the quality thresholds.
	GetThresholds(ctx context.Context) (*QualityThresholds, error)
}

// QualityGateResult represents the result of quality gate evaluation.
type QualityGateResult struct {
	Passed           bool          `json:"passed"`
	OverallScore     float64       `json:"overall_score"`
	Scores           QualityScores `json:"scores"`
	FailedDimensions []string      `json:"failed_dimensions"`
	Recommendations  []string      `json:"recommendations"`
}

// QualityThresholds represents quality thresholds.
type QualityThresholds struct {
	MinimumDimensionScore float64 `json:"minimum_dimension_score"` // default 8
	MinimumCompositeScore float64 `json:"minimum_composite_score"` // default 85
}

// ============================================================================
// Engram Integration Interface
// ============================================================================

// EngramClient provides integration with Engram persistent memory.
type EngramClient interface {
	// Save saves an observation to Engram.
	Save(ctx context.Context, observation EngramObservation) error

	// Search searches for observations.
	Search(ctx context.Context, query string, options *SearchOptions) ([]EngramObservation, error)

	// Get retrieves an observation by ID.
	Get(ctx context.Context, id int) (*EngramObservation, error)

	// Update updates an existing observation.
	Update(ctx context.Context, id int, observation EngramObservation) error

	// Delete deletes an observation.
	Delete(ctx context.Context, id int) error

	// GetSession returns session context.
	GetSession(ctx context.Context) (*SessionContext, error)
}

// EngramObservation represents an observation in Engram.
type EngramObservation struct {
	ID        int       `json:"id,omitempty"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Scope     string    `json:"scope"` // "project" or "personal"
	TopicKey  string    `json:"topic_key,omitempty"`
	Content   string    `json:"content"`
	Project   string    `json:"project,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// SearchOptions represents search options.
type SearchOptions struct {
	Project string    `json:"project,omitempty"`
	Scope   string    `json:"scope,omitempty"`
	Type    string    `json:"type,omitempty"`
	Limit   int       `json:"limit,omitempty"`
	Before  time.Time `json:"before,omitempty"`
	After   time.Time `json:"after,omitempty"`
}

// SessionContext represents session context.
type SessionContext struct {
	ID        string    `json:"id"`
	Project   string    `json:"project"`
	StartTime time.Time `json:"start_time"`
}

// ============================================================================
// Production Readiness Checker Interface
// ============================================================================

// ProductionReadinessChecker checks if a project is production ready.
type ProductionReadinessChecker interface {
	// Check runs all production readiness checks.
	Check(ctx context.Context, projectPath string) (*ReadinessReport, error)

	// CheckUserFlows validates all user flows.
	CheckUserFlows(ctx context.Context, projectPath string) (*FlowCheckResult, error)

	// CheckImports validates all imports.
	CheckImports(ctx context.Context, projectPath string) (*ImportCheckResult, error)

	// CheckDependencies validates all dependencies.
	CheckDependencies(ctx context.Context, projectPath string) (*DependencyCheckResult, error)

	// CheckSpecCompliance validates spec compliance.
	CheckSpecCompliance(ctx context.Context, projectPath string, specs []SpecArtifact) (*SpecComplianceResult, error)
}

// ReadinessReport represents the production readiness report.
type ReadinessReport struct {
	Ready           bool             `json:"ready"`
	TotalChecks     int              `json:"total_checks"`
	PassedChecks    int              `json:"passed_checks"`
	FailedChecks    int              `json:"failed_checks"`
	Warnings        int              `json:"warnings"`
	Issues          []ReadinessIssue `json:"issues,omitempty"`
	Recommendations []string         `json:"recommendations,omitempty"`
}

// ReadinessIssue represents a readiness issue.
type ReadinessIssue struct {
	Severity   string `json:"severity"` // "critical", "major", "minor"
	Category   string `json:"category"`
	Location   string `json:"location"`
	Problem    string `json:"problem"`
	Suggestion string `json:"suggestion,omitempty"`
}

// FlowCheckResult represents user flow check results.
type FlowCheckResult struct {
	Valid       bool         `json:"valid"`
	Flows       []FlowStatus `json:"flows"`
	BrokenFlows []string     `json:"broken_flows,omitempty"`
}

// FlowStatus represents the status of a user flow.
type FlowStatus struct {
	FlowName   string `json:"flow_name"`
	Valid      bool   `json:"valid"`
	StepsCount int    `json:"steps_count"`
}

// ImportCheckResult represents import check results.
type ImportCheckResult struct {
	Valid           bool     `json:"valid"`
	MissingImports  []string `json:"missing_imports,omitempty"`
	CircularImports []string `json:"circular_imports,omitempty"`
}

// DependencyCheckResult represents dependency check results.
type DependencyCheckResult struct {
	Valid           bool     `json:"valid"`
	MissingDeps     []string `json:"missing_deps,omitempty"`
	VersionMismatch []string `json:"version_mismatch,omitempty"`
}

// SpecComplianceResult represents spec compliance check results.
type SpecComplianceResult struct {
	Compliant   bool     `json:"compliant"`
	Implemented []string `json:"implemented,omitempty"`
	Missing     []string `json:"missing,omitempty"`
	Partial     []string `json:"partial,omitempty"`
}
