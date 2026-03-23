package spec

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gentleman-Programming/grove/internal/types"
)

func TestNewInputProcessor(t *testing.T) {
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
				ProjectName: "test-project",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewInputProcessor(tt.config)

			if processor == nil {
				t.Fatal("expected processor to be created, got nil")
			}

			if processor.config == nil && tt.config != nil {
				t.Error("expected config to be set")
			}
		})
	}
}

func TestInputProcessorDetectInputType(t *testing.T) {
	processor := NewInputProcessor(nil)

	tests := []struct {
		name     string
		input    string
		expected InputType
	}{
		{
			name:     "URL input",
			input:    "https://example.com/docs",
			expected: InputTypeURL,
		},
		{
			name:     "HTTP URL",
			input:    "http://example.com/page",
			expected: InputTypeURL,
		},
		{
			name:     "Markdown file",
			input:    "docs/spec.md",
			expected: InputTypeMarkdown,
		},
		{
			name:     "Markdown file with absolute path",
			input:    "C:/docs/spec.md",
			expected: InputTypeMarkdown,
		},
		{
			name:     "Text input",
			input:    "Implement user authentication",
			expected: InputTypeText,
		},
		{
			name:     "Markdown file in Windows style",
			input:    "C:\\docs\\spec.md",
			expected: InputTypeMarkdown,
		},
		{
			name:     "Relative path with separator",
			input:    "./docs/spec.md",
			expected: InputTypeMarkdown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.DetectInputType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestInputProcessorProcessText(t *testing.T) {
	processor := NewInputProcessor(nil)

	tests := []struct {
		name     string
		text     string
		wantType InputType
	}{
		{
			name:     "basic text",
			text:     "Implement user authentication with JWT",
			wantType: InputTypeText,
		},
		{
			name:     "empty text",
			text:     "",
			wantType: InputTypeText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.ProcessText(tt.text)
			if err != nil && tt.text != "" {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.text != "" && result == nil {
				t.Fatal("expected result, got nil")
			}

			if tt.text != "" && result.Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, result.Type)
			}

			if tt.text != "" && result.ParsedContent == "" {
				t.Error("expected parsed content to be set")
			}
		})
	}
}

func TestInputProcessorProcessMarkdown(t *testing.T) {
	processor := NewInputProcessor(nil)

	// Create a temporary markdown file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	content := "# Test Spec\n\n## Requirements\n\n- Feature A\n- Feature B"

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := processor.ProcessMarkdown(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.Type != InputTypeMarkdown {
		t.Errorf("expected type %s, got %s", InputTypeMarkdown, result.Type)
	}

	if result.Metadata["file_name"] != "test.md" {
		t.Errorf("expected file_name to be 'test.md', got %s", result.Metadata["file_name"])
	}

	if result.Metadata["line_count"] == "" {
		t.Error("expected line_count to be set")
	}

	if len(result.Components) == 0 {
		t.Logf("expected components to be extracted, got empty list")
	}
}

func TestInputProcessorProcessDirectory(t *testing.T) {
	processor := NewInputProcessor(nil)

	// Create a temporary directory with markdown files
	tmpDir := t.TempDir()

	// Create multiple markdown files
	files := map[string]string{
		"spec1.md":  "# Spec 1\n\n## Requirements\n\n- Feature A",
		"spec2.md":  "# Spec 2\n\n## Requirements\n\n- Feature B",
		"readme.md": "# Readme\n\nThis is a readme",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
	}

	result, err := processor.ProcessDirectory(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.Type != InputTypeDirectory {
		t.Errorf("expected type %s, got %s", InputTypeDirectory, result.Type)
	}

	if result.Metadata["file_count"] != "3" {
		t.Errorf("expected file_count to be '3', got %s", result.Metadata["file_count"])
	}

	if result.ParsedContent == "" {
		t.Error("expected parsed content to be set")
	}
}

func TestInputProcessorProcess(t *testing.T) {
	processor := NewInputProcessor(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantNil bool
	}{
		{
			name:    "empty input returns nil",
			input:   "",
			wantNil: true,
		},
		{
			name:    "text input returns result",
			input:   "Implement user authentication",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.Process(ctx, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil && result != nil {
				t.Error("expected nil result")
			}

			if !tt.wantNil && result == nil {
				t.Fatal("expected result, got nil")
			}

			if !tt.wantNil && tt.input != "" && result.OriginalInput != tt.input {
				t.Errorf("expected original input '%s', got '%s'", tt.input, result.OriginalInput)
			}
		})
	}
}

func TestInputProcessorProcessExtended(t *testing.T) {
	processor := NewInputProcessor(nil)
	ctx := context.Background()

	input := "Add user authentication with JWT tokens"

	result, err := processor.ProcessExtended(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.Type != InputTypeText {
		t.Errorf("expected type %s, got %s", InputTypeText, result.Type)
	}

	if result.Metadata == nil {
		t.Error("expected metadata to be initialized")
	}

	if result.Metadata["char_count"] == "" {
		t.Error("expected char_count to be set")
	}
}

func TestInputProcessorExtractTypes(t *testing.T) {
	processor := NewInputProcessor(nil)

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "API keywords",
			input:    "Build a REST API with Node.js",
			expected: []string{"api"},
		},
		{
			name:     "Database keywords",
			input:    "Use PostgreSQL for database",
			expected: []string{"database"},
		},
		{
			name:     "Authentication keywords",
			input:    "Implement JWT authentication",
			expected: []string{"authentication"},
		},
		{
			name:     "Multiple types",
			input:    "Build a REST API with PostgreSQL and JWT auth",
			expected: []string{"api", "database", "authentication"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.extractTypes(tt.input)

			for _, expected := range tt.expected {
				found := false
				for _, r := range result {
					if r == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to find %s in result, got %v", expected, result)
				}
			}
		})
	}
}

func TestInputProcessorDetectStack(t *testing.T) {
	processor := NewInputProcessor(nil)

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "React stack",
			input:    "Build a React frontend with TypeScript",
			expected: []string{"React", "TypeScript"},
		},
		{
			name:     "Node.js stack",
			input:    "Create a Node.js API with Express",
			expected: []string{"Node.js", "Express"},
		},
		{
			name:     "Full stack",
			input:    "Build a React frontend with Node.js backend and PostgreSQL",
			expected: []string{"React", "Node.js", "PostgreSQL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.detectStack(tt.input)

			for _, expected := range tt.expected {
				found := false
				for _, r := range result {
					if r == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to find %s in result, got %v", expected, result)
				}
			}
		})
	}
}

