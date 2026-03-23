package spec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Gentleman-Programming/grove/internal/engram"
	"github.com/Gentleman-Programming/grove/internal/sdd"
	"github.com/Gentleman-Programming/grove/internal/types"
)

// Engine is the main Spec engine that orchestrates the SDD workflow.
type Engine struct {
	config    *types.Config
	scorer    *Scorer
	generator *Generator

	// Integration with gentle-ai ecosystem
	sddClient    *sdd.Client
	engramClient *engram.EngramClient

	// LLM client for AI-powered operations
	llmClient *sdd.LLMClient

	// Input processor
	processor *InputProcessor

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

	// Get project directory from config or use default
	projectDir := config.ProjectPath
	if projectDir == "" {
		projectDir = "."
	}

	// Initialize SDD client
	sddClient := sdd.NewClient(projectDir)

	// Initialize Engram client (use localhost as default host)
	engramClient := engram.NewClient(engram.DefaultEngramHost)

	// Initialize LLM client (optional - may be nil if no API key)
	llmClient, _ := sdd.NewLLMClient()

	// Initialize Input processor
	processor := NewInputProcessor(config)

	return &Engine{
		config:       config,
		scorer:       NewScorer(config),
		generator:    NewGenerator(config),
		sddClient:    sddClient,
		engramClient: engramClient,
		llmClient:    llmClient,
		processor:    processor,
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
		// Execute SDD Explore phase before processing input
		if e.sddClient != nil {
			_, _ = e.sddClient.Execute(ctx, sdd.PhaseExplore, map[string]interface{}{
				"input": input,
				"phase": string(phase),
			})
		}

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

// SaveToEngram saves decisions and session summary to Engram after Run completes.
func (e *Engine) SaveToEngram(ctx context.Context, result *types.Result) error {
	if e.engramClient == nil {
		return nil
	}

	// Save spec decisions from iterations
	if len(e.iterations) > 0 {
		changeName := ""
		if e.ctx != nil && e.ctx.Change != nil {
			changeName = e.ctx.Change.Name
		}

		for i, iter := range e.iterations {
			if iter.Question != "" {
				decision := &engram.SpecDecision{
					ID:            fmt.Sprintf("iteration-%d", i+1),
					ChangeName:    changeName,
					Decision:      iter.Question,
					Justification: iter.Answer,
					Timestamp:     iter.Timestamp,
				}
				_ = e.engramClient.SaveSpecDecision(changeName, decision)
			}
		}
	}

	// Save session summary
	sessionID := fmt.Sprintf("spec-%d", time.Now().Unix())
	summary := &engram.SessionSummary{
		SessionID: sessionID,
		Project:   e.config.ProjectName,
		Goal:      "Spec generation completed",
		Discoveries: []string{
			fmt.Sprintf("Generated %d artifacts", len(result.Artifacts)),
			fmt.Sprintf("Completed %d iterations", len(e.iterations)),
		},
		Accomplished: []string{},
		NextSteps:    []string{},
		Timestamp:    time.Now(),
	}

	// Add accomplished items
	if result.Success {
		summary.Accomplished = append(summary.Accomplished, "Spec generation successful")
	}
	if len(result.Errors) > 0 {
		summary.Discoveries = append(summary.Discoveries, fmt.Sprintf("Errors: %s", strings.Join(result.Errors, ", ")))
	}

	return e.engramClient.SaveSessionSummary(summary)
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

	// Execute SDD PhaseSpec after generating specs
	if e.sddClient != nil {
		_, _ = e.sddClient.Execute(ctx, sdd.PhaseSpec, map[string]interface{}{
			"content": improvedContent,
			"phase":   string(types.PhaseSpec),
		})
	}

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
	title := "unnamed-spec"
	if e.ctx != nil && e.ctx.Change != nil {
		title = e.ctx.Change.Name
	}
	return &types.SpecDocument{
		Title:        title,
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
	title := "unnamed-design"
	if e.ctx != nil && e.ctx.Change != nil {
		title = e.ctx.Change.Name
	}
	return &types.DesignDocument{
		Title:        title,
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Architecture: content,
		TechStack:    []types.TechStackItem{},
		Decisions:    []types.Decision{},
	}
}

func (e *Engine) parseTasksFromContent(content string) *types.TasksDocument {
	title := "unnamed-tasks"
	if e.ctx != nil && e.ctx.Change != nil {
		title = e.ctx.Change.Name
	}
	return &types.TasksDocument{
		Title:      title,
		Version:    "1.0.0",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Tasks:      []types.Task{},
		Milestones: []types.Milestone{},
	}
}

// Generate executes the full spec generation workflow using Component Decomposition Loop.
// It processes input, extracts components, runs self-questioning, generates spec via LLM,
// and saves artifacts to the spec/ directory.
func (e *Engine) Generate(ctx context.Context, input string) (*types.SpecDocument, error) {
	// 1. Process input with InputProcessor
	processed, err := e.processor.Process(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("process input: %w", err)
	}

	// 2. Decompose into components
	components, err := e.ComponentDecomposition(ctx, processed)
	if err != nil {
		return nil, fmt.Errorf("component decomposition: %w", err)
	}

	// 3. Self-questioning
	questions := e.GenerateSelfQuestions(ctx, processed, components)
	answers := e.answerSelfQuestionsFromLLM(ctx, questions)

	// 4. Generate spec draft with LLM
	spec, err := e.generateSpecDraft(ctx, processed, components, answers)
	if err != nil {
		return nil, fmt.Errorf("generate spec: %w", err)
	}

	// 5. Save artifacts to spec/ directory
	e.saveArtifacts(spec, components)

	return spec, nil
}

// GenerateSelfQuestions generates self-reflection questions for the input and components.
func (e *Engine) GenerateSelfQuestions(ctx context.Context, processed *types.ProcessedInput, components []string) string {
	questions := []string{
		"¿Están todas las funcionalidades críticas cubiertas con requisitos específicos?",
		"¿Los escenarios de prueba son completos y cubren los casos borde?",
		"¿Hay ambigüedades en la terminología usada?",
		"¿Las dependencias externas están justificadas?",
		"¿Los estimates de tiempo son realistas?",
	}

	// Add component-specific questions
	for _, comp := range components {
		questions = append(questions, fmt.Sprintf("¿El componente %s tiene todas sus interfaces bien definidas?", comp))
	}

	return "Self-Reflection Questions:\n" + joinQuestions(questions)
}

// answerSelfQuestionsFromLLM uses LLM to answer self-reflection questions.
func (e *Engine) answerSelfQuestionsFromLLM(ctx context.Context, questions string) string {
	if e.llmClient == nil {
		return "LLM client not available - using default answers"
	}

	prompt := fmt.Sprintf(`
Responde las siguientes preguntas de auto-reflexión para mejorar la especificación:

%s

Proporciona respuestas concisas y actionable.`, questions)

	response, err := e.llmClient.Send(ctx, prompt)
	if err != nil {
		return fmt.Sprintf("Error getting LLM response: %v", err)
	}

	return response
}

// ComponentDecomposition analyzes the processed input and extracts main components.
// It uses LLM for intelligent decomposition when available, or falls back to pattern matching.
func (e *Engine) ComponentDecomposition(ctx context.Context, input *types.ProcessedInput) ([]string, error) {
	if input == nil {
		return nil, nil
	}

	// If LLM is available, use it for intelligent decomposition
	if e.llmClient != nil {
		return e.decomposeWithLLM(ctx, input)
	}

	// Fallback: use pattern-based decomposition
	return e.decomposeWithPatterns(input), nil
}

// decomposeWithLLM uses the LLM to decompose input into components.
func (e *Engine) decomposeWithLLM(ctx context.Context, input *types.ProcessedInput) ([]string, error) {
	prompt := fmt.Sprintf(`Analiza el siguiente input y descomponlo en componentes principales.
Cada componente debe ser un módulo, servicio o característica distinta del sistema.

Input:
%s

Detected Stack: %s

Lista los componentes identificados con formato:
COMPONENT: nombre
DESCRIPCION: descripción breve
`, input.ParsedContent, strings.Join(input.DetectedStack, ", "))

	response, err := e.llmClient.Send(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM decomposition failed: %w", err)
	}

	return parseComponentsFromResponse(response)
}

// decomposeWithPatterns uses pattern matching to extract components from input.
func (e *Engine) decomposeWithPatterns(input *types.ProcessedInput) []string {
	var components []string

	// Use extracted types as base components
	for _, t := range input.ExtractedTypes {
		components = append(components, capitalize(t))
	}

	// Add detected stack components
	for _, s := range input.DetectedStack {
		if !containsString(components, s) {
			components = append(components, s)
		}
	}

	// If no components found, add a default
	if len(components) == 0 {
		components = append(components, "Core")
	}

	return components
}

// generateSpecDraft generates the spec document using LLM.
func (e *Engine) generateSpecDraft(ctx context.Context, processed *types.ProcessedInput, components []string, answers string) (*types.SpecDocument, error) {
	if e.llmClient == nil {
		// Return a basic spec if LLM is not available
		return e.createBasicSpec(processed, components), nil
	}

	prompt := fmt.Sprintf(`Genera un documento de especificación completo en formato Markdown.

## Input Original:
%s

## Componentes Identificados:
%s

## Respuestas de Auto-Reflexión:
%s

Genera una especificación que incluya:
- Visión general del proyecto
- Lista de componentes con descripción
- Requisitos funcionales para cada componente
- Flujos de usuario principales
- Supuestos y dependencias

 Formato: Markdown`, processed.OriginalInput, strings.Join(components, "\n- "), answers)

	response, err := e.llmClient.Send(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("generate spec draft: %w", err)
	}

	return e.parseSpecFromMarkdown(response, components), nil
}

// createBasicSpec creates a basic spec document without LLM.
func (e *Engine) createBasicSpec(processed *types.ProcessedInput, components []string) *types.SpecDocument {
	overview := ""
	if processed != nil {
		overview = processed.OriginalInput
	}

	spec := &types.SpecDocument{
		Title:        "Generated Specification",
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Overview:     overview,
		Components:   make([]types.Component, 0),
		UserFlows:    []types.UserFlow{},
		Requirements: []types.Requirement{},
	}

	// Convert component strings to Component structs
	for i, compName := range components {
		spec.Components = append(spec.Components, types.Component{
			ID:          fmt.Sprintf("comp-%d", i+1),
			Name:        compName,
			Description: fmt.Sprintf("Component %s extracted from input", compName),
			Type:        types.ComponentTypeBackend,
		})
	}

	return spec
}

// parseSpecFromMarkdown parses LLM response into a SpecDocument.
func (e *Engine) parseSpecFromMarkdown(markdown string, components []string) *types.SpecDocument {
	spec := &types.SpecDocument{
		Title:        "Generated Specification",
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Overview:     markdown,
		Components:   make([]types.Component, 0),
		UserFlows:    []types.UserFlow{},
		Requirements: []types.Requirement{},
	}

	// Convert component strings to Component structs
	for i, compName := range components {
		spec.Components = append(spec.Components, types.Component{
			ID:          fmt.Sprintf("comp-%d", i+1),
			Name:        compName,
			Description: fmt.Sprintf("Component %s from decomposition", compName),
			Type:        types.ComponentTypeBackend,
		})
	}

	return spec
}

// saveArtifacts saves the generated spec and component info to the spec/ directory.
func (e *Engine) saveArtifacts(spec *types.SpecDocument, components []string) {
	// Determine output directory
	outputDir := e.config.OutputPath
	if outputDir == "" {
		outputDir = "spec"
	}

	// Create spec directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: could not create spec directory: %v\n", err)
		return
	}

	// Save SPEC.md
	specPath := filepath.Join(outputDir, "SPEC.md")
	if content, err := e.generator.GenerateSpecMD(spec); err == nil {
		_ = os.WriteFile(specPath, []byte(content), 0644)
	}

	// Save components list
	componentsPath := filepath.Join(outputDir, "components.md")
	componentsContent := "# Componentes Descompuestos\n\n"
	for _, comp := range components {
		componentsContent += fmt.Sprintf("- %s\n", comp)
	}
	_ = os.WriteFile(componentsPath, []byte(componentsContent), 0644)
}

// parseComponentsFromResponse parses LLM response to extract component names.
func parseComponentsFromResponse(response string) ([]string, error) {
	var components []string
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(line), "COMPONENT:") {
			// Extract component name after "COMPONENT:"
			name := strings.TrimPrefix(line, "COMPONENT:")
			name = strings.TrimSpace(name)
			if name != "" {
				components = append(components, name)
			}
		}
	}

	// If no components found with COMPONENT: prefix, try to parse differently
	if len(components) == 0 {
		// Split by common delimiters
		re := regexp.MustCompile(`[,;|\n]`)
		parts := re.Split(response, -1)
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if len(part) > 2 && len(part) < 50 {
				components = append(components, part)
			}
		}
	}

	return components, nil
}

