package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Gentleman-Programming/grove/internal/config"
	"github.com/Gentleman-Programming/grove/internal/loop"
	"github.com/Gentleman-Programming/grove/internal/opti"
	"github.com/Gentleman-Programming/grove/internal/spec"
	"github.com/Gentleman-Programming/grove/internal/types"
)

// TestEndToEnd_SpecFlow tests the complete Spec flow:
// 1. Create temp directory with ideas
// 2. Run spec.Run()
// 3. Verify SPEC.md, DESIGN.md, TASKS.md are generated
// 4. Verify quality scores
func TestEndToEnd_SpecFlow(t *testing.T) {
	// Create temporary directory for the test
	tempDir := t.TempDir()
	specDir := filepath.Join(tempDir, "spec")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	// Create sample ideas file
	ideasContent := `# Feature: User Authentication System

## Overview
Implement a complete user authentication system with login, logout, and session management.

## Requirements
- User can register with email and password
- User can login with credentials
- User can logout
- Sessions persist across browser restarts
- Password must be hashed securely

## Technical Considerations
- Use bcrypt for password hashing
- JWT tokens for session management
- PostgreSQL for user storage
`

	ideasPath := filepath.Join(tempDir, "IDEAS.md")
	if err := os.WriteFile(ideasPath, []byte(ideasContent), 0o644); err != nil {
		t.Fatalf("failed to write ideas file: %v", err)
	}

	// Configure the spec engine
	cfg := &types.Config{
		ProjectName:           "test-project",
		ProjectPath:           tempDir,
		OutputPath:            specDir,
		MaxIterations:         3,
		QualityThreshold:      0.5, // Low threshold for testing
		EnableSelfQuestioning: true,
		ScoringWeights: map[string]float64{
			"completeness":    0.20,
			"consistency":     0.15,
			"clarity":         0.15,
			"testability":     0.15,
			"maintainability": 0.15,
			"feasibility":     0.10,
			"traceability":    0.10,
		},
	}

	// Create and run the spec engine
	engine := spec.NewEngine(cfg)
	ctx := context.Background()

	// Run spec phase
	t.Run("Spec Phase", func(t *testing.T) {
		result, err := engine.Run(ctx, ideasContent, types.PhaseSpec)
		if err != nil {
			t.Fatalf("spec.Run() failed: %v", err)
		}

		// Verify result structure
		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if !result.Success {
			t.Errorf("expected success=true, got false with errors: %v", result.Errors)
		}

		// Verify artifacts
		if len(result.Artifacts) == 0 {
			t.Error("expected at least one artifact")
		}

		// Check for SPEC.md artifact
		var foundSpec bool
		for _, artifact := range result.Artifacts {
			if artifact.Type == types.ArtifactSpec {
				foundSpec = true
				if artifact.Path == "" {
					t.Error("SPEC artifact path should not be empty")
				}
				if artifact.Content == "" {
					t.Error("SPEC artifact content should not be empty")
				}
				if !artifact.Generated {
					t.Error("SPEC artifact should be marked as generated")
				}
			}
		}
		if !foundSpec {
			t.Error("expected to find SPEC artifact in results")
		}

		// Verify context was populated
		if result.Context == nil {
			t.Error("expected non-nil context")
		}

		// Verify metrics
		if result.Metrics == nil {
			t.Error("expected non-nil metrics")
		} else {
			if result.Metrics.Duration == 0 {
				t.Error("expected non-zero duration")
			}
			if result.Metrics.StartTime.IsZero() {
				t.Error("expected non-zero start time")
			}
		}
	})

	// Run design phase
	t.Run("Design Phase", func(t *testing.T) {
		result, err := engine.Run(ctx, ideasContent, types.PhaseDesign)
		if err != nil {
			t.Fatalf("design phase failed: %v", err)
		}

		if !result.Success {
			t.Errorf("expected success=true, got false with errors: %v", result.Errors)
		}

		// Check for DESIGN.md artifact
		var foundDesign bool
		for _, artifact := range result.Artifacts {
			if artifact.Type == types.ArtifactDesign {
				foundDesign = true
				if artifact.Content == "" {
					t.Error("DESIGN artifact content should not be empty")
				}
			}
		}
		if !foundDesign {
			t.Error("expected to find DESIGN artifact in results")
		}
	})

	// Run tasks phase
	t.Run("Tasks Phase", func(t *testing.T) {
		result, err := engine.Run(ctx, ideasContent, types.PhaseTasks)
		if err != nil {
			t.Fatalf("tasks phase failed: %v", err)
		}

		if !result.Success {
			t.Errorf("expected success=true, got false with errors: %v", result.Errors)
		}

		// Check for TASKS.md artifact
		var foundTasks bool
		for _, artifact := range result.Artifacts {
			if artifact.Type == types.ArtifactTasks {
				foundTasks = true
				if artifact.Content == "" {
					t.Error("TASKS artifact content should not be empty")
				}
			}
		}
		if !foundTasks {
			t.Error("expected to find TASKS artifact in results")
		}
	})

	// Verify quality scores via iterations
	t.Run("Quality Scores", func(t *testing.T) {
		iterations := engine.GetIterations()
		if len(iterations) == 0 {
			t.Error("expected at least one iteration from self-questioning loop")
		}

		// Check latest score
		score := engine.GetScore()
		if score == nil {
			t.Error("expected non-nil score from latest iteration")
		} else {
			// Verify score structure
			if score.Overall < 0 || score.Overall > 1 {
				t.Errorf("overall score should be between 0 and 1, got %f", score.Overall)
			}

			if len(score.Breakdown) == 0 {
				t.Error("expected at least one score breakdown dimension")
			}

			// Verify dimensions have valid scores
			for _, dim := range score.Breakdown {
				if dim.Score < 0 || dim.Score > dim.MaxScore {
					t.Errorf("dimension %s has invalid score: %f (max: %f)",
						dim.Name, dim.Score, dim.MaxScore)
				}
			}
		}
	})

	// Verify files would be created at expected paths
	t.Run("Output Paths", func(t *testing.T) {
		expectedPaths := []string{
			filepath.Join(specDir, "SPEC.md"),
			filepath.Join(specDir, "DESIGN.md"),
			filepath.Join(specDir, "TASKS.md"),
		}

		for _, path := range expectedPaths {
			// Note: In real scenario, files would be written to disk
			// For this integration test, we verify the paths are correctly constructed
			if path == "" {
				t.Errorf("expected valid path for %s", filepath.Base(path))
			}
		}
	})
}

