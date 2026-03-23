// Package errors provides custom error types for Grove.
package errors

import (
	"errors"
	"fmt"
	"io"
)

// ============================================================================
// Sentinel Errors (for fast comparisons)
// ============================================================================

var (
	// Spec errors
	ErrSpecNotFound      = errors.New("spec not found")
	ErrSpecInvalidFormat = errors.New("spec has invalid format")
	ErrSpecParseFailed   = errors.New("spec parse failed")
	ErrSpecWriteFailed   = errors.New("spec write failed")

	// Loop errors
	ErrLoopNotFound      = errors.New("loop not found")
	ErrLoopAlreadyExists = errors.New("loop already exists")
	ErrLoopExecution     = errors.New("loop execution failed")
	ErrLoopTimeout       = errors.New("loop execution timeout")
	ErrLoopCancelled     = errors.New("loop execution cancelled")

	// Opti (optimization) errors
	ErrOptiNotFound  = errors.New("optimization not found")
	ErrOptiFailed    = errors.New("optimization failed")
	ErrOptiInvalidOp = errors.New("invalid optimization operation")

	// Config errors
	ErrConfigNotFound    = errors.New("config not found")
	ErrConfigReadFailed  = errors.New("config read failed")
	ErrConfigWriteFailed = errors.New("config write failed")
	ErrConfigInvalid     = errors.New("config invalid")
	ErrConfigValidation  = errors.New("config validation failed")

	// State errors
	ErrStateNotFound   = errors.New("state not found")
	ErrStateCorrupted  = errors.New("state corrupted")
	ErrStateTransition = errors.New("invalid state transition")
	ErrStateLockFailed = errors.New("state lock failed")

	// File errors
	ErrFileNotFound   = errors.New("file not found")
	ErrFileRead       = errors.New("file read error")
	ErrFileWrite      = errors.New("file write error")
	ErrFilePermission = errors.New("file permission denied")
	ErrFileNotDir     = errors.New("path is not a directory")

	// Validation errors
	ErrValidationRequired = errors.New("validation failed: field required")
	ErrValidationType     = errors.New("validation failed: type mismatch")
	ErrValidationRange    = errors.New("validation failed: value out of range")
	ErrValidationPattern  = errors.New("validation failed: pattern mismatch")
)

// ============================================================================
// SpecError - Errors related to spec operations
// ============================================================================

// SpecError represents an error during spec operations.
type SpecError struct {
	SpecName  string
	Operation string
	Cause     error
}

func (e SpecError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("spec %q %s: %v", e.SpecName, e.Operation, e.Cause)
	}
	return fmt.Sprintf("spec %q %s", e.SpecName, e.Operation)
}

func (e SpecError) Unwrap() error {
	return e.Cause
}

func (e SpecError) Is(target error) bool {
	switch target {
	case ErrSpecNotFound, ErrSpecInvalidFormat, ErrSpecParseFailed, ErrSpecWriteFailed:
		return true
	}
	return false
}

// NewSpecError creates a new SpecError.
func NewSpecError(specName, operation string, cause error) SpecError {
	return SpecError{SpecName: specName, Operation: operation, Cause: cause}
}

// ============================================================================
// LoopError - Errors related to loop operations
// ============================================================================

// LoopError represents an error during loop operations.
type LoopError struct {
	LoopID      string
	Operation   string
	Cause       error
	IsTimeout   bool
	IsCancelled bool
}

func (e LoopError) Error() string {
	base := fmt.Sprintf("loop %q %s", e.LoopID, e.Operation)
	switch {
	case e.IsTimeout:
		return base + ": timeout"
	case e.IsCancelled:
		return base + ": cancelled"
	case e.Cause != nil:
		return base + ": " + e.Cause.Error()
	default:
		return base
	}
}

func (e LoopError) Unwrap() error {
	return e.Cause
}

func (e LoopError) Is(target error) bool {
	switch target {
	case ErrLoopNotFound, ErrLoopAlreadyExists, ErrLoopExecution, ErrLoopTimeout, ErrLoopCancelled:
		return true
	}
	return false
}

