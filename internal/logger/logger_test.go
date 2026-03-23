package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// bufferHandler is a slog.Handler that writes to a buffer for testing.
type bufferHandler struct {
	*bytes.Buffer
}

func (h *bufferHandler) Handle(_ context.Context, r slog.Record) error {
	_, err := h.Buffer.WriteString(r.Message)
	return err
}

func (h *bufferHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *bufferHandler) WithAttrs(_ []slog.Attr) slog.Handler         { return h }
func (h *bufferHandler) WithGroup(_ string) slog.Handler              { return h }

func newTestLogger(buf *bytes.Buffer) *GroveLogger {
	return &GroveLogger{
		Logger: slog.New(&bufferHandler{buf}),
		cfg:    DefaultConfig(),
	}
}

// TestDefaultConfig tests that DefaultConfig returns sensible defaults.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != LevelInfo {
		t.Errorf("expected LevelInfo, got %v", cfg.Level)
	}

	if cfg.Output == nil {
		t.Error("expected non-nil Output")
	}

	if cfg.Pretty != true {
		t.Error("expected Pretty to be true by default")
	}
}

// TestNewLogger tests logger creation.
func TestNewLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := Config{
		Level:     LevelDebug,
		Output:    buf,
		AddSource: false,
		Pretty:    true,
	}

	logger := New(cfg)

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	if logger.cfg.Level != LevelDebug {
		t.Errorf("expected LevelDebug, got %v", logger.cfg.Level)
	}
}

// TestNewDefaultLogger tests the NewDefault function.
func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefault()

	if logger == nil {
		t.Fatal("expected non-nil logger from NewDefault")
	}
}

// TestWithComponent tests the WithComponent method.
func TestWithComponent(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	componentLogger := logger.WithComponent("grove-spec")

	if componentLogger == logger {
		t.Error("WithComponent should return a new logger instance")
	}
}

// TestWithChange tests the WithChange method.
func TestWithChange(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	changeLogger := logger.WithChange("feature-auth")

	if changeLogger == logger {
		t.Error("WithChange should return a new logger instance")
	}
}

// TestWithTask tests the WithTask method.
func TestWithTask(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	taskLogger := logger.WithTask("task-001")

	if taskLogger == logger {
		t.Error("WithTask should return a new logger instance")
	}
}

// TestWithLoop tests the WithLoop method.
func TestWithLoop(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	loopLogger := logger.WithLoop(3)

	if loopLogger == logger {
		t.Error("WithLoop should return a new logger instance")
	}
}

// TestWithSDDPhase tests the WithSDDPhase method.
func TestWithSDDPhase(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	sddLogger := logger.WithSDDPhase(SDDPhaseApply)

	if sddLogger == logger {
		t.Error("WithSDDPhase should return a new logger instance")
	}
}

// TestLogSpecOperation tests the LogSpecOperation method.
func TestLogSpecOperation(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogSpecOperation(
		ctx,
		SpecOpInit,
		PhaseParse,
		"feature-auth",
		map[string]any{"input_files": 5},
	)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "spec_operation") {
		t.Errorf("expected 'spec_operation' in output, got: %s", output)
	}
}

// TestLogSpecOperationStart tests LogSpecOperationStart.
func TestLogSpecOperationStart(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogSpecOperationStart(ctx, SpecOpUpdate, "feature-auth")

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "started") {
		t.Errorf("expected 'started' in output, got: %s", output)
	}
}

// TestLogSpecOperationEndSuccess tests LogSpecOperationEnd with success.
func TestLogSpecOperationEndSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogSpecOperationEnd(ctx, SpecOpComplete, "feature-auth", 2*time.Second, true)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "completed") {
		t.Errorf("expected 'completed' in output, got: %s", output)
	}

	if !strings.Contains(output, "true") {
		t.Errorf("expected 'true' success marker in output, got: %s", output)
	}
}

// TestLogSpecOperationEndFailure tests LogSpecOperationEnd with failure.
func TestLogSpecOperationEndFailure(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogSpecOperationEnd(ctx, SpecOpComplete, "feature-auth", 2*time.Second, false)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "failed") {
		t.Errorf("expected 'failed' in output, got: %s", output)
	}
}

// TestLogLoopIteration tests the LogLoopIteration method.
func TestLogLoopIteration(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogLoopIteration(
		ctx,
		1,
		LoopPhasePreLoop,
		"loop_started",
		map[string]any{"total_tasks": 10},
	)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "loop_iteration") {
		t.Errorf("expected 'loop_iteration' in output, got: %s", output)
	}
}

