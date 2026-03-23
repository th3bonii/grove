package spec

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Gentleman-Programming/grove/internal/types"
)

// TestNewEngine tests engine creation.
func TestNewEngine(t *testing.T) {
	tests := []struct {
		name   string
		config *types.Config
	}{
		{
			name:   "with nil config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &types.Config{
				ProjectName:           "test-project",
				MaxIterations:         5,
				QualityThreshold:      0.8,
				EnableSelfQuestioning: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(tt.config)

			if engine == nil {
				t.Fatal("expected engine to be created, got nil")
			}

			if engine.config == nil {
				t.Error("expected config to be set")
			}

			if engine.scorer == nil {
				t.Error("expected scorer to be initialized")
			}

			if engine.generator == nil {
				t.Error("expected generator to be initialized")
			}
		})
	}
}

// TestEngineProcessInput tests input processing.
func TestEngineProcessInput(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	testInput := "Implement user authentication with JWT tokens"

	err := engine.ProcessInput(ctx, testInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	engineCtx := engine.GetContext()
	if engineCtx.InputText != testInput {
		t.Errorf("expected input '%s', got '%s'", testInput, engineCtx.InputText)
	}

	if engineCtx.Metadata == nil {
		t.Error("expected metadata to be initialized")
	}
}

// TestEngineRun tests the full Run method.
func TestEngineRun(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	testInput := "Add dark mode support to the application"

	result, err := engine.Run(ctx, testInput, types.PhaseSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.Context == nil {
		t.Fatal("expected context in result")
	}

	if result.Context.Change == nil {
		t.Fatal("expected change to be created")
	}

	if result.Context.Change.Phase != types.PhaseSpec {
		t.Errorf("expected phase %s, got %s", types.PhaseSpec, result.Context.Change.Phase)
	}

	if result.Metrics == nil {
		t.Error("expected metrics to be populated")
	}

	if result.Metrics.Iterations < 0 {
		t.Error("expected iterations to be >= 0")
	}
}

// TestEngineRunPhases tests different phase executions.
func TestEngineRunPhases(t *testing.T) {
	phases := []types.Phase{
		types.PhaseSpec,
		types.PhaseDesign,
		types.PhaseTasks,
		types.PhaseVerify,
	}

	for _, phase := range phases {
		t.Run(string(phase), func(t *testing.T) {
			engine := NewEngine(nil)
			ctx := context.Background()

			result, err := engine.Run(ctx, "Test input for "+string(phase), phase)
			if err != nil {
				t.Fatalf("unexpected error for phase %s: %v", phase, err)
			}

			if result == nil {
				t.Fatalf("expected result for phase %s", phase)
			}
		})
	}
}

// TestEngineSelfQuestioning tests the self-questioning loop.
func TestEngineSelfQuestioning(t *testing.T) {
	engine := NewEngine(&types.Config{
		EnableSelfQuestioning: true,
		MaxIterations:         3,
	})
	ctx := context.Background()

	content := "## Requirements\n\n- User authentication with JWT\n- Session management\n- Role-based access control"

	iteration, err := engine.SelfQuestioning(ctx, content, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if iteration == nil {
		t.Fatal("expected iteration result")
	}

	if iteration.Number != 1 {
		t.Errorf("expected iteration number 1, got %d", iteration.Number)
	}

	if iteration.Question == "" {
		t.Error("expected questions to be generated")
	}

	if iteration.Score == nil {
		t.Error("expected score to be calculated")
	}

	iterations := engine.GetIterations()
	if len(iterations) != 1 {
		t.Errorf("expected 1 iteration, got %d", len(iterations))
	}
}

// TestEngineSelfQuestioningDisabled tests self-questioning when disabled.
func TestEngineSelfQuestioningDisabled(t *testing.T) {
	engine := NewEngine(&types.Config{
		EnableSelfQuestioning: false,
	})
	ctx := context.Background()

	iteration, err := engine.SelfQuestioning(ctx, "content", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if iteration != nil {
		t.Error("expected nil iteration when self-questioning is disabled")
	}
}

// TestEngineIterationLoop tests the full iteration loop.
func TestEngineIterationLoop(t *testing.T) {
	engine := NewEngine(&types.Config{
		EnableSelfQuestioning: true,
		MaxIterations:         2,
		QualityThreshold:      0.9, // High threshold to force multiple iterations
	})
	ctx := context.Background()

	initialContent := "# Test Spec\n\n## Requirements\n\n- Feature A\n- Feature B\n- Feature C"

	finalContent, score, err := engine.IterationLoop(ctx, initialContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if finalContent == "" {
		t.Error("expected content to be returned")
	}

	// Should have run multiple iterations
	iterations := engine.GetIterations()
	if len(iterations) < 1 {
		t.Error("expected at least 1 iteration")
	}

	_ = score // Score may be nil if no iterations ran
}

// TestEngineGenerateOutput tests output generation.
func TestEngineGenerateOutput(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	// Set up context with a spec
	engine.ctx.Spec = &types.SpecDocument{
		Title:   "test-change",
		Version: "1.0.0",
		Requirements: []types.Requirement{
			{
				ID:          "REQ-001",
				Type:        "functional",
				Description: "Test description",
				Priority:    "high",
			},
		},
		UserFlows: []types.UserFlow{
			{
				ID:          "SCN-001",
				Name:        "Test Scenario",
				Description: "Test scenario description",
				Steps: []types.UserFlowStep{
					{
						StepNumber:     1,
						Action:         "initial state",
						ExpectedResult: "result",
					},
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_ = ctx // unused in this test

	output, err := engine.GenerateOutput(ctx, types.PhaseSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output == "" {
		t.Error("expected output to be generated")
	}

	if !strings.Contains(output, "test-change") {
		t.Error("expected output to contain change name")
	}

	if !strings.Contains(output, "REQ-001") {
		t.Error("expected output to contain requirement ID")
	}
}

// TestEngineCallbacks tests callback functionality.
func TestEngineCallbacks(t *testing.T) {
	engine := NewEngine(nil)

	var iterationCalled bool
	var progressCalled bool
	var lastProgress float64

	engine.SetOnIteration(func(iter *types.Iteration) {
		iterationCalled = true
	})

	engine.SetOnProgress(func(stage string, progress float64) {
		progressCalled = true
		lastProgress = progress
	})

	// Trigger callbacks through iteration
	ctx := context.Background()
	engine.SelfQuestioning(ctx, "test content", 1)

	if !iterationCalled {
		t.Error("expected iteration callback to be called")
	}

	// Progress is called during Run
	engine.Run(ctx, "test", types.PhaseSpec)

	if !progressCalled {
		t.Error("expected progress callback to be called")
	}

	if lastProgress < 0 || lastProgress > 1 {
		t.Errorf("expected progress between 0 and 1, got %f", lastProgress)
	}
}

// TestEngineGetIterations tests iteration retrieval.
func TestEngineGetIterations(t *testing.T) {
	engine := NewEngine(&types.Config{
		EnableSelfQuestioning: true,
	})
	ctx := context.Background()

	// Run multiple iterations
	for i := 1; i <= 3; i++ {
		engine.SelfQuestioning(ctx, "content", i)
	}

	iterations := engine.GetIterations()
	if len(iterations) != 3 {
		t.Errorf("expected 3 iterations, got %d", len(iterations))
	}

	// Verify iteration numbers
	for i, iter := range iterations {
		if iter.Number != i+1 {
			t.Errorf("expected iteration %d, got %d", i+1, iter.Number)
		}
	}
}

// TestEngineGetScore tests score retrieval.
func TestEngineGetScore(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	// No iterations yet
	score := engine.GetScore()
	if score != nil {
		t.Error("expected nil score before iterations")
	}

	// Run an iteration
	engine.SelfQuestioning(ctx, "test content with requirements and scenarios", 1)

	score = engine.GetScore()
	if score == nil {
		t.Fatal("expected score after iteration")
	}

	if score.Overall < 0 || score.Overall > 10 {
		t.Errorf("expected score between 0 and 10, got %f", score.Overall)
	}

	if len(score.Breakdown) == 0 {
		t.Error("expected breakdown to have dimensions")
	}
}

// TestEngineContextThreadSafety tests thread safety of context access.
func TestEngineContextThreadSafety(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	// Run concurrent operations
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			engine.ProcessInput(ctx, "concurrent input")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = engine.GetContext()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			engine.SelfQuestioning(ctx, "content", i)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Should not panic and should have iterations
	iterations := engine.GetIterations()
	if len(iterations) != 100 {
		t.Logf("expected ~100 iterations, got %d", len(iterations))
	}
}

// TestEngineEmptyInput tests handling of empty input.
func TestEngineEmptyInput(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	result, err := engine.Run(ctx, "", types.PhaseSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result")
	}

	// Should still create a change with extracted name
	if result.Context.Change == nil {
		t.Error("expected change to be created even with empty input")
	}
}

// TestEngineOutputPath tests output path configuration.
func TestEngineOutputPath(t *testing.T) {
	customPath := "/custom/output/path"
	engine := NewEngine(&types.Config{
		OutputPath: customPath,
	})

	result, err := engine.Run(context.Background(), "test input", types.PhaseSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that artifacts use configured path
	for _, artifact := range result.Artifacts {
		if !strings.Contains(artifact.Path, customPath) {
			t.Errorf("expected artifact path to contain %s, got %s", customPath, artifact.Path)
		}
	}
}

// BenchmarkEngineRun benchmarks the Run method.
func BenchmarkEngineRun(b *testing.B) {
	engine := NewEngine(nil)
	ctx := context.Background()

	input := "Implement user authentication with JWT tokens and refresh token rotation"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Run(ctx, input, types.PhaseSpec)
	}
}

// BenchmarkEngineSelfQuestioning benchmarks self-questioning.
func BenchmarkEngineSelfQuestioning(b *testing.B) {
	engine := NewEngine(&types.Config{
		EnableSelfQuestioning: true,
	})
	ctx := context.Background()

	content := "# Test Specification\n\n## Requirements\n\n- Feature A with detailed description\n- Feature B with detailed description\n- Feature C with detailed description\n\n## Scenarios\n\n### Scenario 1\n\n**Given** initial state\n**When** action performed\n**Then** result expected\n\n### Scenario 2\n\n**Given** another state\n**When** another action\n**Then** another result"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.SelfQuestioning(ctx, content, 1)
	}
}

// BenchmarkEngineIterationLoop benchmarks the iteration loop.
func BenchmarkEngineIterationLoop(b *testing.B) {
	engine := NewEngine(&types.Config{
		EnableSelfQuestioning: true,
		MaxIterations:         3,
	})
	ctx := context.Background()

	content := "# Specification\n\n## Requirements\n\n- Requirement 1\n- Requirement 2\n- Requirement 3\n\n## Scenarios\n\n### Scenario 1\n\n**Given** context\n**When** action\n**Then** result"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.IterationLoop(ctx, content)
	}
}