// NewLoopError creates a new LoopError.
func NewLoopError(loopID, operation string, cause error) LoopError {
	return LoopError{LoopID: loopID, Operation: operation, Cause: cause}
}

// WithTimeout marks the error as a timeout error.
func (e LoopError) WithTimeout() LoopError {
	e.IsTimeout = true
	return e
}

// WithCancelled marks the error as a cancelled error.
func (e LoopError) WithCancelled() LoopError {
	e.IsCancelled = true
	return e
}

// ============================================================================
// OptiError - Errors related to optimization operations
// ============================================================================

// OptiError represents an error during optimization operations.
type OptiError struct {
	OptiName  string
	Operation string
	Cause     error
}

func (e OptiError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("opti %q %s: %v", e.OptiName, e.Operation, e.Cause)
	}
	return fmt.Sprintf("opti %q %s", e.OptiName, e.Operation)
}

func (e OptiError) Unwrap() error {
	return e.Cause
}

func (e OptiError) Is(target error) bool {
	switch target {
	case ErrOptiNotFound, ErrOptiFailed, ErrOptiInvalidOp:
		return true
	}
	return false
}

// NewOptiError creates a new OptiError.
func NewOptiError(optiName, operation string, cause error) OptiError {
	return OptiError{OptiName: optiName, Operation: operation, Cause: cause}
}

// ============================================================================
// ConfigError - Errors related to configuration
// ============================================================================

// ConfigError represents an error during configuration operations.
type ConfigError struct {
	ConfigKey  string
	Operation  string
	Cause      error
	Validation bool
}

func (e ConfigError) Error() string {
	if e.Validation {
		return fmt.Sprintf("config validation failed for key %q: %v", e.ConfigKey, e.Cause)
	}
	if e.Cause != nil {
		return fmt.Sprintf("config %q %s: %v", e.ConfigKey, e.Operation, e.Cause)
	}
	return fmt.Sprintf("config %q %s", e.ConfigKey, e.Operation)
}

func (e ConfigError) Unwrap() error {
	return e.Cause
}

func (e ConfigError) Is(target error) bool {
	switch target {
	case ErrConfigNotFound, ErrConfigReadFailed, ErrConfigWriteFailed, ErrConfigInvalid, ErrConfigValidation:
		return true
	}
	return false
}

// NewConfigError creates a new ConfigError.
func NewConfigError(key, operation string, cause error) ConfigError {
	return ConfigError{ConfigKey: key, Operation: operation, Cause: cause}
}

// WithValidation marks the error as a validation error.
func (e ConfigError) WithValidation() ConfigError {
	e.Validation = true
	return e
}

// ============================================================================
// StateError - Errors related to state management
// ============================================================================

// StateError represents an error during state operations.
type StateError struct {
	StateName   string
	Operation   string
	Cause       error
	IsCorrupted bool
}

func (e StateError) Error() string {
	if e.IsCorrupted {
		return fmt.Sprintf("state %q is corrupted: %v", e.StateName, e.Cause)
	}
	if e.Cause != nil {
		return fmt.Sprintf("state %q %s: %v", e.StateName, e.Operation, e.Cause)
	}
	return fmt.Sprintf("state %q %s", e.StateName, e.Operation)
}

func (e StateError) Unwrap() error {
	return e.Cause
}

func (e StateError) Is(target error) bool {
	switch target {
	case ErrStateNotFound, ErrStateCorrupted, ErrStateTransition, ErrStateLockFailed:
		return true
	}
	return false
}

// NewStateError creates a new StateError.
func NewStateError(stateName, operation string, cause error) StateError {
	return StateError{StateName: stateName, Operation: operation, Cause: cause}
}

// WithCorrupted marks the error as a corruption error.
func (e StateError) WithCorrupted() StateError {
	e.IsCorrupted = true
	return e
}

// ============================================================================
// FileError - Errors related to file operations
// ============================================================================

// FileError represents an error during file operations.
type FileError struct {
	Path         string
	Operation    string
	Cause        error
	IsNotFound   bool
	IsPermission bool
}

