// Package loop provides LLM error detection and recovery for Ralph Loop.
package loop

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ErrorType categoriza los errores del loop
type ErrorType string

const (
	ErrorTypeLLMResponse ErrorType = "llm_response" // Respuesta malformada
	ErrorTypeNetwork     ErrorType = "network"      // Problemas de red
	ErrorTypeRateLimit   ErrorType = "rate_limit"   // Rate limiting
	ErrorTypeTimeout     ErrorType = "timeout"      // Timeout
	ErrorTypeAuth        ErrorType = "auth"         // Error de autenticación
	ErrorTypeFileSystem  ErrorType = "file_system"  // Problemas de archivo
	ErrorTypeUnknown     ErrorType = "unknown"
)

// ErrorClassifier clasifica errores por tipo
type ErrorClassifier struct{}

func NewErrorClassifier() *ErrorClassifier {
	return &ErrorClassifier{}
}

// Classify clasifica un error y retorna el tipo
func (c *ErrorClassifier) Classify(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	// Detectar errores de red
	if isNetworkError(err) {
		return ErrorTypeNetwork
	}

	// Detectar rate limiting
	if isRateLimitError(err) {
		return ErrorTypeRateLimit
	}

	// Detectar timeout
	if isTimeoutError(err) {
		return ErrorTypeTimeout
	}

	// Detectar errores de auth
	if isAuthError(err) {
		return ErrorTypeAuth
	}

	// Detectar errores de LLM
	if isLLMError(err) {
		return ErrorTypeLLMResponse
	}

	// Detectar errores de sistema de archivos
	if isFileSystemError(err) {
		return ErrorTypeFileSystem
	}

	return ErrorTypeUnknown
}

// IsRetryable indica si un error debe ser reintentado
func (c *ErrorClassifier) IsRetryable(err error) bool {
	et := c.Classify(err)
	switch et {
	case ErrorTypeNetwork, ErrorTypeRateLimit, ErrorTypeTimeout:
		return true
	case ErrorTypeLLMResponse:
		return true // A veces son errores transitorios
	case ErrorTypeAuth, ErrorTypeFileSystem:
		return false // No se resuelven con retry
	default:
		return true // Por defecto, intentar retry
	}
}

// ErrorTypeString retorna la representación en string del tipo de error
func (et ErrorType) String() string {
	return string(et)
}

// RetryStrategy define cómo manejar retries
type RetryStrategy struct {
	MaxRetries       int
	BackoffBase      time.Duration
	ContextReduction float64 // 0.0-1.0 para reducir contexto
}

// Default values for RetryStrategy
const (
	DefaultMaxRetries  = 3
	DefaultBackoffBase = 1 * time.Second
)

// NewRetryStrategy crea una nueva RetryStrategy con valores por defecto
func NewRetryStrategy() *RetryStrategy {
	return &RetryStrategy{
		MaxRetries:       DefaultMaxRetries,
		BackoffBase:      DefaultBackoffBase,
		ContextReduction: 0.0,
	}
}

// NewRetryStrategyWithConfig crea una RetryStrategy con configuración específica
func NewRetryStrategyWithConfig(maxRetries int, backoffBaseMs int64, contextReduction float64) *RetryStrategy {
	backoffBase := time.Duration(backoffBaseMs) * time.Millisecond
	if backoffBase <= 0 {
		backoffBase = DefaultBackoffBase
	}
	return &RetryStrategy{
		MaxRetries:       maxRetries,
		BackoffBase:      backoffBase,
		ContextReduction: contextReduction,
	}
}

