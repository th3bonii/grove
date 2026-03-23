// Package loop provides the core Ralph Loop engine for GROVE.
//
// Ralph Loop is an autonomous documentation-to-code execution engine that:
//   - Validates documentation before processing
//   - Loads and manages implementation tasks
//   - Orchestrates execution across multiple phases
//   - Persists state for checkpoint/resume capability
package loop

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Gentleman-Programming/grove/internal/sdd"
	"github.com/Gentleman-Programming/grove/internal/types"
)

// LoopPhase represents the current execution phase of the loop.
type LoopPhase string

const (
	PhaseInitial    LoopPhase = "initial"
	PhasePreFlight  LoopPhase = "pre-flight"
	PhasePropose    LoopPhase = "propose"
	PhaseAnalyze    LoopPhase = "analyze"
	PhaseSpec       LoopPhase = "spec"
	PhaseDesign     LoopPhase = "design"
	PhaseTasks      LoopPhase = "tasks"
	PhaseImplement  LoopPhase = "implement"
	PhaseVerify     LoopPhase = "verify"
	PhaseProduction LoopPhase = "production"
	PhaseArchive    LoopPhase = "archive"
	PhaseComplete   LoopPhase = "complete"
	PhasePaused     LoopPhase = "paused"
	PhaseFailed     LoopPhase = "failed"
)

func (p LoopPhase) String() string {
	return string(p)
}

// IsTerminal returns true if this is a terminal phase.
func (p LoopPhase) IsTerminal() bool {
	return p == PhaseComplete || p == PhaseFailed
}

// LoopStatus represents the overall status of the loop execution.
type LoopStatus string

const (
	StatusPending   LoopStatus = "pending"
	StatusRunning   LoopStatus = "running"
	StatusPaused    LoopStatus = "paused"
	StatusCompleted LoopStatus = "completed"
	StatusFailed    LoopStatus = "failed"
	StatusBlocked   LoopStatus = "blocked"
)

// OrchestratorState represents the runtime state of the orchestrator.
type OrchestratorState struct {
	Phase       LoopPhase
	Status      LoopStatus
	CurrentTask string
	Progress    float64 // 0.0 to 1.0
	StartedAt   time.Time
	UpdatedAt   time.Time
}

// SDDClientExecutor defines the interface for SDD client execution.
type SDDClientExecutor interface {
	Execute(ctx context.Context, phase sdd.Phase, input map[string]interface{}) (*sdd.Result, error)
}

// Orchestrator manages the execution of Ralph Loop phases and tasks.
type Orchestrator struct {
	mu             sync.RWMutex
	state          *OrchestratorState
	config         *OrchestratorConfig
	validator      *Validator
	stateMgr       *StateManager
	sddClient      SDDClientExecutor
	gentleClient   *GentleClient // Cliente de gentle-ai
	ggaClient      *GGAClient    // Ángel Guardián Caballero
	spec           *types.SpecDocument
	agentsFile     string // Path to scoped AGENTS.md for delegate context
	ctx            context.Context
	cancel         context.CancelFunc
	pauseCh        chan struct{}
	resumeCh       chan struct{}
	tasks          []Task
	completedTasks map[string]bool
	phaseHandlers  map[LoopPhase]PhaseHandler
}

// OrchestratorConfig contains configuration for the orchestrator.
type OrchestratorConfig struct {
	// ProjectPath is the root path of the project being processed.
	ProjectPath string

	// DocsPath is the path to documentation files.
	DocsPath string

	// StateDir is the directory for persisting state.
	StateDir string

	// CheckpointEnabled enables checkpoint/resume functionality.
	CheckpointEnabled bool

	// MaxRetries is the maximum number of retries for failed tasks.
	MaxRetries int

	// BackoffBaseMs is the base delay for exponential backoff (in milliseconds).
	BackoffBaseMs int64

	// ErrorRecovery contains settings for error recovery and retry behavior.
	ErrorRecovery *ErrorRecoveryConfig

	// SDDClient allows passing a custom SDD client (for testing with mocks).
	// If not provided, a real client will be created.
	SDDClient interface {
		Execute(ctx context.Context, phase sdd.Phase, input map[string]interface{}) (*sdd.Result, error)
	}

	// OnPhaseChange is called when the phase changes.
	OnPhaseChange func(from, to LoopPhase)

	// OnTaskComplete is called when a task completes.
	OnTaskComplete func(task *Task, err error)

	// OnError is called when an error occurs.
	OnError func(err error)
}

