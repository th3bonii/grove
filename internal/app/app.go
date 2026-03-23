// Package app provides the main application logic for GROVE.
package app

import (
	"context"
	"fmt"
	"os"

	"github.com/Gentleman-Programming/grove/internal/logger"
	"github.com/Gentleman-Programming/grove/internal/types"
)

// AppConfig holds the global application configuration.
type AppConfig struct {
	ProjectRoot  string
	Verbose      bool
	Quiet        bool
	OutputFormat string // "text", "json", "markdown"
}

// SpecOptions contains the configuration for grove-spec execution.
type SpecOptions struct {
	Input       string
	Output      string
	Update      bool
	Reverse     bool
	LoopMax     int
	Resume      bool
	FullRescore bool
	Feedback    *QualityGateFeedback
}

// QualityGateFeedback is the structured feedback payload from Ralph Loop.
type QualityGateFeedback struct {
	Trigger       string
	LoopNumber    int
	QualityScore  float64
	Missing       []string
	QualityScores map[string]float64
	FailedTasks   []string
	Observations  []string
}

// LoopOptions contains the configuration for grove-loop execution.
type LoopOptions struct {
	PauseAfter       string
	Status           bool
	Report           bool
	AutoCommit       bool
	Resume           bool
	QualityThreshold float64
}

// OptiOptions contains the configuration for grove-opti execution.
type OptiOptions struct {
	Clipboard  bool
	Batch      string
	ExplainAll bool
	MaxTokens  int
	Scope      string
	Templates  string
	Prompt     string
}

// App is the main GROVE application instance.
type App struct {
	config *AppConfig
	log    *logger.GroveLogger
}

// New creates a new GROVE application instance.
func New(config *AppConfig) *App {
	if config == nil {
		config = &AppConfig{}
	}
	if config.OutputFormat == "" {
		config.OutputFormat = "text"
	}

	// Configure logger based on config
	logCfg := logger.DefaultConfig()
	if config.Verbose {
		logCfg.Level = logger.LevelDebug
	}
	if config.Quiet {
		logCfg.Pretty = false
	}

	return &App{
		config: config,
		log:    logger.New(logCfg),
	}
}

// ============================================================================
// Grove Spec Implementation
// ============================================================================

// RunSpec runs grove-spec with command-line arguments.
func (a *App) RunSpec(args []string) error {
	opts, err := parseSpecArgs(args)
	if err != nil {
		return err
	}
	return a.RunSpecWithOptions(opts)
}

// RunSpecWithOptions runs grove-spec with explicit options.
func (a *App) RunSpecWithOptions(opts *SpecOptions) error {
	ctx := context.Background()

	if opts == nil {
		opts = &SpecOptions{}
	}

	// Set defaults
	if opts.Input == "" {
		opts.Input = a.config.ProjectRoot
		if opts.Input == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not determine working directory: %w", err)
			}
			opts.Input = cwd
		}
	}
	if opts.Output == "" {
		opts.Output = "spec"
	}
	if opts.LoopMax == 0 {
		opts.LoopMax = 10
	}

	a.log.LogSpecOperationStart(ctx, logger.SpecOpComplete, "grove-spec")
	a.log.LogInfo(ctx, "GROVE Spec starting", map[string]any{
		"input":  opts.Input,
		"output": opts.Output,
	})

	if opts.Update {
		a.log.LogInfo(ctx, "Mode: Incremental Update", nil)
	} else if opts.Reverse {
		a.log.LogInfo(ctx, "Mode: Reverse Documentation", nil)
	} else {
		a.log.LogInfo(ctx, "Mode: Full Generation", nil)
	}

	// TODO: Implement actual spec generation logic
	// This is a stub that demonstrates the CLI integration

	loopState := &types.LoopState{
		LoopNumber:    0,
		ExitCondition: types.ExitNormal,
	}

	a.log.LogInfo(ctx, "Starting spec loop", map[string]any{
		"max_iterations": opts.LoopMax,
	})

	for loopState.LoopNumber < opts.LoopMax {
		loopState.LoopNumber++
		a.log.LogLoopIteration(ctx, loopState.LoopNumber, logger.LoopPhaseImplementation, "iteration", map[string]any{
			"iteration": loopState.LoopNumber,
		})

		// Simulate spec generation work
		// In real implementation, this would:
		// 1. Read input files
		// 2. Classify components
		// 3. Generate SPEC.md, DESIGN.md, TASKS.md
		// 4. Score quality dimensions
		// 5. Check for convergence

		// For demo, exit after first loop
		if loopState.LoopNumber >= 1 {
			break
		}
	}

	a.log.LogSpecOperationEnd(ctx, logger.SpecOpComplete, "grove-spec", 0, true)
	a.log.LogInfo(ctx, "Spec generation complete", map[string]any{
		"files_generated": []string{"spec/SPEC.md", "spec/DESIGN.md", "spec/TASKS.md"},
	})

	return nil
}

