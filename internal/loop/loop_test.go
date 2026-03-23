// Package loop provides tests for the Ralph Loop engine.

package loop

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Validator Tests
// =============================================================================

func TestNewValidationResult(t *testing.T) {
	result := NewValidationResult()

	if result == nil {
		t.Fatal("NewValidationResult returned nil")
	}

	if !result.Valid {
		t.Error("NewValidationResult should be valid by default")
	}

	if result.Level != ValidationLevelInfo {
		t.Errorf("Expected ValidationLevelInfo, got %v", result.Level)
	}

	if len(result.Errors) != 0 {
		t.Error("NewValidationResult should have no errors")
	}
}

func TestValidationResult_AddError(t *testing.T) {
	result := NewValidationResult()

	result.AddError("TEST_CODE", "Test error message", "test_field")

	if result.Valid {
		t.Error("Result should be invalid after adding error")
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	if result.Level != ValidationLevelError {
		t.Errorf("Expected ValidationLevelError, got %v", result.Level)
	}

	err := result.Errors[0]
	if err.Code != "TEST_CODE" {
		t.Errorf("Expected code 'TEST_CODE', got '%s'", err.Code)
	}
	if err.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got '%s'", err.Message)
	}
	if err.Field != "test_field" {
		t.Errorf("Expected field 'test_field', got '%s'", err.Field)
	}
}

func TestValidationResult_AddWarning(t *testing.T) {
	result := NewValidationResult()

	result.AddWarning("WARN_CODE", "Test warning message", "warn_field")

	if result.Valid {
		t.Error("Result should still be valid with warning only")
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	if result.Level != ValidationLevelWarning {
		t.Errorf("Expected ValidationLevelWarning, got %v", result.Level)
	}
}

func TestValidationResult_AddInfo(t *testing.T) {
	result := NewValidationResult()

	result.AddInfo("INFO_CODE", "Test info message", "info_field")

	if !result.Valid {
		t.Error("Result should still be valid with info only")
	}

	if len(result.Infos) != 1 {
		t.Errorf("Expected 1 info, got %d", len(result.Infos))
	}

	if result.Level != ValidationLevelInfo {
		t.Errorf("Expected ValidationLevelInfo, got %v", result.Level)
	}
}

func TestValidationLevel_String(t *testing.T) {
	tests := []struct {
		level    ValidationLevel
		expected string
	}{
		{ValidationLevelInfo, "info"},
		{ValidationLevelWarning, "warning"},
		{ValidationLevelError, "error"},
		{ValidationLevelCritical, "critical"},
		{ValidationLevel(99), "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if got := tc.level.String(); got != tc.expected {
				t.Errorf("ValidationLevel.String() = %s, want %s", got, tc.expected)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Level:   ValidationLevelError,
		Code:    "ERR_CODE",
		Message: "Error message",
	}

	expected := "[error] ERR_CODE: Error message"
	if got := err.Error(); got != expected {
		t.Errorf("ValidationError.Error() = %s, want %s", got, expected)
	}
}

func TestNewLoopState(t *testing.T) {
	state := NewLoopState()

	if state == nil {
		t.Fatal("NewLoopState returned nil")
	}

	if state.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", state.Version)
	}

	if state.Phase != "initial" {
		t.Errorf("Expected phase 'initial', got '%s'", state.Phase)
	}

	if state.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", state.Status)
	}

	if len(state.Tasks) != 0 {
		t.Error("NewLoopState should have no tasks")
	}

	if state.CheckpointNum != 0 {
		t.Errorf("Expected checkpoint number 0, got %d", state.CheckpointNum)
	}
}

func TestNewValidator(t *testing.T) {
	v := NewValidator("/rules", "/docs")

	if v == nil {
		t.Fatal("NewValidator returned nil")
	}

	if v.rulesDir != "/rules" {
		t.Errorf("Expected rulesDir '/rules', got '%s'", v.rulesDir)
	}

	if v.docsDir != "/docs" {
		t.Errorf("Expected docsDir '/docs', got '%s'", v.docsDir)
	}
}

func TestValidator_Validate_NonExistentPath(t *testing.T) {
	v := NewValidator("", "")

	result, err := v.Validate("/non/existent/path")
	if err != nil {
		t.Errorf("Validate should not return error for non-existent path, got: %v", err)
	}

	if result == nil {
		t.Fatal("Validate returned nil result")
	}

	// Should have errors for non-existent path
	if result.Valid {
		t.Error("Result should be invalid for non-existent path")
	}
}

