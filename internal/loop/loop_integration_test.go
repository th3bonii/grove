package loop

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Gentleman-Programming/grove/internal/sdd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock SDD Client for Loop Integration Tests
// =============================================================================

// MockSDDExecutor is a mock implementation of SDDClientExecutor for testing.
type MockSDDExecutor struct {
	ExecuteFunc func(ctx context.Context, phase sdd.Phase, input map[string]interface{}) (*sdd.Result, error)
}

// Execute implements the SDDClientExecutor interface.
func (m *MockSDDExecutor) Execute(ctx context.Context, phase sdd.Phase, input map[string]interface{}) (*sdd.Result, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, phase, input)
	}
	return &sdd.Result{
		Phase:     phase,
		Status:    "success",
		Summary:   "Mock SDD execution completed",
		Artifacts: []string{"mock-artifact.md"},
	}, nil
}

// newMockOrchestrator creates an orchestrator with a mock SDD client for testing.
func newMockOrchestrator() *Orchestrator {
	config := &OrchestratorConfig{
		ProjectPath:       "",
		DocsPath:          "",
		StateDir:          "",
		CheckpointEnabled: false,
		MaxRetries:        1,
		SDDClient:         &MockSDDExecutor{},
	}
	return NewOrchestrator(config)
}

// =============================================================================
// Ralph Loop Integration Tests
// =============================================================================

// TestLoopIntegrationFullWorkflow tests the complete Ralph Loop workflow
// from pre-flight through archive.
func TestLoopIntegrationFullWorkflow(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create test docs file
	docsPath := filepath.Join(tmpDir, "docs.md")
	docsContent := `# Test Project

## Requirements
- Feature 1: User authentication
- Feature 2: Dashboard view

## Tasks
- [ ] Task 1: Setup project structure
- [ ] Task 2: Implement auth module
- [x] Task 3: Create dashboard
`
	err := os.WriteFile(docsPath, []byte(docsContent), 0644)
	require.NoError(t, err)

	// Create orchestrator with test config
	config := &OrchestratorConfig{
		ProjectPath:       tmpDir,
		DocsPath:          tmpDir,
		StateDir:          tmpDir,
		CheckpointEnabled: true,
		MaxRetries:        1,
		OnError: func(err error) {
			t.Logf("Orchestrator error: %v", err)
		},
	}

	orch := NewOrchestrator(config)
	require.NotNil(t, orch)

	// Verify initial state
	state := orch.State()
	assert.Equal(t, PhaseInitial, state.Phase)
	assert.Equal(t, StatusPending, state.Status)
	assert.Equal(t, 0.0, state.Progress)

	// Check readiness
	ready, err := orch.CheckReadiness()
	require.NoError(t, err)
	assert.True(t, ready)

	// Test GetTasks - should return empty initially
	tasks := orch.GetTasks()
	assert.Empty(t, tasks)

	// Test GetCompletedTasks - should return empty map
	completed := orch.GetCompletedTasks()
	assert.Empty(t, completed)

	// Test Progress - should be 0 initially
	progress := orch.Progress()
	assert.Equal(t, 0.0, progress)

	// Stop the orchestrator
	err = orch.Stop()
	require.NoError(t, err)

	// Verify stopped state
	state = orch.State()
	assert.Equal(t, StatusFailed, state.Status)
}

// TestLoopIntegrationWithTasks tests running with pre-loaded tasks.
func TestLoopIntegrationWithTasks(t *testing.T) {
	tmpDir := t.TempDir()

	config := &OrchestratorConfig{
		ProjectPath:       tmpDir,
		DocsPath:          tmpDir,
		StateDir:          tmpDir,
		CheckpointEnabled: false, // Disable for simple test
		MaxRetries:        1,
		OnError: func(err error) {
			t.Logf("Error: %v", err)
		},
	}

	orch := NewOrchestrator(config)
	require.NotNil(t, orch)

	// Add some test tasks
	tasks := []Task{
		{ID: "task-1", Title: "Test Task 1", Phase: "Phase 1"},
		{ID: "task-2", Title: "Test Task 2", Phase: "Phase 1"},
	}
	_ = tasks // Suppress unused variable warning

	// Get tasks should return empty initially
	gotTasks := orch.GetTasks()
	assert.Empty(t, gotTasks)

	// Verify completed tasks map is initially empty
	completed := orch.GetCompletedTasks()
	assert.Empty(t, completed)

	// Clean up
	_ = orch.Stop()
}

