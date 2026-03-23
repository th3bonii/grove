// Package sdd provides integration tests for SDD client with mocked LLM.
package sdd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock LLM Client for Testing
// =============================================================================

// MockLLMClient implements LLM client interface for testing.
type MockLLMClient struct {
	responses map[string]string
	err       error
}

// NewMockLLMClient creates a new mock LLM client.
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		responses: make(map[string]string),
	}
}

// WithResponse adds a canned response for a prompt pattern.
func (m *MockLLMClient) WithResponse(promptPattern, response string) *MockLLMClient {
	m.responses[promptPattern] = response
	return m
}

// WithError sets a fixed error to return.
func (m *MockLLMClient) WithError(err error) *MockLLMClient {
	m.err = err
	return m
}

// Send implements the LLM client interface.
func (m *MockLLMClient) Send(ctx context.Context, prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	// Return matching response or default
	if response, ok := m.responses[prompt]; ok {
		return response, nil
	}
	// Return any matching prefix response
	for pattern, response := range m.responses {
		if len(prompt) >= len(pattern) && prompt[:len(pattern)] == pattern {
			return response, nil
		}
	}
	return "Mock response: explored the topic and found relevant patterns", nil
}

// SendWithMessages implements the LLM client interface.
func (m *MockLLMClient) SendWithMessages(ctx context.Context, messages []LLMMessage) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	// Build prompt from messages for matching
	var prompt string
	for _, msg := range messages {
		prompt += msg.Content
	}
	return m.Send(ctx, prompt)
}

// Stream implements the LLM client interface (not used in tests).
func (m *MockLLMClient) Stream(ctx context.Context, prompt string, callback StreamCallback) error {
	if m.err != nil {
		return m.err
	}
	response, err := m.Send(ctx, prompt)
	if err != nil {
		return err
	}
	return callback(response)
}

// MockSDDClient is a mock implementation of SDD client for loop testing.
type MockSDDClient struct {
	executeFunc func(ctx context.Context, phase Phase, input map[string]interface{}) (*Result, error)
}

// NewMockSDDClient creates a new mock SDD client.
func NewMockSDDClient() *MockSDDClient {
	return &MockSDDClient{
		executeFunc: func(ctx context.Context, phase Phase, input map[string]interface{}) (*Result, error) {
			return &Result{
				Phase:     phase,
				Status:    "success",
				Summary:   "Mock execution completed",
				Artifacts: []string{"mock-artifact.md"},
			}, nil
		},
	}
}

// WithExecuteFunc sets a custom execute function.
func (m *MockSDDClient) WithExecuteFunc(fn func(ctx context.Context, phase Phase, input map[string]interface{}) (*Result, error)) *MockSDDClient {
	m.executeFunc = fn
	return m
}

// Execute implements the SDD client interface.
func (m *MockSDDClient) Execute(ctx context.Context, phase Phase, input map[string]interface{}) (*Result, error) {
	return m.executeFunc(ctx, phase, input)
}

// =============================================================================
// SDD Integration Tests with Mocks
// =============================================================================

// TestSDDClientExecuteExplore_Mock tests explore phase with mocked LLM.
func TestSDDClientExecuteExplore_Mock(t *testing.T) {
	// Create mock LLM client
	mockLLM := NewMockLLMClient().
		WithResponse("explore", "Exploration findings: analyzed codebase patterns for auth module")

	// Create SDD client with mock LLM
	client := NewClientWithLLM("/test/project", mockLLM)
	require.NotNil(t, client)

	// Execute explore phase
	input := map[string]interface{}{
		"topic":       "user authentication",
		"description": "Add JWT-based auth",
		"project_dir": "/test/project",
	}

	result, err := client.Execute(context.Background(), PhaseExplore, input)

	// Should not panic and should return result
	require.NoError(t, err, "Execute should not panic")
	require.NotNil(t, result)
	require.Equal(t, PhaseExplore, result.Phase)
}