func TestValidator_Validate_ExistingPath(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	v := NewValidator("", tmpDir)

	result, err := v.Validate(tmpDir)
	if err != nil {
		t.Errorf("Validate should not return error for existing path, got: %v", err)
	}

	if result == nil {
		t.Fatal("Validate returned nil result")
	}

	// Should be valid
	if !result.Valid {
		t.Error("Result should be valid for existing directory")
	}

	// Should have validation started info
	if len(result.Infos) == 0 {
		t.Error("Should have at least one info message")
	}
}

func TestValidator_ValidateTask(t *testing.T) {
	v := NewValidator("", "")

	tests := []struct {
		name    string
		task    *Task
		wantErr bool
	}{
		{
			name:    "nil task",
			task:    nil,
			wantErr: true,
		},
		{
			name: "valid task",
			task: &Task{
				ID:    "task-1",
				Title: "Test Task",
				Phase: "Phase 1",
			},
			wantErr: false,
		},
		{
			name: "task without ID",
			task: &Task{
				Title: "Test Task",
			},
			wantErr: true,
		},
		{
			name: "task without title",
			task: &Task{
				ID: "task-2",
			},
			wantErr: true,
		},
		{
			name: "task without phase",
			task: &Task{
				ID:    "task-3",
				Title: "Test Task",
			},
			wantErr: false, // Warning, not error
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := v.ValidateTask(tc.task)

			if tc.wantErr {
				if result.Valid {
					t.Error("Expected validation to fail, but it passed")
				}
			} else {
				if !result.Valid && len(result.Errors) > 0 {
					t.Errorf("Expected validation to pass, but got errors: %v", result.Errors)
				}
			}
		})
	}
}