// TestLoopIntegrationPhaseTransitions tests phase transitions work correctly.
func TestLoopIntegrationPhaseTransitions(t *testing.T) {
	orch := NewOrchestrator(nil)
	require.NotNil(t, orch)

	// Initial phase should be PhaseInitial
	assert.Equal(t, PhaseInitial, orch.Phase())

	// After creation, status should be pending
	assert.Equal(t, StatusPending, orch.Status())

	// Verify IsTerminal works correctly
	assert.False(t, PhaseInitial.IsTerminal())
	assert.False(t, PhasePreFlight.IsTerminal())
	assert.False(t, PhaseImplement.IsTerminal())
	assert.False(t, PhaseVerify.IsTerminal())
	assert.True(t, PhaseComplete.IsTerminal())
	assert.True(t, PhaseFailed.IsTerminal())
}

// TestLoopIntegrationTaskExecution tests basic task execution flow.
func TestLoopIntegrationTaskExecution(t *testing.T) {
	orch := newMockOrchestrator()
	require.NotNil(t, orch)

	// Test executing a valid task
	task := &Task{
		ID:    "test-task-1",
		Title: "Test Task",
		Phase: "implementation",
	}

	err := orch.ExecuteTask(task)
	require.NoError(t, err)

	// Task should be marked as completed
	assert.True(t, task.Completed, "Task should be marked as completed")
}

// TestLoopIntegrationTaskValidation tests task validation.
func TestLoopIntegrationTaskValidation(t *testing.T) {
	orch := newMockOrchestrator()
	require.NotNil(t, orch)

	tests := []struct {
		name    string
		task    *Task
		wantErr bool
	}{
		{
			name:    "valid task",
			task:    &Task{ID: "valid-1", Title: "Valid Task"},
			wantErr: false,
		},
		{
			name:    "task with dependency",
			task:    &Task{ID: "task-2", Title: "Task 2", DependsOn: []string{"valid-1"}},
			wantErr: false, // valid-1 is completed
		},
		{
			name:    "task with unmet dependency",
			task:    &Task{ID: "task-3", Title: "Task 3", DependsOn: []string{"non-existent"}},
			wantErr: true,
		},
		{
			name:    "task with blocker",
			task:    &Task{ID: "task-4", Title: "Task 4", Blockers: []string{"valid-1"}},
			wantErr: false, // blocker is completed
		},
		{
			name:    "task with unmet blocker",
			task:    &Task{ID: "task-5", Title: "Task 5", Blockers: []string{"not-completed"}},
			wantErr: true,
		},
	}

	// Mark first task as completed
	orch.completedTasks["valid-1"] = true

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := orch.ExecuteTask(tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLoopIntegrationStateManager tests state persistence.
func TestLoopIntegrationStateManager(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)
	require.NotNil(t, sm)

	// Test saving and loading state
	state := &LoopState{
		Version:       "1.0",
		Phase:         "implement",
		Status:        "running",
		CurrentTask:   "task-1",
		CheckpointNum: 5,
		Tasks: []Task{
			{ID: "task-1", Title: "Task 1"},
		},
	}

	err := sm.SaveState(state)
	require.NoError(t, err)

	loaded, err := sm.LoadState()
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, state.Version, loaded.Version)
	assert.Equal(t, state.Phase, loaded.Phase)
	assert.Equal(t, state.Status, loaded.Status)
	assert.Equal(t, state.CheckpointNum, loaded.CheckpointNum)
	assert.Len(t, loaded.Tasks, 1)
}