func parseSpecArgs(args []string) (*SpecOptions, error) {
	opts := &SpecOptions{}

	// Simple argument parsing for demo
	// In production, use flag parsing like in cmd/grove-spec/main.go
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--input", "-i":
			if i+1 < len(args) {
				opts.Input = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				opts.Output = args[i+1]
				i++
			}
		case "--update", "-u":
			opts.Update = true
		case "--reverse", "-r":
			opts.Reverse = true
		case "--loop-max":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &opts.LoopMax)
				i++
			}
		case "--resume":
			opts.Resume = true
		case "--full-rescore":
			opts.FullRescore = true
		}
	}

	return opts, nil
}

// ============================================================================
// Grove Loop Implementation
// ============================================================================

// RunLoop runs grove-loop with command-line arguments.
func (a *App) RunLoop(args []string) error {
	opts, err := parseLoopArgs(args)
	if err != nil {
		return err
	}
	return a.RunLoopWithOptions(opts)
}

// RunLoopWithOptions runs grove-loop with explicit options.
func (a *App) RunLoopWithOptions(opts *LoopOptions) error {
	ctx := context.Background()

	if opts == nil {
		opts = &LoopOptions{}
	}

	// Set defaults
	if opts.QualityThreshold == 0 {
		opts.QualityThreshold = 70.0
	}

	// Handle status-only mode
	if opts.Status {
		return a.showLoopStatus(opts)
	}

	// Handle report-only mode
	if opts.Report {
		return a.generateLoopReport(opts)
	}

	a.log.LogInfo(ctx, "GROVE Ralph Loop starting", map[string]any{
		"project":           a.config.ProjectRoot,
		"quality_threshold": opts.QualityThreshold,
	})

	if opts.Resume {
		a.log.LogInfo(ctx, "Mode: Resume from saved state", nil)
	} else if opts.PauseAfter != "" {
		a.log.LogInfo(ctx, "Mode: Pause after task", map[string]any{
			"task_id": opts.PauseAfter,
		})
	} else {
		a.log.LogInfo(ctx, "Mode: New loop", nil)
	}

	// TODO: Implement actual loop execution logic
	// This is a stub that demonstrates the CLI integration

	loopState := &types.LoopState{
		LoopNumber: 0,
	}

	a.log.LogInfo(ctx, "Starting build loop", nil)

	// Simulate loop execution
	for loopState.LoopNumber < 3 {
		loopState.LoopNumber++
		a.log.LogLoopIteration(ctx, loopState.LoopNumber, logger.LoopPhaseImplementation, "iteration", map[string]any{
			"iteration": loopState.LoopNumber,
		})

		// TODO: Implement actual SDD cycle:
		// 1. Pre-loop validation
		// 2. Quality gate check
		// 3. For each unimplemented task:
		//    - Spawn implementation sub-agent
		//    - Spawn verify sub-agent
		//    - Mark task complete or flag for retry
		// 4. Phase-level consistency checks
		// 5. Update state file

		if opts.PauseAfter != "" && loopState.LoopNumber >= 1 {
			a.log.LogWarning(ctx, "Pause requested - saving state", nil)
			break
		}
	}

	a.log.LogInfo(ctx, "Build loop complete", map[string]any{
		"log_file": "GROVE-LOOP-LOG.md",
	})

	return nil
}

func parseLoopArgs(args []string) (*LoopOptions, error) {
	opts := &LoopOptions{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--pause-after":
			if i+1 < len(args) {
				opts.PauseAfter = args[i+1]
				i++
			}
		case "--status", "-s":
			opts.Status = true
		case "--report", "-r":
			opts.Report = true
		case "--auto-commit":
			opts.AutoCommit = true
		case "--resume":
			opts.Resume = true
		case "--quality":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%f", &opts.QualityThreshold)
				i++
			}
		}
	}

	return opts, nil
}

func (a *App) showLoopStatus(opts *LoopOptions) error {
	ctx := context.Background()
	a.log.LogInfo(ctx, "GROVE Loop Status", nil)

	// TODO: Read and display GROVE-LOOP-STATE.json
	stateFile := "GROVE-LOOP-STATE.json"

	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		a.log.LogWarning(ctx, "No loop state found", map[string]any{
			"hint": "Run 'grove loop' to start a new loop",
		})
		return nil
	}

	a.log.LogInfo(ctx, "State file found", map[string]any{
		"file": stateFile,
	})
	a.log.LogInfo(ctx, "Full state display not yet implemented", nil)

	return nil
}