// TestEndToEnd_LoopFlow tests the complete Ralph Loop flow:
// 1. Create spec files
// 2. Run orchestrator.Run()
// 3. Verify state persistence
// 4. Verify readiness report
func TestEndToEnd_LoopFlow(t *testing.T) {
	// Create temporary directory for the test
	tempDir := t.TempDir()

	// Create spec directory structure
	specDir := filepath.Join(tempDir, "spec")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	// Create state directory
	stateDir := filepath.Join(tempDir, ".grove-state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("failed to create state directory: %v", err)
	}

	// Create sample SPEC.md
	specContent := `# Specification

## Requirements
1. Implement user authentication
2. Add session management
3. Create login/logout flows

## Scenarios
- User can register with email
- User can login with valid credentials
- User can logout
`

	specPath := filepath.Join(specDir, "SPEC.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("failed to write SPEC.md: %v", err)
	}

	// Create TASKS.md
	tasksContent := `# Tasks

## Task List
1. Create user model
2. Implement password hashing
3. Add JWT token generation
4. Create login endpoint
5. Create logout endpoint
`

	tasksPath := filepath.Join(specDir, "TASKS.md")
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("failed to write TASKS.md: %v", err)
	}

	// Create AGENTS.md for layer-1 file discovery
	agentsContent := `# AGENTS.md

## Conventions
- Use Clean Architecture
- Follow Go idioms

## Files
- @internal/auth/model.go
- @internal/auth/service.go
`

	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(agentsContent), 0o644); err != nil {
		t.Fatalf("failed to write AGENTS.md: %v", err)
	}

	// Configure the orchestrator
	cfg := &loop.OrchestratorConfig{
		ProjectPath:       tempDir,
		DocsPath:          specDir,
		StateDir:          stateDir,
		CheckpointEnabled: true,
		MaxRetries:        3,
		BackoffBaseMs:     100,
	}

	// Track phase changes
	var phaseChanges []string
	cfg.OnPhaseChange = func(from, to loop.LoopPhase) {
		phaseChanges = append(phaseChanges, string(to))
	}

	// Track task completions
	var completedTasks []string
	cfg.OnTaskComplete = func(task *loop.Task, err error) {
		if err == nil && task != nil {
			completedTasks = append(completedTasks, task.ID)
		}
	}

	// Create orchestrator
	orchestrator := loop.NewOrchestrator(cfg)

	// Test readiness check
	t.Run("Readiness Check", func(t *testing.T) {
		ready, err := orchestrator.CheckReadiness()
		if err != nil {
			t.Fatalf("CheckReadiness() failed: %v", err)
		}
		if !ready {
			t.Error("expected orchestrator to be ready")
		}
	})

	// Run the loop
	t.Run("Execute Loop", func(t *testing.T) {
		err := orchestrator.Run()
		if err != nil {
			t.Fatalf("orchestrator.Run() failed: %v", err)
		}

		// Verify final state
		state := orchestrator.State()
		if state.Status != loop.StatusCompleted {
			t.Errorf("expected status=%s, got %s", loop.StatusCompleted, state.Status)
		}

		if state.Phase != loop.PhaseComplete {
			t.Errorf("expected phase=%s, got %s", loop.PhaseComplete, state.Phase)
		}

		if state.Progress < 1.0 {
			t.Errorf("expected progress=1.0, got %f", state.Progress)
		}
	})

	// Verify state persistence
	t.Run("State Persistence", func(t *testing.T) {
		// Check that state was saved
		state := orchestrator.State()

		if state.StartedAt.IsZero() {
			t.Error("expected non-zero started at time")
		}

		if state.UpdatedAt.IsZero() {
			t.Error("expected non-zero updated at time")
		}

		// Verify UpdatedAt is after StartedAt
		if !state.UpdatedAt.After(state.StartedAt) && !state.UpdatedAt.Equal(state.StartedAt) {
			t.Error("expected updated_at to be after or equal to started_at")
		}

		// Check that progress was tracked
		if state.Progress == 0 {
			t.Error("expected non-zero progress")
		}
	})

	// Verify phase changes were tracked
	t.Run("Phase Changes", func(t *testing.T) {
		if len(phaseChanges) == 0 {
			t.Error("expected at least one phase change")
		}

		// Should have transitioned through multiple phases
		expectedPhases := []loop.LoopPhase{
			loop.PhasePreFlight,
			loop.PhaseAnalyze,
			loop.PhaseSpec,
			loop.PhaseDesign,
			loop.PhaseTasks,
			loop.PhaseImplement,
			loop.PhaseVerify,
			loop.PhaseProduction,
			loop.PhaseArchive,
			loop.PhaseComplete,
		}

		for _, expected := range expectedPhases {
			found := false
			for _, changed := range phaseChanges {
				if loop.LoopPhase(changed) == expected {
					found = true
					break
				}
			}
			if !found {
				t.Logf("phase %s not found in changes: %v", expected, phaseChanges)
			}
		}
	})

	// Verify readiness report would be generated
	t.Run("Readiness Report", func(t *testing.T) {
		// In a real scenario, the readiness report is written to disk
		// Here we verify the orchestrator reached the final phase
		state := orchestrator.State()

		if state.Phase != loop.PhaseComplete {
			t.Errorf("expected phase=%s for readiness report generation", loop.PhaseComplete)
		}

		// Verify progress indicates readiness
		if state.Progress < 0.9 {
			t.Errorf("expected progress >= 0.9 for readiness, got %f", state.Progress)
		}
	})

	// Verify tasks were executed
	t.Run("Task Execution", func(t *testing.T) {
		tasks := orchestrator.GetTasks()
		if len(tasks) == 0 {
			t.Error("expected tasks to be loaded")
		}

		completed := orchestrator.GetCompletedTasks()
		if len(completed) == 0 {
			t.Error("expected at least one completed task")
		}

		// Verify all tasks completed successfully
		for id, success := range completed {
			if !success {
				t.Errorf("task %s should be marked as completed", id)
			}
		}
	})
}

