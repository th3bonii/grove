package spec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Gentleman-Programming/grove/internal/sdd"
	"github.com/Gentleman-Programming/grove/internal/types"
)

// InputType represents the type of input being processed.
type InputType string

const (
	// InputTypeMarkdown represents a markdown file.
	InputTypeMarkdown InputType = "markdown"
	// InputTypeDirectory represents a directory containing markdown files.
	InputTypeDirectory InputType = "directory"
	// InputTypeURL represents a URL to fetch content from.
	InputTypeURL InputType = "url"
	// InputTypeText represents plain text input.
	InputTypeText InputType = "text"
	// InputTypeWireframe represents a wireframe input (future).
	InputTypeWireframe InputType = "wireframe"
)

// ProcessedInput represents the result of processing an input.
type ProcessedInput struct {
	Type          InputType
	RawContent    string
	ParsedContent string
	Metadata      map[string]string
	Components    []string
	Source        string // File path, URL, or inline
}

// InputProcessor processes raw input and extracts structured information.
type InputProcessor struct {
	config *types.Config

	// New fields for extended processing
	llmClient *sdd.LLMClient
	opts      InputProcessorOpts
}

// InputProcessorOpts holds options for the InputProcessor.
type InputProcessorOpts struct {
	EnableLLMExtraction bool
	CacheDir            string
	CacheTTL            time.Duration
	HTTPTimeout         time.Duration
}

// DefaultInputProcessorOpts returns default options.
func DefaultInputProcessorOpts() InputProcessorOpts {
	return InputProcessorOpts{
		EnableLLMExtraction: true,
		CacheDir:            ".grove/cache",
		CacheTTL:            60 * time.Minute,
		HTTPTimeout:         30 * time.Second,
	}
}

// NewInputProcessor creates a new InputProcessor instance.
// This maintains backward compatibility with the existing constructor.
func NewInputProcessor(config *types.Config) *InputProcessor {
	if config == nil {
		config = &types.Config{}
	}

	// Try to initialize LLM client
	var llmClient *sdd.LLMClient
	if llm, err := sdd.NewLLMClient(); err == nil {
		llmClient = llm
	}

	return &InputProcessor{
		config:    config,
		llmClient: llmClient,
		opts:      DefaultInputProcessorOpts(),
	}
}

// NewInputProcessorWithLLM creates a new InputProcessor with a custom LLM client.
func NewInputProcessorWithLLM(config *types.Config, llmClient *sdd.LLMClient) *InputProcessor {
	if config == nil {
		config = &types.Config{}
	}

	return &InputProcessor{
		config:    config,
		llmClient: llmClient,
		opts:      DefaultInputProcessorOpts(),
	}
}

// Process processes raw input and returns a ProcessedInput struct.
// This maintains backward compatibility with the existing method.
func (p *InputProcessor) Process(ctx context.Context, input string) (*types.ProcessedInput, error) {
	if input == "" {
		return nil, nil
	}

	// Use the new ProcessExtended method for full processing
	processed, err := p.ProcessExtended(ctx, input)
	if err != nil {
		// Fall back to basic processing
		return p.processBasic(input), nil
	}

	// Convert to types.ProcessedInput for backward compatibility
	return &types.ProcessedInput{
		OriginalInput:  input,
		ParsedContent:  processed.ParsedContent,
		Metadata:       processed.Metadata,
		ExtractedTypes: p.extractTypes(input),
		DetectedStack:  p.detectStack(input),
	}, nil
}

// ProcessExtended processes the input with full feature support.
// It automatically detects the input type and processes accordingly.
func (p *InputProcessor) ProcessExtended(ctx context.Context, input string) (*ProcessedInput, error) {
	if input == "" {
		return nil, errors.New("input cannot be empty")
	}

	inputType := p.DetectInputType(input)

	switch inputType {
	case InputTypeMarkdown:
		return p.ProcessMarkdown(input)
	case InputTypeDirectory:
		return p.ProcessDirectory(input)
	case InputTypeURL:
		return p.ProcessURL(ctx, input)
	case InputTypeText:
		return p.ProcessText(input)
	case InputTypeWireframe:
		return p.ProcessWireframe(input)
	default:
		return p.ProcessText(input)
	}
}

