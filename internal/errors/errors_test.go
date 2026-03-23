package errors

import (
	"errors"
	"fmt"
	"io"
	"testing"
)

// ============================================================================
// Sentinel Error Tests
// ============================================================================

func TestSentinelErrors(t *testing.T) {
	t.Run("Spec errors are distinct", func(t *testing.T) {
		tests := []struct {
			name   string
			err1   error
			err2   error
			wantEq bool
		}{
			{"ErrSpecNotFound != ErrSpecInvalidFormat", ErrSpecNotFound, ErrSpecInvalidFormat, false},
			{"ErrSpecNotFound == ErrSpecNotFound", ErrSpecNotFound, ErrSpecNotFound, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := errors.Is(tt.err1, tt.err2)
				if got != tt.wantEq {
					t.Errorf("errors.Is() = %v, want %v", got, tt.wantEq)
				}
			})
		}
	})

	t.Run("Loop errors are distinct", func(t *testing.T) {
		tests := []struct {
			name   string
			err1   error
			err2   error
			wantEq bool
		}{
			{"ErrLoopNotFound != ErrLoopExecution", ErrLoopNotFound, ErrLoopExecution, false},
			{"ErrLoopAlreadyExists != ErrLoopTimeout", ErrLoopAlreadyExists, ErrLoopTimeout, false},
			{"ErrLoopTimeout == ErrLoopTimeout", ErrLoopTimeout, ErrLoopTimeout, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := errors.Is(tt.err1, tt.err2)
				if got != tt.wantEq {
					t.Errorf("errors.Is() = %v, want %v", got, tt.wantEq)
				}
			})
		}
	})

	t.Run("Opti errors are distinct", func(t *testing.T) {
		tests := []struct {
			name   string
			err1   error
			err2   error
			wantEq bool
		}{
			{"ErrOptiNotFound != ErrOptiFailed", ErrOptiNotFound, ErrOptiFailed, false},
			{"ErrOptiInvalidOp != ErrOptiNotFound", ErrOptiInvalidOp, ErrOptiNotFound, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := errors.Is(tt.err1, tt.err2)
				if got != tt.wantEq {
					t.Errorf("errors.Is() = %v, want %v", got, tt.wantEq)
				}
			})
		}
	})

	t.Run("Config errors are distinct", func(t *testing.T) {
		tests := []struct {
			name   string
			err1   error
			err2   error
			wantEq bool
		}{
			{"ErrConfigNotFound != ErrConfigInvalid", ErrConfigNotFound, ErrConfigInvalid, false},
			{"ErrConfigValidation != ErrConfigWriteFailed", ErrConfigValidation, ErrConfigWriteFailed, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := errors.Is(tt.err1, tt.err2)
				if got != tt.wantEq {
					t.Errorf("errors.Is() = %v, want %v", got, tt.wantEq)
				}
			})
		}
	})

	t.Run("State errors are distinct", func(t *testing.T) {
		tests := []struct {
			name   string
			err1   error
			err2   error
			wantEq bool
		}{
			{"ErrStateNotFound != ErrStateCorrupted", ErrStateNotFound, ErrStateCorrupted, false},
			{"ErrStateTransition != ErrStateLockFailed", ErrStateTransition, ErrStateLockFailed, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := errors.Is(tt.err1, tt.err2)
				if got != tt.wantEq {
					t.Errorf("errors.Is() = %v, want %v", got, tt.wantEq)
				}
			})
		}
	})

	t.Run("File errors are distinct", func(t *testing.T) {
		tests := []struct {
			name   string
			err1   error
			err2   error
			wantEq bool
		}{
			{"ErrFileNotFound != ErrFileRead", ErrFileNotFound, ErrFileRead, false},
			{"ErrFileWrite != ErrFilePermission", ErrFileWrite, ErrFilePermission, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := errors.Is(tt.err1, tt.err2)
				if got != tt.wantEq {
					t.Errorf("errors.Is() = %v, want %v", got, tt.wantEq)
				}
			})
		}
	})

	t.Run("Validation errors are distinct", func(t *testing.T) {
		tests := []struct {
			name   string
			err1   error
			err2   error
			wantEq bool
		}{
			{"ErrValidationRequired != ErrValidationType", ErrValidationRequired, ErrValidationType, false},
			{"ErrValidationRange != ErrValidationPattern", ErrValidationRange, ErrValidationPattern, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := errors.Is(tt.err1, tt.err2)
				if got != tt.wantEq {
					t.Errorf("errors.Is() = %v, want %v", got, tt.wantEq)
				}
			})
		}
	})
}