// TestSDDClientExecutePropose_Mock tests propose phase with mocked LLM.
func TestSDDClientExecutePropose_Mock(t *testing.T) {
	mockLLM := NewMockLLMClient().
		WithResponse("propose", "Proposal: implement JWT auth with login/logout endpoints")

	client := NewClientWithLLM("/test/project", mockLLM)
	require.NotNil(t, client)

	input := map[string]interface{}{
		"topic":       "user authentication",
		"exploration": "Found existing auth patterns in codebase",
	}

	result, err := client.Execute(context.Background(), PhasePropose, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, PhasePropose, result.Phase)
}

// TestSDDClientExecuteSpec_Mock tests spec phase with mocked LLM.
func TestSDDClientExecuteSpec_Mock(t *testing.T) {
	mockLLM := NewMockLLMClient().
		WithResponse("spec", "Specification: JWT tokens, 1h expiry, refresh token support")

	client := NewClientWithLLM("/test/project", mockLLM)
	require.NotNil(t, client)

	input := map[string]interface{}{
		"topic":    "user authentication",
		"proposal": "Implement JWT auth",
	}

	result, err := client.Execute(context.Background(), PhaseSpec, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, PhaseSpec, result.Phase)
}

// TestSDDClientExecuteDesign_Mock tests design phase with mocked LLM.
func TestSDDClientExecuteDesign_Mock(t *testing.T) {
	mockLLM := NewMockLLMClient().
		WithResponse("design", "Design: middleware pattern, JWT validation in interceptor")

	client := NewClientWithLLM("/test/project", mockLLM)
	require.NotNil(t, client)

	input := map[string]interface{}{
		"topic": "user authentication",
		"spec":  "JWT tokens with 1h expiry",
	}

	result, err := client.Execute(context.Background(), PhaseDesign, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, PhaseDesign, result.Phase)
}

// TestSDDClientExecuteTasks_Mock tests tasks phase with mocked LLM.
func TestSDDClientExecuteTasks_Mock(t *testing.T) {
	mockLLM := NewMockLLMClient().
		WithResponse("tasks", "Tasks: 1) Setup JWT middleware, 2) Add login endpoint, 3) Add logout")

	client := NewClientWithLLM("/test/project", mockLLM)
	require.NotNil(t, client)

	input := map[string]interface{}{
		"topic":  "user authentication",
		"design": "Middleware pattern with JWT",
	}

	result, err := client.Execute(context.Background(), PhaseTasks, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, PhaseTasks, result.Phase)
}

// TestSDDClientExecuteApply_Mock tests apply phase with mocked LLM.
func TestSDDClientExecuteApply_Mock(t *testing.T) {
	mockLLM := NewMockLLMClient().
		WithResponse("apply", "Applied: created auth/middleware.go, auth/login.go")

	client := NewClientWithLLM("/test/project", mockLLM)
	require.NotNil(t, client)

	input := map[string]interface{}{
		"task_id":     "task-1",
		"task_name":   "Setup JWT middleware",
		"description": "Create JWT validation middleware",
		"spec":        "JWT with 1h expiry",
		"design":      "Middleware pattern",
	}

	result, err := client.Execute(context.Background(), PhaseApply, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, PhaseApply, result.Phase)
}

// TestSDDClientExecuteVerify_Mock tests verify phase with mocked LLM.
func TestSDDClientExecuteVerify_Mock(t *testing.T) {
	mockLLM := NewMockLLMClient().
		WithResponse("verify", "Verification: PASS - all requirements implemented correctly")

	client := NewClientWithLLM("/test/project", mockLLM)
	require.NotNil(t, client)

	input := map[string]interface{}{
		"task_id":        "task-1",
		"implementation": "Created auth/middleware.go",
		"spec":           "JWT with 1h expiry",
	}

	result, err := client.Execute(context.Background(), PhaseVerify, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, PhaseVerify, result.Phase)
}