// processBasic provides basic processing for backward compatibility.
func (p *InputProcessor) processBasic(input string) *types.ProcessedInput {
	return &types.ProcessedInput{
		OriginalInput:  input,
		ParsedContent:  p.parseContent(input),
		Metadata:       make(map[string]string),
		ExtractedTypes: p.extractTypes(input),
		DetectedStack:  p.detectStack(input),
	}
}

// DetectInputType detects the type of input based on the input string.
func (p *InputProcessor) DetectInputType(input string) InputType {
	// Check if it's a URL
	if isURL(input) {
		return InputTypeURL
	}

	// Check if it's a file path
	if isFilePath(input) {
		// Check if it's a markdown file
		if isMarkdownFile(input) {
			// Check if it's a directory
			if isDirectory(input) {
				return InputTypeDirectory
			}
			return InputTypeMarkdown
		}

		// Check if it's a directory (even if not .md)
		if isDirectory(input) {
			return InputTypeDirectory
		}
	}

	// Check for wireframe patterns (future detection)
	if isWireframePattern(input) {
		return InputTypeWireframe
	}

	// Default to text
	return InputTypeText
}

// ProcessMarkdown reads and processes a markdown file.
func (p *InputProcessor) ProcessMarkdown(path string) (*ProcessedInput, error) {
	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read markdown file: %w", err)
	}

	rawContent := string(content)

	// Parse markdown content (basic parsing - could be extended)
	parsedContent := p.parseMarkdown(rawContent)

	// Extract metadata
	metadata := p.extractMetadata(rawContent)
	metadata["file_name"] = filepath.Base(path)
	metadata["file_size"] = fmt.Sprintf("%d", len(rawContent))
	metadata["line_count"] = fmt.Sprintf("%d", strings.Count(rawContent, "\n"))

	// Extract components using LLM if available
	components := p.extractComponents(rawContent)

	// Also run backward-compatible extraction
	extractedTypes := p.extractTypes(rawContent)
	for _, t := range extractedTypes {
		metadata["extracted_type"] = t
	}

	return &ProcessedInput{
		Type:          InputTypeMarkdown,
		RawContent:    rawContent,
		ParsedContent: parsedContent,
		Metadata:      metadata,
		Components:    components,
		Source:        path,
	}, nil
}

// ProcessDirectory processes a directory by scanning for markdown files.
func (p *InputProcessor) ProcessDirectory(path string) (*ProcessedInput, error) {
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}

	// Scan for markdown files
	mdFiles, err := scanMarkdownFiles(path)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(mdFiles) == 0 {
		return nil, errors.New("no markdown files found in directory")
	}

	// Process each markdown file and combine content
	var combinedContent strings.Builder
	var allComponents []string

	for _, mdFile := range mdFiles {
		content, err := os.ReadFile(mdFile)
		if err != nil {
			continue // Skip files that can't be read
		}

		combinedContent.WriteString(fmt.Sprintf("\n# %s\n\n", filepath.Base(mdFile)))
		combinedContent.Write(content)
		combinedContent.WriteString("\n\n---\n\n")

		// Extract components from each file
		components := p.extractComponents(string(content))
		allComponents = append(allComponents, components...)
	}

	rawContent := combinedContent.String()
	parsedContent := p.parseMarkdown(rawContent)

	// Extract metadata
	metadata := make(map[string]string)
	metadata["directory"] = path
	metadata["file_count"] = fmt.Sprintf("%d", len(mdFiles))
	metadata["files"] = strings.Join(mdFiles, ", ")

	return &ProcessedInput{
		Type:          InputTypeDirectory,
		RawContent:    rawContent,
		ParsedContent: parsedContent,
		Metadata:      metadata,
		Components:    allComponents,
		Source:        path,
	}, nil
}