func (a *App) generateLoopReport(opts *LoopOptions) error {
	ctx := context.Background()
	a.log.LogInfo(ctx, "Generating loop report...", nil)

	// TODO: Generate GROVE-READY-REPORT.md
	reportFile := "GROVE-READY-REPORT.md"

	a.log.LogInfo(ctx, "Report generated", map[string]any{
		"file": reportFile,
	})

	return nil
}

// ============================================================================
// Grove Opti Prompt Implementation
// ============================================================================

// RunOpti runs grove-opti with command-line arguments.
func (a *App) RunOpti(args []string) error {
	opts, err := parseOptiArgs(args)
	if err != nil {
		return err
	}
	return a.RunOptiWithOptions(opts)
}

// RunOptiWithOptions runs grove-opti with explicit options.
func (a *App) RunOptiWithOptions(opts *OptiOptions) error {
	ctx := context.Background()

	if opts == nil {
		return fmt.Errorf("options cannot be nil")
	}

	// Set defaults
	if opts.MaxTokens == 0 {
		opts.MaxTokens = 2000
	}

	a.log.LogInfo(ctx, "GROVE Opti Prompt starting", nil)

	// Batch mode
	if opts.Batch != "" {
		return a.runBatchMode(opts)
	}

	// Single prompt mode
	if opts.Prompt == "" {
		return fmt.Errorf("no prompt provided. Use argument or --clipboard flag")
	}

	a.log.LogInfo(ctx, "Input prompt", map[string]any{
		"prompt": truncate(opts.Prompt, 50),
	})

	// TODO: Implement actual prompt optimization
	// This is a stub that demonstrates the CLI integration

	// Step 1: Classify intent
	a.log.LogInfo(ctx, "Classifying intent...", nil)
	intent := &types.Intent{
		Type:       types.IntentFeatureAddition,
		Confidence: 0.85,
	}
	a.log.LogInfo(ctx, "Intent classified", map[string]any{
		"type":       intent.Type,
		"confidence": intent.Confidence,
	})

	// Step 2: Collect context
	a.log.LogInfo(ctx, "Collecting context", map[string]any{
		"token_budget": opts.MaxTokens,
	})
	files := []string{
		"src/components/Example.tsx",
		"src/hooks/useTheme.ts",
	}
	for _, f := range files {
		a.log.LogInfo(ctx, "File selected", map[string]any{"file": f})
	}

	// Step 3: Optimize prompt
	a.log.LogInfo(ctx, "Optimizing prompt...", nil)

	optimized := &types.OptimizedPrompt{
		Original:   opts.Prompt,
		Optimized:  opts.Prompt + " [optimized with @file references and skill() calls]",
		Intent:     *intent,
		TokenCount: 1250,
	}

	// Step 4: Display results
	a.log.LogInfo(ctx, "=== OPTIMIZED PROMPT ===", nil)
	a.log.LogInfo(ctx, optimized.Optimized, nil)
	a.log.LogInfo(ctx, "Optimization complete", map[string]any{
		"token_count":  optimized.TokenCount,
		"token_budget": opts.MaxTokens,
		"explain_all":  opts.ExplainAll,
	})

	return nil
}

func parseOptiArgs(args []string) (*OptiOptions, error) {
	opts := &OptiOptions{}

	// Collect non-flag arguments as the prompt
	var promptArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--clipboard", "-c":
			opts.Clipboard = true
		case "--batch", "-b":
			if i+1 < len(args) {
				opts.Batch = args[i+1]
				i++
			}
		case "--explain-all", "-e":
			opts.ExplainAll = true
		case "--max-tokens":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &opts.MaxTokens)
				i++
			}
		case "--scope", "-s":
			if i+1 < len(args) {
				opts.Scope = args[i+1]
				i++
			}
		case "--templates", "-t":
			if i+1 < len(args) {
				opts.Templates = args[i+1]
				i++
			}
		default:
			if len(arg) > 0 && arg[0] != '-' {
				promptArgs = append(promptArgs, arg)
			}
		}
	}

	if len(promptArgs) > 0 {
		opts.Prompt = joinArgs(promptArgs)
	}

	return opts, nil
}

func (a *App) runBatchMode(opts *OptiOptions) error {
	ctx := context.Background()
	a.log.LogInfo(ctx, "Batch mode", map[string]any{
		"input_file": opts.Batch,
	})

	// TODO: Read file and process each line as a prompt
	// Output to GROVE-OPTI-BATCH-<timestamp>.md

	batchFile := fmt.Sprintf("GROVE-OPTI-BATCH-%d.md", os.Getpid())
	a.log.LogInfo(ctx, "Output file", map[string]any{
		"file": batchFile,
	})
	a.log.LogInfo(ctx, "Batch processing not yet fully implemented", nil)

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func joinArgs(args []string) string {
	result := ""
	for _, arg := range args {
		if result != "" {
			result += " "
		}
		result += arg
	}
	return result
}
