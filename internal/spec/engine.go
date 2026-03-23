package spec

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Gentleman-Programming/grove/internal/types"
)

// Engine is the main Spec engine that orchestrates the SDD workflow.
type Engine struct {
	config    *types.Config
	scorer    *Scorer
	generator *Generator

	mu         sync.RWMutex
	ctx        *types.Context
	iterations []types.Iteration

	// Callbacks for extensibility
	onIteration func(*types.Iteration)
	onProgress  func(string, float64)
}

// NewEngine creates a new Spec engine instance.
func NewEngine(config *types.Config) *Engine {
	if config == nil {
		config = defaultConfig()
	}

	return &Engine{
		config:    config,
		scorer:    NewScorer(config),
		generator: NewGenerator(config),
		ctx: &types.Context{
			Metadata: make(map[string]interface{}),
		},
		iterations: make([]types.Iteration, 0),
	}
}

// defaultConfig returns default configuration.
func defaultConfig() *types.Config {
	return &types.Config{
		MaxIterations:         3,
		QualityThreshold:      0.7,
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
}

// Run executes the full SDD workflow for a change.
func (e *Engine) Run(ctx context.Context, input string, phase types.Phase) (*types.Result, error) {
	startTime := time.Now()

	// Initialize context
	e.mu.Lock()
	e.ctx = &types.Context{
		InputText: input,
		Metadata:  make(map[string]interface{}),
	}
	e.iterations = make([]types.Iteration, 0)
	e.mu.Unlock()

	// Create change
	change := &types.Change{
		Name:      extractChangeName(input),
		Phase:     phase,
		Status:    types.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	e.mu.Lock()
	e.ctx.Change = change
	e.mu.Unlock()

	var artifacts []types.Artifact
	var operationErrors []string

	// Execute based on phase
	switch phase {
	case types.PhaseSpec:
		result, err := e.runSpecPhase(ctx, input)
		if err != nil {
			operationErrors = append(operationErrors, err.Error())
		}
		if result != nil {
			artifacts = append(artifacts, result...)
		}

	case types.PhaseDesign:
		result, err := e.runDesignPhase(ctx, input)
		if err != nil {
			operationErrors = append(operationErrors, err.Error())
		}
		if result != nil {
			artifacts = append(artifacts, result...)
		}

	case types.PhaseTasks:
		result, err := e.runTasksPhase(ctx, input)
		if err != nil {
			operationErrors = append(operationErrors, err.Error())
		}
		if result != nil {
			artifacts = append(artifacts, result...)
		}

	case types.PhaseVerify:
		result, err := e.runVerifyPhase(ctx, input)
		if err != nil {
			operationErrors = append(operationErrors, err.Error())
		}
		if result != nil {
			artifacts = append(artifacts, result...)
		}

	default:
		// Process input and generate output
		if err := e.ProcessInput(ctx, input); err != nil {
			operationErrors = append(operationErrors, err.Error())
		}
	}

	e.mu.RLock()
	finalCtx := e.ctx
	e.mu.RUnlock()

	return &types.Result{
		Success:   len(operationErrors) == 0,
		Context:   finalCtx,
		Artifacts: artifacts,
		Errors:    operationErrors,
		Metrics: &types.Metrics{
			StartTime:  startTime,
			EndTime:    time.Now(),
			Duration:   time.Since(startTime),
			Iterations: len(e.iterations),
		},
	}, nil
}

// ProcessInput processes the input text and prepares context.
func (e *Engine) ProcessInput(ctx context.Context, input string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.ctx == nil {
		e.ctx = &types.Context{
			Metadata: make(map[string]interface{}),
		}
	}

	e.ctx.InputText = input

	// Parse and extract key information
	e.ctx.Metadata["processed_at"] = time.Now()
	e.ctx.Metadata["input_length"] = len(input)

	return nil
}

// SelfQuestioning implements the self-questioning loop to improve quality.
func (e *Engine) SelfQuestioning(ctx context.Context, content string, iteration int) (*types.Iteration, error) {
	if !e.config.EnableSelfQuestioning {
		return nil, nil
	}

	// Generate self-reflection questions
	questions := e.generateSelfQuestions(content, iteration)

	iterationResult := types.Iteration{
		Number:    iteration,
		Question:  questions,
		Timestamp: time.Now(),
	}

	// Simulate answering questions (in real implementation, this would use AI)
	answers := e.answerSelfQuestions(content, questions)
	iterationResult.Answer = answers

	// Calculate score after self-questioning
	score := e.scorer.ScoreContent(content)
	iterationResult.Score = score

	// Determine if improvement was made
	previousScore := 0.0
	if len(e.iterations) > 0 && e.iterations[len(e.iterations)-1].Score != nil {
		previousScore = e.iterations[len(e.iterations)-1].Score.Overall
	}
	iterationResult.Improved = score.Overall > previousScore

	e.mu.Lock()
	e.iterations = append(e.iterations, iterationResult)
	e.mu.Unlock()

	// Trigger callback
	if e.onIteration != nil {
		e.onIteration(&iterationResult)
	}

	return &iterationResult, nil
}

// generateSelfQuestions generates reflection questions for self-questioning.
func (e *Engine) generateSelfQuestions(content string, iteration int) string {
	questions := []string{
		"¿Están todas las funcionalidades críticas cubiertas con requisitos específicos?",
		"¿Los escenarios de prueba son completos y cubran los casos borde?",
		"¿Hay ambigüedades en la terminología usada?",
		"¿Las dependencias externas están justificadas?",
		"¿Los estimates de tiempo son realistas?",
	}

	if iteration == 1 {
		questions = append(questions,
			"¿El alcance está claramente definido?",
			"¿Hay requisitos contradictorios?",
		)
	}

	return fmt.Sprintf("Iteration %d Questions:\n- %s", iteration, joinQuestions(questions))
}

// answerSelfQuestions provides answers to self-reflection questions.
func (e *Engine) answerSelfQuestions(content string, questions string) string {
	// In a real implementation, this would use AI to analyze and answer
	// For now, we return a placeholder
	return "Auto-generated answers based on content analysis. " +
		"Review recommended for critical requirements."
}

// IterationLoop runs the iterative improvement loop.
func (e *Engine) IterationLoop(ctx context.Context, initialContent string) (string, *types.Score, error) {
	currentContent := initialContent
	var bestScore *types.Score
	bestContent := currentContent

	for i := 1; i <= e.config.MaxIterations; i++ {
		// Update progress
		if e.onProgress != nil {
			e.onProgress(fmt.Sprintf("Iteration %d/%d", i, e.config.MaxIterations), float64(i)/float64(e.config.MaxIterations))
		}

		// Run self-questioning
		iteration, err := e.SelfQuestioning(ctx, currentContent, i)
		if err != nil {
			return currentContent, bestScore, err
		}

		if iteration != nil && iteration.Score != nil {
			if bestScore == nil || iteration.Score.Overall > bestScore.Overall {
				bestScore = iteration.Score
				bestContent = currentContent
			}

			// Check if quality threshold is met
			if iteration.Score.Overall >= e.config.QualityThreshold {
				break
			}
		}

		// Improve content based on iteration feedback
		improved, err := e.improveContent(ctx, currentContent, iteration)
		if err != nil {
			return currentContent, bestScore, err
		}
		currentContent = improved
	}

	return bestContent, bestScore, nil
}

// improveContent improves the content based on iteration feedback.
func (e *Engine) improveContent(ctx context.Context, content string, iteration *types.Iteration) (string, error) {
	if iteration == nil || iteration.Score == nil {
		return content, nil
	}

	// Generate improvement suggestions based on low-scoring dimensions
	var improvements []string

	for _, dim := range iteration.Score.Breakdown {
		if dim.Score/dim.MaxScore < 0.5 {
			improvements = append(improvements,
				fmt.Sprintf("Improve %s: %s", dim.Name, dim.Details))
		}
	}

	if len(improvements) > 0 {
		// In real implementation, this would regenerate content with improvements
		return content + "\n\n## Improvements Suggested:\n" + joinQuestions(improvements), nil
	}

	return content, nil
}

// GenerateOutput generates the final output based on context.
func (e *Engine) GenerateOutput(ctx context.Context, phase types.Phase) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	switch phase {
	case types.PhaseSpec:
		if e.ctx.Spec != nil {
			return e.generator.GenerateSpecMD(e.ctx.Spec)
		}

	case types.PhaseDesign:
		if e.ctx.Design != nil {
			return e.generator.GenerateDesignMD(e.ctx.Design)
		}

	case types.PhaseTasks:
		if e.ctx.Tasks != nil {
			return e.generator.GenerateTasksMD(e.ctx.Tasks)
		}
	}

	// Fallback: generate from input text
	return e.generator.GenerateFromTemplate(e.ctx.InputText, phase)
}

// GetIterations returns all iterations from the self-questioning loop.
func (e *Engine) GetIterations() []types.Iteration {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]types.Iteration, len(e.iterations))
	copy(result, e.iterations)
	return result
}

