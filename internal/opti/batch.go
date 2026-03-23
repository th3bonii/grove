package opti

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// BatchResult represents the result of processing a single prompt in batch mode.
type BatchResult struct {
	LineNumber     int
	Original       string
	Optimized      string
	Classification string
	Tokens         int
	Files          []string
	Status         string // "success", "unclassified"
	Error          string
}

// BatchProcessor handles batch processing of multiple prompts from a file.
type BatchProcessor struct {
	optimizer   *Optimizer
	classifier  *Classifier
	collector   *Collector
	explainer   *Explainer
	maxTokens   int
	explainAll  bool
	noTeach     bool
	projectRoot string
}

// NewBatchProcessor creates a new BatchProcessor.
func NewBatchProcessor(projectRoot string, maxTokens int, explainAll, noTeach bool) *BatchProcessor {
	if maxTokens <= 0 {
		maxTokens = 2000
	}

	return &BatchProcessor{
		optimizer:   NewOptimizer(maxTokens),
		classifier:  NewClassifier(),
		collector:   NewCollector(projectRoot),
		explainer:   NewExplainer(projectRoot, explainAll, noTeach),
		maxTokens:   maxTokens,
		explainAll:  explainAll,
		noTeach:     noTeach,
		projectRoot: projectRoot,
	}
}

// ProcessFile processes all prompts from an input file and writes results to a timestamped output file.
func (p *BatchProcessor) ProcessFile(ctx context.Context, inputPath string) (string, error) {
	// Generate output filename with timestamp
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	outputPath := fmt.Sprintf("GROVE-OPTI-BATCH-%s.md", timestamp)

	// Read prompts from file
	prompts, err := readPromptsFromFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompts file: %w", err)
	}

	if len(prompts) == 0 {
		return "", fmt.Errorf("no prompts found in file")
	}

	// Process each prompt
	results := make([]BatchResult, 0, len(prompts))
	for i, prompt := range prompts {
		result := p.processPrompt(ctx, prompt, i+1)
		results = append(results, result)
	}

	// Generate output content
	content := p.generateOutput(results)

	// Write output file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write output file: %w", err)
	}

	return outputPath, nil
}

// processPrompt processes a single prompt and returns the result.
func (p *BatchProcessor) processPrompt(ctx context.Context, prompt string, lineNumber int) BatchResult {
	result := BatchResult{
		LineNumber:     lineNumber,
		Original:       prompt,
		Optimized:      "",
		Classification: "",
		Tokens:         0,
		Files:          []string{},
		Status:         "unclassified",
	}

	// Step 1: Classify intent
	classification := p.classifier.Classify(prompt)

	// Check if classification was successful
	if classification.Intent == IntentOther && classification.Confidence < 0.3 {
		result.Classification = "unclassified"
		result.Status = "unclassified"
		result.Optimized = prompt // Use original if unclassified
		return result
	}

	result.Classification = string(classification.Intent)

	// Step 2: Collect context
	contextResult, err := p.collector.Collect(ctx, classification)
	if err != nil {
		result.Error = fmt.Sprintf("context collection failed: %v", err)
		result.Optimized = prompt
		return result
	}

	// Collect file paths
	for _, file := range contextResult.Files {
		result.Files = append(result.Files, file.Path)
	}

	// Step 3: Optimize prompt
	optimized, err := p.optimizer.Optimize(ctx, prompt, classification, contextResult)
	if err != nil {
		result.Error = fmt.Sprintf("optimization failed: %v", err)
		result.Optimized = prompt
		return result
	}

	result.Optimized = optimized.Optimized
	result.Tokens = optimized.TokenCount
	result.Status = "success"

	return result
}

// generateOutput creates the markdown output content from batch results.
func (p *BatchProcessor) generateOutput(results []BatchResult) string {
	var buf bytes.Buffer

	buf.WriteString("# GROVE Opti Prompt - Batch Results\n\n")
	buf.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("**Total prompts:** %d\n\n", len(results)))

	// Summary section
	buf.WriteString("## Summary\n\n")
	successCount := 0
	unclassifiedCount := 0
	for _, r := range results {
		if r.Status == "success" {
			successCount++
		} else {
			unclassifiedCount++
		}
	}
	buf.WriteString(fmt.Sprintf("- Processed: %d\n", len(results)))
	buf.WriteString(fmt.Sprintf("- Success: %d\n", successCount))
	buf.WriteString(fmt.Sprintf("- Unclassified: %d\n\n", unclassifiedCount))

	// Results section
	buf.WriteString("## Results\n\n")
	for _, r := range results {
		buf.WriteString(fmt.Sprintf("### Prompt #%d\n\n", r.LineNumber))
		buf.WriteString(fmt.Sprintf("**Status:** %s\n\n", r.Status))

		if r.Status == "unclassified" {
			buf.WriteString("⚠️ Could not classify this prompt - using original\n\n")
		}

		buf.WriteString("**Original:**\n```\n")
		buf.WriteString(r.Original)
		buf.WriteString("\n```\n\n")

		buf.WriteString("**Optimized:**\n```\n")
		buf.WriteString(r.Optimized)
		buf.WriteString("\n```\n\n")

		buf.WriteString(fmt.Sprintf("**Classification:** %s\n", r.Classification))
		buf.WriteString(fmt.Sprintf("**Tokens:** %d\n", r.Tokens))

		if len(r.Files) > 0 {
			buf.WriteString("**Files:**\n")
			for _, f := range r.Files {
				buf.WriteString(fmt.Sprintf("  - %s\n", f))
			}
		} else {
			buf.WriteString("**Files:** None\n")
		}

		if r.Error != "" {
			buf.WriteString(fmt.Sprintf("\n**Error:** %s\n", r.Error))
		}

		buf.WriteString("\n---\n\n")
	}

	return buf.String()
}

// readPromptsFromFile reads prompts from a file in the format:
// # Prompt 1
// [optimize this...]
//
// ---
//
// # Prompt 2
// [another prompt...]
//
// Each prompt is separated by "---" on its own line or surrounded by whitespace.
func readPromptsFromFile(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Split by "---" delimiter (with optional whitespace around it)
	// This handles formats like:
	// ---
	// or
	// \n---\n
	// or
	// \n---\n\n
	segments := strings.Split(string(content), "---")

	var prompts []string

	for _, segment := range segments {
		// Clean up each segment
		lines := strings.Split(segment, "\n")
		var cleanedLines []string

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Skip comments (lines starting with #) but keep them in context
			// Actually, per spec, we skip comments entirely as they are just labels
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			if trimmed != "" {
				cleanedLines = append(cleanedLines, line)
			}
		}

		// Join non-empty lines to form the prompt
		prompt := strings.TrimSpace(strings.Join(cleanedLines, "\n"))
		if prompt != "" {
			prompts = append(prompts, prompt)
		}
	}

	return prompts, nil
}

// ReadPromptsFromFile is exported for testing purposes
func ReadPromptsFromFile(path string) ([]string, error) {
	return readPromptsFromFile(path)
}
