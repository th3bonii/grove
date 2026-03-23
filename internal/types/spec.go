// Package types provides Go type definitions for the GROVE ecosystem.
//
// GROVE (Gentleman's Robust Orchestration & Verification Engine) is a suite of
// tools for specification-driven development, autonomous code generation, and
// prompt optimization. This package contains the core type definitions used
// across all GROVE components.
//
// # GROVE Spec
//
// GROVE Spec transforms raw, unstructured project input into tightly structured
// specification packages ready for autonomous AI-driven development. It evaluates
// documentation quality across 7 dimensions with a target score of ≥85/100.
//
// # GROVE Ralph Loop
//
// GROVE Ralph Loop is an autonomous build engine that transforms complete
// documentation into production-ready code through iterative loops of validation,
// implementation, and verification.
//
// # GROVE Opti Prompt
//
// GROVE Opti Prompt optimizes natural language prompts into precise, project-aware
// instructions with file references, skill calls, and scope boundaries.
package types

import (
	"time"
)

// =============================================================================
// GROVE Spec Types
// =============================================================================

// ExitCondition represents the reason why a spec loop terminated.
type ExitCondition string

const (
	ExitNormal    ExitCondition = "normal"
	ExitSafetyNet ExitCondition = "safety_net"
	ExitManual    ExitCondition = "manual"
	ExitError     ExitCondition = "error"
)

// String returns a human-readable description of the exit condition.
func (ec ExitCondition) String() string {
	switch ec {
	case ExitNormal:
		return "normal termination"
	case ExitSafetyNet:
		return "safety net triggered (max iterations)"
	case ExitManual:
		return "manual intervention"
	case ExitError:
		return "error occurred"
	default:
		return "unknown"
	}
}

// DimensionKey identifies one of the 7 quality scoring dimensions.
type DimensionKey string

const (
	DimensionFlowCoverage               DimensionKey = "flow_coverage"
	DimensionComponentDecomposition     DimensionKey = "component_decomposition_depth"
	DimensionLogicalConsistency         DimensionKey = "logical_consistency"
	DimensionInterComponentConnectivity DimensionKey = "inter_component_connectivity"
	DimensionEdgeCaseCoverage           DimensionKey = "edge_case_coverage"
	DimensionDecisionJustification      DimensionKey = "decision_justification"
	DimensionAgentConsumability         DimensionKey = "agent_consumability"
)

// String returns the dimension name in a human-readable format.
func (dk DimensionKey) String() string {
	switch dk {
	case DimensionFlowCoverage:
		return "Flow Coverage"
	case DimensionComponentDecomposition:
		return "Component Decomposition"
	case DimensionLogicalConsistency:
		return "Logical Consistency"
	case DimensionInterComponentConnectivity:
		return "Inter-component Connectivity"
	case DimensionEdgeCaseCoverage:
		return "Edge Case Coverage"
	case DimensionDecisionJustification:
		return "Decision Justification"
	case DimensionAgentConsumability:
		return "Agent Consumability"
	default:
		return "unknown"
	}
}

// QualityScore represents the evaluation result for a single quality dimension.
type QualityScore struct {
	Dimension DimensionKey `json:"dimension"`
	Score     int          `json:"score"`     // 0-10
	MaxScore  int          `json:"max_score"` // Always 10 for dimensions
	Notes     string       `json:"notes,omitempty"`
}

// CompositeScore represents the overall quality evaluation result.
type CompositeScore struct {
	Scores       []QualityScore `json:"scores"`
	Composite    int            `json:"composite"`     // Sum of weighted scores (0-100)
	MinDimension int            `json:"min_dimension"` // Lowest single dimension score
	Passed       bool           `json:"passed"`        // All dims ≥8 AND composite ≥85
}

// LoopState represents the current state of a spec generation loop.
type LoopState struct {
	LoopNumber               int            `json:"loop_number"`
	DimensionScores          []QualityScore `json:"dimension_scores"`
	CompositeScore           int            `json:"composite_score"`
	ContentDeltaPct          float64        `json:"content_delta_pct"`
	ConsecutiveLowDeltaCount int            `json:"consecutive_low_delta_count"`
	ExitCondition            ExitCondition  `json:"exit_condition"`
	StartedAt                time.Time      `json:"started_at"`
	CompletedAt              *time.Time     `json:"completed_at,omitempty"`
}

// WebCacheEntry represents a cached web or MCP query result.
type WebCacheEntry struct {
	Query     string    `json:"query"`
	Source    string    `json:"source"` // "web" or "context7"
	Summary   string    `json:"summary"`
	Timestamp time.Time `json:"timestamp"`
}

