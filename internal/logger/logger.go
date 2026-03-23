// Package logger provides a structured logging wrapper around Go's slog package,
// specifically designed for the GROVE ecosystem with domain-specific logging methods.
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

// OperationPhase represents the phase of a spec operation.
type OperationPhase string

const (
	PhaseParse    OperationPhase = "parse"
	PhaseAnalyze  OperationPhase = "analyze"
	PhaseGenerate OperationPhase = "generate"
	PhaseMerge    OperationPhase = "merge"
	PhaseValidate OperationPhase = "validate"
	PhaseComplete OperationPhase = "complete"
)

// SpecOperation represents the type of spec operation being performed.
type SpecOperation string

const (
	SpecOpInit     SpecOperation = "init"
	SpecOpUpdate   SpecOperation = "update"
	SpecOpReverse  SpecOperation = "reverse"
	SpecOpComplete SpecOperation = "complete"
)

// LoopPhase represents the current phase of a Ralph Loop iteration.
type LoopPhase string

const (
	LoopPhaseValidation     LoopPhase = "validation"
	LoopPhasePreLoop        LoopPhase = "pre_loop"
	LoopPhaseTaskLoad       LoopPhase = "task_load"
	LoopPhaseImplementation LoopPhase = "implementation"
	LoopPhaseVerification   LoopPhase = "verification"
	LoopPhaseStatePersist   LoopPhase = "state_persist"
	LoopPhaseCompletion     LoopPhase = "completion"
)

// TaskStatus represents the status of a task within the loop.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusDeferred  TaskStatus = "deferred"
	TaskStatusSkipped   TaskStatus = "skipped"
)

// SDDPhase represents the current SDD phase being executed.
type SDDPhase string

const (
	SDDPhaseExplore SDDPhase = "explore"
	SDDPhasePropose SDDPhase = "propose"
	SDDPhaseSpec    SDDPhase = "spec"
	SDDPhaseDesign  SDDPhase = "design"
	SDDPhaseTasks   SDDPhase = "tasks"
	SDDPhaseApply   SDDPhase = "apply"
	SDDPhaseVerify  SDDPhase = "verify"
	SDDPhaseArchive SDDPhase = "archive"
)

// LogLevel defines the logging level for the GROVE logger.
// It implements slog.Leveler for compatibility with slog.HandlerOptions.
type LogLevel slog.Level

// Level implements slog.Leveler interface.
func (l LogLevel) Level() slog.Level {
	return slog.Level(l)
}

const (
	LevelDebug LogLevel = LogLevel(slog.LevelDebug)
	LevelInfo  LogLevel = LogLevel(slog.LevelInfo)
	LevelWarn  LogLevel = LogLevel(slog.LevelWarn)
	LevelError LogLevel = LogLevel(slog.LevelError)
)

// Config holds the configuration for the logger.
type Config struct {
	// Level sets the minimum log level (default: Info).
	Level LogLevel
	// Output specifies the destination writer (default: os.Stderr).
	Output io.Writer
	// AddSource enables source location in logs (default: false).
	AddSource bool
	// Pretty enables human-readable formatting (default: true in development).
	Pretty bool
}

// DefaultConfig returns a default logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:     LevelInfo,
		Output:    os.Stderr,
		AddSource: false,
		Pretty:    true,
	}
}

// GroveLogger is a structured logger wrapper with GROVE-specific methods.
type GroveLogger struct {
	*slog.Logger
	cfg Config
}

// toAttrs converts a slice of key-value pairs ([]any) to []slog.Attr.
// Assumes args are alternating: key1, value1, key2, value2, ...
func toAttrs(args []any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		attrs = append(attrs, slog.Any(key, args[i+1]))
	}
	return attrs
}

// New creates a new GroveLogger with the given configuration.
func New(cfg Config) *GroveLogger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	if cfg.Pretty {
		handler = slog.NewTextHandler(cfg.Output, opts)
	} else {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	}

	return &GroveLogger{
		Logger: slog.New(handler),
		cfg:    cfg,
	}
}