// ErrorRecoveryConfig contains settings for error recovery.
type ErrorRecoveryConfig struct {
	// MaxRetries is the maximum number of retries for failed tasks.
	MaxRetries int

	// BackoffBaseMs is the base delay for exponential backoff in milliseconds.
	BackoffBaseMs int64

	// ContextReduction is the percentage (0.0-1.0) to reduce context on retry.
	ContextReduction float64

	// EnableSmartRetry enables intelligent retry based on error type.
	EnableSmartRetry bool
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		CheckpointEnabled: true,
		MaxRetries:        3,
		BackoffBaseMs:     1000,
		ErrorRecovery: &ErrorRecoveryConfig{
			MaxRetries:       3,
			BackoffBaseMs:    1000,
			ContextReduction: 0.3,
			EnableSmartRetry: true,
		},
	}
}

// PhaseHandler defines the interface for phase-specific behavior.
type PhaseHandler interface {
	// Execute runs the phase logic.
	Execute(ctx context.Context, orch *Orchestrator) error

	// CanProceed checks if the phase can transition to the next.
	CanProceed(ctx context.Context) (bool, error)

	// Name returns the phase name.
	Name() string
}

// NewOrchestrator creates a new Orchestrator instance.
func NewOrchestrator(config *OrchestratorConfig) *Orchestrator {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Use provided SDD client or create a real one
	var sddClient SDDClientExecutor
	if config.SDDClient != nil {
		sddClient = config.SDDClient.(SDDClientExecutor)
	} else {
		sddClient = sdd.NewClient(config.ProjectPath)
	}

	// Initialize Gentle AI client
	var gentleClient *GentleClient
	if endpoint := os.Getenv("GENTLE_ENDPOINT"); endpoint != "" {
		gentleClient = NewGentleClient(endpoint, os.Getenv("GENTLE_API_KEY"))
	}

	// Initialize GGA client for provider switching
	ggaClient := NewGGAClient(nil)

	orch := &Orchestrator{
		state: &OrchestratorState{
			Phase:     PhaseInitial,
			Status:    StatusPending,
			Progress:  0.0,
			StartedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		config:         config,
		validator:      NewValidator(config.DocsPath, config.DocsPath),
		stateMgr:       NewStateManager(config.StateDir),
		sddClient:      sddClient,
		gentleClient:   gentleClient,
		ggaClient:      ggaClient,
		ctx:            ctx,
		cancel:         cancel,
		pauseCh:        make(chan struct{}),
		resumeCh:       make(chan struct{}),
		tasks:          make([]Task, 0),
		completedTasks: make(map[string]bool),
		phaseHandlers:  make(map[LoopPhase]PhaseHandler),
	}

	// Register default phase handlers
	orch.registerDefaultHandlers()

	return orch
}

// registerDefaultHandlers registers built-in phase handlers.
func (o *Orchestrator) registerDefaultHandlers() {
	// The orchestrator uses ExecuteTask for task execution
	// Phase-specific behavior can be added here
}

// State returns the current orchestrator state (thread-safe).
func (o *Orchestrator) State() *OrchestratorState {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Return a copy to prevent external modification
	state := *o.state
	return &state
}

// Phase returns the current phase.
func (o *Orchestrator) Phase() LoopPhase {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.state.Phase
}

// Status returns the current status.
func (o *Orchestrator) Status() LoopStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.state.Status
}

// Run executes the Ralph Loop from the current phase to completion.
// It blocks until the loop completes, fails, or is cancelled.
func (o *Orchestrator) Run() error {
	return o.RunWithTasks(nil)
}

