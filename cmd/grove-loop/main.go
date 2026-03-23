// Package main implements the grove-loop CLI command.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Gentleman-Programming/grove/internal/app"
	"github.com/Gentleman-Programming/grove/internal/logger"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		logger.Error("grove-loop failed: %v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("grove-loop", flag.ContinueOnError)
	fs.Usage = usage

	var (
		pauseAfterFlag = fs.String("pause-after", "", "Pause loop after completing specified task ID")
		statusFlag     = fs.Bool("status", false, "Show current loop status and progress")
		reportFlag     = fs.Bool("report", false, "Generate and display loop execution report")
		autoCommitFlag = fs.Bool("auto-commit", false, "Auto-commit changes after each completed task")
		resumeFlag     = fs.Bool("resume", false, "Resume loop from saved state")
		qualityFlag    = fs.Float64("quality", 70.0, "Minimum documentation quality threshold (0-100)")
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

	// Build options
	opts := &app.LoopOptions{
		PauseAfter:       *pauseAfterFlag,
		Status:           *statusFlag,
		Report:           *reportFlag,
		AutoCommit:       *autoCommitFlag,
		Resume:           *resumeFlag,
		QualityThreshold: *qualityFlag,
	}

	// Create config
	cfg := &app.AppConfig{
		ProjectRoot: getProjectRoot(),
		Verbose:     *verboseFlag,
		Quiet:       *quietFlag,
	}

	// Create app and run
	groveApp := app.New(cfg)
	return groveApp.RunLoopWithOptions(opts)
}

func getProjectRoot() string {
	// Try to find project root by looking for AGENTS.md or spec directory
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Check common locations
	paths := []string{
		cwd,
		cwd + "/..",
		cwd + "/../..",
	}

	for _, path := range paths {
		if _, err := os.Stat(path + "/AGENTS.md"); err == nil {
			return path
		}
		if _, err := os.Stat(path + "/spec"); err == nil {
			return path
		}
	}

	return cwd
}

func usage() {
	fmt.Printf(`GROVE Ralph Loop - Autonomous Build Engine

Execute autonomous build loops that transform specifications into
production-ready code through iterative validation and verification.

Usage:
	grove-loop [flags]

Flags:
`)
	flag.PrintDefaults()

	fmt.Println(`
Examples:
	# Start a new build loop
	grove-loop

	# Check current status without starting
	grove-loop --status

	# Pause after specific task
	grove-loop --pause-after task-15

	# Resume from saved state
	grove-loop --resume

	# Generate execution report
	grove-loop --report

	# Resume with auto-commit enabled
	grove-loop --resume --auto-commit

Exit Codes:
	0 - Success (PRODUCTION READY)
	1 - Error
	2 - Paused (check status and resume)
	3 - Quality gate failed (documentation needs improvement)
`)
}