// ProcessURL fetches and processes content from a URL.
func (p *InputProcessor) ProcessURL(ctx context.Context, url string) (*ProcessedInput, error) {
	// Validate URL
	if !isURL(url) {
		return nil, fmt.Errorf("invalid URL: %s", url)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: p.opts.HTTPTimeout,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("User-Agent", "GROVE-Spec/1.0")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: status %d", resp.StatusCode)
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	rawContent := string(body)

	// Convert HTML to markdown (basic conversion)
	parsedContent := p.htmlToMarkdown(rawContent)

	// Extract metadata
	metadata := make(map[string]string)
	metadata["url"] = url
	metadata["status_code"] = fmt.Sprintf("%d", resp.StatusCode)
	metadata["content_length"] = fmt.Sprintf("%d", len(rawContent))
	metadata["content_type"] = resp.Header.Get("Content-Type")

	// Extract components
	components := p.extractComponents(parsedContent)

	return &ProcessedInput{
		Type:          InputTypeURL,
		RawContent:    rawContent,
		ParsedContent: parsedContent,
		Metadata:      metadata,
		Components:    components,
		Source:        url,
	}, nil
}

// ProcessText processes plain text input.
func (p *InputProcessor) ProcessText(text string) (*ProcessedInput, error) {
	if text == "" {
		return nil, errors.New("text cannot be empty")
	}

	rawContent := text
	parsedContent := p.parseMarkdown(rawContent)

	// Extract metadata
	metadata := make(map[string]string)
	metadata["char_count"] = fmt.Sprintf("%d", len(rawContent))
	metadata["word_count"] = fmt.Sprintf("%d", countWordsInText(rawContent))

	// Extract components
	components := p.extractComponents(rawContent)

	return &ProcessedInput{
		Type:          InputTypeText,
		RawContent:    rawContent,
		ParsedContent: parsedContent,
		Metadata:      metadata,
		Components:    components,
		Source:        "inline",
	}, nil
}

// ProcessWireframe processes a wireframe input (placeholder for future).
func (p *InputProcessor) ProcessWireframe(input string) (*ProcessedInput, error) {
	// Placeholder for future wireframe processing
	metadata := make(map[string]string)
	metadata["wireframe_type"] = "placeholder"
	metadata["note"] = "Wireframe processing not yet implemented"

	return &ProcessedInput{
		Type:          InputTypeWireframe,
		RawContent:    input,
		ParsedContent: input,
		Metadata:      metadata,
		Components:    []string{},
		Source:        "wireframe",
	}, nil
}

// extractComponents extracts component names from content using LLM.
// If LLM is not available, falls back to heuristic extraction.
func (p *InputProcessor) extractComponents(content string) []string {
	// If LLM is available, use it for extraction
	if p.llmClient != nil && p.opts.EnableLLMExtraction {
		return p.extractComponentsWithLLM(content)
	}

	// Fallback: heuristic extraction
	return p.extractComponentsHeuristic(content)
}

// extractComponentsWithLLM uses the LLM to extract components.
func (p *InputProcessor) extractComponentsWithLLM(content string) []string {
	prompt := fmt.Sprintf(`Extract all component names from the following specification content. 
Return only a JSON array of component names, nothing else.

Content:
%s

Respond with only the JSON array.`, content)

	resp, err := p.llmClient.Send(context.Background(), prompt)
	if err != nil {
		// Fall back to heuristic
		return p.extractComponentsHeuristic(content)
	}

	// Parse JSON array from response
	var components []string
	start := strings.Index(resp, "[")
	end := strings.LastIndex(resp, "]")

	if start != -1 && end != -1 && end > start {
		jsonStr := resp[start : end+1]
		// Simple JSON parsing (without json package to avoid issues)
		re := regexp.MustCompile(`"([^"]+)"`)
		matches := re.FindAllStringSubmatch(jsonStr, -1)
		for _, match := range matches {
			if len(match) > 1 {
				components = append(components, match[1])
			}
		}
	}

	return components
}

// extractComponentsHeuristic extracts components using simple heuristics.
func (p *InputProcessor) extractComponentsHeuristic(content string) []string {
	var components []string
	seen := make(map[string]bool)

	// Patterns to find component names
	patterns := []string{
		// Headers (## Component Name)
		`##\s+([A-Z][a-zA-Z0-9\s]+)`,
		// Component definitions
		`[Cc]omponent:\s*([A-Z][a-zA-Z0-9]+)`,
		// Interface definitions
		`type\s+([A-Z][a-zA-Z0-9]+)\s+(?:interface|struct)`,
		// Class definitions
		`class\s+([A-Z][a-zA-Z0-9]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				name := strings.TrimSpace(match[1])
				if name != "" && !seen[name] {
					seen[name] = true
					components = append(components, name)
				}
			}
		}
	}

	return components
}

// parseMarkdown performs basic markdown parsing.
func (p *InputProcessor) parseMarkdown(content string) string {
	// Basic markdown parsing - remove excessive whitespace
	// and normalize headers

	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		// Remove trailing whitespace
		line = strings.TrimRight(line, " \t")

		// Skip empty lines that are repeated
		if len(result) > 0 && line == "" && result[len(result)-1] == "" {
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// extractMetadata extracts metadata from markdown content.
func (p *InputProcessor) extractMetadata(content string) map[string]string {
	metadata := make(map[string]string)

	// Extract title from first heading
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			metadata["title"] = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Extract frontmatter if present
	if strings.HasPrefix(content, "---") {
		metadata["has_frontmatter"] = "true"
	}

	return metadata
}

// htmlToMarkdown performs basic HTML to Markdown conversion.
func (p *InputProcessor) htmlToMarkdown(html string) string {
	content := html

	// Convert headings
	content = regexp.MustCompile(`<h1[^>]*>(.*?)</h1>`).ReplaceAllString(content, "# $1")
	content = regexp.MustCompile(`<h2[^>]*>(.*?)</h2>`).ReplaceAllString(content, "## $1")
	content = regexp.MustCompile(`<h3[^>]*>(.*?)</h3>`).ReplaceAllString(content, "### $1")
	content = regexp.MustCompile(`<h4[^>]*>(.*?)</h4>`).ReplaceAllString(content, "#### $1")

	// Convert paragraph
	content = regexp.MustCompile(`<p[^>]*>(.*?)</p>`).ReplaceAllString(content, "$1\n")

	// Convert links
	content = regexp.MustCompile(`<a[^>]*href=["']([^"']*)["'][^>]*>(.*?)</a>`).ReplaceAllString(content, "[$2]($1)")

	// Convert bold
	content = regexp.MustCompile(`<strong[^>]*>(.*?)</strong>`).ReplaceAllString(content, "**$1**")
	content = regexp.MustCompile(`<b[^>]*>(.*?)</b>`).ReplaceAllString(content, "**$1**")

	// Convert italic
	content = regexp.MustCompile(`<em[^>]*>(.*?)</em>`).ReplaceAllString(content, "*$1*")
	content = regexp.MustCompile(`<i[^>]*>(.*?)</i>`).ReplaceAllString(content, "*$1*")

	// Convert code blocks
	content = regexp.MustCompile(`<pre[^>]*><code[^>]*>(.*?)</code></pre>`).ReplaceAllString(content, "```\n$1\n```")
	content = regexp.MustCompile(`<code[^>]*>(.*?)</code>`).ReplaceAllString(content, "`$1`")

	// Convert lists
	content = regexp.MustCompile(`<li[^>]*>(.*?)</li>`).ReplaceAllString(content, "- $1")
	content = regexp.MustCompile(`<ul[^>]*>|</ul>`).ReplaceAllString(content, "")
	content = regexp.MustCompile(`<ol[^>]*>|</ol>`).ReplaceAllString(content, "")

	// Remove remaining tags
	content = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(content, "")

	// Decode HTML entities
	content = decodeHTMLEntities(content)

	return content
}

// parseContent normalizes and cleans the input content (backward compatibility).
func (p *InputProcessor) parseContent(input string) string {
	// Normalize whitespace
	re := regexp.MustCompile(`\s+`)
	content := re.ReplaceAllString(input, " ")

	// Trim leading/trailing whitespace
	content = strings.TrimSpace(content)

	return content
}

// extractTypes extracts the type of application/feature from input.
func (p *InputProcessor) extractTypes(input string) []string {
	var types []string
	lowerInput := strings.ToLower(input)

	typeIndicators := map[string]string{
		"api":       "api",
		"rest":      "api",
		"graphql":   "api",
		"frontend":  "frontend",
		"ui":        "ui",
		"dashboard": "ui",
		"admin":     "ui",
		"backend":   "backend",
		"database":  "database",
		"db":        "database",
		"postgres":  "database",
		"mysql":     "database",
		"mongodb":   "database",
		"auth":      "authentication",
		"login":     "authentication",
		"register":  "authentication",
		"jwt":       "authentication",
		"oauth":     "authentication",
		"webhook":   "integration",
		"payment":   "payment",
		"stripe":    "payment",
		"email":     "notification",
		"sms":       "notification",
		"push":      "notification",
		"websocket": "realtime",
		"realtime":  "realtime",
	}

	for keyword, featureType := range typeIndicators {
		if strings.Contains(lowerInput, keyword) {
			if !contains(types, featureType) {
				types = append(types, featureType)
			}
		}
	}

	return types
}

// detectStack detects the technology stack from input.
func (p *InputProcessor) detectStack(input string) []string {
	var stack []string
	lowerInput := strings.ToLower(input)

	stackIndicators := map[string][]string{
		"react":      {"React"},
		"vue":        {"Vue.js"},
		"angular":    {"Angular"},
		"svelte":     {"Svelte"},
		"next.js":    {"Next.js"},
		"nextjs":     {"Next.js"},
		"nuxt":       {"Nuxt.js"},
		"node":       {"Node.js"},
		"express":    {"Express"},
		"go":         {"Go"},
		"golang":     {"Go"},
		"python":     {"Python"},
		"django":     {"Django"},
		"flask":      {"Flask"},
		"fastapi":    {"FastAPI"},
		"java":       {"Java"},
		"spring":     {"Spring"},
		"rust":       {"Rust"},
		"typescript": {"TypeScript"},
		"javascript": {"JavaScript"},
		"postgres":   {"PostgreSQL"},
		"postgresql": {"PostgreSQL"},
		"mysql":      {"MySQL"},
		"mongodb":    {"MongoDB"},
		"redis":      {"Redis"},
		"docker":     {"Docker"},
		"kubernetes": {"Kubernetes"},
		"aws":        {"AWS"},
		"gcp":        {"GCP"},
		"azure":      {"Azure"},
	}

	for keyword, tech := range stackIndicators {
		if strings.Contains(lowerInput, keyword) {
			for _, t := range tech {
				if !contains(stack, t) {
					stack = append(stack, t)
				}
			}
		}
	}

	return stack
}

// contains checks if a slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper functions

func isURL(input string) bool {
	pattern := `^https?://[^\s]+$`
	matched, _ := regexp.MatchString(pattern, input)
	return matched
}

func isFilePath(input string) bool {
	// Check for absolute or relative path patterns
	if filepath.IsAbs(input) {
		return true
	}

	// Check for relative path with directory separators
	if strings.Contains(input, string(filepath.Separator)) || strings.Contains(input, "/") {
		return true
	}

	return false
}

func isMarkdownFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isWireframePattern(input string) bool {
	// Simple pattern detection for wireframes (future expansion)
	// For now, always return false
	_ = input
	return false
}

func scanMarkdownFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if it's a markdown file
		if isMarkdownFile(path) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func countWordsInText(text string) int {
	words := strings.Fields(text)
	return len(words)
}

func decodeHTMLEntities(text string) string {
	// Simple HTML entity decoding
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&apos;", "'")
	return text
}