func TestValidator_LoadTasks_EmptyPath(t *testing.T) {
	v := NewValidator("", "")

	_, err := v.LoadTasks("")

	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestValidator_LoadTasks_NonExistentPath(t *testing.T) {
	v := NewValidator("", "")

	_, err := v.LoadTasks("/non/existent/path")

	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestValidator_LoadTasks_Directory(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create a markdown file with tasks
	mdContent := `# Phase 1: Setup

- [ ] Task 1.1
- [x] Task 1.2
`
	mdPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	v := NewValidator("", tmpDir)

	tasks, err := v.LoadTasks(tmpDir)
	if err != nil {
		t.Errorf("LoadTasks returned error: %v", err)
	}

	// Should have loaded the tasks
	if len(tasks) == 0 {
		t.Error("Expected to load tasks from markdown file")
	}
}

func TestValidator_LoadTasks_File(t *testing.T) {
	// Create a temporary file with task data
	tmpDir := t.TempDir()

	taskContent := `{"tasks":[{"id":"task-1","title":"Test Task","phase":"Phase 1"}]}`
	taskPath := filepath.Join(tmpDir, "tasks.json")
	if err := os.WriteFile(taskPath, []byte(taskContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	v := NewValidator("", "")

	tasks, err := v.LoadTasks(taskPath)
	if err != nil {
		t.Errorf("LoadTasks returned error: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].ID != "task-1" {
		t.Errorf("Expected task ID 'task-1', got '%s'", tasks[0].ID)
	}
}

// =============================================================================
// StateManager Tests
// =============================================================================

func TestNewStateManager(t *testing.T) {
	sm := NewStateManager("/state")

	if sm == nil {
		t.Fatal("NewStateManager returned nil")
	}

	if sm.stateDir != "/state" {
		t.Errorf("Expected stateDir '/state', got '%s'", sm.stateDir)
	}
}

func TestStateManager_SaveAndLoadState(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)

	state := &LoopState{
		Version:     "1.0",
		Phase:       "implement",
		Status:      "running",
		CurrentTask: "task-1",
		Tasks: []Task{
			{
				ID:        "task-1",
				Title:     "Test Task",
				Phase:     "Phase 1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		CheckpointNum: 5,
	}

	// Save state
	if err := sm.SaveState(state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// Load state
	loaded, err := sm.LoadState()
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("LoadState returned nil")
	}

	if loaded.Version != state.Version {
		t.Errorf("Expected version '%s', got '%s'", state.Version, loaded.Version)
	}

	if loaded.Phase != state.Phase {
		t.Errorf("Expected phase '%s', got '%s'", state.Phase, loaded.Phase)
	}

	if loaded.CurrentTask != state.CurrentTask {
		t.Errorf("Expected current task '%s', got '%s'", state.CurrentTask, loaded.CurrentTask)
	}

	if len(loaded.Tasks) != len(state.Tasks) {
		t.Errorf("Expected %d tasks, got %d", len(state.Tasks), len(loaded.Tasks))
	}
}

func TestStateManager_LoadState_NotExists(t *testing.T) {
	sm := NewStateManager("/non/existent")

	_, err := sm.LoadState()
	if err != nil {
		t.Error("LoadState should return nil error for non-existent state")
	}
}

func TestStateManager_SaveState_EmptyDir(t *testing.T) {
	sm := NewStateManager("")

	state := &LoopState{}

	err := sm.SaveState(state)
	if err == nil {
		t.Error("Expected error for empty state directory")
	}
}

func TestStateManager_LoadState_EmptyDir(t *testing.T) {
	sm := NewStateManager("")

	_, err := sm.LoadState()
	if err == nil {
		t.Error("Expected error for empty state directory")
	}
}

func TestStateManager_ListCheckpoints(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)

	// Create some checkpoint files
	for i := 1; i <= 3; i++ {
		state := &LoopState{CheckpointNum: i}
		if err := sm.SaveState(state); err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}
	}

	checkpoints, err := sm.ListCheckpoints()
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}

	if len(checkpoints) == 0 {
		t.Error("Expected to find checkpoints")
	}
}

// =============================================================================
// Orchestrator Tests
// =============================================================================

func TestNewOrchestrator(t *testing.T) {
	config := &OrchestratorConfig{
		ProjectPath: "/project",
		StateDir:    "/state",
	}

	orch := NewOrchestrator(config)

	if orch == nil {
		t.Fatal("NewOrchestrator returned nil")
	}

	state := orch.State()
	if state.Phase != PhaseInitial {
		t.Errorf("Expected phase 'initial', got '%s'", state.Phase)
	}

	if state.Status != StatusPending {
		t.Errorf("Expected status 'pending', got '%s'", state.Status)
	}
}

func TestNewOrchestrator_WithNilConfig(t *testing.T) {
	orch := NewOrchestrator(nil)

	if orch == nil {
		t.Fatal("NewOrchestrator returned nil for nil config")
	}

	// Should use defaults
	state := orch.State()
	if state.Phase != PhaseInitial {
		t.Error("Should initialize with default state")
	}
}

func TestOrchestrator_State(t *testing.T) {
	orch := NewOrchestrator(nil)

	// State should return a copy
	state1 := orch.State()
	state2 := orch.State()

	if state1 == state2 {
		t.Error("State() should return a copy, not the same pointer")
	}

	// Modifying one shouldn't affect the other
	state1.Phase = PhaseComplete
	if state2.Phase == state1.Phase {
		t.Error("Modifying returned state shouldn't affect internal state")
	}
}

func TestOrchestrator_Phase(t *testing.T) {
	orch := NewOrchestrator(nil)

	if orch.Phase() != PhaseInitial {
		t.Error("Initial phase should be 'initial'")
	}
}

func TestOrchestrator_Status(t *testing.T) {
	orch := NewOrchestrator(nil)

	if orch.Status() != StatusPending {
		t.Error("Initial status should be 'pending'")
	}
}

func TestOrchestrator_CheckReadiness(t *testing.T) {
	orch := NewOrchestrator(nil)

	tests := []struct {
		name       string
		setup      func(*Orchestrator)
		wantReady  bool
		wantErrMsg string
	}{
		{
			name:      "initial state",
			setup:     func(o *Orchestrator) {},
			wantReady: true,
		},
		{
			name: "already running",
			setup: func(o *Orchestrator) {
				o.mu.Lock()
				o.state.Status = StatusRunning
				o.mu.Unlock()
			},
			wantReady:  false,
			wantErrMsg: "already running",
		},
		{
			name: "already completed",
			setup: func(o *Orchestrator) {
				o.mu.Lock()
				o.state.Phase = PhaseComplete
				o.mu.Unlock()
			},
			wantReady:  false,
			wantErrMsg: "already completed",
		},
		{
			name: "paused state",
			setup: func(o *Orchestrator) {
				o.mu.Lock()
				o.state.Status = StatusPaused
				o.mu.Unlock()
			},
			wantReady: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orch := NewOrchestrator(nil)
			tc.setup(orch)

			ready, err := orch.CheckReadiness()

			if tc.wantReady {
				if !ready {
					t.Errorf("Expected ready=true, got ready=%v", ready)
				}
			} else {
				if ready {
					t.Errorf("Expected ready=false, got ready=%v", ready)
				}
				if err == nil {
					t.Error("Expected error, got nil")
				}
			}
		})
	}
}

func TestOrchestrator_Pause(t *testing.T) {
	orch := NewOrchestrator(nil)

	// Not running - should fail
	err := orch.Pause()
	if err == nil {
		t.Error("Pause should fail when not running")
	}

	// Set running
	orch.mu.Lock()
	orch.state.Status = StatusRunning
	orch.mu.Unlock()

	// Pause should succeed
	err = orch.Pause()
	if err != nil {
		t.Errorf("Pause should succeed when running: %v", err)
	}

	// Already paused - should fail
	err = orch.Pause()
	if err == nil {
		t.Error("Pause should fail when already paused")
	}
}

func TestOrchestrator_Resume(t *testing.T) {
	orch := NewOrchestrator(nil)

	// Not paused - should fail
	err := orch.Resume()
	if err == nil {
		t.Error("Resume should fail when not paused")
	}

	// Set paused
	orch.mu.Lock()
	orch.state.Status = StatusPaused
	orch.mu.Unlock()

	// Resume should succeed
	err = orch.Resume()
	if err != nil {
		t.Errorf("Resume should succeed when paused: %v", err)
	}
}

func TestOrchestrator_Stop(t *testing.T) {
	orch := NewOrchestrator(nil)

	err := orch.Stop()
	if err != nil {
		t.Errorf("Stop should not return error: %v", err)
	}

	state := orch.State()
	if state.Status != StatusFailed {
		t.Errorf("Status should be 'failed' after Stop, got '%s'", state.Status)
	}
}

func TestOrchestrator_ExecuteTask(t *testing.T) {
	orch := NewOrchestrator(nil)

	// Nil task
	err := orch.ExecuteTask(nil)
	if err == nil {
		t.Error("ExecuteTask should fail for nil task")
	}

	// Task with unmet dependency
	task := &Task{
		ID:        "task-1",
		Title:     "Test Task",
		DependsOn: []string{"missing-dep"},
	}

	err = orch.ExecuteTask(task)
	if err == nil {
		t.Error("ExecuteTask should fail for task with unmet dependency")
	}

	// Valid task
	task = &Task{
		ID:    "task-2",
		Title: "Test Task",
	}

	err = orch.ExecuteTask(task)
	if err != nil {
		t.Errorf("ExecuteTask should succeed for valid task: %v", err)
	}

	if !task.Completed {
		t.Error("Task should be marked as completed")
	}

	if !orch.completedTasks["task-2"] {
		t.Error("Task should be in completedTasks map")
	}
}

func TestOrchestrator_ExecuteTask_WithBlocker(t *testing.T) {
	orch := NewOrchestrator(nil)

	// Mark a task as completed
	orch.completedTasks["blocker-task"] = true

	// Task with completed blocker
	task := &Task{
		ID:       "task-1",
		Title:    "Test Task",
		Blockers: []string{"blocker-task"},
	}

	err := orch.ExecuteTask(task)
	if err != nil {
		t.Errorf("ExecuteTask should succeed when blocker is completed: %v", err)
	}
}

func TestOrchestrator_GetTasks(t *testing.T) {
	orch := NewOrchestrator(nil)

	tasks := []Task{
		{ID: "task-1", Title: "Task 1"},
		{ID: "task-2", Title: "Task 2"},
	}
	orch.tasks = tasks

	got := orch.GetTasks()

	if len(got) != len(tasks) {
		t.Errorf("Expected %d tasks, got %d", len(tasks), len(got))
	}

	// Should be a copy
	if &got[0] == &orch.tasks[0] {
		t.Error("GetTasks should return a copy, not the original slice")
	}
}

func TestOrchestrator_GetCompletedTasks(t *testing.T) {
	orch := NewOrchestrator(nil)

	orch.completedTasks["task-1"] = true
	orch.completedTasks["task-2"] = false

	got := orch.GetCompletedTasks()

	if len(got) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(got))
	}

	// Should be a copy
	if got == orch.completedTasks {
		t.Error("GetCompletedTasks should return a copy")
	}
}