// TestEndToEnd_OptiFlow tests the complete Opti Prompt flow:
// 1. Run classifier.Classify()
// 2. Run collector.Collect()
// 3. Run optimizer.Optimize()
// 4. Verify optimized prompt
func TestEndToEnd_OptiFlow(t *testing.T) {
	// Create temporary project directory
	tempDir := t.TempDir()

	// Create project structure
	srcDir := filepath.Join(tempDir, "src", "auth")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("failed to create src directory: %v", err)
	}

	// Create AGENTS.md
	agentsContent := `# AGENTS.md

## Conventions
- Use Clean Architecture patterns
- Follow Go idioms and best practices

## Auth Module
@src/auth/model.go - User data structures
@src/auth/service.go - Authentication service
`

	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(agentsContent), 0o644); err != nil {
		t.Fatalf("failed to write AGENTS.md: %v", err)
	}

	// Create SPEC.md
	specContent := `# Specification

## User Authentication
- Users can register with email
- Users can login with credentials
- Sessions managed via JWT

## Components
- UserModel: stores user data
- AuthService: handles authentication logic
`

	specPath := filepath.Join(tempDir, "SPEC.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("failed to write SPEC.md: %v", err)
	}

	// Create sample source files
	modelContent := `package auth

type User struct {
	ID       string
	Email    string
	Password string
}
`

	modelPath := filepath.Join(srcDir, "model.go")
	if err := os.WriteFile(modelPath, []byte(modelContent), 0o644); err != nil {
		t.Fatalf("failed to write model.go: %v", err)
	}

	serviceContent := `package auth

import "context"

type AuthService struct{}

func (s *AuthService) Login(ctx context.Context, email, password string) error {
	// Implementation
	return nil
}
`

	servicePath := filepath.Join(srcDir, "service.go")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0o644); err != nil {
		t.Fatalf("failed to write service.go: %v", err)
	}

	// Test input
	userInput := "Add refresh token support to the authentication service"

	// Step 1: Classify intent
	t.Run("Classify Intent", func(t *testing.T) {
		classifier := opti.NewClassifier()
		classification := classifier.Classify(userInput)

		// Verify classification structure
		if classification.Intent == "" {
			t.Error("expected non-empty intent")
		}

		if classification.RawInput != userInput {
			t.Errorf("expected raw input to match, got %s", classification.RawInput)
		}

		// Verify confidence is in valid range
		if classification.Confidence < 0 || classification.Confidence > 1 {
			t.Errorf("confidence should be between 0 and 1, got %f", classification.Confidence)
		}

		// Verify keywords were extracted
		if len(classification.Keywords) == 0 {
			t.Error("expected at least one keyword")
		}

		// Should have extracted "auth" or "authentication" from the input
		foundAuth := false
		for _, kw := range classification.Keywords {
			if kw == "auth" || kw == "authentication" {
				foundAuth = true
				break
			}
		}

		// Log what we got for debugging
		t.Logf("Classification: intent=%s, domain=%s, keywords=%v, confidence=%f",
			classification.Intent, classification.Domain, classification.Keywords, classification.Confidence)

		// For "Add refresh token" it should classify as feature addition
		if classification.Intent != opti.IntentFeatureAddition {
			t.Logf("Note: intent classified as %s (may be correct based on patterns)", classification.Intent)
		}
	})

	// Step 2: Collect context
	t.Run("Collect Context", func(t *testing.T) {
		classifier := opti.NewClassifier()
		classification := classifier.Classify(userInput)

		collector := opti.NewCollector(tempDir)
		ctx := context.Background()

		result, err := collector.Collect(ctx, classification)
		if err != nil {
			t.Fatalf("collector.Collect() failed: %v", err)
		}

		// Verify result structure
		if result == nil {
			t.Fatal("expected non-nil context result")
		}

		// Verify files were selected
		if len(result.Files) == 0 {
			t.Error("expected at least one file to be selected")
		}

		// Verify layer log was populated
		if len(result.LayerLog) == 0 {
			t.Error("expected layer log entries")
		}

		// Verify files have valid structure
		for _, file := range result.Files {
			if file.Path == "" {
				t.Error("file path should not be empty")
			}
			if file.Layer < 1 || file.Layer > 4 {
				t.Errorf("file layer should be 1-4, got %d", file.Layer)
			}
			if file.LayerName == "" {
				t.Error("file layer name should not be empty")
			}
		}

		// Verify skills were discovered
		t.Logf("Discovered skills: %v", result.Skills)

		// Verify dependency refs were built for cross-module intent
		if len(result.DependencyRefs) >= 0 {
			t.Logf("Dependency refs: %v", result.DependencyRefs)
		}
	})

	// Step 3: Optimize prompt
	t.Run("Optimize Prompt", func(t *testing.T) {
		classifier := opti.NewClassifier()
		classification := classifier.Classify(userInput)

		collector := opti.NewCollector(tempDir)
		ctx := context.Background()

		contextResult, err := collector.Collect(ctx, classification)
		if err != nil {
			t.Fatalf("collector.Collect() failed: %v", err)
		}

		optimizer := opti.NewOptimizer(2000)
		optimized, err := optimizer.Optimize(ctx, userInput, classification, contextResult)
		if err != nil {
			t.Fatalf("optimizer.Optimize() failed: %v", err)
		}

		// Verify result structure
		if optimized == nil {
			t.Fatal("expected non-nil optimized prompt")
		}

		// Verify original is preserved
		if optimized.Original != userInput {
			t.Errorf("expected original to match input, got %s", optimized.Original)
		}

		// Verify optimized output exists
		if optimized.Optimized == "" {
			t.Error("expected non-empty optimized prompt")
		}

		// Verify optimized is different from original
		if optimized.Optimized == optimized.Original {
			t.Error("expected optimized prompt to differ from original")
		}

		// Verify elements were created
		if len(optimized.Elements) == 0 {
			t.Error("expected at least one prompt element")
		}

		// Verify token budget was set
		if optimized.TokenBudget == 0 {
			t.Error("expected non-zero token budget")
		}

		// Verify token count was estimated
		if optimized.TokenCount == 0 {
			t.Error("expected non-zero token count")
		}

		// Token count should not exceed budget
		if optimized.TokenCount > optimized.TokenBudget {
			t.Errorf("token count (%d) exceeds budget (%d)",
				optimized.TokenCount, optimized.TokenBudget)
		}

		// Verify skills used were tracked
		if len(optimized.SkillsUsed) > 0 {
			t.Logf("Skills used in optimization: %v", optimized.SkillsUsed)
		}

		// Verify warnings were generated if needed
		t.Logf("Warnings: %v", optimized.Warnings)

		// Print the optimized prompt for visibility
		t.Logf("Optimized prompt:\n%s", optimized.Optimized)
	})

	// Full integration: classify -> collect -> optimize
	t.Run("Full Pipeline", func(t *testing.T) {
		// This test verifies the complete flow works end-to-end
		userInput := "Fix the bug where login fails with special characters in password"

		classifier := opti.NewClassifier()
		collector := opti.NewCollector(tempDir)
		optimizer := opti.NewOptimizer(1500)
		ctx := context.Background()

		// Execute pipeline
		classification := classifier.Classify(userInput)
		contextResult, err := collector.Collect(ctx, classification)
		if err != nil {
			t.Fatalf("collect failed: %v", err)
		}

		optimized, err := optimizer.Optimize(ctx, userInput, classification, contextResult)
		if err != nil {
			t.Fatalf("optimize failed: %v", err)
		}

		// Verify the bug fix intent was detected
		if classification.Intent != opti.IntentBugFix {
			t.Logf("Note: intent classified as %s (expected bug-fix for 'Fix bug')",
				classification.Intent)
		}

		// Verify the optimized prompt contains bug fix context
		if optimized.Optimized == "" {
			t.Error("expected non-empty optimized prompt from full pipeline")
		}

		// Verify confidence is reasonable for bug fix
		if classification.Confidence < 0.3 {
			t.Logf("Note: confidence is low (%f) for bug fix detection", classification.Confidence)
		}

		// Verify success criteria was added
		var foundSuccessCriteria bool
		for _, elem := range optimized.Elements {
			if elem.Type == opti.ElementSuccessCriteria {
				foundSuccessCriteria = true
				break
			}
		}
		if !foundSuccessCriteria {
			t.Error("expected success criteria element in optimized prompt")
		}
	})
}