// RunWithTasks executes the Ralph Loop with the provided tasks.
func (o *Orchestrator) RunWithTasks(tasks []Task) error {
	if len(tasks) > 0 {
		o.tasks = tasks
	}

	// Check readiness first
	if ready, err := o.CheckReadiness(); !ready {
		return fmt.Errorf("orchestrator not ready: %w", err)
	}

	o.setState(PhasePreFlight, StatusRunning)

	// Load previous state if available
	if o.config.CheckpointEnabled {
		if err := o.loadCheckpoint(); err != nil {
			if o.config.OnError != nil {
				o.config.OnError(fmt.Errorf("failed to load checkpoint: %w", err))
			}
		}
	}

	// Execute phases in order
	phases := []LoopPhase{
		PhasePreFlight,
		PhaseAnalyze,
		PhaseSpec,
		PhaseDesign,
		PhaseTasks,
		PhaseImplement,
		PhaseVerify,
		PhaseProduction,
		PhaseArchive,
	}

	for _, phase := range phases {
		select {
		case <-o.ctx.Done():
			return o.ctx.Err()
		default:
		}

		if err := o.executePhase(phase); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			return fmt.Errorf("phase %s failed: %w", phase, err)
		}
	}

	o.setState(PhaseComplete, StatusCompleted)
	o.updateProgress(1.0)

	// Save final state
	if o.config.CheckpointEnabled {
		if err := o.saveCheckpoint(); err != nil {
			if o.config.OnError != nil {
				o.config.OnError(fmt.Errorf("failed to save checkpoint: %w", err))
			}
		}
	}

	return nil
}

// executePhase executes a single phase of the loop.
func (o *Orchestrator) executePhase(phase LoopPhase) error {
	previousPhase := o.state.Phase
	o.setState(phase, StatusRunning)

	// Notify phase change
	if o.config.OnPhaseChange != nil {
		o.config.OnPhaseChange(previousPhase, phase)
	}

	// Execute phase-specific logic
	switch phase {
	case PhasePreFlight:
		return o.executePreFlight()
	case PhaseAnalyze:
		return o.executeAnalyze()
	case PhaseImplement:
		return o.executeTasks()
	case PhaseVerify:
		// Pasamos task nil y result nil para verificación general
		// En implementación real, se passing la task específica siendo verificada
		report, err := o.executeVerify(types.Task{}, nil)
		if err != nil {
			return err
		}
		if report.Status == types.VerifyStatusFailed {
			return errors.New("verification failed: " + report.Message)
		}
		o.updateProgressForPhase(PhaseVerify)
		return nil
	default:
		// For other phases, just update progress
		o.updateProgressForPhase(phase)
	}

	return nil
}

// executePreFlight runs the pre-flight validation phase.
func (o *Orchestrator) executePreFlight() error {
	result, err := o.validator.Validate(o.config.DocsPath)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid {
		o.addError(ValidationError{
			Level:   ValidationLevelError,
			Code:    "VALIDATION_FAILED",
			Message: "Pre-flight validation failed",
		})
		return errors.New("pre-flight validation failed")
	}

	// Load tasks after successful validation
	tasks, err := o.validator.LoadTasks(o.config.DocsPath)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	if len(tasks) > 0 {
		o.tasks = tasks
	}

	o.updateProgressForPhase(PhasePreFlight)
	return nil
}

// executeAnalyze runs the analysis phase.
func (o *Orchestrator) executeAnalyze() error {
	// Analysis phase processes documentation
	o.updateProgressForPhase(PhaseAnalyze)
	return nil
}

// executeTasks executes all pending tasks.
func (o *Orchestrator) executeTasks() error {
	var pendingTasks []Task

	for _, task := range o.tasks {
		if o.completedTasks[task.ID] {
			continue
		}
		pendingTasks = append(pendingTasks, task)
	}

	if len(pendingTasks) == 0 {
		return nil
	}

	for _, task := range pendingTasks {
		select {
		case <-o.ctx.Done():
			return o.ctx.Err()
		case <-o.pauseCh:
			o.setState(PhasePaused, StatusPaused)
			<-o.resumeCh
			o.setState(PhaseImplement, StatusRunning)
		default:
		}

		o.state.CurrentTask = task.ID

		if err := o.ExecuteTask(&task); err != nil {
			o.completedTasks[task.ID] = false
			o.addError(ValidationError{
				Level:   ValidationLevelError,
				Code:    "TASK_FAILED",
				Message: fmt.Sprintf("Task %s failed: %v", task.ID, err),
				Field:   task.ID,
			})

			// Retry logic
			for i := 0; i < o.config.MaxRetries; i++ {
				backoff := time.Duration(o.config.BackoffBaseMs) * time.Millisecond * time.Duration(1<<i)
				time.Sleep(backoff)

				if retryErr := o.ExecuteTask(&task); retryErr == nil {
					break
				}
			}
		} else {
			o.completedTasks[task.ID] = true
			o.setState(PhaseImplement, StatusRunning)

			// GUARDAR STATE DESPUÉS DE CADA TAREA
			if err := o.saveCheckpoint(); err != nil {
				slog.Warn("failed to save checkpoint", slog.String("task", task.ID), slog.String("error", err.Error()))
			}
		}
	}

	o.updateProgressForPhase(PhaseImplement)
	return nil
}