// TestLogLoopIterationStart tests LogLoopIterationStart.
func TestLogLoopIterationStart(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogLoopIterationStart(ctx, 1, 15)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "loop_started") {
		t.Errorf("expected 'loop_started' in output, got: %s", output)
	}

	if !strings.Contains(output, "15") {
		t.Errorf("expected '15' total_tasks in output, got: %s", output)
	}
}

// TestLogLoopIterationEnd tests LogLoopIterationEnd.
func TestLogLoopIterationEnd(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogLoopIterationEnd(ctx, 1, 5*time.Minute, 8, 2)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "loop_completed") {
		t.Errorf("expected 'loop_completed' in output, got: %s", output)
	}
}

// TestLogTaskExecution tests the LogTaskExecution method.
func TestLogTaskExecution(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogTaskExecution(
		ctx,
		"task-001",
		TaskStatusRunning,
		map[string]any{"description": "implement auth"},
	)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "task_execution") {
		t.Errorf("expected 'task_execution' in output, got: %s", output)
	}
}

// TestLogTaskExecutionStart tests LogTaskExecutionStart.
func TestLogTaskExecutionStart(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogTaskExecutionStart(ctx, "task-001", "feature-auth", "Implement login handler")

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "started") {
		t.Errorf("expected 'started' in output, got: %s", output)
	}
}

// TestLogTaskExecutionEndSuccess tests LogTaskExecutionEnd with success.
func TestLogTaskExecutionEndSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogTaskExecutionEnd(ctx, "task-001", TaskStatusSucceeded, 3*time.Second, 1, nil)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "ended") {
		t.Errorf("expected 'ended' in output, got: %s", output)
	}
}

// TestLogTaskExecutionEndWithError tests LogTaskExecutionEnd with an error.
func TestLogTaskExecutionEndWithError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogTaskExecutionEnd(
		ctx,
		"task-001",
		TaskStatusFailed,
		2*time.Second,
		2,
		&testError{msg: "compilation failed"},
	)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "compilation failed") {
		t.Errorf("expected error message in output, got: %s", output)
	}
}

// testError is a simple error implementation for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestLogSDDPhase tests the LogSDDPhase method.
func TestLogSDDPhase(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogSDDPhase(
		ctx,
		SDDPhaseApply,
		"feature-auth",
		"completed",
		map[string]any{"duration": "5s"},
	)

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "sdd_phase_transition") {
		t.Errorf("expected 'sdd_phase_transition' in output, got: %s", output)
	}
}

// TestLogVerificationPassed tests LogVerification with passed=true.
func TestLogVerificationPassed(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogVerification(ctx, "task-001", true, "all checks passed")

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "passed") {
		t.Errorf("expected 'passed' in output, got: %s", output)
	}
}

// TestLogVerificationFailed tests LogVerification with passed=false.
func TestLogVerificationFailed(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogVerification(ctx, "task-001", false, "missing test coverage")

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "failed") {
		t.Errorf("expected 'failed' in output, got: %s", output)
	}
}

// TestLogError tests the LogError method.
func TestLogError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogError(ctx, "operation failed", &testError{msg: "file not found"}, map[string]any{
		"file": "/path/to/file.go",
	})

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}

	if !strings.Contains(output, "operation failed") {
		t.Errorf("expected 'operation failed' in output, got: %s", output)
	}

	if !strings.Contains(output, "file not found") {
		t.Errorf("expected error message in output, got: %s", output)
	}
}

// TestLogWarning tests the LogWarning method.
func TestLogWarning(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogWarning(ctx, "deprecated method used", map[string]any{"method": "oldMethod"})

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}
}

// TestLogDebug tests the LogDebug method.
func TestLogDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogDebug(ctx, "debug info", map[string]any{"key": "value"})

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}
}

// TestLogInfo tests the LogInfo method.
func TestLogInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)
	ctx := context.Background()

	logger.LogInfo(ctx, "info message", map[string]any{"count": 42})

	output := buf.String()
	if output == "" {
		t.Error("expected log output")
	}
}