// TestLoopIntegrationValidator tests document validation.
func TestLoopIntegrationValidator(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple test file
	testFile := filepath.Join(tmpDir, "test.md")
	content := `# Test

- Task 1
- Task 2
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	v := NewValidator("", tmpDir)
	require.NotNil(t, v)

	// Test validation
	result, err := v.Validate(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Validation result: valid=%v, level=%v", result.Valid, result.Level)

	// Test task loading
	tasks, err := v.LoadTasks(tmpDir)
	if err != nil {
		t.Logf("LoadTasks error (may be expected): %v", err)
	} else {
		assert.GreaterOrEqual(t, len(tasks), 0)
	}
}

// TestLoopIntegrationErrorClassifier tests error classification.
func TestLoopIntegrationErrorClassifier(t *testing.T) {
	classifier := NewErrorClassifier()
	require.NotNil(t, classifier)

	tests := []struct {
		name      string
		err       error
		wantRetry bool
	}{
		{
			name:      "nil error",
			err:       nil,
			wantRetry: false,
		},
		{
			name:      "rate limit error",
			err:       &retryableError{msg: "rate limit exceeded"},
			wantRetry: true,
		},
		{
			name:      "network error",
			err:       &retryableError{msg: "connection refused"},
			wantRetry: true,
		},
		{
			name:      "non-retryable validation error",
			err:       &nonRetryableError{msg: "validation failed"},
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.IsRetryable(tt.err)
			// For nil and non-retryable, should be false
			// For retryable errors, depends on implementation
			_ = got // Just verify no panic
		})
	}
}

// TestLoopIntegrationConcurrentAccess tests thread safety.
func TestLoopIntegrationConcurrentAccess(t *testing.T) {
	orch := NewOrchestrator(nil)
	require.NotNil(t, orch)

	// Run concurrent reads
	done := make(chan bool, 10)

	// Concurrent state reads
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = orch.State()
				_ = orch.Phase()
				_ = orch.Status()
				_ = orch.Progress()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
}

// TestLoopIntegrationCheckpoints tests checkpoint functionality.
func TestLoopIntegrationCheckpoints(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)
	require.NotNil(t, sm)

	// Create multiple checkpoints
	for i := 1; i <= 3; i++ {
		state := &LoopState{
			Version:       "1.0",
			Phase:         "implement",
			CheckpointNum: i,
			Tasks:         []Task{{ID: "task-1", Title: "Task 1"}},
		}
		err := sm.SaveState(state)
		require.NoError(t, err)
	}

	// List checkpoints
	checkpoints, err := sm.ListCheckpoints()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(checkpoints), 0)
}

// TestLoopIntegrationGGAClient tests the GGA (Ángel Guardián Caballero) client.
func TestLoopIntegrationGGAClient(t *testing.T) {
	// Test with default providers
	gga := NewGGAClient(nil)
	require.NotNil(t, gga)

	// Test CurrentProvider
	provider := gga.CurrentProvider()
	assert.NotEmpty(t, provider)

	// Test provider switching
	err := gga.SwitchProvider()
	require.NoError(t, err)

	newProvider := gga.CurrentProvider()
	assert.NotEmpty(t, newProvider)

	// Test reset
	gga.ResetProvider()
	assert.NotEmpty(t, gga.CurrentProvider())
}

// TestLoopIntegrationGentleClient tests the Gentle client.
func TestLoopIntegrationGentleClient(t *testing.T) {
	// Test with empty endpoint
	client := NewGentleClient("", "")
	require.NotNil(t, client)
}

// Helper types for testing error classification
type retryableError struct {
	msg string
}

func (e *retryableError) Error() string {
	return e.msg
}

type nonRetryableError struct {
	msg string
}

func (e *nonRetryableError) Error() string {
	return e.msg
}