// TestSDDClientExecuteArchive_Mock tests archive phase with mocked LLM.
func TestSDDClientExecuteArchive_Mock(t *testing.T) {
	mockLLM := NewMockLLMClient().
		WithResponse("archive", "Archive: synced SPEC.md to main, generated archive-report.md")

	client := NewClientWithLLM("/test/project", mockLLM)
	require.NotNil(t, client)

	input := map[string]interface{}{
		"topic": "user authentication",
	}

	result, err := client.Execute(context.Background(), PhaseArchive, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, PhaseArchive, result.Phase)
}

// TestSDDClientLLMErrorHandling tests error handling when LLM fails.
func TestSDDClientLLMErrorHandling(t *testing.T) {
	mockLLM := NewMockLLMClient().
		WithError(errors.New("mock LLM error: rate limit exceeded"))

	client := NewClientWithLLM("/test/project", mockLLM)
	require.NotNil(t, client)

	input := map[string]interface{}{
		"topic": "test topic",
	}

	result, err := client.Execute(context.Background(), PhaseExplore, input)

	// Should handle error gracefully, not panic
	require.NoError(t, err, "Should handle LLM error gracefully")
	require.NotNil(t, result)
	assert.Equal(t, PhaseExplore, result.Phase)
	assert.Equal(t, "failure", result.Status, "Should return failure status on LLM error")
	// Error code is "llm_error" (lowercase)
	assert.Contains(t, result.Error, "llm", "Error should mention llm")
}

// TestSDDClientWithNilLLM tests behavior when LLM is not provided.
func TestSDDClientWithNilLLM(t *testing.T) {
	// Client without LLM - will try to initialize from env
	client := NewClient("/test/project")
	require.NotNil(t, client)

	// Execute should handle missing LLM gracefully (skill check happens first)
	input := map[string]interface{}{
		"topic": "test",
	}

	result, _ := client.Execute(context.Background(), PhaseExplore, input)

	// Either succeeds or returns graceful failure - no panic
	require.NotNil(t, result)
}

// TestSDDClientResultStructure tests the result structure fields.
func TestSDDClientResultStructure(t *testing.T) {
	mockLLM := NewMockLLMClient().
		WithResponse("explore", "test response")

	client := NewClientWithLLM("/test/project", mockLLM)
	require.NotNil(t, client)

	result, err := client.Execute(context.Background(), PhaseExplore, map[string]interface{}{
		"topic": "test",
	})

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify all result fields are populated
	assert.NotEmpty(t, result.Phase)
	assert.NotEmpty(t, result.Status)
	assert.Greater(t, result.Duration.Milliseconds(), int64(0), "Duration should be set")
}