// Apply aplica la estrategia de retry con backoff
func (s *RetryStrategy) Apply(ctx context.Context, attempt int, op func() error) error {
	if attempt >= s.MaxRetries {
		return fmt.Errorf("max retries (%d) exceeded", s.MaxRetries)
	}

	// Exponential backoff: base * 2^attempt
	backoff := s.BackoffBase * time.Duration(1<<attempt)
	if backoff > 60*time.Second {
		backoff = 60 * time.Second // Cap at 60 seconds
	}

	select {
	case <-time.After(backoff):
		return op()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// BackoffDuration calcula el tiempo de espera para un intento específico
func (s *RetryStrategy) BackoffDuration(attempt int) time.Duration {
	backoff := s.BackoffBase * time.Duration(1<<attempt)
	if backoff > 60*time.Second {
		return 60 * time.Second
	}
	return backoff
}

// RetryResult contiene el resultado de una operación con retry
type RetryResult struct {
	Success   bool
	Error     error
	Attempts  int
	LastError error
	ErrorType ErrorType
}

// RetryWithStrategy ejecuta una operación con reintentos usando la estrategia dada
func (s *RetryStrategy) RetryWithStrategy(ctx context.Context, op func() error) *RetryResult {
	classifier := NewErrorClassifier()
	result := &RetryResult{
		Success:  false,
		Attempts: 0,
	}

	var lastErr error
	for attempt := 0; attempt <= s.MaxRetries; attempt++ {
		result.Attempts = attempt + 1

		if attempt > 0 {
			// Aplicar backoff
			if err := s.Apply(ctx, attempt, func() error { return nil }); err != nil {
				result.Error = err
				result.LastError = err
				return result
			}
		}

		err := op()
		if err == nil {
			result.Success = true
			return result
		}

		lastErr = err
		result.LastError = err

		// Clasificar error
		errType := classifier.Classify(err)
		result.ErrorType = errType

		if !classifier.IsRetryable(err) {
			// Error no retryable, salir inmediatamente
			result.Error = err
			return result
		}

		// Log del retry
		fmt.Printf("Retry attempt %d/%d for error: %v (type: %s)\n", attempt+1, s.MaxRetries+1, err, errType)
	}

	result.Error = fmt.Errorf("max retries (%d) exceeded, last error: %w", s.MaxRetries, lastErr)
	return result
}

// ErrRetryExhausted representa un error cuando se agotan los retries
type ErrRetryExhausted struct {
	Attempts  int
	LastError error
	ErrorType ErrorType
	Retryable bool
}

func (e *ErrRetryExhausted) Error() string {
	return fmt.Sprintf("retry exhausted after %d attempts, last error: %v", e.Attempts, e.LastError)
}

func (e *ErrRetryExhausted) Unwrap() error {
	return e.LastError
}

// NewErrRetryExhausted crea un nuevo error de retry agotado
func NewErrRetryExhausted(attempts int, lastError error, errorType ErrorType, retryable bool) *ErrRetryExhausted {
	return &ErrRetryExhausted{
		Attempts:  attempts,
		LastError: lastError,
		ErrorType: errorType,
		Retryable: retryable,
	}
}

// Funciones helper para detección de tipos de error

func isNetworkError(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Temporary() || urlErr.Timeout()
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "network is unreachable")
}

func isRateLimitError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "too many requests") ||
		strings.Contains(errStr, "rate_limit") ||
		strings.Contains(errStr, "quota exceeded")
}

func isTimeoutError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "i/o timeout")
}

func isAuthError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "api key") ||
		strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "403")
}

func isLLMError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "llm") ||
		strings.Contains(errStr, "model") ||
		strings.Contains(errStr, "anthropic") ||
		strings.Contains(errStr, "openai") ||
		strings.Contains(errStr, "google.ai") ||
		strings.Contains(errStr, "generation") ||
		strings.Contains(errStr, "content filter")
}

func isFileSystemError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "no such file") ||
		strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "read-only") ||
		strings.Contains(errStr, "disk full") ||
		strings.Contains(errStr, "i/o error")
}

// IsRetryableError es una función standalone para verificar si un error es reintentable
// Esta es la versión que se usará en el orchestrator
func IsRetryableError(err error) bool {
	classifier := NewErrorClassifier()
	return classifier.IsRetryable(err)
}

// ClassifyError clasifica un error y retorna su tipo
// Esta es la versión standalone para uso en el orchestrator
func ClassifyError(err error) ErrorType {
	classifier := NewErrorClassifier()
	return classifier.Classify(err)
}