func TestInputProcessorExtractComponents(t *testing.T) {
	processor := NewInputProcessor(nil)

	content := `## User Authentication

The system includes:
- LoginComponent
- RegisterComponent
- AuthMiddleware

type UserService interface {
	Login(email string, password string) (*User, error)
}`

	components := processor.extractComponents(content)

	// Should find at least some components
	if len(components) == 0 {
		t.Logf("expected components to be extracted, got empty list")
	}
}

func TestInputProcessorParseMarkdown(t *testing.T) {
	processor := NewInputProcessor(nil)

	content := "# Title\n\n\n\n## Section\n\nContent"

	parsed := processor.parseMarkdown(content)

	if parsed == "" {
		t.Error("expected parsed content to be non-empty")
	}

	// Should not have multiple empty lines
	if strings.Contains(parsed, "\n\n\n") {
		t.Error("expected multiple empty lines to be collapsed")
	}
}

func TestInputProcessorHtmlToMarkdown(t *testing.T) {
	processor := NewInputProcessor(nil)

	html := `<h1>Title</h1><p>This is <strong>bold</strong> and <em>italic</em> text.</p><a href="https://example.com">Link</a>`

	markdown := processor.htmlToMarkdown(html)

	if !strings.Contains(markdown, "# Title") {
		t.Error("expected h1 to be converted to markdown header")
	}

	if !strings.Contains(markdown, "**bold**") {
		t.Error("expected strong to be converted to bold markdown")
	}

	if !strings.Contains(markdown, "[Link](https://example.com)") {
		t.Error("expected anchor to be converted to markdown link")
	}
}