// TestMockSDDClientIntegration tests the mock for loop integration.
func TestMockSDDClientIntegration(t *testing.T) {
	mockSDD := NewMockSDDClient()

	// Test with default execute function
	result, err := mockSDD.Execute(context.Background(), PhaseApply, map[string]interface{}{
		"task_id": "test-1",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "success", result.Status)

	// Test with custom execute function
	mockSDD.WithExecuteFunc(func(ctx context.Context, phase Phase, input map[string]interface{}) (*Result, error) {
		return &Result{
			Phase:   phase,
			Status:  "failure",
			Summary: "Custom failure",
			Error:   "custom_error",
		}, errors.New("custom error")
	})

	result, err = mockSDD.Execute(context.Background(), PhaseApply, nil)

	require.Error(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "failure", result.Status)
}

// TestSDDPromptBuilding tests that prompts are built correctly.
func TestSDDPromptBuilding(t *testing.T) {
	mockLLM := NewMockLLMClient()
	client := NewClientWithLLM("/test/project", mockLLM)

	// Test explore prompt building
	prompt := client.buildExplorePrompt("# Skill content", map[string]interface{}{
		"topic":       "auth",
		"description": "Add login",
	})

	assert.Contains(t, prompt, "auth")
	assert.Contains(t, prompt, "Add login")
	assert.Contains(t, prompt, "SKILL.md")
}

// TestSDDPhaseConstants verifies phase constants are correct.
func TestSDDPhaseConstants(t *testing.T) {
	assert.Equal(t, Phase("explore"), PhaseExplore)
	assert.Equal(t, Phase("propose"), PhasePropose)
	assert.Equal(t, Phase("spec"), PhaseSpec)
	assert.Equal(t, Phase("design"), PhaseDesign)
	assert.Equal(t, Phase("tasks"), PhaseTasks)
	assert.Equal(t, Phase("apply"), PhaseApply)
	assert.Equal(t, Phase("verify"), PhaseVerify)
	assert.Equal(t, Phase("archive"), PhaseArchive)
}

// TestSDDResponseParsers tests the response parsers don't panic.
func TestSDDResponseParsers(t *testing.T) {
	client := NewClient("/test/project")

	// Test each parser with sample content
	tests := []struct {
		name    string
		parseFn func(string) []string
		input   string
	}{
		{"parseExploreResponse", client.parseExploreResponse, "explore content"},
		{"parseProposeResponse", client.parseProposeResponse, "propose content"},
		{"parseSpecResponse", client.parseSpecResponse, "spec content"},
		{"parseDesignResponse", client.parseDesignResponse, "design content"},
		{"parseTasksResponse", client.parseTasksResponse, "tasks content"},
		{"parseApplyResponse", func(s string) []string { return client.parseApplyResponse(s, "task-1") }, "apply content"},
		{"parseArchiveResponse", client.parseArchiveResponse, "archive content"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := tt.parseFn(tt.input)
			assert.NotNil(t, result)
		})
	}
}

// TestSDDVerifyResponseParser tests verify response parsing.
func TestSDDVerifyResponseParser(t *testing.T) {
	client := NewClient("/test/project")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"pass", "This task PASSES all requirements", "PASS"},
		{"fail", "This task FAILS due to missing tests", "FAIL"},
		{"warning", "This task has some warnings", "WARNING"},
		{"mixed pass", "passed but with issues", "PASS"},  // Parser returns PASS because it finds "pass"
		{"mixed fail", "failed with some passes", "FAIL"}, // Parser returns FAIL because it finds "fail"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.parseVerifyResponse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSDDArtifactSaving tests artifact saving doesn't panic.
func TestSDDArtifactSaving(t *testing.T) {
	client := NewClient("/test/project")

	// Test with various artifact paths
	artifacts := []string{
		"/test/project/spec/explore.md",
		"/test/project/spec/SPEC.md",
		"/test/project/spec/DESIGN.md",
	}

	saved := client.saveArtifacts(PhaseExplore, artifacts)

	assert.Equal(t, len(artifacts), len(saved))
	for i, a := range artifacts {
		assert.Equal(t, a, saved[i])
	}
}

// TestSDDClientContextCancellation tests graceful handling of context cancellation.
func TestSDDClientContextCancellation(t *testing.T) {
	mockLLM := NewMockLLMClient().WithResponse("test", "response")

	client := NewClientWithLLM("/test/project", mockLLM)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := client.Execute(ctx, PhaseExplore, map[string]interface{}{"topic": "test"})

	// Should handle cancelled context gracefully
	require.NoError(t, err, "Should handle context cancellation")
	require.NotNil(t, result)
}

// TestSDDWithEmptyInput tests handling of empty inputs.
func TestSDDWithEmptyInput(t *testing.T) {
	mockLLM := NewMockLLMClient().WithResponse("explore", "explored")

	client := NewClientWithLLM("/test/project", mockLLM)

	// Test with empty input
	result, err := client.Execute(context.Background(), PhaseExplore, nil)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, PhaseExplore, result.Phase)

	// Test with empty topic
	result, err = client.Execute(context.Background(), PhaseExplore, map[string]interface{}{
		"topic": "",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
}
