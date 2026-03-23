package types

import (
	"time"
)

// =============================================================================
// GROVE Ralph Loop Types
// =============================================================================

// LoopRunState represents the current state of a Ralph Loop execution.
type LoopRunState struct {
	ProjectPath   string          `json:"project_path"`
	StartedAt     time.Time       `json:"started_at"`
	LastUpdatedAt time.Time       `json:"last_updated_at"`
	LoopNumber    int             `json:"loop_number"`
	CurrentTaskID string          `json:"current_task_id,omitempty"`
	Tasks         []TaskExecution `json:"tasks"`
	Status        LoopStatus      `json:"status"`
	Errors        []LoopError     `json:"errors,omitempty"`
	Checkpoints   []Checkpoint    `json:"checkpoints,omitempty"`
}

// LoopStatus represents the overall status of the loop.
type LoopStatus string

const (
	LoopStatusInitializing LoopStatus = "initializing"
	LoopStatusRunning      LoopStatus = "running"
	LoopStatusPaused       LoopStatus = "paused"
	LoopStatusCompleted    LoopStatus = "completed"
	LoopStatusFailed       LoopStatus = "failed"
	LoopStatusRecovering   LoopStatus = "recovering"
)

// String returns the loop status as a readable string.
func (ls LoopStatus) String() string {
	switch ls {
	case LoopStatusInitializing:
		return "Initializing"
	case LoopStatusRunning:
		return "Running"
	case LoopStatusPaused:
		return "Paused"
	case LoopStatusCompleted:
		return "Completed"
	case LoopStatusFailed:
		return "Failed"
	case LoopStatusRecovering:
		return "Recovering"
	default:
		return "Unknown"
	}
}

// TaskExecution represents the execution state of a single task.
type TaskExecution struct {
	TaskID        string        `json:"task_id"`
	Status        TaskStatus    `json:"status"`
	Attempts      int           `json:"attempts"`
	MaxAttempts   int           `json:"max_attempts"` // Default: 3
	StartedAt     *time.Time    `json:"started_at,omitempty"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
	AssignedAgent string        `json:"assigned_agent,omitempty"`
	Result        *TaskResult   `json:"result,omitempty"`
	VerifyReport  *VerifyReport `json:"verify_report,omitempty"`
	Error         *LoopError    `json:"error,omitempty"`
}

// TaskResult represents the result of a task execution.
type TaskResult struct {
	Success      bool     `json:"success"`
	FilesChanged []string `json:"files_changed,omitempty"`
	FilesCreated []string `json:"files_created,omitempty"`
	Output       string   `json:"output,omitempty"`
	Message      string   `json:"message,omitempty"`
}

// VerifyReport represents the verification result after task implementation.
type VerifyReport struct {
	TaskID      string        `json:"task_id"`
	Timestamp   time.Time     `json:"timestamp"`
	Status      VerifyStatus  `json:"status"`
	Checks      []VerifyCheck `json:"checks"`
	PassedCount int           `json:"passed_count"`
	FailedCount int           `json:"failed_count"`
	Message     string        `json:"message,omitempty"`
	Suggestions []string      `json:"suggestions,omitempty"`
}

// VerifyStatus represents the outcome of verification.
type VerifyStatus string

const (
	VerifyStatusPassed  VerifyStatus = "passed"
	VerifyStatusFailed  VerifyStatus = "failed"
	VerifyStatusWarning VerifyStatus = "warning"
	VerifyStatusSkipped VerifyStatus = "skipped"
)

// VerifyCheck represents a single verification check.
type VerifyCheck struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	Status      VerifyStatus `json:"status"`
	Error       string       `json:"error,omitempty"`
	Details     string       `json:"details,omitempty"`
}

// Checkpoint represents a saved execution state for recovery.
type Checkpoint struct {
	ID         string        `json:"id"`
	Timestamp  time.Time     `json:"timestamp"`
	LoopNumber int           `json:"loop_number"`
	TaskID     string        `json:"task_id"`
	State      *LoopRunState `json:"state"`
	Reason     string        `json:"reason"`
}

// LoopError represents an error that occurred during loop execution.
type LoopError struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        ErrorType `json:"type"`
	Message     string    `json:"message"`
	TaskID      string    `json:"task_id,omitempty"`
	Recoverable bool      `json:"recoverable"`
	RetryCount  int       `json:"retry_count"`
}

// ErrorType categorizes loop errors.
type ErrorType string

const (
	ErrorTypeLLMResponse  ErrorType = "llm_response" // Malformed or empty LLM response
	ErrorTypeNetwork      ErrorType = "network"      // Network connectivity issues
	ErrorTypeFileSystem   ErrorType = "file_system"  // File read/write failures
	ErrorTypeVerification ErrorType = "verification" // Verification check failures
	ErrorTypeTimeout      ErrorType = "timeout"      // Operation timeout
	ErrorTypeUnknown      ErrorType = "unknown"      // Unclassified errors
)

// ErrorRecovery represents the error recovery strategy configuration.
type ErrorRecovery struct {
	MaxRetries       int     `json:"max_retries"`       // Default: 3
	BackoffBase      int     `json:"backoff_base"`      // Base for exponential backoff in seconds
	ReducedContext   bool    `json:"reduced_context"`   // Use reduced context on retry
	ContextReduction float64 `json:"context_reduction"` // Percentage of context to remove (0.0-1.0)
}

// AgentAction represents an action taken by a sub-agent.
type AgentAction struct {
	Timestamp   time.Time     `json:"timestamp"`
	AgentID     string        `json:"agent_id"`
	ActionType  ActionType    `json:"action_type"`
	TaskID      string        `json:"task_id,omitempty"`
	Description string        `json:"description"`
	Success     bool          `json:"success"`
	Duration    time.Duration `json:"duration"`
}

// ActionType categorizes agent actions.
type ActionType string

const (
	ActionTypeImplement ActionType = "implement"
	ActionTypeVerify    ActionType = "verify"
	ActionTypeRefactor  ActionType = "refactor"
	ActionTypeDocument  ActionType = "document"
	ActionTypeTest      ActionType = "test"
	ActionTypeReview    ActionType = "review"
)

// LogEntry represents a single entry in the loop audit log.
type LogEntry struct {
	Timestamp  time.Time         `json:"timestamp"`
	Level      LogLevel          `json:"level"`
	Component  string            `json:"component"`
	Message    string            `json:"message"`
	TaskID     string            `json:"task_id,omitempty"`
	LoopNumber int               `json:"loop_number,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// LogLevel represents log severity.
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)