// NewDefault creates a logger with default configuration.
func NewDefault() *GroveLogger {
	return New(DefaultConfig())
}

// WithContext returns a logger that includes context values in log attributes.
func (l *GroveLogger) WithContext(ctx context.Context) *GroveLogger {
	return l
}

// WithComponent returns a logger with an added component attribute.
func (l *GroveLogger) WithComponent(component string) *GroveLogger {
	return &GroveLogger{
		Logger: l.With("component", component),
		cfg:    l.cfg,
	}
}

// WithChange returns a logger with an added change attribute.
func (l *GroveLogger) WithChange(change string) *GroveLogger {
	return &GroveLogger{
		Logger: l.With("change", change),
		cfg:    l.cfg,
	}
}

// WithTask returns a logger with an added task attribute.
func (l *GroveLogger) WithTask(taskID string) *GroveLogger {
	return &GroveLogger{
		Logger: l.With("task_id", taskID),
		cfg:    l.cfg,
	}
}

// WithLoop returns a logger with an added loop number attribute.
func (l *GroveLogger) WithLoop(loopNumber int) *GroveLogger {
	return &GroveLogger{
		Logger: l.With("loop", loopNumber),
		cfg:    l.cfg,
	}
}

// WithSDDPhase returns a logger with an added SDD phase attribute.
func (l *GroveLogger) WithSDDPhase(phase SDDPhase) *GroveLogger {
	return &GroveLogger{
		Logger: l.With("sdd_phase", phase),
		cfg:    l.cfg,
	}
}

// LogSpecOperation logs a spec operation with detailed attributes.
func (l *GroveLogger) LogSpecOperation(
	ctx context.Context,
	operation SpecOperation,
	phase OperationPhase,
	change string,
	details map[string]any,
) {
	args := []any{
		"operation", operation,
		"phase", phase,
		"change", change,
	}
	for k, v := range details {
		args = append(args, k, v)
	}
	l.LogAttrs(ctx, slog.LevelInfo, "spec_operation", toAttrs(args)...)
}

// LogSpecOperationStart logs the start of a spec operation.
func (l *GroveLogger) LogSpecOperationStart(
	ctx context.Context,
	operation SpecOperation,
	change string,
) {
	l.LogSpecOperation(ctx, operation, PhaseParse, change, map[string]any{
		"status": "started",
	})
}

// LogSpecOperationEnd logs the end of a spec operation.
func (l *GroveLogger) LogSpecOperationEnd(
	ctx context.Context,
	operation SpecOperation,
	change string,
	duration time.Duration,
	success bool,
) {
	status := "completed"
	if !success {
		status = "failed"
	}
	l.LogSpecOperation(ctx, operation, PhaseComplete, change, map[string]any{
		"status":   status,
		"duration": duration.String(),
		"success":  success,
	})
}

// LogLoopIteration logs a Ralph Loop iteration event.
func (l *GroveLogger) LogLoopIteration(
	ctx context.Context,
	loopNumber int,
	phase LoopPhase,
	event string,
	details map[string]any,
) {
	args := []any{
		"loop", loopNumber,
		"phase", phase,
		"event", event,
	}
	for k, v := range details {
		args = append(args, k, v)
	}
	l.LogAttrs(ctx, slog.LevelInfo, "loop_iteration", toAttrs(args)...)
}

// LogLoopIterationStart logs the start of a loop iteration.
func (l *GroveLogger) LogLoopIterationStart(
	ctx context.Context,
	loopNumber int,
	totalTasks int,
) {
	l.LogLoopIteration(ctx, loopNumber, LoopPhasePreLoop, "loop_started", map[string]any{
		"total_tasks": totalTasks,
	})
}