// ============================================================================
// SpecError Tests
// ============================================================================

func TestSpecError(t *testing.T) {
	t.Run("Error() returns formatted message", func(t *testing.T) {
		err := NewSpecError("test-spec", "read", io.EOF)
		want := `spec "test-spec" read: EOF`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() returns message without cause", func(t *testing.T) {
		err := NewSpecError("test-spec", "read", nil)
		want := `spec "test-spec" read`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Unwrap() returns underlying cause", func(t *testing.T) {
		cause := io.EOF
		err := NewSpecError("test-spec", "read", cause)
		if got := err.Unwrap(); got != cause {
			t.Errorf("Unwrap() = %v, want %v", got, cause)
		}
	})

	t.Run("Is() matches sentinel errors", func(t *testing.T) {
		err := NewSpecError("test-spec", "read", io.EOF)

		if !err.Is(ErrSpecNotFound) {
			t.Error("Is(ErrSpecNotFound) = false, want true")
		}
		if !err.Is(ErrSpecInvalidFormat) {
			t.Error("Is(ErrSpecInvalidFormat) = false, want true")
		}
		if err.Is(ErrLoopNotFound) {
			t.Error("Is(ErrLoopNotFound) = true, want false")
		}
	})

	t.Run("errors.Is() works with wrapped errors", func(t *testing.T) {
		err := NewSpecError("test-spec", "read", ErrSpecNotFound)
		if !errors.Is(err, ErrSpecNotFound) {
			t.Error("errors.Is(err, ErrSpecNotFound) = false, want true")
		}
	})

	t.Run("errors.As() works correctly", func(t *testing.T) {
		cause := io.EOF
		err := NewSpecError("test-spec", "read", cause)

		var specErr SpecError
		if !errors.As(err, &specErr) {
			t.Fatal("errors.As() failed to extract SpecError")
		}
		if specErr.SpecName != "test-spec" {
			t.Errorf("SpecName = %q, want %q", specErr.SpecName, "test-spec")
		}
		if specErr.Operation != "read" {
			t.Errorf("Operation = %q, want %q", specErr.Operation, "read")
		}
	})
}

// ============================================================================
// LoopError Tests
// ============================================================================

func TestLoopError(t *testing.T) {
	t.Run("Error() returns formatted message", func(t *testing.T) {
		err := NewLoopError("loop-1", "execute", io.EOF)
		want := `loop "loop-1" execute: EOF`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() with timeout", func(t *testing.T) {
		err := NewLoopError("loop-1", "execute", nil).WithTimeout()
		want := `loop "loop-1" execute: timeout`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() with cancelled", func(t *testing.T) {
		err := NewLoopError("loop-1", "execute", nil).WithCancelled()
		want := `loop "loop-1" execute: cancelled`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Is() matches sentinel errors", func(t *testing.T) {
		err := NewLoopError("loop-1", "execute", nil)

		if !err.Is(ErrLoopNotFound) {
			t.Error("Is(ErrLoopNotFound) = false, want true")
		}
		if !err.Is(ErrLoopExecution) {
			t.Error("Is(ErrLoopExecution) = false, want true")
		}
		if err.Is(ErrOptiNotFound) {
			t.Error("Is(ErrOptiNotFound) = true, want false")
		}
	})

	t.Run("WithTimeout() sets IsTimeout flag", func(t *testing.T) {
		err := NewLoopError("loop-1", "execute", nil).WithTimeout()
		if !err.IsTimeout {
			t.Error("IsTimeout = false, want true")
		}
	})

	t.Run("WithCancelled() sets IsCancelled flag", func(t *testing.T) {
		err := NewLoopError("loop-1", "execute", nil).WithCancelled()
		if !err.IsCancelled {
			t.Error("IsCancelled = false, want true")
		}
	})
}

// ============================================================================
// OptiError Tests
// ============================================================================