// executeVerify runs the verification phase.
func (o *Orchestrator) executeVerify(task types.Task, result *types.TaskResult) (*types.VerifyReport, error) {
	// El orchestrator NO evalúa calidad de código directamente
	// Delega la verificación al skill sdd-verify
	// y solo lee el verify-report con el veredicto (PASS/FAIL)

	report := &types.VerifyReport{
		TaskID:    task.ID,
		Status:    types.VerifyStatusPassed,
		Timestamp: time.Now(),
	}

	// TODO: En implementación real, esto sería:
	// 1. Cargar skill: skill({ name: 'sdd-verify' })
	// 2. Pasar task + spec sections + apply-progress
	// 3. El skill retorna verify-report con PASS/FAIL
	// 4. Orchestrator solo lee el veredicto

	return report, nil
}

// ExecuteTask executes a single task with bounded context and retry logic.
// Returns an error if the task fails after all retries.
func (o *Orchestrator) ExecuteTask(task *Task) error {
	return o.ExecuteTaskWithContext(context.Background(), task)
}

// ExecuteTaskWithContext executes a task with an external context and full retry logic.
func (o *Orchestrator) ExecuteTaskWithContext(ctx context.Context, task *Task) error {
	start := time.Now()

	// Step 1: Validate task
	if err := o.validateTask(task); err != nil {
		return err
	}

	// Step 2: Build bounded context for sub-agent delegation
	delegateCtx := o.buildDelegateContext(task)

	// Step 3: Retry loop with exponential backoff
	var lastErr error
	for attempt := 0; attempt < o.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: base * 2^attempt
			backoff := time.Duration(o.config.BackoffBaseMs) * time.Millisecond * time.Duration(1<<attempt)
			slog.Info("retrying task",
				slog.String("task", task.ID),
				slog.Int("attempt", attempt+1),
				slog.Duration("backoff", backoff))

			select {
			case <-time.After(backoff):
				// Continue with retry
			case <-ctx.Done():
				return ctx.Err()
			}

			// Reduce context on retry to avoid token bloat
			delegateCtx = o.reduceContext(delegateCtx)
		}

		// Step 4: Execute the task via SDD client
		err := o.executeTaskWithContext(ctx, task, delegateCtx)
		if err == nil {
			// Success - update task state
			task.Completed = true
			task.UpdatedAt = time.Now()

			// Mark task as completed in the internal map
			o.completedTasks[task.ID] = true

			slog.Info("task completed",
				slog.String("task", task.ID),
				slog.Duration("duration", time.Since(start)))
			return nil
		}

		lastErr = err

		// Step 5: Check if error is retryable
		if !isRetryable(err) {
			slog.Warn("non-retryable error",
				slog.String("task", task.ID),
				slog.String("error", err.Error()))
			break
		}

		slog.Warn("task attempt failed",
			slog.String("task", task.ID),
			slog.Int("attempt", attempt+1),
			slog.String("error", err.Error()))
	}

	// All retries exhausted - mark task as failed
	task.Completed = false
	task.UpdatedAt = time.Now()

	// Log final failure - only if we have an error
	if lastErr != nil {
		slog.Error("task failed after retries",
			slog.String("task", task.ID),
			slog.String("error", lastErr.Error()),
			slog.Duration("duration", time.Since(start)))

		// Notify error callback
		if o.config.OnError != nil {
			o.config.OnError(lastErr)
		}
	}

	return lastErr
}

// validateTask validates task prerequisites before execution.
func (o *Orchestrator) validateTask(task *Task) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	// Update current task in state
	o.mu.Lock()
	o.state.CurrentTask = task.ID
	o.mu.Unlock()

	// Check blockers are completed
	for _, blocker := range task.Blockers {
		if !o.completedTasks[blocker] {
			return fmt.Errorf("blocker not completed: %s", blocker)
		}
	}

	// Check dependencies are met
	for _, dep := range task.DependsOn {
		if !o.completedTasks[dep] {
			return fmt.Errorf("dependency not met: %s", dep)
		}
	}

	return nil
}