func (e FileError) Error() string {
	base := fmt.Sprintf("file %q %s", e.Path, e.Operation)
	switch {
	case e.IsNotFound:
		return base + ": not found"
	case e.IsPermission:
		return base + ": permission denied"
	case e.Cause != nil:
		return base + ": " + e.Cause.Error()
	default:
		return base
	}
}

func (e FileError) Unwrap() error {
	return e.Cause
}

func (e FileError) Is(target error) bool {
	switch target {
	case ErrFileNotFound, ErrFileRead, ErrFileWrite, ErrFilePermission, ErrFileNotDir:
		return true
	}
	return false
}

// NewFileError creates a new FileError.
func NewFileError(path, operation string, cause error) FileError {
	return FileError{Path: path, Operation: operation, Cause: cause}
}

// WithNotFound marks the error as a not-found error.
func (e FileError) WithNotFound() FileError {
	e.IsNotFound = true
	return e
}

// WithPermission marks the error as a permission error.
func (e FileError) WithPermission() FileError {
	e.IsPermission = true
	return e
}

// ============================================================================
// ValidationError - Errors related to data validation
// ============================================================================

// ValidationError represents an error during data validation.
type ValidationError struct {
	Field   string
	Rule    string
	Value   any
	Message string
	Cause   error
}

func (e ValidationError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("validation failed for field %q: %s", e.Field, e.Message)
	}
	base := fmt.Sprintf("validation failed for field %q", e.Field)
	if e.Rule != "" {
		base += fmt.Sprintf(" (rule: %s)", e.Rule)
	}
	if e.Value != nil {
		base += fmt.Sprintf(", got: %v", e.Value)
	}
	return base
}

func (e ValidationError) Unwrap() error {
	return e.Cause
}

func (e ValidationError) Is(target error) bool {
	switch target {
	case ErrValidationRequired, ErrValidationType, ErrValidationRange, ErrValidationPattern:
		return true
	}
	return false
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, rule string, cause error) ValidationError {
	return ValidationError{Field: field, Rule: rule, Cause: cause}
}

// WithValue adds the invalid value to the error.
func (e ValidationError) WithValue(value any) ValidationError {
	e.Value = value
	return e
}

// WithMessage sets a custom message for the error.
func (e ValidationError) WithMessage(msg string) ValidationError {
	e.Message = msg
	return e
}

// ============================================================================
// Helper Functions
// ============================================================================

// IsNotFound checks if an error indicates a "not found" condition.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrSpecNotFound) ||
		errors.Is(err, ErrLoopNotFound) ||
		errors.Is(err, ErrOptiNotFound) ||
		errors.Is(err, ErrConfigNotFound) ||
		errors.Is(err, ErrStateNotFound) ||
		errors.Is(err, ErrFileNotFound)
}

// IsPermission checks if an error indicates a permission problem.
func IsPermission(err error) bool {
	return errors.Is(err, ErrFilePermission) ||
		errors.Is(err, ErrConfigValidation)
}

// IsTimeout checks if an error indicates a timeout.
func IsTimeout(err error) bool {
	var loopErr LoopError
	if errors.As(err, &loopErr) {
		return loopErr.IsTimeout
	}
	return errors.Is(err, ErrLoopTimeout)
}

// IsCancelled checks if an error indicates a cancelled operation.
func IsCancelled(err error) bool {
	var loopErr LoopError
	if errors.As(err, &loopErr) {
		return loopErr.IsCancelled
	}
	return errors.Is(err, ErrLoopCancelled)
}

// IsCorrupted checks if an error indicates data corruption.
func IsCorrupted(err error) bool {
	var stateErr StateError
	if errors.As(err, &stateErr) {
		return stateErr.IsCorrupted
	}
	return errors.Is(err, ErrStateCorrupted)
}

// Cause returns the underlying cause of an error, unwrapping all layers.
func Cause(err error) error {
	for err != nil {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
	return nil
}

// IsIOError checks if an error is related to I/O operations.
func IsIOError(err error) bool {
	return errors.Is(err, ErrFileRead) ||
		errors.Is(err, ErrFileWrite) ||
		errors.Is(err, io.EOF)
}