// UpdateSpec updates an existing spec with new inputs.
// It reads the existing spec, compares with new inputs, and generates a diff
// while preserving sections that are not affected by the new inputs.
func (e *Engine) UpdateSpec(ctx context.Context, existingSpecPath string, newInputs string) (*types.SpecDocument, error) {
	// Read existing spec
	existingContent, err := os.ReadFile(existingSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read existing spec: %w", err)
	}

	// Parse existing spec
	existingSpec := e.parseSpecFromMarkdown(string(existingContent), []string{})

	// Process new inputs
	processed, err := e.processor.Process(ctx, newInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to process new inputs: %w", err)
	}

	// Decompose new inputs into components
	newComponents, err := e.ComponentDecomposition(ctx, processed)
	if err != nil {
		return nil, fmt.Errorf("failed to decompose new inputs: %w", err)
	}

	// Generate diff by comparing existing and new components
	updatedSpec := e.generateSpecDiff(existingSpec, processed, newComponents)

	// Update timestamp
	updatedSpec.UpdatedAt = time.Now()

	return updatedSpec, nil
}

// generateSpecDiff generates a diff between existing and new spec.
func (e *Engine) generateSpecDiff(existingSpec *types.SpecDocument, processed *types.ProcessedInput, newComponents []string) *types.SpecDocument {
	// Create updated spec
	updatedSpec := &types.SpecDocument{
		Title:        existingSpec.Title,
		Version:      incrementVersion(existingSpec.Version),
		CreatedAt:    existingSpec.CreatedAt,
		UpdatedAt:    time.Now(),
		Overview:     existingSpec.Overview,
		Components:   existingSpec.Components,
		UserFlows:    existingSpec.UserFlows,
		Requirements: existingSpec.Requirements,
		Assumptions:  existingSpec.Assumptions,
	}

	// Track which existing components are still relevant
	existingComponentNames := make(map[string]bool)
	for _, comp := range existingSpec.Components {
		existingComponentNames[comp.Name] = true
	}

	// Add new components that don't exist in the spec
	for _, newComp := range newComponents {
		if !existingComponentNames[newComp] {
			// New component from inputs
			updatedSpec.Components = append(updatedSpec.Components, types.Component{
				ID:          fmt.Sprintf("comp-%d", len(updatedSpec.Components)+1),
				Name:        newComp,
				Description: fmt.Sprintf("Component %s extracted from new inputs", newComp),
				Type:        types.ComponentTypeBackend,
				Inferred:    true,
			})
		}
	}

	// Add requirements from new inputs
	if len(processed.ExtractedTypes) > 0 {
		for _, et := range processed.ExtractedTypes {
			reqID := fmt.Sprintf("REQ-%d", len(updatedSpec.Requirements)+1)
			updatedSpec.Requirements = append(updatedSpec.Requirements, types.Requirement{
				ID:          reqID,
				Type:        "functional",
				Description: fmt.Sprintf("Support for %s functionality", et),
				Priority:    "should",
			})
		}
	}

	return updatedSpec
}