// buildDelegateContext constructs the bounded context for sub-agent delegation.
func (o *Orchestrator) buildDelegateContext(task *Task) map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Extract relevant spec sections based on task
	var specSections []string
	if o.spec != nil {
		specSections = extractRelevantSpecSections(o.spec, task)
	}

	// Build skills list from task metadata
	skills := extractSkillsFromTask(task)

	// Build constraints from config and task
	constraints := buildConstraints(task, o.config)

	// Build bounded context as map for flexibility
	return map[string]interface{}{
		"task_id":          task.ID,
		"task_title":       task.Title,
		"task_description": task.Description,
		"spec_sections":    specSections,
		"relevant_skills":  skills,
		"agents_file":      o.agentsFile,
		"constraints":      constraints,
		"success_criteria": buildSuccessCriteria(task),
	}
}

// executeTaskWithContext executes the task using the SDD client.
func (o *Orchestrator) executeTaskWithContext(ctx context.Context, task *Task, delegateCtx map[string]interface{}) error {
	// Prepare input for SDD apply phase
	input := map[string]interface{}{
		"task":          task,
		"spec":          o.spec,
		"agents_file":   delegateCtx["agents_file"],
		"spec_sections": delegateCtx["spec_sections"],
		"skills":        delegateCtx["relevant_skills"],
		"constraints":   delegateCtx["constraints"],
		"success":       delegateCtx["success_criteria"],
		"project":       o.config.ProjectPath,
		"context":       delegateCtx, // Pass full bounded context
	}

	// Execute via SDD client
	result, err := o.sddClient.Execute(ctx, sdd.PhaseApply, input)
	if err != nil {
		return fmt.Errorf("SDD execute failed: %w", err)
	}

	// Check result status
	if result.Status == "failure" {
		if result.Error != "" {
			return errors.New(result.Error)
		}
		return errors.New("task execution returned failure status")
	}

	slog.Debug("SDD task execution completed",
		slog.String("task", task.ID),
		slog.String("status", result.Status),
		slog.String("summary", result.Summary))

	return nil
}

// reduceContext reduces context size for retry attempts.
// This helps prevent token bloat when retrying failed tasks.
func (o *Orchestrator) reduceContext(ctx map[string]interface{}) map[string]interface{} {
	// Create a reduced copy
	reduced := make(map[string]interface{})

	for k, v := range ctx {
		reduced[k] = v
	}

	// Truncate spec sections to first 3
	if sections, ok := ctx["spec_sections"].([]string); ok && len(sections) > 3 {
		reduced["spec_sections"] = sections[:3]
	}

	// Keep only essential skills (first 2)
	if skills, ok := ctx["relevant_skills"].([]string); ok && len(skills) > 2 {
		reduced["relevant_skills"] = skills[:2]
	}

	// Truncate constraints
	if constraints, ok := ctx["constraints"].([]string); ok && len(constraints) > 5 {
		reduced["constraints"] = constraints[:5]
	}

	return reduced
}

// handleLLMError handles LLM errors by attempting provider switching via GGA.
// Returns the error if switching fails or no more providers are available.
func (o *Orchestrator) handleLLMError(err error, attempt int) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// Check if this is an LLM-related error
	isLLMError := false
	llmPatterns := []string{"llm", "api", "rate limit", "timeout", "connection", "model"}
	for _, pattern := range llmPatterns {
		if strings.Contains(errStr, pattern) {
			isLLMError = true
			break
		}
	}

	if !isLLMError {
		return err
	}

	// Attempt to switch provider via GGA
	slog.Warn("LLM error detected, attempting provider switch",
		slog.String("error", err.Error()),
		slog.Int("attempt", attempt+1),
		slog.String("current_provider", o.ggaClient.CurrentProvider()))

	if switchErr := o.ggaClient.SwitchProvider(); switchErr != nil {
		slog.Error("provider switch failed",
			slog.String("error", switchErr.Error()),
			slog.String("current_provider", o.ggaClient.CurrentProvider()))
		return fmt.Errorf("LLM error and provider switch failed: %w", err)
	}

	slog.Info("provider switched successfully",
		slog.String("new_provider", o.ggaClient.CurrentProvider()))

	return nil
}