// GetScore returns the current score from the latest iteration.
func (e *Engine) GetScore() *types.Score {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.iterations) == 0 {
		return nil
	}

	return e.iterations[len(e.iterations)-1].Score
}

// SetOnIteration sets a callback for iteration events.
func (e *Engine) SetOnIteration(callback func(*types.Iteration)) {
	e.onIteration = callback
}

// SetOnProgress sets a callback for progress updates.
func (e *Engine) SetOnProgress(callback func(string, float64)) {
	e.onProgress = callback
}

// GetContext returns the current execution context.
func (e *Engine) GetContext() *types.Context {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ctx := *e.ctx
	return &ctx
}

// Phase execution methods

// runSpecPhase executes the spec generation phase.
func (e *Engine) runSpecPhase(ctx context.Context, input string) ([]types.Artifact, error) {
	e.reportProgress("Generating specification", 0.1)

	// Generate spec content
	content, err := e.generator.GenerateFromTemplate(input, types.PhaseSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to generate spec: %w", err)
	}

	e.reportProgress("Running quality checks", 0.4)

	// Run quality checks (result used in IterationLoop)
	_ = e.scorer.ScoreContent(content)

	e.reportProgress("Self-questioning loop", 0.6)

	// Run self-questioning
	improvedContent, finalScore, err := e.IterationLoop(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed in self-questioning: %w", err)
	}

	e.reportProgress("Finalizing specification", 0.9)

	// Update context
	e.mu.Lock()
	e.ctx.Spec = e.parseSpecFromContent(improvedContent)
	e.ctx.Score = finalScore
	e.ctx.OutputText = improvedContent
	e.mu.Unlock()

	return []types.Artifact{
		{
			Type:      types.ArtifactSpec,
			Path:      fmt.Sprintf("%s/SPEC.md", e.config.OutputPath),
			Content:   improvedContent,
			Generated: true,
		},
	}, nil
}