// incrementVersion increments the semantic version.
func incrementVersion(version string) string {
	// Simple version increment - in production would use semver
	parts := strings.Split(version, ".")
	if len(parts) == 3 {
		// Parse patch version and increment
		var patch int
		if _, err := fmt.Sscanf(parts[2], "%d", &patch); err == nil {
			return fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch+1)
		}
	}
	return "1.0.1"
}

// ReverseSpec analyzes existing code and generates an initial spec.
// It uses AST parsing or regex to extract functions, types, and dependencies.
func (e *Engine) ReverseSpec(ctx context.Context, codePath string) (*types.SpecDocument, error) {
	// Validate path
	info, err := os.Stat(codePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat code path: %w", err)
	}

	var codeFiles []string

	if info.IsDir() {
		// Scan for code files
		codeFiles, err = scanCodeFiles(codePath)
		if err != nil {
			return nil, fmt.Errorf("failed to scan code files: %w", err)
		}
	} else {
		codeFiles = []string{codePath}
	}

	if len(codeFiles) == 0 {
		return nil, errors.New("no code files found")
	}

	// Analyze each code file
	var allFunctions []string
	var allTypes []string
	var allImports []string

	for _, file := range codeFiles {
		funcs, types, imports := e.analyzeCodeFile(file)
		allFunctions = append(allFunctions, funcs...)
		allTypes = append(allTypes, types...)
		allImports = append(allImports, imports...)
	}

	// Deduplicate
	allFunctions = deduplicateStrings(allFunctions)
	allTypes = deduplicateStrings(allTypes)
	allImports = deduplicateStrings(allImports)

	// Generate spec from analyzed code
	spec := e.generateSpecFromCode(allFunctions, allTypes, allImports, codePath)

	return spec, nil
}