func TestOptiError(t *testing.T) {
	t.Run("Error() returns formatted message", func(t *testing.T) {
		err := NewOptiError("opti-1", "apply", io.EOF)
		want := `opti "opti-1" apply: EOF`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() returns message without cause", func(t *testing.T) {
		err := NewOptiError("opti-1", "apply", nil)
		want := `opti "opti-1" apply`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Is() matches sentinel errors", func(t *testing.T) {
		err := NewOptiError("opti-1", "apply", nil)

		if !err.Is(ErrOptiNotFound) {
			t.Error("Is(ErrOptiNotFound) = false, want true")
		}
		if !err.Is(ErrOptiFailed) {
			t.Error("Is(ErrOptiFailed) = false, want true")
		}
		if err.Is(ErrSpecNotFound) {
			t.Error("Is(ErrSpecNotFound) = true, want false")
		}
	})
}

// ============================================================================
// ConfigError Tests
// ============================================================================

func TestConfigError(t *testing.T) {
	t.Run("Error() returns formatted message", func(t *testing.T) {
		err := NewConfigError("database.url", "read", io.EOF)
		want := `config "database.url" read: EOF`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() with validation flag", func(t *testing.T) {
		err := NewConfigError("database.url", "validate", io.EOF).WithValidation()
		want := `config validation failed for key "database.url": EOF`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Is() matches sentinel errors", func(t *testing.T) {
		err := NewConfigError("database.url", "read", nil)

		if !err.Is(ErrConfigNotFound) {
			t.Error("Is(ErrConfigNotFound) = false, want true")
		}
		if !err.Is(ErrConfigReadFailed) {
			t.Error("Is(ErrConfigReadFailed) = false, want true")
		}
		if err.Is(ErrStateNotFound) {
			t.Error("Is(ErrStateNotFound) = true, want false")
		}
	})

	t.Run("WithValidation() sets Validation flag", func(t *testing.T) {
		err := NewConfigError("key", "validate", nil).WithValidation()
		if !err.Validation {
			t.Error("Validation = false, want true")
		}
	})
}

// ============================================================================
// StateError Tests
// ============================================================================

func TestStateError(t *testing.T) {
	t.Run("Error() returns formatted message", func(t *testing.T) {
		err := NewStateError("session", "load", io.EOF)
		want := `state "session" load: EOF`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() with corrupted flag", func(t *testing.T) {
		err := NewStateError("session", "load", io.EOF).WithCorrupted()
		want := `state "session" is corrupted: EOF`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Is() matches sentinel errors", func(t *testing.T) {
		err := NewStateError("session", "load", nil)

		if !err.Is(ErrStateNotFound) {
			t.Error("Is(ErrStateNotFound) = false, want true")
		}
		if !err.Is(ErrStateCorrupted) {
			t.Error("Is(ErrStateCorrupted) = false, want true")
		}
		if err.Is(ErrConfigNotFound) {
			t.Error("Is(ErrConfigNotFound) = true, want false")
		}
	})

	t.Run("WithCorrupted() sets IsCorrupted flag", func(t *testing.T) {
		err := NewStateError("session", "load", nil).WithCorrupted()
		if !err.IsCorrupted {
			t.Error("IsCorrupted = false, want true")
		}
	})
}

// ============================================================================
// FileError Tests
// ============================================================================

func TestFileError(t *testing.T) {
	t.Run("Error() returns formatted message", func(t *testing.T) {
		err := NewFileError("/path/to/file.txt", "read", io.EOF)
		want := `file "/path/to/file.txt" read: EOF`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() with not found flag", func(t *testing.T) {
		err := NewFileError("/path/to/file.txt", "read", nil).WithNotFound()
		want := `file "/path/to/file.txt" read: not found`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() with permission flag", func(t *testing.T) {
		err := NewFileError("/path/to/file.txt", "read", nil).WithPermission()
		want := `file "/path/to/file.txt" read: permission denied`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Is() matches sentinel errors", func(t *testing.T) {
		err := NewFileError("/path/to/file.txt", "read", nil)

		if !err.Is(ErrFileNotFound) {
			t.Error("Is(ErrFileNotFound) = false, want true")
		}
		if !err.Is(ErrFileRead) {
			t.Error("Is(ErrFileRead) = false, want true")
		}
		if err.Is(ErrLoopNotFound) {
			t.Error("Is(ErrLoopNotFound) = true, want false")
		}
	})

	t.Run("WithNotFound() sets IsNotFound flag", func(t *testing.T) {
		err := NewFileError("/path/to/file.txt", "read", nil).WithNotFound()
		if !err.IsNotFound {
			t.Error("IsNotFound = false, want true")
		}
	})

	t.Run("WithPermission() sets IsPermission flag", func(t *testing.T) {
		err := NewFileError("/path/to/file.txt", "read", nil).WithPermission()
		if !err.IsPermission {
			t.Error("IsPermission = false, want true")
		}
	})
}

