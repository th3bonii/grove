package types

import (
	"time"
)

// Phase represents a phase in the SDD workflow.
type Phase string

const (
	PhaseExplore Phase = "explore"
	PhasePropose Phase = "propose"
	PhaseSpec    Phase = "spec"
	PhaseDesign  Phase = "design"
	PhaseTasks   Phase = "tasks"
	PhaseApply   Phase = "apply"
	PhaseVerify  Phase = "verify"
	PhaseArchive Phase = "archive"
)

// Change represents a planned change in the system.
type Change struct {
	Name      string                 `json:"name" yaml:"name"`
	Phase     Phase                  `json:"phase" yaml:"phase"`
	Status    ChangeStatus           `json:"status" yaml:"status"`
	Intent    string                 `json:"intent,omitempty" yaml:"intent,omitempty"`
	Scope     []string               `json:"scope,omitempty" yaml:"scope,omitempty"`
	Approach  string                 `json:"approach,omitempty" yaml:"approach,omitempty"`
	CreatedAt time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" yaml:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ChangeStatus represents the status of a change.
type ChangeStatus string

const (
	StatusPending   ChangeStatus = "pending"
	StatusActive    ChangeStatus = "active"
	StatusCompleted ChangeStatus = "completed"
	StatusArchived  ChangeStatus = "archived"
)

// Spec represents a specification document.
type Spec struct {
	ChangeName   string            `json:"change_name" yaml:"change_name"`
	Version      string            `json:"version" yaml:"version"`
	Requirements []Requirement     `json:"requirements" yaml:"requirements"`
	Scenarios    []Scenario        `json:"scenarios" yaml:"scenarios"`
	Constraints  []string          `json:"constraints,omitempty" yaml:"constraints,omitempty"`
	Glossary     map[string]string `json:"glossary,omitempty" yaml:"glossary,omitempty"`
	CreatedAt    time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" yaml:"updated_at"`
}

// Scenario represents a test scenario.
type Scenario struct {
	ID       string              `json:"id" yaml:"id"`
	Title    string              `json:"title" yaml:"title"`
	Given    string              `json:"given" yaml:"given"`
	When     string              `json:"when" yaml:"when"`
	Then     string              `json:"then" yaml:"then"`
	Examples []map[string]string `json:"examples,omitempty" yaml:"examples,omitempty"`
	Tags     []string            `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Priority represents requirement priority.
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

// Design represents a technical design document.
type Design struct {
	ChangeName   string                 `json:"change_name" yaml:"change_name"`
	Architecture string                 `json:"architecture" yaml:"architecture"`
	Components   []Component            `json:"components" yaml:"components"`
	DataFlow     string                 `json:"data_flow,omitempty" yaml:"data_flow,omitempty"`
	Decisions    []ArchitectureDecision `json:"decisions" yaml:"decisions"`
	Dependencies []Dependency           `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	CreatedAt    time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" yaml:"updated_at"`
}

// ArchitectureDecision represents an architectural decision.
type ArchitectureDecision struct {
	ID           string `json:"id" yaml:"id"`
	Title        string `json:"title" yaml:"title"`
	Problem      string `json:"problem" yaml:"problem"`
	Decision     string `json:"decision" yaml:"decision"`
	Consequences string `json:"consequences,omitempty" yaml:"consequences,omitempty"`
}

// Dependency represents an external dependency.
type Dependency struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
	Purpose string `json:"purpose" yaml:"purpose"`
}

// TaskList represents a list of tasks for a change.
type TaskList struct {
	ChangeName string    `json:"change_name" yaml:"change_name"`
	Tasks      []Task    `json:"tasks" yaml:"tasks"`
	CreatedAt  time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" yaml:"updated_at"`
}

// Report represents a verification or archive report.
type Report struct {
	ChangeName string        `json:"change_name" yaml:"change_name"`
	Type       ReportType    `json:"type" yaml:"type"`
	Summary    string        `json:"summary" yaml:"summary"`
	Checks     []ReportCheck `json:"checks" yaml:"checks"`
	Issues     []Issue       `json:"issues,omitempty" yaml:"issues,omitempty"`
	CreatedAt  time.Time     `json:"created_at" yaml:"created_at"`
}

// ReportType represents the type of report.
type ReportType string

const (
	ReportVerify  ReportType = "verify"
	ReportArchive ReportType = "archive"
)