// ProductionReadyReport represents the final production readiness assessment.
type ProductionReadyReport struct {
	ProjectPath         string            `json:"project_path"`
	GeneratedAt         time.Time         `json:"generated_at"`
	OverallStatus       ReadyStatus       `json:"overall_status"`
	TotalTasks          int               `json:"total_tasks"`
	CompletedTasks      int               `json:"completed_tasks"`
	FailedTasks         int               `json:"failed_tasks"`
	BlockedTasks        int               `json:"blocked_tasks"`
	VerificationResults []VerifyReport    `json:"verification_results"`
	QualityGate         QualityGateResult `json:"quality_gate"`
	DocumentationScore  *CompositeScore   `json:"documentation_score,omitempty"`
	Summary             string            `json:"summary"`
	Recommendations     []string          `json:"recommendations,omitempty"`
}

// ReadyStatus represents the production readiness result.
type ReadyStatus string

const (
	ReadyStatusProductionReady ReadyStatus = "production_ready"
	ReadyStatusNeedsWork       ReadyStatus = "needs_work"
	ReadyStatusBlocked         ReadyStatus = "blocked"
	ReadyStatusUnknown         ReadyStatus = "unknown"
)

// QualityGateResult represents the documentation quality gate assessment.
type QualityGateResult struct {
	Passed           bool           `json:"passed"`
	ScoreThreshold   int            `json:"score_threshold"` // Default: 85
	ActualScore      int            `json:"actual_score"`
	WeakDimensions   []DimensionKey `json:"weak_dimensions,omitempty"`
	NeedsEscalation  bool           `json:"needs_escalation"`
	EscalationReason string         `json:"escalation_reason,omitempty"`
}

// DocumentationQuality represents the quality assessment of project documentation.
type DocumentationQuality struct {
	SpecPresent       bool            `json:"spec_present"`
	DesignPresent     bool            `json:"design_present"`
	TasksPresent      bool            `json:"tasks_present"`
	AgentsPresent     bool            `json:"agents_present"`
	SkillsPresent     bool            `json:"skills_present"`
	QualityScore      *CompositeScore `json:"quality_score,omitempty"`
	MissingComponents []string        `json:"missing_components,omitempty"`
}

// ValidationResult represents the pre-loop validation outcome.
type ValidationResult struct {
	Valid           bool                 `json:"valid"`
	Documentation   DocumentationQuality `json:"documentation"`
	DependenciesMet bool                 `json:"dependencies_met"`
	MissingFiles    []string             `json:"missing_files,omitempty"`
	Warnings        []string             `json:"warnings,omitempty"`
	Suggestions     []string             `json:"suggestions,omitempty"`
}

// OrchestratorConfig represents Ralph Loop orchestrator configuration.
type OrchestratorConfig struct {
	ProjectPath          string        `json:"project_path"`
	MaxParallelTasks     int           `json:"max_parallel_tasks"` // Default: 1 (serial execution)
	VerifyAfterEach      bool          `json:"verify_after_each"`  // Default: true
	ErrorRecovery        ErrorRecovery `json:"error_recovery"`
	QualityGateThreshold int           `json:"quality_gate_threshold"` // Default: 85
	AutoEscalate         bool          `json:"auto_escalate"`          // Auto-invoke spec on quality gate fail
}

// DelegateContext represents the bounded context passed to sub-agents.
type DelegateContext struct {
	Task            *Task    `json:"task"`
	SpecSections    []string `json:"spec_sections"`   // File paths to relevant spec sections
	RelevantSkills  []string `json:"relevant_skills"` // Skill names to load
	AgentsFile      string   `json:"agents_file"`     // Path to relevant scoped AGENTS.md
	Constraints     []string `json:"constraints,omitempty"`
	SuccessCriteria string   `json:"success_criteria"`
}
