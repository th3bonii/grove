// Package engram provides integration with the Engram persistent memory system.
//
// Engram is a long-term memory system for AI agents that persists across sessions
// and context compactions. This package provides a client for communicating with
// the Engram HTTP API and GROVE-specific integration functions.
//
// # Engram Client
//
// The EngramClient communicates with the Engram HTTP API (default port 7437).
// It provides basic CRUD operations: Save, Load, Search, Delete, and List.
//
// # Graceful Degradation
//
// The client implements graceful degradation - if Engram is unavailable, operations
// return a human-friendly error without crashing. All methods include retry logic
// with exponential backoff (3 attempts by default, 5 second timeout).
//
// # GROVE Integration
//
// The integration functions provide GROVE-specific operations:
//   - SaveSpecDecision/LoadSpecDecisions: Persist specification decisions
//   - SaveLoopCheckpoint/LoadLoopCheckpoint: Persist Ralph Loop state
//   - SaveOptiPatterns/LoadOptiPatterns: Persist prompt optimization patterns
package engram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	gerrors "github.com/Gentleman-Programming/grove/internal/errors"
)

// Default Engram configuration
const (
	DefaultEngramHost = "localhost"
	DefaultEngramPort = 7437
	DefaultTimeout    = 5 * time.Second
	DefaultMaxRetries = 3
	backoffBase       = 100 * time.Millisecond
	backoffMultiplier = 2.0
)

// =============================================================================
// Engram Client
// =============================================================================

// EngramClient provides a client for the Engram HTTP API.
type EngramClient struct {
	host    string
	port    int
	timeout time.Duration
	client  *http.Client
	baseURL string // Custom base URL (for testing with mock servers)
}

// NewClient creates a new EngramClient with the specified host.
// Default port (7437) and timeout (5s) are used.
func NewClient(host string) *EngramClient {
	return NewClientWithConfig(host, DefaultEngramPort, DefaultTimeout)
}

