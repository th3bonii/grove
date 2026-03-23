// Package main is the entry point for the GROVE CLI.
package main

import (
	"fmt"
	"os"

	"github.com/Gentleman-Programming/grove/internal/app"
	"github.com/Gentleman-Programming/grove/internal/logger"
)

const (
	version = "0.1.0"
	name    = "grove"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse CLI arguments
	args := os.Args[1:]

	if len(args) == 0 {
		return printHelp()
	}

	// Handle global flags
	cfg := &app.AppConfig{
		Verbose: false,
		Quiet:   false,
	}

	// Filter out global flags
	filteredArgs := make([]string, 0, len(args))
	for _, arg := range args {
		switch arg {
		case "-v", "--verbose":
			cfg.Verbose = true
			logger.SetVerbose(true)
		case "-q", "--quiet":
			cfg.Quiet = true
			logger.SetQuiet(true)
		case "-h", "--help":
			return printHelp()
		case "-V", "--version":
			fmt.Printf("%s version %s\n", name, version)
			return nil
		default:
			filteredArgs = append(filteredArgs, arg)
			break
		}
	}

	// Get subcommand
	if len(filteredArgs) == 0 {
		return printHelp()
	}

	subcommand := filteredArgs[0]
	subArgs := filteredArgs[1:]

	// Create app instance
	groveApp := app.New(cfg)

	// Route to appropriate command
	switch subcommand {
	case "spec":
		return groveApp.RunSpec(subArgs)
	case "loop":
		return groveApp.RunLoop(subArgs)
	case "opti", "prompt":
		return groveApp.RunOpti(subArgs)
	default:
		return fmt.Errorf("unknown command: %s\nRun '%s --help' for usage", subcommand, name)
	}
}

func printHelp() error {
	help := `GROVE - Gentleman's Robust Orchestration & Verification Engine

Usage:
	grove [global flags] <command> [command flags]

GROVE is a suite of tools for specification-driven development, autonomous 
code generation, and prompt optimization.

Commands:
	spec, s       Transform raw ideas into complete specifications
	loop, l       Autonomous build loop for code implementation
	opti, prompt  Optimize natural language prompts for OpenCode

Global Flags:
	-v, --verbose    Enable verbose output
	-q, --quiet      Suppress non-essential output
	-h, --help       Show this help message
	-V, --version    Show version number

Examples:
	grove spec --input ./ideas
	grove loop --status
	grove opti "add dark mode to settings"

For more information on a specific command:
	grove <command> --help
`
	fmt.Print(help)
	return nil
}
