// Package main implements the grove-spec CLI command.
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
		logger.Error("grove-spec failed: %v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("grove-spec", flag.ContinueOnError)
	fs.Usage = usage

	var (
		inputFlag       = fs.String("input", "", "Input folder path containing raw ideas, wireframes, or documentation")
		outputFlag      = fs.String("output", "./spec", "Output directory for generated specifications")
		updateFlag      = fs.Bool("update", false, "Incremental update mode - only process changed components")
		reverseFlag     = fs.Bool("reverse", false, "Reverse documentation mode - analyze existing codebase")
		loopMaxFlag     = fs.Int("loop-max", 10, "Maximum number of spec generation loops")
		resumeFlag      = fs.Bool("resume", false, "Resume from previous loop state")
		fullRescoreFlag = fs.Bool("full-rescore", false, "Force complete re-score of all dimensions")
		verboseFlag     = fs.Bool("v", false, "Enable verbose output")
		quietFlag       = fs.Bool("q", false, "Suppress non-essential output")
		helpFlag        = fs.Bool("h", false, "Show help message")
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

	// Validate input
	if *inputFlag == "" && !*reverseFlag {
		inputDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine working directory: %w", err)
		}
		inputFlag = &inputDir
		logger.Warn("No input specified, using current directory: %s", *inputFlag)
	}

	// Build options
	opts := &app.SpecOptions{
		Input:       *inputFlag,
		Output:      *outputFlag,
		Update:      *updateFlag,
		Reverse:     *reverseFlag,
		LoopMax:     *loopMaxFlag,
		Resume:      *resumeFlag,
		FullRescore: *fullRescoreFlag,
	}

	// Create config
	cfg := &app.AppConfig{
		ProjectRoot: *inputFlag,
		Verbose:     *verboseFlag,
		Quiet:       *quietFlag,
	}

	// Create app and run
	groveApp := app.New(cfg)
	return groveApp.RunSpecWithOptions(opts)
}

func usage() {
	fmt.Printf(`GROVE Spec - Specification Generator

Transform raw, unstructured project input into complete specifications
ready for autonomous AI-driven development.

Usage:
	grove-spec [flags]

Flags:
`)
	flag.PrintDefaults()

	fmt.Println(`
Examples:
	# Basic usage with input folder
	grove-spec --input ./my-ideas

	# Incremental update mode
	grove-spec --input ./my-ideas --update

	# Reverse documentation (analyze existing codebase)
	grove-spec --reverse --input ./existing-project

	# Resume from previous state
	grove-spec --resume

	# Custom output directory
	grove-spec --input ./ideas --output ./documentation

Exit Codes:
	0 - Success
	1 - Error
	2 - Quality threshold not met (warning)
`)
}
