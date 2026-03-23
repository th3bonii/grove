// Package main implements the grove-opti CLI command.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Gentleman-Programming/grove/internal/app"
	"github.com/Gentleman-Programming/grove/internal/logger"
)

const (
	defaultMaxTokens = 2000
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		logger.Error("grove-opti failed: %v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("grove-opti", flag.ContinueOnError)
	fs.Usage = usage

	var (
		clipboardFlag  = fs.Bool("clipboard", false, "Read input prompt from system clipboard")
		batchFlag      = fs.String("batch", "", "Process prompts from specified file (one per line)")
		explainAllFlag = fs.Bool("explain-all", false, "Force full explanations regardless of history")
		maxTokensFlag  = fs.Int("max-tokens", defaultMaxTokens, "Maximum token budget for context collection")
		scopeFlag      = fs.String("scope", "", "Limit context collection to specified module/component")
		templatesFlag  = fs.String("templates", "", "Custom templates directory")
		verboseFlag    = fs.Bool("v", false, "Enable verbose output")
		quietFlag      = fs.Bool("q", false, "Suppress non-essential output")
		helpFlag       = fs.Bool("h", false, "Show help message")
	)

	// Parse flags
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	// Show help if requested
	if *helpFlag {
		fs.Usage()
		return nil
	}

	// Configure logger
	logger.SetVerbose(*verboseFlag)
	logger.SetQuiet(*quietFlag)

	// Validate token budget
	if *maxTokensFlag <= 0 || *maxTokensFlag > 10000 {
		return fmt.Errorf("max-tokens must be between 1 and 10000")
	}

	// Get the prompt from remaining args or clipboard
	var prompt string
	var err error

	switch {
	case *batchFlag != "":
		// Batch mode - pass the file path
		prompt = *batchFlag

	case *clipboardFlag:
		// Read from clipboard
		prompt, err = readClipboard()
		if err != nil {
			return fmt.Errorf("failed to read clipboard: %w", err)
		}

	case fs.NArg() > 0:
		// Use remaining args as prompt
		prompt = fs.Arg(0)

	default:
		// Interactive mode - prompt for input
		fmt.Println("Enter your prompt (Ctrl+D to finish, Ctrl+C to cancel):")
		var lines []string
		for {
			var line string
			_, err := fmt.Scanln(&line)
			if err != nil {
				break
			}
			lines = append(lines, line)
		}
		prompt = ""
		for _, line := range lines {
			prompt += line + " "
		}
		prompt = trimSpace(prompt)
	}

	// Build options
	opts := &app.OptiOptions{
		Clipboard:  *clipboardFlag,
		Batch:      *batchFlag,
		ExplainAll: *explainAllFlag,
		MaxTokens:  *maxTokensFlag,
		Scope:      *scopeFlag,
		Templates:  *templatesFlag,
		Prompt:     prompt,
	}

	// Create config
	cfg := &app.AppConfig{
		ProjectRoot: getProjectRoot(),
		Verbose:     *verboseFlag,
		Quiet:       *quietFlag,
	}

	// Create app and run
	groveApp := app.New(cfg)
	return groveApp.RunOptiWithOptions(opts)
}

func getProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Walk up looking for project markers
	for {
		if _, err := os.Stat(cwd + "/AGENTS.md"); err == nil {
			return cwd
		}
		if _, err := os.Stat(cwd + "/.opencode"); err == nil {
			return cwd
		}

		parent := cwd + "/.."
		if parent == cwd {
			break
		}
		cwd = parent
	}

	cwd2, _ := os.Getwd()
	return cwd2
}

func readClipboard() (string, error) {
	// Platform-specific clipboard reading would go here
	// For now, return a placeholder
	return "", fmt.Errorf("clipboard reading not yet implemented for this platform")
}

func trimSpace(s string) string {
	// Simple whitespace trimming
	result := ""
	for _, c := range s {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			result += string(c)
		}
	}
	return result
}

func usage() {
	fmt.Printf(`GROVE Opti Prompt - Prompt Optimizer

Transform natural language requests into precise, project-aware prompts
optimized for OpenCode agents.

Usage:
	grove-opti [flags] [prompt]

Flags:
`)
	flag.PrintDefaults()

	fmt.Println(`
Examples:
	# Optimize a single prompt
	grove-opti "add dark mode toggle to settings"

	# Read prompt from clipboard
	grove-opti --clipboard

	# Batch mode from file
	grove-opti --batch ./prompts.txt

	# Custom token budget
	grove-opti --max-tokens 3000 "implement user authentication"

	# Force full explanations
	grove-opti --explain-all "fix the login bug"

Exit Codes:
	0 - Success
	1 - Error
	2 - User rejected optimized prompt`)
}