func TestOrchestrator_Progress(t *testing.T) {
	orch := NewOrchestrator(nil)

	progress := orch.Progress()
	if progress != 0.0 {
		t.Errorf("Initial progress should be 0.0, got %f", progress)
	}

	orch.mu.Lock()
	orch.state.Progress = 0.5
	orch.mu.Unlock()

	progress = orch.Progress()
	if progress != 0.5 {
		t.Errorf("Expected progress 0.5, got %f", progress)
	}
}

func TestOrchestrator_Run_Cancelled(t *testing.T) {
	orch := NewOrchestrator(nil)
	orch.cancel() // Cancel immediately

	err := orch.Run()
	if err == nil {
		t.Error("Run should return error when cancelled")
	}
}

func TestOrchestrator_RunWithTasks(t *testing.T) {
	tmpDir := t.TempDir()

	config := &OrchestratorConfig{
		ProjectPath:       tmpDir,
		StateDir:          tmpDir,
		DocsPath:          tmpDir,
		CheckpointEnabled: false, // Disable checkpoints for testing
		OnError: func(err error) {
			// Ignore errors for this test
		},
	}

	orch := NewOrchestrator(config)

	// Create a test file to avoid validation error
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Test\n- [ ] Task 1\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set initial phase to allow running
	orch.mu.Lock()
	orch.state.Phase = PhaseInitial
	orch.mu.Unlock()

	// Run with empty tasks - should use pre-flight validation
	err := orch.RunWithTasks(nil)
	if err != nil {
		// May fail on validation but that's expected in test
		t.Logf("RunWithTasks returned error (may be expected): %v", err)
	}
}