// scanCodeFiles recursively scans a directory for code files.
func scanCodeFiles(dir string) ([]string, error) {
	var files []string

	// Common code file extensions
	codeExtensions := map[string]bool{
		".go": true, ".ts": true, ".tsx": true, ".js": true,
		".jsx": true, ".py": true, ".java": true, ".rs": true,
	}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if codeExtensions[ext] {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// analyzeCodeFile analyzes a single code file and extracts functions, types, and imports.
func (e *Engine) analyzeCodeFile(filePath string) (functions []string, types []string, imports []string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		functions, types, imports = e.analyzeGoFile(string(content))
	case ".ts", ".tsx", ".js", ".jsx":
		functions, types, imports = e.analyzeTSJSFile(string(content))
	case ".py":
		functions, types, imports = e.analyzePythonFile(string(content))
	default:
		// Generic analysis
		functions = e.extractGenericFunctions(string(content))
		types = e.extractGenericTypes(string(content))
	}

	return
}

// analyzeGoFile analyzes Go source code.
func (e *Engine) analyzeGoFile(content string) (functions []string, types []string, imports []string) {
	// Extract imports
	importRe := regexp.MustCompile(`import\s+"(.*?)"`)
	for _, match := range importRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}

	// Extract named imports (import (
	importBlockRe := regexp.MustCompile(`import\s+\(([\s\S]*?)\)`)
	importBlock := importBlockRe.FindStringSubmatch(content)
	if len(importBlock) > 1 {
		namedImportRe := regexp.MustCompile(`"([^"]+)"`)
		for _, match := range namedImportRe.FindAllStringSubmatch(importBlock[1], -1) {
			if len(match) > 1 {
				imports = append(imports, match[1])
			}
		}
	}

	// Extract functions (func name(...) or func (receiver) name(...))
	funcRe := regexp.MustCompile(`func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(`)
	for _, match := range funcRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			functions = append(functions, match[1])
		}
	}

	// Extract types (type Name struct or type Name interface)
	typeRe := regexp.MustCompile(`type\s+(\w+)\s+(?:struct|interface|enum)`)
	for _, match := range typeRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			types = append(types, match[1])
		}
	}

	return
}