// LogLoopIterationEnd logs the end of a loop iteration.
func (l *GroveLogger) LogLoopIterationEnd(
	ctx context.Context,
	loopNumber int,
	duration time.Duration,
	completedTasks int,
	failedTasks int,
) {
	l.LogLoopIteration(ctx, loopNumber, LoopPhaseCompletion, "loop_completed", map[string]any{
		"duration":     duration.String(),
		"completed":    completedTasks,
		"failed":       failedTasks,
		"success_rate": float64(completedTasks) / float64(completedTasks+failedTasks),
	})
}

// LogTaskExecution logs a task execution event.
func (l *GroveLogger) LogTaskExecution(
	ctx context.Context,
	taskID string,
	status TaskStatus,
	details map[string]any,
) {
	args := []any{
		"task_id", taskID,
		"status", status,
	}
	for k, v := range details {
		args = append(args, k, v)
	}
	level := slog.LevelInfo
	if status == TaskStatusFailed {
		level = slog.LevelError
	}
	l.LogAttrs(ctx, level, "task_execution", toAttrs(args)...)
}

// LogTaskExecutionStart logs the start of a task execution.
func (l *GroveLogger) LogTaskExecutionStart(
	ctx context.Context,
	taskID string,
	changeName string,
	taskDescription string,
) {
	l.LogTaskExecution(ctx, taskID, TaskStatusRunning, map[string]any{
		"change":      changeName,
		"description": taskDescription,
		"event":       "started",
	})
}

// LogTaskExecutionEnd logs the end of a task execution.
func (l *GroveLogger) LogTaskExecutionEnd(
	ctx context.Context,
	taskID string,
	status TaskStatus,
	duration time.Duration,
	attempts int,
	err error,
) {
	details := map[string]any{
		"duration": duration.String(),
		"attempts": attempts,
		"event":    "ended",
	}
	if err != nil {
		details["error"] = err.Error()
	}
	l.LogTaskExecution(ctx, taskID, status, details)
}

// LogSDDPhase logs an SDD phase transition.
func (l *GroveLogger) LogSDDPhase(
	ctx context.Context,
	phase SDDPhase,
	change string,
	status string,
	details map[string]any,
) {
	args := []any{
		"sdd_phase", phase,
		"change", change,
		"status", status,
	}
	for k, v := range details {
		args = append(args, k, v)
	}
	l.LogAttrs(ctx, slog.LevelInfo, "sdd_phase_transition", toAttrs(args)...)
}

// LogVerification logs a verification result.
func (l *GroveLogger) LogVerification(
	ctx context.Context,
	taskID string,
	passed bool,
	reason string,
) {
	level := slog.LevelInfo
	status := "passed"
	if !passed {
		level = slog.LevelWarn
		status = "failed"
	}
	l.LogAttrs(ctx, level, "verification_result",
		slog.String("task_id", taskID),
		slog.String("status", status),
		slog.String("reason", reason),
	)
}

// LogError logs an error with optional context.
func (l *GroveLogger) LogError(
	ctx context.Context,
	msg string,
	err error,
	details map[string]any,
) {
	args := []any{}
	if err != nil {
		args = append(args, "error", err.Error())
	}
	for k, v := range details {
		args = append(args, k, v)
	}
	l.LogAttrs(ctx, slog.LevelError, msg, toAttrs(args)...)
}

// LogWarning logs a warning message.
func (l *GroveLogger) LogWarning(
	ctx context.Context,
	msg string,
	details map[string]any,
) {
	args := []any{}
	for k, v := range details {
		args = append(args, k, v)
	}
	l.LogAttrs(ctx, slog.LevelWarn, msg, toAttrs(args)...)
}

// LogDebug logs a debug message.
func (l *GroveLogger) LogDebug(
	ctx context.Context,
	msg string,
	details map[string]any,
) {
	args := []any{}
	for k, v := range details {
		args = append(args, k, v)
	}
	l.LogAttrs(ctx, slog.LevelDebug, msg, toAttrs(args)...)
}