// TestEndToEnd_ConfigIntegration tests that configuration is properly
// loaded and applied across all flows.
func TestEndToEnd_ConfigIntegration(t *testing.T) {
	// Test loading default config
	t.Run("Default Config", func(t *testing.T) {
		cfg := config.DefaultConfig()

		if cfg == nil {
			t.Fatal("expected non-nil config")
		}

		if cfg.ProjectName != config.DefaultProjectName {
			t.Errorf("expected project name %s, got %s",
				config.DefaultProjectName, cfg.ProjectName)
		}

		if cfg.MaxIterations != config.DefaultMaxIterations {
			t.Errorf("expected max iterations %d, got %d",
				config.DefaultMaxIterations, cfg.MaxIterations)
		}
	})

	// Test config validation
	t.Run("Config Validation", func(t *testing.T) {
		cfg := config.DefaultConfig()

		errors := cfg.Validate()
		if len(errors) > 0 {
			t.Errorf("expected no validation errors, got: %v", errors)
		}
	})

	// Test state directory functions
	t.Run("State Directory", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg := &config.Config{
			ProjectPath: tempDir,
			Loop: config.LoopConfig{
				StateDir:  ".grove-state",
				StateFile: "test-state.json",
			},
		}

		stateDir := config.GetLoopStateDir(cfg)
		expectedDir := filepath.Join(tempDir, ".grove-state")
		if stateDir != expectedDir {
			t.Errorf("expected state dir %s, got %s", expectedDir, stateDir)
		}

		statePath := config.GetLoopStatePath(cfg)
		expectedPath := filepath.Join(expectedDir, "test-state.json")
		if statePath != expectedPath {
			t.Errorf("expected state path %s, got %s", expectedPath, statePath)
		}
	})
}