// analyzeTSJSFile analyzes TypeScript/JavaScript source code.
func (e *Engine) analyzeTSJSFile(content string) (functions []string, types []string, imports []string) {
	// Extract imports (import { ... } from "..." or import ... from "...")
	importRe := regexp.MustCompile(`import\s+(?:\{[^}]*\}|\w+)\s+from\s+"([^"]+)"`)
	for _, match := range importRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}

	// Extract require
	requireRe := regexp.MustCompile(`require\s*\(\s*"([^"]+)"\s*\)`)
	for _, match := range requireRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}

	// Extract functions (function name(...) or const name = (...)=> ...)
	funcRe := regexp.MustCompile(`function\s+(\w+)\s*\(`)
	for _, match := range funcRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			functions = append(functions, match[1])
		}
	}

	// Extract arrow functions assigned to const
	arrowRe := regexp.MustCompile(`const\s+(\w+)\s*=\s*(?:\([^)]*\)|[^=])\s*=>`)
	for _, match := range arrowRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			functions = append(functions, match[1])
		}
	}

	// Extract classes
	classRe := regexp.MustCompile(`class\s+(\w+)`)
	for _, match := range classRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			types = append(types, match[1])
		}
	}

	// Extract interfaces
	interfaceRe := regexp.MustCompile(`interface\s+(\w+)`)
	for _, match := range interfaceRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			types = append(types, match[1])
		}
	}

	return
}