// Component represents a decomposed feature or UI element.
type Component struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        ComponentType `json:"type"`
	ParentID    string        `json:"parent_id,omitempty"`

	// Decomposition details
	States        []ComponentState `json:"states,omitempty"`
	Behaviors     []Behavior       `json:"behaviors,omitempty"`
	EdgeCases     []EdgeCase       `json:"edge_cases,omitempty"`
	SubComponents []string         `json:"sub_components,omitempty"` // IDs of child components

	// Technical details
	FrontendDetails *FrontendDetails `json:"frontend_details,omitempty"`
	BackendDetails  *BackendDetails  `json:"backend_details,omitempty"`

	// Connections
	Dependencies []string `json:"dependencies,omitempty"` // IDs of components this depends on
	ConnectedTo  []string `json:"connected_to,omitempty"` // IDs of connected components

	// Documentation
	Justification string `json:"justification,omitempty"`
	Inferred      bool   `json:"inferred,omitempty"` // True if reverse documentation mode
}

// ComponentType categorizes a component.
type ComponentType string

const (
	ComponentTypeUI          ComponentType = "ui"
	ComponentTypeBackend     ComponentType = "backend"
	ComponentTypeService     ComponentType = "service"
	ComponentTypeData        ComponentType = "data"
	ComponentTypeIntegration ComponentType = "integration"
	ComponentTypeUtility     ComponentType = "utility"
)

// String returns the component type name.
func (ct ComponentType) String() string {
	switch ct {
	case ComponentTypeUI:
		return "UI"
	case ComponentTypeBackend:
		return "Backend"
	case ComponentTypeService:
		return "Service"
	case ComponentTypeData:
		return "Data"
	case ComponentTypeIntegration:
		return "Integration"
	case ComponentTypeUtility:
		return "Utility"
	default:
		return "unknown"
	}
}

// ComponentState represents a distinct state of a component.
type ComponentState struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	IsDefault   bool         `json:"is_default"`
	Transitions []Transition `json:"transitions,omitempty"`
}

// Transition represents a state transition.
type Transition struct {
	FromState string `json:"from_state"`
	ToState   string `json:"to_state"`
	Trigger   string `json:"trigger"`
	Condition string `json:"condition,omitempty"`
}

// Behavior represents an action or interaction of a component.
type Behavior struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Trigger        string   `json:"trigger"`
	Preconditions  []string `json:"preconditions,omitempty"`
	Postconditions []string `json:"postconditions,omitempty"`
}

// EdgeCase represents an unusual or boundary condition.
type EdgeCase struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // "low", "medium", "high", "critical"
	Handling    string `json:"handling"`
}

// FrontendDetails describes frontend-specific implementation details.
type FrontendDetails struct {
	Framework     string   `json:"framework,omitempty"`
	ComponentFile string   `json:"component_file,omitempty"`
	Props         []Prop   `json:"props,omitempty"`
	Events        []Event  `json:"events,omitempty"`
	Accessibility []string `json:"accessibility,omitempty"`
}

// BackendDetails describes backend-specific implementation details.
type BackendDetails struct {
	Language    string `json:"language,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	Method      string `json:"method,omitempty"`
	RequestBody string `json:"request_body,omitempty"`
	Response    string `json:"response,omitempty"`
}

// Prop represents a component property.
type Prop struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Default  string `json:"default,omitempty"`
}

// Event represents a component event handler.
type Event struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Payload     string `json:"payload,omitempty"`
}

// UserFlow represents a complete user interaction path through the application.
type UserFlow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []UserFlowStep `json:"steps"`
	EntryPoint  string         `json:"entry_point"`
	ExitPoint   string         `json:"exit_point"`
	Covers      []string       `json:"covers,omitempty"` // Component IDs covered by this flow
}

// UserFlowStep represents a single step in a user flow.
type UserFlowStep struct {
	StepNumber     int      `json:"step_number"`
	Action         string   `json:"action"`
	ComponentID    string   `json:"component_id"`
	State          string   `json:"state,omitempty"`
	ExpectedResult string   `json:"expected_result,omitempty"`
	Alternatives   []string `json:"alternatives,omitempty"`
}

// SpecDocument represents the complete specification document.
type SpecDocument struct {
	Title        string        `json:"title"`
	Version      string        `json:"version"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Overview     string        `json:"overview"`
	Components   []Component   `json:"components"`
	UserFlows    []UserFlow    `json:"user_flows"`
	Requirements []Requirement `json:"requirements"`
	Assumptions  []Assumption  `json:"assumptions,omitempty"`
}

