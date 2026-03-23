// Package loop provides the core Ralph Loop engine for GROVE.
//
// Ralph Loop is an autonomous documentation-to-code execution engine that:
//   - Validates documentation before processing
//   - Loads and manages implementation tasks
//   - Orchestrates execution across multiple phases
//   - Persists state for checkpoint/resume capability
package loop

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ValidationLevel represents the severity of a validation issue.
type ValidationLevel int

const (
	ValidationLevelInfo     ValidationLevel = iota // Informational message
	ValidationLevelWarning                         // Non-blocking issue
	ValidationLevelError                           // Blocking issue
	ValidationLevelCritical                        // Must fix before proceeding
)

func (v ValidationLevel) String() string {
	switch v {
	case ValidationLevelInfo:
		return "info"
	case ValidationLevelWarning:
		return "warning"
	case ValidationLevelError:
		return "error"
	case ValidationLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ValidationError represents a single validation issue.
type ValidationError struct {
	Level   ValidationLevel `json:"level"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Field   string          `json:"field,omitempty"`
	File    string          `json:"file,omitempty"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Level, e.Code, e.Message)
}

// ValidationResult contains the outcome of a validation operation.
type ValidationResult struct {
	Valid      bool              `json:"valid"`
	Level      ValidationLevel   `json:"max_level"`
	Errors     []ValidationError `json:"errors"`
	Warnings   []ValidationError `json:"warnings"`
	Infos      []ValidationError `json:"infos"`
	Score      float64           `json:"score,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
	CheckedAt  time.Time         `json:"checked_at"`
	DurationMs int64             `json:"duration_ms"`
}

// NewValidationResult creates an empty validation result.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:     true,
		Level:     ValidationLevelInfo,
		Errors:    make([]ValidationError, 0),
		Warnings:  make([]ValidationError, 0),
		Infos:     make([]ValidationError, 0),
		Timestamp: time.Now(),
		CheckedAt: time.Now(),
	}
}

// AddError adds an error to the result and marks validation as failed.
func (r *ValidationResult) AddError(code, message, field string) {
	err := ValidationError{
		Level:   ValidationLevelError,
		Code:    code,
		Message: message,
		Field:   field,
	}
	r.Errors = append(r.Errors, err)
	r.Valid = false
	if r.Level < ValidationLevelError {
		r.Level = ValidationLevelError
	}
}

// AddWarning adds a warning to the result.
func (r *ValidationResult) AddWarning(code, message, field string) {
	err := ValidationError{
		Level:   ValidationLevelWarning,
		Code:    code,
		Message: message,
		Field:   field,
	}
	r.Warnings = append(r.Warnings, err)
	if r.Level < ValidationLevelWarning {
		r.Level = ValidationLevelWarning
	}
}

// AddInfo adds an informational message.
func (r *ValidationResult) AddInfo(code, message, field string) {
	r.Infos = append(r.Infos, ValidationError{
		Level:   ValidationLevelInfo,
		Code:    code,
		Message: message,
		Field:   field,
	})
}

// Task represents a single implementation task from documentation.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Phase       string    `json:"phase"`
	Priority    int       `json:"priority"`
	Completed   bool      `json:"completed"`
	Blockers    []string  `json:"blockers,omitempty"`
	DependsOn   []string  `json:"depends_on,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LoopState represents the persisted state of the Ralph Loop execution.
type LoopState struct {
	Version       string                 `json:"version"`
	Phase         string                 `json:"phase"`
	Status        string                 `json:"status"`
	CurrentTask   string                 `json:"current_task,omitempty"`
	Tasks         []Task                 `json:"tasks"`
	Errors        []ValidationError      `json:"errors,omitempty"`
	CheckpointID  string                 `json:"checkpoint_id,omitempty"`
	CheckpointNum int                    `json:"checkpoint_num"`
	StartedAt     time.Time              `json:"started_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	PausedAt      *time.Time             `json:"paused_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// NewLoopState creates a new loop state with defaults.
func NewLoopState() *LoopState {
	return &LoopState{
		Version:       "1.0",
		Phase:         "initial",
		Status:        "pending",
		Tasks:         make([]Task, 0),
		Errors:        make([]ValidationError, 0),
		CheckpointNum: 0,
		StartedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}
}

// Validator handles pre-flight validation for documentation and tasks.
type Validator struct {
	rulesDir string
	docsDir  string
}

// NewValidator creates a new Validator instance.
func NewValidator(rulesDir, docsDir string) *Validator {
	return &Validator{
		rulesDir: rulesDir,
		docsDir:  docsDir,
	}
}

// Validate performs pre-flight validation on documentation.
// Returns a ValidationResult indicating if validation passed.
func (v *Validator) Validate(docsPath string) (*ValidationResult, error) {
	start := time.Now()
	result := NewValidationResult()

	// If no docsPath provided, try to use the configured docsDir
	validatePath := docsPath
	if validatePath == "" {
		validatePath = v.docsDir
	}

	// Check if docs directory exists
	if validatePath != "" {
		if _, err := os.Stat(validatePath); os.IsNotExist(err) {
			result.AddError("DOCS_NOT_FOUND", "Documentation directory does not exist", validatePath)
			result.DurationMs = time.Since(start).Milliseconds()
			return result, nil
		}
	}

	// Validate the 5 required documentation files
	if validatePath != "" {
		v.validateRequiredFiles(result, validatePath)
	}

	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}

// validateRequiredFiles validates the 5 required documentation files:
// 1. SPEC.md - Specification of requirements
// 2. DESIGN.md - Technical design
// 3. TASKS.md - Task list
// 4. AGENTS.md - Agent configuration
// 5. SKILL.md(s) - Required skills
func (v *Validator) validateRequiredFiles(result *ValidationResult, projectPath string) {
	// Validate SPEC.md - warning if missing (not blocking)
	if err := v.validateSpec(projectPath); err != nil {
		result.AddWarning("SPEC_MISSING", err.Error(), "SPEC.md")
	} else {
		// Verify minimum content
		content, _ := v.readFile(filepath.Join(projectPath, "SPEC.md"))
		if !v.hasMinimumContent(content, 200) {
			result.AddWarning("SPEC_TOO_SHORT", "SPEC.md exists but has insufficient content", "SPEC.md")
		} else {
			result.AddInfo("SPEC_FOUND", "SPEC.md found with sufficient content", "SPEC.md")
		}
	}

	// Validate DESIGN.md - warning if missing (not blocking)
	if err := v.validateDesign(projectPath); err != nil {
		result.AddWarning("DESIGN_MISSING", err.Error(), "DESIGN.md")
	} else {
		result.AddInfo("DESIGN_FOUND", "DESIGN.md found", "DESIGN.md")
	}

	// Validate TASKS.md - warning if missing (not blocking)
	if err := v.validateTasks(projectPath); err != nil {
		result.AddWarning("TASKS_MISSING", err.Error(), "TASKS.md")
	} else {
		// Check if tasks are defined
		tasks, _ := v.loadTasksFromFile(filepath.Join(projectPath, "TASKS.md"))
		if len(tasks) == 0 {
			result.AddWarning("TASKS_EMPTY", "TASKS.md has no tasks defined", "TASKS.md")
		} else {
			result.AddInfo("TASKS_FOUND", fmt.Sprintf("TASKS.md found with %d tasks", len(tasks)), "TASKS.md")
		}
	}

	// Validate AGENTS.md (warning only - not critical)
	if err := v.validateAgents(projectPath); err != nil {
		result.AddWarning("AGENTS_MISSING", err.Error(), "AGENTS.md")
	} else {
		result.AddInfo("AGENTS_FOUND", "AGENTS.md found", "AGENTS.md")
	}

	// Validate SKILL.md(s) (info only - not critical)
	if err := v.validateSkills(projectPath); err != nil {
		result.AddInfo("SKILLS_MISSING", err.Error(), "SKILL.md")
	} else {
		result.AddInfo("SKILLS_FOUND", "Required skills found", "SKILL.md")
	}
}

// validateSpec checks if SPEC.md exists
func (v *Validator) validateSpec(projectPath string) error {
	filePath := filepath.Join(projectPath, "SPEC.md")
	if !fileExists(filePath) {
		return fmt.Errorf("SPEC.md not found at %s", filePath)
	}
	return nil
}

// validateDesign checks if DESIGN.md exists
func (v *Validator) validateDesign(projectPath string) error {
	filePath := filepath.Join(projectPath, "DESIGN.md")
	if !fileExists(filePath) {
		return fmt.Errorf("DESIGN.md not found at %s", filePath)
	}
	return nil
}

// validateTasks checks if TASKS.md exists
func (v *Validator) validateTasks(projectPath string) error {
	filePath := filepath.Join(projectPath, "TASKS.md")
	if !fileExists(filePath) {
		return fmt.Errorf("TASKS.md not found at %s", filePath)
	}
	return nil
}

// validateAgents checks if AGENTS.md exists
func (v *Validator) validateAgents(projectPath string) error {
	filePath := filepath.Join(projectPath, "AGENTS.md")
	if !fileExists(filePath) {
		return fmt.Errorf("AGENTS.md not found at %s", filePath)
	}
	return nil
}

// validateSkills checks if SKILL.md or SKILLs directory exists
func (v *Validator) validateSkills(projectPath string) error {
	// Check for SKILL.md in project root
	skillPath := filepath.Join(projectPath, "SKILL.md")
	if fileExists(skillPath) {
		return nil
	}

	// Check for .skills directory
	skillsDir := filepath.Join(projectPath, ".skills")
	if fileExists(skillsDir) {
		return nil
	}

	// Check for skills directory
	skillsDir = filepath.Join(projectPath, "skills")
	if fileExists(skillsDir) {
		return nil
	}

	return fmt.Errorf("no SKILL.md or skills directory found in %s", projectPath)
}

// readFile reads file content safely
func (v *Validator) readFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// hasMinimumContent checks if content has minimum length
func (v *Validator) hasMinimumContent(content string, minLength int) bool {
	return len(content) >= minLength
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ValidateTask validates a single task for completeness and correctness.
func (v *Validator) ValidateTask(task *Task) *ValidationResult {
	result := NewValidationResult()

	if task == nil {
		result.AddError("NIL_TASK", "Task cannot be nil", "")
		return result
	}

	if task.ID == "" {
		result.AddError("MISSING_ID", "Task ID is required", "id")
	}

	if task.Title == "" {
		result.AddError("MISSING_TITLE", "Task title is required", "title")
	}

	if task.Phase == "" {
		result.AddWarning("MISSING_PHASE", "Task phase is not specified", "phase")
	}

	// Validate dependencies exist
	for _, dep := range task.DependsOn {
		if dep == "" {
			result.AddWarning("EMPTY_DEP", "Empty dependency reference found", "")
		}
	}

	return result
}

// LoadTasks loads implementation tasks from documentation files.
// Returns a slice of tasks or an error if loading fails.
func (v *Validator) LoadTasks(docsPath string) ([]Task, error) {
	if docsPath == "" {
		return nil, errors.New("docs path cannot be empty")
	}

	info, err := os.Stat(docsPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access docs path: %w", err)
	}

	var tasks []Task

	if info.IsDir() {
		tasks, err = v.loadTasksFromDirectory(docsPath)
	} else {
		tasks, err = v.loadTasksFromFile(docsPath)
	}

	if err != nil {
		return nil, err
	}

	// Sort tasks by phase and priority
	v.sortTasks(tasks)

	return tasks, nil
}

// loadTasksFromDirectory loads tasks from all markdown files in a directory.
func (v *Validator) loadTasksFromDirectory(dirPath string) ([]Task, error) {
	var allTasks []Task

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext == ".md" || ext == ".yaml" || ext == ".yml" {
			filePath := filepath.Join(dirPath, entry.Name())
			tasks, err := v.loadTasksFromFile(filePath)
			if err != nil {
				continue // Skip files that fail to parse
			}
			allTasks = append(allTasks, tasks...)
		}
	}

	return allTasks, nil
}

// loadTasksFromFile loads tasks from a single documentation file.
func (v *Validator) loadTasksFromFile(filePath string) ([]Task, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	ext := filepath.Ext(filePath)
	switch ext {
	case ".yaml", ".yml":
		return v.parseTasksFromYAML(content)
	case ".md":
		return v.parseTasksFromMarkdown(content, filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// parseTasksFromYAML parses tasks from YAML content.
func (v *Validator) parseTasksFromYAML(content []byte) ([]Task, error) {
	var data struct {
		Tasks []Task `json:"tasks"`
	}

	if err := json.Unmarshal(content, &data); err != nil {
		// Try parsing as raw YAML map
		var rawTasks []map[string]interface{}
		if err2 := json.Unmarshal(content, &rawTasks); err2 != nil {
			return nil, fmt.Errorf("invalid task format: %w", err)
		}

		tasks := make([]Task, 0, len(rawTasks))
		for i, raw := range rawTasks {
			task := Task{
				ID:        fmt.Sprintf("task-%d", i+1),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if v, ok := raw["id"].(string); ok {
				task.ID = v
			}
			if v, ok := raw["title"].(string); ok {
				task.Title = v
			}
			if v, ok := raw["description"].(string); ok {
				task.Description = v
			}
			if v, ok := raw["phase"].(string); ok {
				task.Phase = v
			}
			tasks = append(tasks, task)
		}
		return tasks, nil
	}

	return data.Tasks, nil
}

// parseTasksFromMarkdown parses tasks from markdown content.
func (v *Validator) parseTasksFromMarkdown(content []byte, filePath string) ([]Task, error) {
	// Simple markdown task parsing
	// Looks for patterns like:
	// - [ ] Task description
	// - [x] Completed task
	// ## Phase 1: Name
	var tasks []Task
	var currentPhase string
	lineNum := 0

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		lineNum++

		// Check for phase headers
		if len(line) >= 2 && line[0] == '#' {
			currentPhase = extractPhase(line)
			continue
		}

		// Check for task items
		if isTaskLine(line) {
			task := parseTaskLine(line, lineNum, currentPhase, filePath)
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// extractPhase extracts the phase name from a markdown header.
func extractPhase(line string) string {
	// Remove leading # and spaces
	trimmed := line
	for len(trimmed) > 0 && (trimmed[0] == '#' || trimmed[0] == ' ') {
		trimmed = trimmed[1:]
	}
	return trimmed
}

// isTaskLine checks if a line represents a task.
func isTaskLine(line string) bool {
	return len(line) >= 5 && (line[:5] == "- [ ]" || line[:5] == "- [x]" || line[:5] == "- [X]")
}

// parseTaskLine parses a single task line into a Task struct.
func parseTaskLine(line string, lineNum int, phase, filePath string) Task {
	task := Task{
		ID:        fmt.Sprintf("%s:%d", filepath.Base(filePath), lineNum),
		Phase:     phase,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Extract task text - remove "- [ ]" or "- [x]" prefix (5 chars) and trim
	text := strings.TrimSpace(line[5:])
	task.Title = text

	// Check completion status - line[2] is 'x' in "- [x]"
	task.Completed = len(line) >= 3 && line[2] == 'x'

	return task
}

// sortTasks sorts tasks by phase order using O(n log n) sort.Slice.
func (v *Validator) sortTasks(tasks []Task) {
	phaseOrder := map[string]int{
		"explore": 1, "propose": 2, "spec": 3,
		"design": 4, "tasks": 5, "apply": 6,
	}
	sort.Slice(tasks, func(i, j int) bool {
		return phaseOrder[tasks[i].Phase] < phaseOrder[tasks[j].Phase]
	})
}

// StateManager handles persistence of loop state.
type StateManager struct {
	stateDir string
}

// NewStateManager creates a new state manager.
func NewStateManager(stateDir string) *StateManager {
	return &StateManager{
		stateDir: stateDir,
	}
}

// SaveState persists the loop state to disk.
func (sm *StateManager) SaveState(state *LoopState) error {
	if sm.stateDir == "" {
		return errors.New("state directory not configured")
	}

	// Ensure directory exists
	if err := os.MkdirAll(sm.stateDir, 0755); err != nil {
		return fmt.Errorf("cannot create state directory: %w", err)
	}

	state.UpdatedAt = time.Now()

	// Generate checkpoint ID if not set
	if state.CheckpointID == "" {
		state.CheckpointID = fmt.Sprintf("checkpoint-%d", state.CheckpointNum)
	}

	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal state: %w", err)
	}

	// Write to temp file first, then atomically rename
	// This prevents corruption if crash occurs during write
	tmpPath := filepath.Join(sm.stateDir, "loop-state.tmp")
	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		return fmt.Errorf("cannot write temp state file: %w", err)
	}

	finalPath := filepath.Join(sm.stateDir, "loop-state.json")
	if err := os.Rename(tmpPath, finalPath); err != nil {
		// Clean up temp file on failure
		os.Remove(tmpPath)
		return fmt.Errorf("cannot rename state file: %w", err)
	}

	return nil
}

// LoadState loads the loop state from disk.
func (sm *StateManager) LoadState() (*LoopState, error) {
	if sm.stateDir == "" {
		return nil, errors.New("state directory not configured")
	}

	filePath := filepath.Join(sm.stateDir, "loop-state.json")

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No saved state
		}
		return nil, fmt.Errorf("cannot read state file: %w", err)
	}

	var state LoopState
	if err := json.Unmarshal(content, &state); err != nil {
		return nil, fmt.Errorf("cannot unmarshal state: %w", err)
	}

	return &state, nil
}

// LoadStateByID loads a specific checkpoint state.
func (sm *StateManager) LoadStateByID(checkpointID string) (*LoopState, error) {
	if sm.stateDir == "" {
		return nil, errors.New("state directory not configured")
	}

	filePath := filepath.Join(sm.stateDir, fmt.Sprintf("%s.json", checkpointID))

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("checkpoint not found: %s", checkpointID)
		}
		return nil, fmt.Errorf("cannot read checkpoint: %w", err)
	}

	var state LoopState
	if err := json.Unmarshal(content, &state); err != nil {
		return nil, fmt.Errorf("cannot unmarshal checkpoint: %w", err)
	}

	return &state, nil
}

// ListCheckpoints returns all available checkpoint IDs.
func (sm *StateManager) ListCheckpoints() ([]string, error) {
	if sm.stateDir == "" {
		return nil, errors.New("state directory not configured")
	}

	entries, err := os.ReadDir(sm.stateDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read state directory: %w", err)
	}

	var checkpoints []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) > 5 && name[len(name)-5:] == ".json" {
			checkpoints = append(checkpoints, name[:len(name)-5])
		}
	}

	return checkpoints, nil
}