// ============================================================================
// ValidationError Tests
// ============================================================================

func TestValidationError(t *testing.T) {
	t.Run("Error() returns formatted message", func(t *testing.T) {
		err := NewValidationError("email", "format", nil)
		want := `validation failed for field "email" (rule: format)`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() with custom message", func(t *testing.T) {
		err := NewValidationError("email", "", nil).WithMessage("must be a valid email")
		want := `validation failed for field "email": must be a valid email`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() with value", func(t *testing.T) {
		err := NewValidationError("age", "range", nil).WithValue(-5)
		want := `validation failed for field "age", got: -5`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Error() with rule and value", func(t *testing.T) {
		err := NewValidationError("age", "positive", nil).WithValue(-5)
		want := `validation failed for field "age" (rule: positive), got: -5`
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Is() matches sentinel errors", func(t *testing.T) {
		err := NewValidationError("email", "format", nil)

		if !err.Is(ErrValidationRequired) {
			t.Error("Is(ErrValidationRequired) = false, want true")
		}
		if !err.Is(ErrValidationType) {
			t.Error("Is(ErrValidationType) = false, want true")
		}
		if err.Is(ErrFileNotFound) {
			t.Error("Is(ErrFileNotFound) = true, want false")
		}
	})

	t.Run("WithValue() sets Value field", func(t *testing.T) {
		err := NewValidationError("name", "required", nil).WithValue("")
		if err.Value != "" {
			t.Errorf("Value = %q, want %q", err.Value, "")
		}
	})

	t.Run("WithMessage() sets Message field", func(t *testing.T) {
		err := NewValidationError("name", "required", nil).WithMessage("name is required")
		if err.Message != "name is required" {
			t.Errorf("Message = %q, want %q", err.Message, "name is required")
		}
	})
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"SpecError matches", NewSpecError("test", "read", nil), true},
		{"LoopError matches", NewLoopError("test", "execute", nil), true},
		{"OptiError matches", NewOptiError("test", "apply", nil), true},
		{"ConfigError matches", NewConfigError("test", "read", nil), true},
		{"StateError matches", NewStateError("test", "load", nil), true},
		{"FileError matches", NewFileError("/path", "read", nil), true},
		{"generic error does not match", io.EOF, false},
		{"nil does not match", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPermission(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"FileError with permission", NewFileError("/path", "read", nil).WithPermission(), true},
		{"ConfigError with validation", NewConfigError("test", "validate", nil).WithValidation(), true},
		{"SpecError does not match", NewSpecError("test", "read", nil), false},
		{"generic error does not match", io.EOF, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPermission(tt.err); got != tt.want {
				t.Errorf("IsPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTimeout(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"LoopError with timeout", NewLoopError("test", "execute", nil).WithTimeout(), true},
		{"LoopError without timeout", NewLoopError("test", "execute", nil), false},
		{"ErrLoopTimeout sentinel", ErrLoopTimeout, true},
		{"SpecError does not match", NewSpecError("test", "read", nil), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTimeout(tt.err); got != tt.want {
				t.Errorf("IsTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCancelled(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"LoopError with cancelled", NewLoopError("test", "execute", nil).WithCancelled(), true},
		{"LoopError without cancelled", NewLoopError("test", "execute", nil), false},
		{"ErrLoopCancelled sentinel", ErrLoopCancelled, true},
		{"SpecError does not match", NewSpecError("test", "read", nil), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCancelled(tt.err); got != tt.want {
				t.Errorf("IsCancelled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCorrupted(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"StateError with corrupted", NewStateError("test", "load", nil).WithCorrupted(), true},
		{"StateError without corrupted", NewStateError("test", "load", nil), false},
		{"ErrStateCorrupted sentinel", ErrStateCorrupted, true},
		{"SpecError does not match", NewSpecError("test", "read", nil), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCorrupted(tt.err); got != tt.want {
				t.Errorf("IsCorrupted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCause(t *testing.T) {
	t.Run("returns unwrapped error", func(t *testing.T) {
		cause := io.EOF
		err := NewSpecError("test", "read", cause)
		if got := Cause(err); got != cause {
			t.Errorf("Cause() = %v, want %v", got, cause)
		}
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		if got := Cause(nil); got != nil {
			t.Errorf("Cause(nil) = %v, want nil", got)
		}
	})
}

func TestIsIOError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"FileError read", NewFileError("/path", "read", nil), true},
		{"FileError write", NewFileError("/path", "write", nil), true},
		{"io.EOF", io.EOF, true},
		{"SpecError does not match", NewSpecError("test", "read", nil), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIOError(tt.err); got != tt.want {
				t.Errorf("IsIOError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Error Wrapping Tests
// ============================================================================

func TestErrorWrapping(t *testing.T) {
	t.Run("sentinel error wrapping with fmt.Errorf", func(t *testing.T) {
		err := fmt.Errorf("failed to read spec: %w", ErrSpecNotFound)
		if !errors.Is(err, ErrSpecNotFound) {
			t.Error("errors.Is(err, ErrSpecNotFound) = false, want true")
		}
	})

	t.Run("nested error wrapping", func(t *testing.T) {
		baseErr := NewFileError("/path", "read", nil).WithNotFound()
		wrappedErr := fmt.Errorf("could not process: %w", baseErr)

		if !errors.Is(wrappedErr, ErrFileNotFound) {
			t.Error("errors.Is(wrappedErr, ErrFileNotFound) = false, want true")
		}

		var fileErr FileError
		if !errors.As(wrappedErr, &fileErr) {
			t.Error("errors.As(wrappedErr, &fileErr) = false, want true")
		}
		if fileErr.Path != "/path" {
			t.Errorf("fileErr.Path = %q, want %q", fileErr.Path, "/path")
		}
	})

	t.Run("multi-layer unwrapping", func(t *testing.T) {
		err1 := NewConfigError("key", "read", io.EOF)
		err2 := fmt.Errorf("outer: %w", err1)
		err3 := fmt.Errorf("middle: %w", err2)
		err4 := fmt.Errorf("inner: %w", err3)

		if !errors.Is(err4, ErrConfigReadFailed) {
			t.Error("errors.Is(err4, ErrConfigReadFailed) = false, want true")
		}
	})
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestEdgeCases(t *testing.T) {
	t.Run("empty spec name", func(t *testing.T) {
		err := NewSpecError("", "read", nil)
		if got := err.Error(); got != `spec "" read` {
			t.Errorf("Error() = %q, want %q", got, `spec "" read`)
		}
	})

	t.Run("empty loop id", func(t *testing.T) {
		err := NewLoopError("", "execute", nil)
		if got := err.Error(); got != `loop "" execute` {
			t.Errorf("Error() = %q, want %q", got, `loop "" execute`)
		}
	})

	t.Run("special characters in paths", func(t *testing.T) {
		err := NewFileError(`/path/with "quotes"`, "read", nil)
		if got := err.Error(); got != `file "/path/with \"quotes\"" read` {
			t.Errorf("Error() = %q, want %q", got, `file "/path/with \"quotes\"" read`)
		}
	})

	t.Run("complex values in validation", func(t *testing.T) {
		type Complex struct {
			Name string
			Age  int
		}
		err := NewValidationError("user", "type", nil).WithValue(Complex{Name: "test", Age: 25})
		got := err.Error()
		if got == "" {
			t.Error("Error() returned empty string")
		}
		// Just verify it doesn't panic and returns something
	})
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkErrorCreation(b *testing.B) {
	b.Run("SpecError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewSpecError("test-spec", "read", io.EOF)
		}
	})

	b.Run("LoopError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewLoopError("loop-1", "execute", io.EOF)
		}
	})

	b.Run("FileError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewFileError("/path/to/file.txt", "read", io.EOF)
		}
	})

	b.Run("ValidationError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewValidationError("email", "format", nil)
		}
	})
}

func BenchmarkErrorIs(b *testing.B) {
	err := NewSpecError("test-spec", "read", ErrSpecNotFound)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = errors.Is(err, ErrSpecNotFound)
	}
}

func BenchmarkErrorAs(b *testing.B) {
	err := NewSpecError("test-spec", "read", io.EOF)
	var specErr SpecError

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = errors.As(err, &specErr)
	}
}