// LogInfo logs an info message.
func (l *GroveLogger) LogInfo(
	ctx context.Context,
	msg string,
	details map[string]any,
) {
	args := []any{}
	for k, v := range details {
		args = append(args, k, v)
	}
	l.LogAttrs(ctx, slog.LevelInfo, msg, toAttrs(args)...)
}

// Global logger instance
var defaultLogger = NewDefault()

// Default returns the default global logger.
func Default() *GroveLogger {
	return defaultLogger
}

// SetDefault sets a new default logger.
func SetDefault(l *GroveLogger) {
	defaultLogger = l
}

// Package-level convenience functions using the default logger

// LogSpecOperation logs a spec operation using the default logger.
func LogSpecOperation(
	ctx context.Context,
	operation SpecOperation,
	phase OperationPhase,
	change string,
	details map[string]any,
) {
	defaultLogger.LogSpecOperation(ctx, operation, phase, change, details)
}

// LogLoopIteration logs a loop iteration using the default logger.
func LogLoopIteration(
	ctx context.Context,
	loopNumber int,
	phase LoopPhase,
	event string,
	details map[string]any,
) {
	defaultLogger.LogLoopIteration(ctx, loopNumber, phase, event, details)
}

// LogTaskExecution logs a task execution using the default logger.
func LogTaskExecution(
	ctx context.Context,
	taskID string,
	status TaskStatus,
	details map[string]any,
) {
	defaultLogger.LogTaskExecution(ctx, taskID, status, details)
}

// LogError logs an error using the default logger.
func LogError(ctx context.Context, msg string, err error, details map[string]any) {
	defaultLogger.LogError(ctx, msg, err, details)
}

// LogWarning logs a warning using the default logger.
func LogWarning(ctx context.Context, msg string, details map[string]any) {
	defaultLogger.LogWarning(ctx, msg, details)
}

// LogDebug logs a debug message using the default logger.
func LogDebug(ctx context.Context, msg string, details map[string]any) {
	defaultLogger.LogDebug(ctx, msg, details)
}

// LogInfo logs an info message using the default logger.
func LogInfo(ctx context.Context, msg string, details map[string]any) {
	defaultLogger.LogInfo(ctx, msg, details)
}

// CLI flags for controlling output
var cliVerbose bool
var cliQuiet bool

// SetVerbose enables verbose output for CLI functions.
func SetVerbose(v bool) { cliVerbose = v }

// SetQuiet suppresses non-essential output for CLI functions.
func SetQuiet(q bool) { cliQuiet = q }

// Info prints an info-level message for CLIs.
func Info(format string, args ...interface{}) {
	if cliQuiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	defaultLogger.LogInfo(context.Background(), msg, nil)
}

// Success prints a success message for CLIs.
func Success(format string, args ...interface{}) {
	if cliQuiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	defaultLogger.LogInfo(context.Background(), "✓ "+msg, nil)
}

// Warn prints a warning message for CLIs.
func Warn(format string, args ...interface{}) {
	if cliQuiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	defaultLogger.LogWarning(context.Background(), "⚠ "+msg, nil)
}

// Error prints an error message for CLIs.
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	defaultLogger.LogError(context.Background(), msg, nil, nil)
}

// Debug prints a debug message for CLIs (only if verbose is enabled).
func Debug(format string, args ...interface{}) {
	if !cliVerbose || cliQuiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	defaultLogger.LogDebug(context.Background(), msg, nil)
}

// Progress prints a progress indicator for CLIs.
func Progress(current, total int, message string) {
	if cliQuiet {
		return
	}
	Info("[%d/%d] %s", current, total, message)
}

// Section prints a section header for CLIs.
func Section(title string) {
	if cliQuiet {
		return
	}
	Info("")
	Info(title)
}

// Fatal prints an error and exits with code 1.
func Fatal(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}