// ReportCheck represents a single check in a report.
type ReportCheck struct {
	Name    string      `json:"name" yaml:"name"`
	Status  CheckStatus `json:"status" yaml:"status"`
	Message string      `json:"message,omitempty" yaml:"message,omitempty"`
}

// CheckStatus represents the status of a check.
type CheckStatus string

const (
	CheckPass    CheckStatus = "pass"
	CheckFail    CheckStatus = "fail"
	CheckWarning CheckStatus = "warning"
	CheckSkip    CheckStatus = "skip"
)

// Issue represents an issue found during verification.
type Issue struct {
	Type        IssueType `json:"type" yaml:"type"`
	Title       string    `json:"title" yaml:"title"`
	Description string    `json:"description" yaml:"description"`
	Severity    Severity  `json:"severity" yaml:"severity"`
	File        string    `json:"file,omitempty" yaml:"file,omitempty"`
	Line        int       `json:"line,omitempty" yaml:"line,omitempty"`
}

// IssueType represents the type of issue.
type IssueType string

const (
	IssueMissing    IssueType = "missing"
	IssueMismatch   IssueType = "mismatch"
	IssueIncomplete IssueType = "incomplete"
)

// Severity represents issue severity.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// Score represents a quality score result.
type Score struct {
	Overall         float64            `json:"overall"`
	Dimensions      map[string]float64 `json:"dimensions"`
	Breakdown       []ScoreDimension   `json:"breakdown"`
	Recommendations []string           `json:"recommendations,omitempty"`
}

// ScoreDimension represents a scoring dimension.
type ScoreDimension struct {
	Name     string   `json:"name" yaml:"name"`
	Score    float64  `json:"score" yaml:"score"`
	MaxScore float64  `json:"max_score" yaml:"max_score"`
	Weight   float64  `json:"weight" yaml:"weight"`
	Details  string   `json:"details,omitempty" yaml:"details,omitempty"`
	Issues   []string `json:"issues,omitempty" yaml:"issues,omitempty"`
}

// Config represents the spec engine configuration.
type Config struct {
	ProjectName           string             `json:"project_name" yaml:"project_name"`
	ProjectPath           string             `json:"project_path" yaml:"project_path"`
	OutputPath            string             `json:"output_path" yaml:"output_path"`
	MaxIterations         int                `json:"max_iterations" yaml:"max_iterations"`
	QualityThreshold      float64            `json:"quality_threshold" yaml:"quality_threshold"`
	EnableSelfQuestioning bool               `json:"enable_self_questioning" yaml:"enable_self_questioning"`
	ScoringWeights        map[string]float64 `json:"scoring_weights,omitempty" yaml:"scoring_weights,omitempty"`
}

// Result represents the result of an engine operation.
type Result struct {
	Success   bool       `json:"success"`
	Context   *Context   `json:"context,omitempty"`
	Artifacts []Artifact `json:"artifacts,omitempty"`
	Errors    []string   `json:"errors,omitempty"`
	Metrics   *Metrics   `json:"metrics,omitempty"`
}

// Artifact represents a generated artifact.
type Artifact struct {
	Type      ArtifactType `json:"type" yaml:"type"`
	Path      string       `json:"path" yaml:"path"`
	Content   string       `json:"content,omitempty" yaml:"content,omitempty"`
	Generated bool         `json:"generated" yaml:"generated"`
}

// ArtifactType represents the type of artifact.
type ArtifactType string

const (
	ArtifactSpec   ArtifactType = "spec"
	ArtifactDesign ArtifactType = "design"
	ArtifactTasks  ArtifactType = "tasks"
	ArtifactReport ArtifactType = "report"
	ArtifactAgents ArtifactType = "agents"
	ArtifactState  ArtifactType = "state"
)

// Metrics represents execution metrics.
type Metrics struct {
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Duration      time.Duration `json:"duration"`
	Iterations    int           `json:"iterations"`
	TokensUsed    int64         `json:"tokens_used,omitempty"`
	FilesModified int           `json:"files_modified,omitempty"`
}

// Iteration represents a single iteration in the self-questioning loop.
type Iteration struct {
	Number    int       `json:"number"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Score     *Score    `json:"score,omitempty"`
	Improved  bool      `json:"improved"`
	Timestamp time.Time `json:"timestamp"`
}