// analyzePythonFile analyzes Python source code.
func (e *Engine) analyzePythonFile(content string) (functions []string, types []string, imports []string) {
	// Extract imports
	importRe := regexp.MustCompile(`(?:from\s+(\S+)\s+import|import\s+(\S+))`)
	for _, match := range importRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 && match[1] != "" {
			imports = append(imports, match[1])
		} else if len(match) > 2 && match[2] != "" {
			imports = append(imports, match[2])
		}
	}

	// Extract functions (def name(...))
	funcRe := regexp.MustCompile(`def\s+(\w+)\s*\(`)
	for _, match := range funcRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			functions = append(functions, match[1])
		}
	}

	// Extract classes (class Name(...))
	classRe := regexp.MustCompile(`class\s+(\w+)`)
	for _, match := range classRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			types = append(types, match[1])
		}
	}

	return
}

// extractGenericFunctions extracts functions using generic patterns.
func (e *Engine) extractGenericFunctions(content string) []string {
	var funcs []string
	re := regexp.MustCompile(`(?i)(?:function|def|func|proc|sub)\s+(\w+)`)
	for _, match := range re.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			funcs = append(funcs, match[1])
		}
	}
	return funcs
}

// extractGenericTypes extracts types using generic patterns.
func (e *Engine) extractGenericTypes(content string) []string {
	var types []string
	re := regexp.MustCompile(`(?i)(?:class|type|struct|interface|enum)\s+(\w+)`)
	for _, match := range re.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			types = append(types, match[1])
		}
	}
	return types
}

// generateSpecFromCode generates a spec document from analyzed code.
func (e *Engine) generateSpecFromCode(functions []string, codeTypes []string, imports []string, codePath string) *types.SpecDocument {
	spec := &types.SpecDocument{
		Title:        filepath.Base(codePath),
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Overview:     "Auto-generated specification from reverse engineering of existing code",
		Components:   make([]types.Component, 0),
		UserFlows:    []types.UserFlow{},
		Requirements: make([]types.Requirement, 0),
	}

	// Add components for each type found
	for i, t := range codeTypes {
		spec.Components = append(spec.Components, types.Component{
			ID:          fmt.Sprintf("comp-%d", i+1),
			Name:        t,
			Description: fmt.Sprintf("Type/Class %s identified from code", t),
			Type:        types.ComponentTypeBackend,
			Inferred:    true,
		})
	}

	// Add components for each major function group
	// Group functions by potential module/service
	moduleMap := make(map[string][]string)
	for _, f := range functions {
		// Skip private functions (starting with _ or lowercase in Go)
		if len(f) > 0 && (f[0] >= 'A' && f[0] <= 'Z') {
			moduleMap["Core"] = append(moduleMap["Core"], f)
		}
	}

	for module, funcs := range moduleMap {
		compID := fmt.Sprintf("comp-%d", len(spec.Components)+1)
		spec.Components = append(spec.Components, types.Component{
			ID:          compID,
			Name:        module,
			Description: fmt.Sprintf("Module with %d functions: %s", len(funcs), strings.Join(funcs[:min(3, len(funcs))], ", ")),
			Type:        types.ComponentTypeService,
			Inferred:    true,
		})
	}

	// Add requirements based on detected functionality
	if len(functions) > 0 {
		spec.Requirements = append(spec.Requirements, types.Requirement{
			ID:          "REQ-1",
			Type:        "functional",
			Description: fmt.Sprintf("Implement %d identified functions", len(functions)),
			Priority:    "must",
		})
	}

	// Add dependency requirements from imports
	for _, imp := range imports {
		spec.Requirements = append(spec.Requirements, types.Requirement{
			ID:          fmt.Sprintf("REQ-%d", len(spec.Requirements)+1),
			Type:        "non-functional",
			Description: fmt.Sprintf("Dependency: %s", imp),
			Priority:    "must",
		})
	}

	// Add assumptions about the codebase
	spec.Assumptions = []types.Assumption{
		{
			ID:        "ASM-1",
			Statement: "Code analysis performed via static analysis",
			Rationale: "No runtime information available",
		},
		{
			ID:        "ASM-2",
			Statement: fmt.Sprintf("Found %d functions and %d types", len(functions), len(codeTypes)),
			Rationale: "Extracted from source code",
		},
	}

	return spec
}

// deduplicateStrings removes duplicates from a string slice.
func deduplicateStrings(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// Helper functions

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[0:1]) + s[1:]
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