// extractRelevantSpecSections extracts spec sections relevant to a task.
func extractRelevantSpecSections(spec *types.SpecDocument, task *Task) []string {
	var sections []string

	if spec == nil {
		return sections
	}

	// Add overview if available
	if spec.Overview != "" {
		sections = append(sections, "SPEC.md#overview")
	}

	// Add requirements section
	if len(spec.Requirements) > 0 {
		sections = append(sections, "SPEC.md#requirements")
	}

	// Add components if task references any
	if task.Description != "" {
		sections = append(sections, "SPEC.md#components")
	}

	return sections
}

// extractSkillsFromTask extracts skill names from task metadata.
func extractSkillsFromTask(task *Task) []string {
	// For now, extract from task phase as skill category
	// In a real implementation, this would parse task metadata or tags
	skillMap := map[string][]string{
		"frontend": {"react", "angular", "vue"},
		"backend":  {"go", "nodejs", "python"},
		"database": {"sql", "mongodb", "redis"},
		"devops":   {"docker", "kubernetes", "terraform"},
		"test":     {"testing", "cypress", "playwright"},
	}

	phase := strings.ToLower(task.Phase)
	if skills, ok := skillMap[phase]; ok {
		return skills
	}

	// Default skill based on task phase
	return []string{"sdd-apply", phase}
}

// buildConstraints builds constraint list for task execution.
func buildConstraints(task *Task, config *OrchestratorConfig) []string {
	constraints := []string{
		"Follow project coding conventions",
		"Write tests for new functionality",
	}

	// Add project-specific constraints
	if config.ProjectPath != "" {
		constraints = append(constraints, fmt.Sprintf("Project: %s", config.ProjectPath))
	}

	return constraints
}

// buildSuccessCriteria builds the success criteria string for a task.
func buildSuccessCriteria(task *Task) string {
	if task.Title != "" {
		return fmt.Sprintf("Task '%s' must be completed with all code changes committed", task.Title)
	}
	return "Task must be completed successfully"
}

// isRetryable determines if an error should trigger a retry.
// isRetryable determines if an error should trigger a retry.
// Uses the new ErrorClassifier with additional fatal error detection.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Non-retryable errors (fatal) - these override classifier decision
	nonRetryable := []string{
		"skill not found",
		"task cannot be nil",
		"blocker not completed",
		"dependency not met",
		"validation failed",
		"permission denied",
	}

	for _, pattern := range nonRetryable {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	// Use the ErrorClassifier for intelligent retry decision
	classifier := NewErrorClassifier()
	return classifier.IsRetryable(err)
}

// CheckReadiness verifies that the orchestrator is ready to run.
// Returns true if ready, false otherwise.
func (o *Orchestrator) CheckReadiness() (bool, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Check if already running or completed
	if o.state.Status == StatusRunning {
		return false, errors.New("orchestrator is already running")
	}

	if o.state.Phase == PhaseComplete || o.state.Phase == PhaseFailed {
		return false, errors.New("orchestrator has already completed")
	}

	if o.state.Status == StatusFailed {
		return false, errors.New("orchestrator has failed")
	}

	if o.state.Status == StatusPaused {
		return true, nil // Can resume from paused state
	}

	// Check project path exists
	if o.config.ProjectPath != "" {
		// Project path validation would happen here
	}

	return true, nil
}

// Pause pauses the orchestrator execution.
// The orchestrator will pause after completing the current task.
func (o *Orchestrator) Pause() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.state.Status != StatusRunning {
		return errors.New("orchestrator is not running")
	}

	if o.state.Phase == PhasePaused {
		return errors.New("orchestrator is already paused")
	}

	// Set status to paused
	o.state.Status = StatusPaused
	o.state.Phase = PhasePaused
	o.state.UpdatedAt = time.Now()

	// Signal pause
	go func() {
		o.pauseCh <- struct{}{}
	}()

	// Save checkpoint before pausing
	if o.config.CheckpointEnabled {
		if err := o.saveCheckpoint(); err != nil {
			if o.config.OnError != nil {
				o.config.OnError(fmt.Errorf("failed to save checkpoint: %w", err))
			}
		}
	}

	return nil
}