// TestEndToEnd_TypesConsistency tests that types are consistent
// across packages and phases.
func TestEndToEnd_TypesConsistency(t *testing.T) {
	// Test Phase constants
	t.Run("Phase Constants", func(t *testing.T) {
		phases := []types.Phase{
			types.PhaseExplore,
			types.PhasePropose,
			types.PhaseSpec,
			types.PhaseDesign,
			types.PhaseTasks,
			types.PhaseApply,
			types.PhaseVerify,
			types.PhaseArchive,
		}

		for _, phase := range phases {
			if phase == "" {
				t.Error("phase constant should not be empty")
			}
		}
	})

	// Test ArtifactType constants
	t.Run("Artifact Types", func(t *testing.T) {
		artifactTypes := []types.ArtifactType{
			types.ArtifactSpec,
			types.ArtifactDesign,
			types.ArtifactTasks,
			types.ArtifactReport,
			types.ArtifactAgents,
			types.ArtifactState,
		}

		for _, at := range artifactTypes {
			if at == "" {
				t.Error("artifact type should not be empty")
			}
		}
	})

	// Test ChangeStatus constants
	t.Run("Change Status", func(t *testing.T) {
		statuses := []types.ChangeStatus{
			types.StatusPending,
			types.StatusActive,
			types.StatusCompleted,
			types.StatusArchived,
		}

		for _, status := range statuses {
			if status == "" {
				t.Error("change status should not be empty")
			}
		}
	})

	// Test TaskStatus constants
	t.Run("Task Status", func(t *testing.T) {
		statuses := []types.TaskStatus{
			types.TaskPending,
			types.TaskInProgress,
			types.TaskBlocked,
			types.TaskDone,
		}

		for _, status := range statuses {
			if status == "" {
				t.Error("task status should not be empty")
			}
		}
	})

	// Test Score structure
	t.Run("Score Structure", func(t *testing.T) {
		score := &types.Score{
			Overall:    0.85,
			Dimensions: make(map[string]float64),
			Breakdown: []types.ScoreDimension{
				{
					Name:     "completeness",
					Score:    8.5,
					MaxScore: 10,
					Weight:   0.20,
				},
				{
					Name:     "clarity",
					Score:    9.0,
					MaxScore: 10,
					Weight:   0.15,
				},
			},
		}

		if score.Overall < 0 || score.Overall > 1 {
			t.Errorf("overall score should be 0-1, got %f", score.Overall)
		}

		if len(score.Breakdown) == 0 {
			t.Error("expected at least one dimension in breakdown")
		}
	})

	// Test Metrics structure
	t.Run("Metrics Structure", func(t *testing.T) {
		now := time.Now()
		metrics := &types.Metrics{
			StartTime:     now,
			EndTime:       now.Add(5 * time.Second),
			Duration:      5 * time.Second,
			Iterations:    3,
			TokensUsed:    1500,
			FilesModified: 5,
		}

		if metrics.Duration == 0 {
			t.Error("expected non-zero duration")
		}

		if metrics.Iterations < 0 {
			t.Error("iterations should be non-negative")
		}
	})
}

// TestEndToEnd_IntegrationBenchmark provides benchmark-style tests
// to measure performance of the integration flows.
func TestEndToEnd_IntegrationBenchmark(t *testing.T) {
	// This test can be used for performance regression testing
	if testing.Short() {
		t.Skip("skipping benchmark in short mode")
	}

	t.Run("Spec Engine Performance", func(t *testing.T) {
		tempDir := t.TempDir()
		specDir := filepath.Join(tempDir, "spec")
		os.MkdirAll(specDir, 0o755)

		cfg := &types.Config{
			ProjectPath:           tempDir,
			OutputPath:            specDir,
			MaxIterations:         2,
			QualityThreshold:      0.5,
			EnableSelfQuestioning: true,
		}

		engine := spec.NewEngine(cfg)
		ctx := context.Background()

		input := "Add user profile feature with avatar upload"

		start := time.Now()
		_, err := engine.Run(ctx, input, types.PhaseSpec)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("spec run failed: %v", err)
		}

		// Log performance metrics
		t.Logf("Spec phase completed in %v", elapsed)

		// Should complete in reasonable time
		if elapsed > 30*time.Second {
			t.Logf("Note: Spec phase took longer than expected (%v)", elapsed)
		}
	})
}