// NewClientWithConfig creates a new EngramClient with custom configuration.
func NewClientWithConfig(host string, port int, timeout time.Duration) *EngramClient {
	return &EngramClient{
		host:    host,
		port:    port,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// SetBaseURL sets a custom base URL for the client (useful for testing with mock servers).
func (c *EngramClient) SetBaseURL(url string) {
	c.baseURL = url
}

// URL returns the base URL for the Engram API.
func (c *EngramClient) URL() string {
	if c.baseURL != "" {
		return c.baseURL
	}
	return fmt.Sprintf("http://%s:%d", c.host, c.port)
}

// =============================================================================
// HTTP Methods with Retry
// =============================================================================

// httpError represents an HTTP-related error.
type httpError struct {
	StatusCode int
	Message    string
	Op         string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("engram %s: HTTP %d - %s", e.Op, e.StatusCode, e.Message)
}

// doRequest performs an HTTP request with retry logic and exponential backoff.
func (c *EngramClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	var lastErr error
	for attempt := 0; attempt < DefaultMaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			sleepTime := time.Duration(float64(backoffBase) * pow(backoffMultiplier, float64(attempt-1)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(sleepTime):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, c.URL()+path, reqBody)
		if err != nil {
			lastErr = fmt.Errorf("create request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("do request: %w", err)
			continue
		}

		// Check for server errors (5xx) - retry
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: HTTP %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	// All retries failed - return last error with context
	return nil, &httpError{
		StatusCode: 0,
		Message:    fmt.Sprintf("max retries exceeded: %v", lastErr),
		Op:         method,
	}
}

// pow calculates base to the power of exp.
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// =============================================================================
// Public API Methods
// =============================================================================

// Save stores a value in Engram with the specified key.
func (c *EngramClient) Save(key string, value interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Validate key
	if key == "" {
		return gerrors.NewValidationError("key", "required", fmt.Errorf("key cannot be empty"))
	}

	reqBody := map[string]interface{}{
		"value": value,
	}

	resp, err := c.doRequest(ctx, "POST", "/api/memory/"+key, reqBody)
	if err != nil {
		// Check if it's a connection error (Engram unavailable)
		if isConnectionError(err) {
			return &engramUnavailableError{Message: "Engram service unavailable - graceful degradation: operations will work without persistent memory"}
		}
		// Check if it's a server error (5xx) after retries - treat as unavailable
		if isServerError(err) {
			return &engramUnavailableError{Message: "Engram service unavailable - graceful degradation: operations will work without persistent memory"}
		}
		return fmt.Errorf("save: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return c.parseErrorResponse(resp, "save")
	}

	return nil
}

// Load retrieves a value from Engram by key.
func (c *EngramClient) Load(key string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Validate key
	if key == "" {
		return nil, gerrors.NewValidationError("key", "required", fmt.Errorf("key cannot be empty"))
	}

	resp, err := c.doRequest(ctx, "GET", "/api/memory/"+key, nil)
	if err != nil {
		if isConnectionError(err) {
			return nil, &engramUnavailableError{Message: "Engram service unavailable - graceful degradation: operations will work without persistent memory"}
		}
		if isServerError(err) {
			return nil, &engramUnavailableError{Message: "Engram service unavailable - graceful degradation: operations will work without persistent memory"}
		}
		return nil, fmt.Errorf("load: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseErrorResponse(resp, "load")
	}

	var result struct {
		Value interface{} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Value, nil
}

// Search queries Engram for keys matching the specified search query.
func (c *EngramClient) Search(query string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Validate query
	if query == "" {
		return nil, gerrors.NewValidationError("query", "required", fmt.Errorf("search query cannot be empty"))
	}

	reqBody := map[string]string{
		"query": query,
	}

	resp, err := c.doRequest(ctx, "POST", "/api/search", reqBody)
	if err != nil {
		if isConnectionError(err) {
			return nil, &engramUnavailableError{Message: "Engram service unavailable - graceful degradation: operations will work without persistent memory"}
		}
		if isServerError(err) {
			return nil, &engramUnavailableError{Message: "Engram service unavailable - graceful degradation: operations will work without persistent memory"}
		}
		return nil, fmt.Errorf("search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseErrorResponse(resp, "search")
	}

	var result struct {
		Results []string `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Results, nil
}

// Delete removes a key from Engram.
func (c *EngramClient) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Validate key
	if key == "" {
		return gerrors.NewValidationError("key", "required", fmt.Errorf("key cannot be empty"))
	}

	resp, err := c.doRequest(ctx, "DELETE", "/api/memory/"+key, nil)
	if err != nil {
		if isConnectionError(err) {
			return &engramUnavailableError{Message: "Engram service unavailable - graceful degradation: operations will work without persistent memory"}
		}
		if isServerError(err) {
			return &engramUnavailableError{Message: "Engram service unavailable - graceful degradation: operations will work without persistent memory"}
		}
		return fmt.Errorf("delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return c.parseErrorResponse(resp, "delete")
	}

	return nil
}

// List returns all keys stored in Engram.
func (c *EngramClient) List() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	resp, err := c.doRequest(ctx, "GET", "/api/keys", nil)
	if err != nil {
		if isConnectionError(err) {
			return nil, &engramUnavailableError{Message: "Engram service unavailable - graceful degradation: operations will work without persistent memory"}
		}
		if isServerError(err) {
			return nil, &engramUnavailableError{Message: "Engram service unavailable - graceful degradation: operations will work without persistent memory"}
		}
		return nil, fmt.Errorf("list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseErrorResponse(resp, "list")
	}

	var result struct {
		Keys []string `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Keys, nil
}

// =============================================================================
// Error Helpers
// =============================================================================

// engramUnavailableError indicates Engram service is not available.
type engramUnavailableError struct {
	Message string
}

func (e *engramUnavailableError) Error() string {
	return e.Message
}

// IsEngramUnavailable checks if an error indicates Engram is unavailable.
func IsEngramUnavailable(err error) bool {
	_, ok := err.(*engramUnavailableError)
	return ok
}

// isConnectionError checks if an error is a connection-related error.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsAny(errStr, []string{"connection refused", "no such host", "timeout", "no route to host", "network is unreachable"})
}

// isServerError checks if an error indicates a server error (5xx) after retries.
func isServerError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Check for server errors in the error message (e.g., "server error: HTTP 503")
	return contains(errStr, "server error: HTTP")
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstring(s, sub))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// parseErrorResponse parses an error response from Engram.
func (c *EngramClient) parseErrorResponse(resp *http.Response, op string) error {
	var errResp struct {
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return &httpError{
			StatusCode: resp.StatusCode,
			Message:    "unknown error",
			Op:         op,
		}
	}

	return &httpError{
		StatusCode: resp.StatusCode,
		Message:    errResp.Error,
		Op:         op,
	}
}