// runDesignPhase executes the design generation phase.
func (e *Engine) runDesignPhase(ctx context.Context, input string) ([]types.Artifact, error) {
	e.reportProgress("Generating design", 0.1)

	content, err := e.generator.GenerateFromTemplate(input, types.PhaseDesign)
	if err != nil {
		return nil, fmt.Errorf("failed to generate design: %w", err)
	}

	e.reportProgress("Validating architecture", 0.4)
	_ = e.scorer.ScoreContent(content)

	e.reportProgress("Self-questioning loop", 0.6)
	improvedContent, finalScore, err := e.IterationLoop(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed in self-questioning: %w", err)
	}

	e.mu.Lock()
	e.ctx.Design = e.parseDesignFromContent(improvedContent)
	e.ctx.Score = finalScore
	e.ctx.OutputText = improvedContent
	e.mu.Unlock()

	return []types.Artifact{
		{
			Type:      types.ArtifactDesign,
			Path:      fmt.Sprintf("%s/DESIGN.md", e.config.OutputPath),
			Content:   improvedContent,
			Generated: true,
		},
	}, nil
}

// runTasksPhase executes the tasks generation phase.
func (e *Engine) runTasksPhase(ctx context.Context, input string) ([]types.Artifact, error) {
	e.reportProgress("Generating tasks", 0.1)

	content, err := e.generator.GenerateFromTemplate(input, types.PhaseTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tasks: %w", err)
	}

	e.reportProgress("Validating task structure", 0.4)
	_ = e.scorer.ScoreContent(content)

	e.reportProgress("Self-questioning loop", 0.6)
	improvedContent, finalScore, err := e.IterationLoop(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed in self-questioning: %w", err)
	}

	e.mu.Lock()
	e.ctx.Tasks = e.parseTasksFromContent(improvedContent)
	e.ctx.Score = finalScore
	e.ctx.OutputText = improvedContent
	e.mu.Unlock()

	return []types.Artifact{
		{
			Type:      types.ArtifactTasks,
			Path:      fmt.Sprintf("%s/TASKS.md", e.config.OutputPath),
			Content:   improvedContent,
			Generated: true,
		},
	}, nil
}

// runVerifyPhase executes the verification phase.
func (e *Engine) runVerifyPhase(ctx context.Context, input string) ([]types.Artifact, error) {
	e.reportProgress("Running verification", 0.2)

	report, err := e.generator.GenerateVerifyReport(e.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification report: %w", err)
	}

	e.mu.Lock()
	e.ctx.Report = report
	e.mu.Unlock()

	return []types.Artifact{
		{
			Type:      types.ArtifactReport,
			Path:      fmt.Sprintf("%s/VERIFY.md", e.config.OutputPath),
			Content:   report.Summary,
			Generated: true,
		},
	}, nil
}

// Helper methods

func (e *Engine) reportProgress(stage string, progress float64) {
	if e.onProgress != nil {
		e.onProgress(stage, progress)
	}
}

func extractChangeName(input string) string {
	// Extract name from first line or first heading
	// This is a simplified version
	if len(input) == 0 {
		return "unnamed-change"
	}
	return "change-" + fmt.Sprintf("%d", time.Now().Unix())
}

func (e *Engine) parseSpecFromContent(content string) *types.SpecDocument {
	// Parse markdown content into SpecDocument struct
	// Simplified implementation
	return &types.SpecDocument{
		Title:        e.ctx.Change.Name,
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Overview:     content,
		Components:   []types.Component{},
		UserFlows:    []types.UserFlow{},
		Requirements: []types.Requirement{},
	}
}

func (e *Engine) parseDesignFromContent(content string) *types.DesignDocument {
	return &types.DesignDocument{
		Title:        e.ctx.Change.Name,
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Architecture: content,
		TechStack:    []types.TechStackItem{},
		Decisions:    []types.Decision{},
	}
}

func (e *Engine) parseTasksFromContent(content string) *types.TasksDocument {
	return &types.TasksDocument{
		Title:      e.ctx.Change.Name,
		Version:    "1.0.0",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Tasks:      []types.Task{},
		Milestones: []types.Milestone{},
	}
}

func joinQuestions(questions []string) string {
	result := ""
	for i, q := range questions {
		if i > 0 {
			result += "\n- "
		}
		result += q
	}
	return result
}
