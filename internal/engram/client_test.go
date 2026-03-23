// Package engram provides integration tests with mock Engram server.
package engram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock Engram Server
// =============================================================================

// MockEngramServer simulates an Engram HTTP API server for testing.
type MockEngramServer struct {
	mu      sync.RWMutex
	data    map[string]interface{}
	queries map[string][]string // search index
	server  *httptest.Server
	t       *testing.T
}

// NewMockEngramServer creates a new mock Engram server.
func NewMockEngramServer(t *testing.T) *MockEngramServer {
	m := &MockEngramServer{
		data:    make(map[string]interface{}),
		queries: make(map[string][]string),
		t:       t,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/memory/", m.handleMemory)
	mux.HandleFunc("/api/search", m.handleSearch)
	mux.HandleFunc("/api/keys", m.handleKeys)

	m.server = httptest.NewServer(mux)
	return m
}

// URL returns the base URL of the mock server.
func (m *MockEngramServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server.
func (m *MockEngramServer) Close() {
	m.server.Close()
}

// Client returns an EngramClient configured to use the mock server.
func (m *MockEngramServer) Client() *EngramClient {
	client := NewClientWithConfig(
		"invalid-host-for-testing", // won't be used
		80,
		5*time.Second,
	)
	client.SetBaseURL(m.server.URL)
	return client
}

// handleMemory handles /api/memory/* endpoints.
func (m *MockEngramServer) handleMemory(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	key := path[len("/api/memory/"):]

	switch r.Method {
	case "GET":
		m.handleGet(w, key)
	case "POST":
		m.handlePost(w, r, key)
	case "DELETE":
		m.handleDelete(w, key)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (m *MockEngramServer) handleGet(w http.ResponseWriter, key string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.data[key]
	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"value": val})
}

func (m *MockEngramServer) handlePost(w http.ResponseWriter, r *http.Request, key string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Value interface{} `json:"value"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	m.data[key] = req.Value

	// Update search index - index individual words
	for _, word := range extractWords(key) {
		m.queries[word] = append(m.queries[word], key)
	}
	// Also index the full key for exact matches
	m.queries[key] = append(m.queries[key], key)

	// Index path prefixes (e.g., "spec-decision" from "spec-decision/auth-model")
	for _, prefix := range extractPathPrefixes(key) {
		m.queries[prefix] = append(m.queries[prefix], key)
	}

	m.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
}

func (m *MockEngramServer) handleDelete(w http.ResponseWriter, key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.data[key]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	delete(m.data, key)
	w.WriteHeader(http.StatusOK)
}

// handleSearch handles /api/search endpoint.
func (m *MockEngramServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	m.mu.RLock()
	results := m.queries[req.Query]
	m.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string][]string{"results": results})
}

// handleKeys handles /api/keys endpoint.
func (m *MockEngramServer) handleKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	m.mu.RLock()
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	m.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string][]string{"keys": keys})
}

// extractWords splits a key into searchable words.
func extractWords(s string) []string {
	// Simple word extraction - split on common separators
	var words []string
	word := ""
	for _, c := range s {
		if c == '/' || c == '_' || c == '-' {
			if word != "" {
				words = append(words, word)
			}
			word = ""
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

// extractPathPrefixes extracts all path prefixes from a key.
// For example: "spec-decision/auth-model/dec-001" -> ["spec-decision", "spec-decision/auth-model"]
func extractPathPrefixes(key string) []string {
	var prefixes []string
	parts := splitPath(key)
	for i := 1; i < len(parts); i++ {
		prefixes = append(prefixes, joinPath(parts[:i]))
	}
	return prefixes
}

// splitPath splits a key into path components.
func splitPath(key string) []string {
	var parts []string
	var current string
	for _, c := range key {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// joinPath joins path components into a single string.
func joinPath(parts []string) string {
	result := parts[0]
	for _, p := range parts[1:] {
		result += "/" + p
	}
	return result
}

// =============================================================================
// Failing Mock Server
// =============================================================================

// NewFailingMockServer creates a mock server that always fails.
func NewFailingMockServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate service unavailable
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	})
	return httptest.NewServer(mux)
}

// =============================================================================
// Client Tests
// =============================================================================

func TestNewClient(t *testing.T) {
	client := NewClient("localhost")
	assert.NotNil(t, client)
	assert.Equal(t, "localhost", client.host)
	assert.Equal(t, DefaultEngramPort, client.port)
}

func TestNewClientWithConfig(t *testing.T) {
	client := NewClientWithConfig("example.com", 8080, 10*time.Second)
	assert.NotNil(t, client)
	assert.Equal(t, "example.com", client.host)
	assert.Equal(t, 8080, client.port)
	assert.Equal(t, 10*time.Second, client.timeout)
}

func TestClient_URL(t *testing.T) {
	client := NewClient("localhost")
	assert.Equal(t, "http://localhost:7437", client.URL())

	client = NewClientWithConfig("example.com", 8080, time.Second)
	assert.Equal(t, "http://example.com:8080", client.URL())
}

func TestClient_SaveAndLoad(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()

	// Override the host to use the full URL (strip http://)
	client.SetBaseURL(mock.server.URL)

	// Test save and load string
	err := client.Save("test-key", "test-value")
	require.NoError(t, err)

	val, err := client.Load("test-key")
	require.NoError(t, err)
	assert.Equal(t, "test-value", val)

	// Test save and load complex object
	complex := map[string]interface{}{
		"name":   "test",
		"count":  42,
		"nested": map[string]string{"key": "value"},
	}
	err = client.Save("complex-key", complex)
	require.NoError(t, err)

	val, err = client.Load("complex-key")
	require.NoError(t, err)

	loaded, ok := val.(map[string]interface{})
	require.True(t, ok, "expected map[string]interface{}")
	assert.Equal(t, "test", loaded["name"])
	assert.Equal(t, float64(42), loaded["count"])
}

func TestClient_LoadNotFound(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	_, err := client.Load("nonexistent-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestClient_Delete(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	// Save a key
	err := client.Save("delete-me", "value")
	require.NoError(t, err)

	// Delete it
	err = client.Delete("delete-me")
	require.NoError(t, err)

	// Verify it's gone
	_, err = client.Load("delete-me")
	assert.Error(t, err)
}

func TestClient_DeleteNotFound(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	// Deleting non-existent key should not error
	err := client.Delete("nonexistent")
	require.NoError(t, err)
}

func TestClient_List(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	// Add some keys
	client.Save("key1", "value1")
	client.Save("key2", "value2")
	client.Save("key3", "value3")

	// List should return all keys
	keys, err := client.List()
	require.NoError(t, err)
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
	assert.Contains(t, keys, "key3")
}

func TestClient_Search(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	// Add keys with searchable components
	client.Save("spec-decision/auth-model", "value1")
	client.Save("spec-decision/db-schema", "value2")
	client.Save("loop-checkpoint/main", "value3")

	// Search for spec decisions
	keys, err := client.Search("spec-decision")
	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestClient_SaveEmptyKey(t *testing.T) {
	client := NewClient("localhost")
	err := client.Save("", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key")
	assert.Contains(t, err.Error(), "required")
}

func TestClient_LoadEmptyKey(t *testing.T) {
	client := NewClient("localhost")
	_, err := client.Load("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key")
	assert.Contains(t, err.Error(), "required")
}

func TestClient_SearchEmptyQuery(t *testing.T) {
	client := NewClient("localhost")
	_, err := client.Search("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query")
	assert.Contains(t, err.Error(), "required")
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestIntegration_SaveAndLoadSpecDecisions(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	decision := &SpecDecision{
		ID:            "dec-001",
		ChangeName:    "feature-auth",
		Decision:      "Use JWT for authentication",
		Justification: "Stateless and scalable",
		Alternatives:  []string{"Sessions", "OAuth2"},
		Timestamp:     time.Now(),
	}

	err := client.SaveSpecDecision("feature-auth", decision)
	require.NoError(t, err)

	decisions, err := client.LoadSpecDecisions("feature-auth")
	require.NoError(t, err)
	assert.Len(t, decisions, 1)
	assert.Equal(t, "Use JWT for authentication", decisions[0].Decision)
}

func TestIntegration_LoopCheckpoint(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	checkpoint := NewLoopCheckpoint("my-change", 3, "implementation")
	checkpoint.State = map[string]interface{}{
		"currentFile": "main.go",
		"lineNumber":  42,
	}
	checkpoint.Artifacts = []string{"file1.go", "file2.go"}

	err := client.SaveLoopCheckpoint(checkpoint)
	require.NoError(t, err)

	loaded, err := client.LoadLoopCheckpoint("my-change")
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, 3, loaded.LoopNumber)
	assert.Equal(t, "implementation", loaded.Phase)
	assert.Equal(t, "main.go", loaded.State["currentFile"])
}

func TestIntegration_LoopCheckpointExpiration(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	// Create expired checkpoint
	checkpoint := &LoopCheckpoint{
		ChangeName: "expired-change",
		LoopNumber: 1,
		Phase:      "spec",
		CreatedAt:  time.Now().Add(-48 * time.Hour),
		ExpiresAt:  time.Now().Add(-24 * time.Hour), // Expired
	}

	err := client.SaveLoopCheckpoint(checkpoint)
	require.NoError(t, err)

	// Loading should return nil for expired checkpoint
	loaded, err := client.LoadLoopCheckpoint("expired-change")
	require.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestIntegration_OptiPattern(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	pattern := &OptiPattern{
		ID:          "pat-001",
		Name:        "Use file references",
		Category:    "context",
		Pattern:     "Reference files with {{file:path}}",
		Description: "Include file paths in prompts for better context",
		Examples:    []string{"See {{file:src/main.go}}"},
		SuccessRate: 0.85,
		UsageCount:  10,
	}

	err := client.SaveOptiPattern(pattern)
	require.NoError(t, err)

	patterns, err := client.LoadOptiPatterns("context")
	require.NoError(t, err)
	assert.Len(t, patterns, 1)
	assert.Equal(t, "Use file references", patterns[0].Name)
}

func TestIntegration_MultipleOptiPatterns(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	patterns := []OptiPattern{
		{ID: "p1", Name: "Pattern 1", Category: "test", SuccessRate: 0.9},
		{ID: "p2", Name: "Pattern 2", Category: "test", SuccessRate: 0.7},
		{ID: "p3", Name: "Pattern 3", Category: "test", SuccessRate: 0.8},
	}

	err := client.SaveOptiPatterns(patterns)
	require.NoError(t, err)

	// Get top 2
	best, err := client.GetBestPatterns("test", 2)
	require.NoError(t, err)
	assert.Len(t, best, 2)
	assert.Equal(t, "Pattern 1", best[0].Name) // Highest success rate
}

func TestIntegration_ChangeMetadata(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	meta := &ChangeMetadata{
		ChangeName: "new-feature",
		Project:    "my-project",
		Status:     "specifying",
		Artifacts:  []string{"spec.json", "design.md"},
		CreatedAt:  time.Now(),
	}

	err := client.SaveChangeMetadata(meta)
	require.NoError(t, err)

	loaded, err := client.LoadChangeMetadata("new-feature")
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, "specifying", loaded.Status)
	assert.Len(t, loaded.Artifacts, 2)
}

func TestIntegration_ListChanges(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	// Add multiple changes
	client.SaveChangeMetadata(&ChangeMetadata{ChangeName: "change1", Project: "proj", Status: "specifying"})
	client.SaveChangeMetadata(&ChangeMetadata{ChangeName: "change2", Project: "proj", Status: "implementing"})
	client.SaveChangeMetadata(&ChangeMetadata{ChangeName: "change3", Project: "other", Status: "archived"})

	changes, err := client.ListChanges("proj")
	require.NoError(t, err)
	assert.Len(t, changes, 2) // Only proj changes
}

func TestIntegration_SessionSummary(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	summary := &SessionSummary{
		SessionID:    "session-123",
		Project:      "test-project",
		Goal:         "Implement auth feature",
		Discoveries:  []string{"Found JWT library", "Need middleware"},
		Accomplished: []string{"Created spec", "Designed architecture"},
		NextSteps:    []string{"Implement handlers", "Write tests"},
		Timestamp:    time.Now(),
	}

	err := client.SaveSessionSummary(summary)
	require.NoError(t, err)

	loaded, err := client.LoadSessionSummary("session-123")
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, "Implement auth feature", loaded.Goal)
	assert.Len(t, loaded.Discoveries, 2)
}

func TestIntegration_GetRecentSessions(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	// Add sessions with different timestamps
	client.SaveSessionSummary(&SessionSummary{
		SessionID: "s1", Project: "proj", Goal: "Goal 1", Timestamp: time.Now().Add(-1 * time.Hour),
	})
	client.SaveSessionSummary(&SessionSummary{
		SessionID: "s2", Project: "proj", Goal: "Goal 2", Timestamp: time.Now(),
	})
	client.SaveSessionSummary(&SessionSummary{
		SessionID: "s3", Project: "other", Goal: "Goal 3", Timestamp: time.Now(),
	})

	sessions, err := client.GetRecentSessions("proj", 10)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
	// Most recent first
	assert.Equal(t, "Goal 2", sessions[0].Goal)
}

// =============================================================================
// Graceful Degradation Tests
// =============================================================================

func TestGracefulDegradation_SaveWhenUnavailable(t *testing.T) {
	failing := NewFailingMockServer(t)
	defer failing.Close()

	// Create client pointing to failing server
	client := NewClientWithConfig("test", 80, time.Second)
	client.SetBaseURL(failing.URL)

	err := client.Save("test", "value")
	assert.Error(t, err)
	assert.True(t, IsEngramUnavailable(err), "expected engram unavailable error")
}

func TestGracefulDegradation_LoadWhenUnavailable(t *testing.T) {
	failing := NewFailingMockServer(t)
	defer failing.Close()

	client := NewClientWithConfig("test", 80, time.Second)
	client.SetBaseURL(failing.URL)

	_, err := client.Load("test")
	assert.Error(t, err)
	assert.True(t, IsEngramUnavailable(err))
}

func TestGracefulDegradation_SearchWhenUnavailable(t *testing.T) {
	failing := NewFailingMockServer(t)
	defer failing.Close()

	client := NewClientWithConfig("test", 80, time.Second)
	client.SetBaseURL(failing.URL)

	_, err := client.Search("test")
	assert.Error(t, err)
	assert.True(t, IsEngramUnavailable(err))
}

func TestGracefulDegradation_LoadSpecDecisionsWhenUnavailable(t *testing.T) {
	failing := NewFailingMockServer(t)
	defer failing.Close()

	client := NewClientWithConfig("test", 80, time.Second)
	client.SetBaseURL(failing.URL)

	decisions, err := client.LoadSpecDecisions("test")
	// Should return empty slice, not error
	assert.NoError(t, err)
	assert.Empty(t, decisions)
}

func TestGracefulDegradation_LoadOptiPatternsWhenUnavailable(t *testing.T) {
	failing := NewFailingMockServer(t)
	defer failing.Close()

	client := NewClientWithConfig("test", 80, time.Second)
	client.SetBaseURL(failing.URL)

	patterns, err := client.LoadOptiPatterns("")
	assert.NoError(t, err)
	assert.Empty(t, patterns)
}

// =============================================================================
// Validation Tests
// =============================================================================

func TestValidation_SaveNilValue(t *testing.T) {
	mock := NewMockEngramServer(t)
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	// This should work (nil value is valid JSON)
	err := client.Save("key", nil)
	assert.NoError(t, err)
}

func TestValidation_SaveSpecDecisionNil(t *testing.T) {
	client := NewClient("localhost")
	err := client.SaveSpecDecision("change", nil)
	assert.Error(t, err)
}

func TestValidation_SaveSpecDecisionEmptyChange(t *testing.T) {
	client := NewClient("localhost")
	err := client.SaveSpecDecision("", &SpecDecision{ID: "test"})
	assert.Error(t, err)
}

func TestValidation_LoopCheckpointNil(t *testing.T) {
	client := NewClient("localhost")
	err := client.SaveLoopCheckpoint(nil)
	assert.Error(t, err)
}

func TestValidation_OptiPatternNil(t *testing.T) {
	client := NewClient("localhost")
	err := client.SaveOptiPattern(nil)
	assert.Error(t, err)
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkClient_Save(b *testing.B) {
	mock := NewMockEngramServer(&testing.T{})
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Save(fmt.Sprintf("bench-%d", i), "value")
	}
}

func BenchmarkClient_Load(b *testing.B) {
	mock := NewMockEngramServer(&testing.T{})
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		client.Save(fmt.Sprintf("bench-%d", i), "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Load(fmt.Sprintf("bench-%d", i%1000))
	}
}

func BenchmarkClient_Search(b *testing.B) {
	mock := NewMockEngramServer(&testing.T{})
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	// Pre-populate with searchable keys
	for i := 0; i < 100; i++ {
		client.Save(fmt.Sprintf("spec-decision/%d", i), "value")
		client.Save(fmt.Sprintf("loop-checkpoint/%d", i), "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Search("spec-decision")
	}
}

func BenchmarkIntegration_SaveSpecDecision(b *testing.B) {
	mock := NewMockEngramServer(&testing.T{})
	defer mock.Close()
	client := mock.Client()
	client.SetBaseURL(mock.server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.SaveSpecDecision("bench-change", &SpecDecision{
			ID:        fmt.Sprintf("dec-%d", i),
			Decision:  "Test decision",
			Timestamp: time.Now(),
		})
	}
}

// =============================================================================
// Context Tests
// =============================================================================

func TestClient_WithContextCancellation(t *testing.T) {
	// Create a slow mock server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClientWithConfig(server.URL[7:], 80, 50*time.Millisecond)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.doRequest(ctx, "GET", "/test", nil)
	// Should fail due to context cancellation
	assert.Error(t, err)
}