// TestDefaultLogger tests the package-level default logger functions.
func TestDefaultLogger(t *testing.T) {
	// Save original logger
	orig := defaultLogger

	// Create a new default logger
	buf := &bytes.Buffer{}
	newLogger := newTestLogger(buf)
	SetDefault(newLogger)

	// Test package-level functions
	ctx := context.Background()

	LogSpecOperation(ctx, SpecOpInit, PhaseParse, "test", nil)
	LogLoopIteration(ctx, 1, LoopPhasePreLoop, "test", nil)
	LogTaskExecution(ctx, "task-1", TaskStatusRunning, nil)
	LogError(ctx, "test error", &testError{msg: "test"}, nil)
	LogWarning(ctx, "test warning", nil)
	LogDebug(ctx, "test debug", nil)
	LogInfo(ctx, "test info", nil)

	// Restore original logger
	SetDefault(orig)

	output := buf.String()
	if output == "" {
		t.Error("expected log output from default logger")
	}
}

// TestDefaultLoggerFunction tests that Default() returns the default logger.
func TestDefaultLoggerFunction(t *testing.T) {
	logger := Default()

	if logger == nil {
		t.Error("expected non-nil default logger")
	}
}

// TestSetDefault tests setting a new default logger.
func TestSetDefault(t *testing.T) {
	orig := defaultLogger
	buf := &bytes.Buffer{}
	newLogger := newTestLogger(buf)

	SetDefault(newLogger)

	if defaultLogger != newLogger {
		t.Error("SetDefault did not update default logger")
	}

	// Restore
	SetDefault(orig)
}

// TestConstants tests that all constants are defined correctly.
func TestConstants(t *testing.T) {
	// Operation phases
	if PhaseParse != "parse" {
		t.Errorf("expected PhaseParse='parse', got '%s'", PhaseParse)
	}
	if PhaseAnalyze != "analyze" {
		t.Errorf("expected PhaseAnalyze='analyze', got '%s'", PhaseAnalyze)
	}
	if PhaseGenerate != "generate" {
		t.Errorf("expected PhaseGenerate='generate', got '%s'", PhaseGenerate)
	}
	if PhaseMerge != "merge" {
		t.Errorf("expected PhaseMerge='merge', got '%s'", PhaseMerge)
	}
	if PhaseValidate != "validate" {
		t.Errorf("expected PhaseValidate='validate', got '%s'", PhaseValidate)
	}

	// Spec operations
	if SpecOpInit != "init" {
		t.Errorf("expected SpecOpInit='init', got '%s'", SpecOpInit)
	}
	if SpecOpUpdate != "update" {
		t.Errorf("expected SpecOpUpdate='update', got '%s'", SpecOpUpdate)
	}
	if SpecOpReverse != "reverse" {
		t.Errorf("expected SpecOpReverse='reverse', got '%s'", SpecOpReverse)
	}
	if SpecOpComplete != "complete" {
		t.Errorf("expected SpecOpComplete='complete', got '%s'", SpecOpComplete)
	}

	// Loop phases
	if LoopPhaseValidation != "validation" {
		t.Errorf("expected LoopPhaseValidation='validation', got '%s'", LoopPhaseValidation)
	}
	if LoopPhasePreLoop != "pre_loop" {
		t.Errorf("expected LoopPhasePreLoop='pre_loop', got '%s'", LoopPhasePreLoop)
	}

	// Task statuses
	if TaskStatusPending != "pending" {
		t.Errorf("expected TaskStatusPending='pending', got '%s'", TaskStatusPending)
	}
	if TaskStatusRunning != "running" {
		t.Errorf("expected TaskStatusRunning='running', got '%s'", TaskStatusRunning)
	}
	if TaskStatusSucceeded != "succeeded" {
		t.Errorf("expected TaskStatusSucceeded='succeeded', got '%s'", TaskStatusSucceeded)
	}
	if TaskStatusFailed != "failed" {
		t.Errorf("expected TaskStatusFailed='failed', got '%s'", TaskStatusFailed)
	}

	// SDD phases
	if SDDPhaseExplore != "explore" {
		t.Errorf("expected SDDPhaseExplore='explore', got '%s'", SDDPhaseExplore)
	}
	if SDDPhaseApply != "apply" {
		t.Errorf("expected SDDPhaseApply='apply', got '%s'", SDDPhaseApply)
	}
	if SDDPhaseVerify != "verify" {
		t.Errorf("expected SDDPhaseVerify='verify', got '%s'", SDDPhaseVerify)
	}
}

// TestLogLevels tests that log levels work correctly.
func TestLogLevels(t *testing.T) {
	if LevelDebug != slog.LevelDebug {
		t.Error("LevelDebug mismatch")
	}
	if LevelInfo != slog.LevelInfo {
		t.Error("LevelInfo mismatch")
	}
	if LevelWarn != slog.LevelWarn {
		t.Error("LevelWarn mismatch")
	}
	if LevelError != slog.LevelError {
		t.Error("LevelError mismatch")
	}
}