// Resume resumes a paused orchestrator.
func (o *Orchestrator) Resume() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.state.Status != StatusPaused {
		return errors.New("orchestrator is not paused")
	}

	// Signal resume
	go func() {
		o.resumeCh <- struct{}{}
	}()

	o.state.Status = StatusRunning
	o.state.UpdatedAt = time.Now()

	return nil
}

// Stop stops the orchestrator immediately.
func (o *Orchestrator) Stop() error {
	o.cancel()

	o.mu.Lock()
	defer o.mu.Unlock()

	o.state.Status = StatusFailed
	o.state.UpdatedAt = time.Now()

	// Save final checkpoint
	if o.config.CheckpointEnabled {
		o.saveCheckpoint()
	}

	return nil
}

// GetTasks returns the current list of tasks.
func (o *Orchestrator) GetTasks() []Task {
	o.mu.RLock()
	defer o.mu.RUnlock()

	tasks := make([]Task, len(o.tasks))
	copy(tasks, o.tasks)
	return tasks
}

// GetCompletedTasks returns a map of completed task IDs.
func (o *Orchestrator) GetCompletedTasks() map[string]bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	completed := make(map[string]bool, len(o.completedTasks))
	for k, v := range o.completedTasks {
		completed[k] = v
	}
	return completed
}

// SetSpec sets the specification document for bounded context.
func (o *Orchestrator) SetSpec(spec *types.SpecDocument) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.spec = spec
}

// SetAgentsFile sets the path to the scoped AGENTS.md file.
func (o *Orchestrator) SetAgentsFile(path string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.agentsFile = path
}

// Progress returns the current progress as a percentage (0.0 to 1.0).
func (o *Orchestrator) Progress() float64 {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.state.Progress
}

// setState updates the orchestrator state (thread-safe).
func (o *Orchestrator) setState(phase LoopPhase, status LoopStatus) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.state.Phase = phase
	o.state.Status = status
	o.state.UpdatedAt = time.Now()
}

// updateProgress updates the progress percentage.
func (o *Orchestrator) updateProgress(progress float64) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.state.Progress = progress
	o.state.UpdatedAt = time.Now()
}

// updateProgressForPhase updates progress based on current phase.
func (o *Orchestrator) updateProgressForPhase(phase LoopPhase) {
	progress := float64(phaseIndex(phase)) / float64(phaseIndex(PhaseComplete))
	o.updateProgress(progress)
}

// phaseIndex returns the index of a phase in the execution order.
func phaseIndex(phase LoopPhase) int {
	phases := []LoopPhase{
		PhaseInitial,
		PhasePreFlight,
		PhaseAnalyze,
		PhaseSpec,
		PhaseDesign,
		PhaseTasks,
		PhaseImplement,
		PhaseVerify,
		PhaseProduction,
		PhaseArchive,
		PhaseComplete,
	}

	for i, p := range phases {
		if p == phase {
			return i
		}
	}
	return 0
}

// addError adds an error to the state.
func (o *Orchestrator) addError(err ValidationError) {
	if o.config.OnError != nil {
		o.config.OnError(err)
	}
}

// saveCheckpoint saves the current state to disk.
func (o *Orchestrator) saveCheckpoint() error {
	state := &LoopState{
		Version:       "1.0",
		Phase:         string(o.state.Phase),
		Status:        string(o.state.Status),
		CurrentTask:   o.state.CurrentTask,
		Tasks:         o.tasks,
		CheckpointNum: phaseIndex(o.state.Phase),
		StartedAt:     o.state.StartedAt,
		UpdatedAt:     time.Now(),
	}

	return o.stateMgr.SaveState(state)
}

// loadCheckpoint loads a previous state from disk.
func (o *Orchestrator) loadCheckpoint() error {
	state, err := o.stateMgr.LoadState()
	if err != nil || state == nil {
		return err
	}

	// Restore state
	o.tasks = state.Tasks
	o.state.Phase = LoopPhase(state.Phase)
	o.state.Status = LoopStatus(state.Status)
	o.state.CurrentTask = state.CurrentTask
	o.state.Progress = float64(state.CheckpointNum) / float64(phaseIndex(PhaseComplete))

	// Restore completed tasks
	for _, task := range state.Tasks {
		if task.Completed {
			o.completedTasks[task.ID] = true
		}
	}

	return nil
}

// Errors returns all errors encountered during execution.
func (o *Orchestrator) Errors() []ValidationError {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// This would track errors in a real implementation
	return []ValidationError{}
}