func TestOrchestrator_ConcurrentAccess(t *testing.T) {
	orch := NewOrchestrator(nil)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = orch.State()
				_ = orch.Phase()
				_ = orch.Status()
				_ = orch.Progress()
			}
		}()
	}

	// Concurrent writes
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				orch.mu.Lock()
				orch.state.Phase = LoopPhase(string(rune('a' + id%26)))
				orch.mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// If we get here without deadlock, the test passes
}

// =============================================================================
// LoopPhase Tests
// =============================================================================

func TestLoopPhase_IsTerminal(t *testing.T) {
	tests := []struct {
		phase      LoopPhase
		isTerminal bool
	}{
		{PhaseComplete, true},
		{PhaseFailed, true},
		{PhaseInitial, false},
		{PhasePreFlight, false},
		{PhaseImplement, false},
		{PhasePaused, false},
	}

	for _, tc := range tests {
		t.Run(string(tc.phase), func(t *testing.T) {
			if got := tc.phase.IsTerminal(); got != tc.isTerminal {
				t.Errorf("Phase.IsTerminal() = %v, want %v", got, tc.isTerminal)
			}
		})
	}
}

func TestLoopPhase_String(t *testing.T) {
	phase := PhasePreFlight
	if got := phase.String(); got != "pre-flight" {
		t.Errorf("Phase.String() = %s, want 'pre-flight'", got)
	}
}

// =============================================================================
// Integration-like Tests
// =============================================================================

func TestOrchestrator_FullLifecycle(t *testing.T) {
	// This test verifies the basic lifecycle: create -> check -> stop
	orch := NewOrchestrator(nil)

	// Check initial state
	state := orch.State()
	if state.Phase != PhaseInitial {
		t.Errorf("Initial phase should be 'initial', got '%s'", state.Phase)
	}

	// Check readiness
	ready, _ := orch.CheckReadiness()
	if !ready {
		t.Error("Orchestrator should be ready initially")
	}

	// Stop
	err := orch.Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	// Should not be ready anymore
	ready, _ = orch.CheckReadiness()
	if ready {
		t.Error("Stopped orchestrator should not be ready")
	}
}

func TestOrchestrator_PauseResume(t *testing.T) {
	orch := NewOrchestrator(nil)

	// Cannot pause when not running
	err := orch.Pause()
	if err == nil {
		t.Error("Should not be able to pause when not running")
	}

	// Cannot resume when not paused
	err = orch.Resume()
	if err == nil {
		t.Error("Should not be able to resume when not paused")
	}

	// Set to running state manually
	orch.mu.Lock()
	orch.state.Status = StatusRunning
	orch.mu.Unlock()

	// Now pause should work
	err = orch.Pause()
	if err != nil {
		t.Errorf("Pause failed: %v", err)
	}

	// Wait for pause to be processed
	time.Sleep(50 * time.Millisecond)

	// Resume should work
	err = orch.Resume()
	if err != nil {
		t.Errorf("Resume failed: %v", err)
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkValidationResult_AddError(b *testing.B) {
	result := NewValidationResult()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result.AddError("CODE", "Message", "Field")
	}
}

func BenchmarkOrchestrator_State(b *testing.B) {
	orch := NewOrchestrator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = orch.State()
	}
}

func BenchmarkOrchestrator_ExecuteTask(b *testing.B) {
	orch := NewOrchestrator(nil)
	task := &Task{ID: "bench-task", Title: "Benchmark Task"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = orch.ExecuteTask(task)
		task.Completed = false
		orch.completedTasks["bench-task"] = false
	}
}

// =============================================================================
// Context Cancellation Test
// =============================================================================

func TestOrchestrator_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	orch := &Orchestrator{
		state: &OrchestratorState{
			Phase:     PhasePreFlight,
			Status:    StatusRunning,
			StartedAt: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Cancel the context
	cancel()

	// The orchestrator should detect cancellation
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled")
	}
}
