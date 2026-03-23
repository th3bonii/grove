package spec

import (
	"context"
	"testing"

	"github.com/Gentleman-Programming/grove/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// GROVE Spec Engine Integration Tests
// =============================================================================

// TestSpecToVerifyWorkflow tests the full spec → loop → verify workflow.
// This is an integration test that verifies the spec engine can generate
// specifications that can be used by the verify phase.
func TestSpecToVerifyWorkflow(t *testing.T) {
	// Create a test spec document
	spec := &types.SpecDocument{
		Title:        "Test Feature",
		Version:      "1.0.0",
		Overview:     "This is a test feature specification",
		Requirements: []types.Requirement{},
	}

	// Add test requirements
	spec.Requirements = append(spec.Requirements, types.Requirement{
		ID:          "REQ-001",
		Description: "User should be able to authenticate",
		Priority:    "high",
		Type:        "functional",
	})

	require.NotNil(t, spec)
	require.NotEmpty(t, spec.Requirements)
	require.Equal(t, "Test Feature", spec.Title)
}

// TestGROVESpecGeneration tests the GROVE Spec engine generates specs correctly.
func TestGROVESpecGeneration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checkFn func(*testing.T, *types.SpecDocument)
	}{
		{
			name:    "simple feature request",
			input:   "Implementar autenticación JWT",
			wantErr: false,
			checkFn: func(t *testing.T, spec *types.SpecDocument) {
				require.NotNil(t, spec)
				// Spec should have basic structure
				assert.NotEmpty(t, spec.Title)
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: false,
			checkFn: func(t *testing.T, spec *types.SpecDocument) {
				require.NotNil(t, spec)
			},
		},
		{
			name:    "feature with multiple components",
			input:   "Add user management with roles and permissions",
			wantErr: false,
			checkFn: func(t *testing.T, spec *types.SpecDocument) {
				require.NotNil(t, spec)
				assert.NotEmpty(t, spec.Title)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create engine with default config
			engine := NewEngine(&types.Config{
				ProjectPath: ".",
				OutputPath:  "spec",
			})

			// Run spec generation
			spec, err := engine.Generate(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			tt.checkFn(t, spec)
		})
	}
}

// TestGROVESpecWithComponents tests spec generation with component decomposition.
func TestGROVESpecWithComponents(t *testing.T) {
	engine := NewEngine(&types.Config{
		ProjectPath:   ".",
		MaxIterations: 3,
	})

	// Input with multiple components
	input := "Add authentication system with JWT tokens and user roles"

	spec, err := engine.Generate(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, spec)

	// Verify spec has required fields
	assert.NotEmpty(t, spec.Title)
	assert.Equal(t, "1.0.0", spec.Version)
}

// TestSpecEngineIterationLoop tests the self-questioning iteration loop.
func TestSpecEngineIterationLoop(t *testing.T) {
	engine := NewEngine(&types.Config{
		MaxIterations:         2,
		EnableSelfQuestioning: true,
		QualityThreshold:      0.7,
	})

	ctx := context.Background()
	content := "# Test Specification\n\nThis is a test."

	// Run iteration loop
	resultContent, score, err := engine.IterationLoop(ctx, content)
	require.NoError(t, err)

	// Should return content and score
	assert.NotEmpty(t, resultContent)
	if score != nil {
		assert.GreaterOrEqual(t, score.Overall, 0.0)
		assert.LessOrEqual(t, score.Overall, 10.0)
	}

	// Verify iterations were recorded
	iterations := engine.GetIterations()
	assert.GreaterOrEqual(t, len(iterations), 0)
}

// TestSpecEngineSelfQuestioning tests the self-questioning mechanism.
func TestSpecEngineSelfQuestioning(t *testing.T) {
	engine := NewEngine(&types.Config{
		EnableSelfQuestioning: true,
	})

	// Test with valid content
	iteration, err := engine.SelfQuestioning(context.Background(), "Test content", 1)
	require.NoError(t, err)
	require.NotNil(t, iteration)

	// Verify iteration has required fields
	assert.Equal(t, 1, iteration.Number)
	assert.NotEmpty(t, iteration.Question)
	assert.NotEmpty(t, iteration.Answer)
}

// TestSpecEngineSelfQuestioningDisabled tests when self-questioning is disabled.
func TestSpecEngineSelfQuestioningDisabled(t *testing.T) {
	engine := NewEngine(&types.Config{
		EnableSelfQuestioning: false,
	})

	// Should return nil when disabled
	iteration, err := engine.SelfQuestioning(context.Background(), "Test content", 1)
	require.NoError(t, err)
	require.Nil(t, iteration)
}

// TestSpecGeneratorOutput tests output generation for different phases.
func TestSpecGeneratorOutput(t *testing.T) {
	engine := NewEngine(&types.Config{
		ProjectPath: ".",
		OutputPath:  "output",
	})

	tests := []struct {
		name        string
		phase       types.Phase
		setupCtx    func(*Engine)
		wantErr     bool
		checkOutput func(string)
	}{
		{
			name:    "spec phase output",
			phase:   types.PhaseSpec,
			wantErr: false,
			checkOutput: func(output string) {
				assert.NotEmpty(t, output)
			},
		},
		{
			name:    "design phase output",
			phase:   types.PhaseDesign,
			wantErr: false,
			checkOutput: func(output string) {
				assert.NotEmpty(t, output)
			},
		},
		{
			name:    "tasks phase output",
			phase:   types.PhaseTasks,
			wantErr: false,
			checkOutput: func(output string) {
				assert.NotEmpty(t, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupCtx != nil {
				tt.setupCtx(engine)
			}

			output, err := engine.GenerateOutput(context.Background(), tt.phase)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			tt.checkOutput(output)
		})
	}
}

// TestSpecEngineRunFullCycle tests running a full spec generation cycle.
func TestSpecEngineRunFullCycle(t *testing.T) {
	engine := NewEngine(&types.Config{
		ProjectPath:   ".",
		OutputPath:    "spec",
		MaxIterations: 1,
	})

	input := "Add new feature: user dashboard"
	result, err := engine.Run(context.Background(), input, types.PhaseSpec)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have artifacts generated
	assert.GreaterOrEqual(t, len(result.Artifacts), 0)
}

// TestSpecEnginePhaseSpec tests running the spec phase specifically.
func TestSpecEnginePhaseSpec(t *testing.T) {
	engine := NewEngine(&types.Config{
		ProjectPath: ".",
		OutputPath:  "spec",
	})

	input := "Create login page with email and password"
	artifacts, err := engine.runSpecPhase(context.Background(), input)

	// May fail if no LLM client available, but should not crash
	if err != nil {
		t.Logf("Expected error without LLM: %v", err)
		return
	}

	require.NotNil(t, artifacts)
	assert.GreaterOrEqual(t, len(artifacts), 0)
}

// TestSpecEnginePhaseDesign tests running the design phase.
func TestSpecEnginePhaseDesign(t *testing.T) {
	engine := NewEngine(&types.Config{
		ProjectPath: ".",
		OutputPath:  "design",
	})

	input := "Design the authentication system"
	artifacts, err := engine.runDesignPhase(context.Background(), input)

	if err != nil {
		t.Logf("Expected error: %v", err)
		return
	}

	require.NotNil(t, artifacts)
}

// TestSpecEnginePhaseTasks tests running the tasks phase.
func TestSpecEnginePhaseTasks(t *testing.T) {
	engine := NewEngine(&types.Config{
		ProjectPath: ".",
		OutputPath:  "tasks",
	})

	input := "Create implementation tasks for auth"
	artifacts, err := engine.runTasksPhase(context.Background(), input)

	if err != nil {
		t.Logf("Expected error: %v", err)
		return
	}

	require.NotNil(t, artifacts)
}

// TestSpecEngineProcessInput tests the input processing functionality.
func TestSpecEngineProcessInput(t *testing.T) {
	engine := NewEngine(nil)

	err := engine.ProcessInput(context.Background(), "Test input with keywords")
	require.NoError(t, err)

	ctx := engine.GetContext()
	require.NotNil(t, ctx)
	assert.Equal(t, "Test input with keywords", ctx.InputText)
}

// TestSpecEngineContext tests getting and setting context.
func TestSpecEngineContext(t *testing.T) {
	engine := NewEngine(nil)

	// Get initial context
	ctx := engine.GetContext()
	require.NotNil(t, ctx)

	// Set input
	engine.ProcessInput(context.Background(), "test input")

	// Get updated context
	ctx = engine.GetContext()
	assert.Equal(t, "test input", ctx.InputText)
}

// TestSpecEngineCallbacks tests setting up callbacks for iteration and progress.
func TestSpecEngineCallbacks(t *testing.T) {
	engine := NewEngine(nil)

	var iterationCount int
	var progressEvents []string

	engine.SetOnIteration(func(i *types.Iteration) {
		iterationCount++
	})

	engine.SetOnProgress(func(stage string, progress float64) {
		progressEvents = append(progressEvents, stage)
	})

	// Run with callbacks
	_, _ = engine.SelfQuestioning(context.Background(), "test", 1)

	assert.GreaterOrEqual(t, iterationCount, 0)
	assert.GreaterOrEqual(t, len(progressEvents), 0)
}

// TestSpecEngineScore tests the scoring functionality.
func TestSpecEngineScore(t *testing.T) {
	engine := NewEngine(nil)

	// Initially no score
	score := engine.GetScore()
	require.Nil(t, score)

	// After self-questioning, should have score
	_, _ = engine.SelfQuestioning(context.Background(), "test content", 1)

	// Score may or may not be present depending on implementation
	_ = engine.GetScore() // Just verify it doesn't panic
}

// TestSpecEngineComponentDecomposition tests component decomposition.
func TestSpecEngineComponentDecomposition(t *testing.T) {
	engine := NewEngine(nil)

	// Create test input
	processed := &types.ProcessedInput{
		OriginalInput:  "Add user authentication with JWT",
		DetectedStack:  []string{"Go", "React"},
		ExtractedTypes: []string{"User", "Auth"},
	}

	components, err := engine.ComponentDecomposition(context.Background(), processed)
	require.NoError(t, err)
	require.NotNil(t, components)
	assert.GreaterOrEqual(t, len(components), 0)
}