// Requirement represents a functional or non-functional requirement.
type Requirement struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // "functional", "non-functional", "business"
	Description string `json:"description"`
	Priority    string `json:"priority"` // "must", "should", "could", "won't"
}

// Assumption documents an explicit assumption made during specification.
type Assumption struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
	Rationale string `json:"rationale,omitempty"`
}

// DesignDocument represents the technical architecture and design decisions.
type DesignDocument struct {
	Title              string          `json:"title"`
	Version            string          `json:"version"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	Architecture       string          `json:"architecture"`
	TechStack          []TechStackItem `json:"tech_stack"`
	Decisions          []Decision      `json:"decisions"`
	DirectoryStructure string          `json:"directory_structure,omitempty"`
}

// TechStackItem represents a technology in the stack.
type TechStackItem struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Purpose      string   `json:"purpose"`
	Alternatives []string `json:"alternatives,omitempty"`
}

// Decision represents an architectural or design decision.
type Decision struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Decision      string   `json:"decision"`
	Alternatives  []string `json:"alternatives,omitempty"`
	Justification string   `json:"justification"`
	Consequences  string   `json:"consequences,omitempty"`
}

// TasksDocument represents the implementation task breakdown.
type TasksDocument struct {
	Title      string      `json:"title"`
	Version    string      `json:"version"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Tasks      []Task      `json:"tasks"`
	Milestones []Milestone `json:"milestones,omitempty"`
}

// Task represents a single implementation task.
type Task struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	ComponentID     string     `json:"component_id,omitempty"`
	Priority        string     `json:"priority"` // "critical", "high", "medium", "low"
	Status          TaskStatus `json:"status"`
	EstimatedEffort string     `json:"estimated_effort,omitempty"`
	Dependencies    []string   `json:"dependencies,omitempty"`
	Skills          []string   `json:"skills,omitempty"`
}

// TaskStatus represents the status of an implementation task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusBlocked    TaskStatus = "blocked"
)

// String returns the task status as a readable string.
func (ts TaskStatus) String() string {
	switch ts {
	case TaskStatusPending:
		return "Pending"
	case TaskStatusInProgress:
		return "In Progress"
	case TaskStatusCompleted:
		return "Completed"
	case TaskStatusBlocked:
		return "Blocked"
	default:
		return "Unknown"
	}
}

// Milestone represents a grouping of tasks.
type Milestone struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TaskIDs     []string `json:"task_ids"`
}

// SpecLoopLog represents a single iteration's log entry.
type SpecLoopLog struct {
	LoopNumber     int            `json:"loop_number"`
	Timestamp      time.Time      `json:"timestamp"`
	Scores         []QualityScore `json:"scores"`
	CompositeScore int            `json:"composite_score"`
	DeltaPct       float64        `json:"delta_pct"`
	Actions        []string       `json:"actions"`
	IssuesFound    []string       `json:"issues_found,omitempty"`
	IssuesResolved []string       `json:"issues_resolved,omitempty"`
}

// CompletionReport represents the final GROVE Spec completion report.
type CompletionReport struct {
	TotalLoops      int             `json:"total_loops"`
	ExitCondition   ExitCondition   `json:"exit_condition"`
	FinalScores     CompositeScore  `json:"final_scores"`
	FilesGenerated  []string        `json:"files_generated"`
	FilesMerged     []string        `json:"files_merged,omitempty"`
	AssumptionsMade []Assumption    `json:"assumptions_made,omitempty"`
	SkillConflicts  []SkillConflict `json:"skill_conflicts,omitempty"`
	Timestamp       time.Time       `json:"timestamp"`
}

// SkillConflict documents a naming conflict with an existing skill.
type SkillConflict struct {
	DesiredName   string `json:"desired_name"`
	ResolvedName  string `json:"resolved_name"`
	OriginalSkill string `json:"original_skill"`
}

// Context represents the execution context for the spec engine.
type Context struct {
	Change     *Change                `json:"change,omitempty"`
	InputText  string                 `json:"input_text,omitempty"`
	OutputText string                 `json:"output_text,omitempty"`
	Score      *Score                 `json:"score,omitempty"`
	Spec       *SpecDocument          `json:"spec,omitempty"`
	Design     *DesignDocument        `json:"design,omitempty"`
	Tasks      *TasksDocument         `json:"tasks,omitempty"`
	Report     *Report                `json:"report,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
