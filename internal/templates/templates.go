// Package templates provides project templates for various technology stacks.
// It integrates with gentle-ai skills and supports SDD (Spec-Driven Development).
//
// Templates include:
//   - AGENTS.md with relevant skills
//   - SKILL.md patterns for the stack
//   - Recommended folder structure
//   - Naming conventions
//   - Testing patterns
package templates

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed skills/*
var skillsFS embed.FS

// Stack represents a technology stack identifier.
type Stack string

const (
	StackReact         Stack = "react"
	StackVue           Stack = "vue"
	StackAngular       Stack = "angular"
	StackNextJS        Stack = "nextjs"
	StackGoChi         Stack = "go-chi"
	StackGoEcho        Stack = "go-echo"
	StackGoGin         Stack = "go-gin"
	StackPythonFastAPI Stack = "python-fastapi"
	StackPythonDjango  Stack = "python-django"
	StackNodeExpress   Stack = "node-express"
)

// Template represents a complete project template.
type Template struct {
	Name        string            `json:"name" yaml:"name"`
	Stack       Stack             `json:"stack" yaml:"stack"`
	Description string            `json:"description" yaml:"description"`
	Skills      []TemplateSkill   `json:"skills" yaml:"skills"`
	Folders     []FolderConfig    `json:"folders" yaml:"folders"`
	Files       []FileConfig      `json:"files" yaml:"files"`
	Conventions NamingConventions `json:"conventions" yaml:"conventions"`
	Testing     TestingPatterns   `json:"testing" yaml:"testing"`
}

// TemplateSkill represents a skill reference for the template.
type TemplateSkill struct {
	Name        string   `json:"name" yaml:"name"`
	Path        string   `json:"path" yaml:"path"`
	Required    bool     `json:"required" yaml:"required"`
	Description string   `json:"description" yaml:"description"`
	Aliases     []string `json:"aliases,omitempty" yaml:"aliases,omitempty"`
}

// FolderConfig defines a folder and its purpose.
type FolderConfig struct {
	Path        string         `json:"path" yaml:"path"`
	Purpose     string         `json:"purpose" yaml:"purpose"`
	Permissions []string       `json:"permissions,omitempty" yaml:"permissions,omitempty"`
	Children    []FolderConfig `json:"children,omitempty" yaml:"children,omitempty"`
}

// FileConfig defines a file to be generated.
type FileConfig struct {
	Path         string `json:"path" yaml:"path"`
	Content      string `json:"content,omitempty" yaml:"content,omitempty"`
	TemplateName string `json:"templateName,omitempty" yaml:"templateName,omitempty"`
	SkipIfExists bool   `json:"skipIfExists" yaml:"skipIfExists"`
}

// NamingConventions defines naming rules for the stack.
type NamingConventions struct {
	Files      NamingRule `json:"files" yaml:"files"`
	Components NamingRule `json:"components" yaml:"components"`
	Functions  NamingRule `json:"functions" yaml:"functions"`
	Types      NamingRule `json:"types" yaml:"types"`
	Constants  NamingRule `json:"constants" yaml:"constants"`
	Tests      NamingRule `json:"tests" yaml:"tests"`
}

// NamingRule defines a specific naming convention.
type NamingRule struct {
	Pattern     string `json:"pattern" yaml:"pattern"`
	Example     string `json:"example" yaml:"example"`
	Description string `json:"description" yaml:"description"`
}

// TestingPatterns defines testing conventions for the stack.
type TestingPatterns struct {
	Framework    string   `json:"framework" yaml:"framework"`
	Location     string   `json:"location" yaml:"location"`
	Naming       string   `json:"naming" yaml:"naming"`
	SetupFile    string   `json:"setupFile,omitempty" yaml:"setupFile,omitempty"`
	Utilities    []string `json:"utilities,omitempty" yaml:"utilities,omitempty"`
	CoverageTool string   `json:"coverageTool,omitempty" yaml:"coverageTool,omitempty"`
}

// TemplateRegistry holds all available templates.
type TemplateRegistry map[Stack]*Template

// GetTemplate returns the template for a given stack identifier.
// Returns nil if the stack is not found.
func GetTemplate(stack string) *Template {
	registry := getDefaultRegistry()
	normalized := normalizeStack(stack)
	return registry[Stack(normalized)]
}

// ApplyTemplate applies a template to a project directory.
// Creates all folders and files defined in the template.
// Does not overwrite existing files unless explicitly configured.
func ApplyTemplate(projectDir string, template *Template) error {
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	// Create base project structure
	if err := createFolders(projectDir, template.Folders); err != nil {
		return fmt.Errorf("failed to create folders: %w", err)
	}

	// Generate files from templates
	if err := generateFiles(projectDir, template); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	return nil
}

// DetectStack attempts to detect the technology stack from a project idea or description.
// Uses keyword matching to identify the most likely stack.
func DetectStack(idea string) string {
	lowerIdea := strings.ToLower(idea)

	// Priority order for detection
	keywords := map[Stack][]string{
		StackNextJS:        {"next", "next.js", "nextjs", "ssr", "isr", "app router", "pages router"},
		StackReact:         {"react", "jsx", "tsx", "create react app", "vite react"},
		StackVue:           {"vue", "vue.js", "vuejs", "nuxt", "nuxt.js"},
		StackAngular:       {"angular", "angularjs", "@angular", "standalone components"},
		StackGoGin:         {"gin", "golang gin", "go-gin"},
		StackGoEcho:        {"echo", "golang echo", "go-echo", "labstack echo"},
		StackGoChi:         {"chi", "golang chi", "go-chi", "go router chi"},
		StackPythonFastAPI: {"fastapi", "fast api", "python fastapi"},
		StackPythonDjango:  {"django", "python django", "django orm"},
		StackNodeExpress:   {"express", "node express", "node.js express", "expressjs"},
	}

	bestMatch := ""
	highestScore := 0

	for stack, words := range keywords {
		score := 0
		for _, word := range words {
			// For short keywords (2-3 chars), require word boundary
			// For longer keywords, substring match is fine
			if len(word) <= 3 {
				// Check if word is surrounded by word boundaries
				if containsWord(lowerIdea, word) {
					score++
				}
			} else if strings.Contains(lowerIdea, word) {
				score++
			}
		}
		if score > highestScore {
			highestScore = score
			bestMatch = string(stack)
		}
	}

	// If no match found, default to Node/Express as most common
	if bestMatch == "" {
		bestMatch = string(StackNodeExpress)
	}

	return bestMatch
}

// containsWord checks if a word exists as a standalone word (with word boundaries).
func containsWord(s, word string) bool {
	for i := 0; i <= len(s)-len(word); i++ {
		// Check if word starts at position i
		if s[i:i+len(word)] == word {
			// Check word boundary before
			if i > 0 && isAlphanumeric(s[i-1]) {
				continue
			}
			// Check word boundary after
			if i+len(word) < len(s) && isAlphanumeric(s[i+len(word)]) {
				continue
			}
			return true
		}
	}
	return false
}

// isAlphanumeric checks if a character is alphanumeric.
func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// GetAllStacks returns a list of all available stack identifiers.
func GetAllStacks() []Stack {
	return []Stack{
		StackReact,
		StackVue,
		StackAngular,
		StackNextJS,
		StackGoChi,
		StackGoEcho,
		StackGoGin,
		StackPythonFastAPI,
		StackPythonDjango,
		StackNodeExpress,
	}
}

// normalizeStack normalizes a stack string to match registry keys.
func normalizeStack(stack string) string {
	normalized := strings.ToLower(strings.TrimSpace(stack))

	// Handle common aliases
	aliases := map[string]string{
		"react":          string(StackReact),
		"react+ts":       string(StackReact),
		"react-ts":       string(StackReact),
		"vue":            string(StackVue),
		"vue+ts":         string(StackVue),
		"vue-ts":         string(StackVue),
		"angular":        string(StackAngular),
		"ng":             string(StackAngular),
		"next":           string(StackNextJS),
		"next.js":        string(StackNextJS),
		"nextjs":         string(StackNextJS),
		"go":             string(StackGoGin),
		"golang":         string(StackGoGin),
		"chi":            string(StackGoChi),
		"go-chi":         string(StackGoChi),
		"echo":           string(StackGoEcho),
		"go-echo":        string(StackGoEcho),
		"gin":            string(StackGoGin),
		"go-gin":         string(StackGoGin),
		"python":         string(StackPythonFastAPI),
		"fastapi":        string(StackPythonFastAPI),
		"python-fastapi": string(StackPythonFastAPI),
		"django":         string(StackPythonDjango),
		"python-django":  string(StackPythonDjango),
		"node":           string(StackNodeExpress),
		"nodejs":         string(StackNodeExpress),
		"node.js":        string(StackNodeExpress),
		"express":        string(StackNodeExpress),
		"expressjs":      string(StackNodeExpress),
	}

	if mapped, ok := aliases[normalized]; ok {
		return mapped
	}

	return normalized
}

// createFolders creates the folder structure for a template.
func createFolders(baseDir string, folders []FolderConfig) error {
	for _, folder := range folders {
		fullPath := filepath.Join(baseDir, folder.Path)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create folder %s: %w", folder.Path, err)
		}

		// Create children recursively
		if len(folder.Children) > 0 {
			if err := createFolders(fullPath, folder.Children); err != nil {
				return err
			}
		}
	}
	return nil
}

// generateFiles creates files from template configuration.
func generateFiles(baseDir string, template *Template) error {
	for _, file := range template.Files {
		fullPath := filepath.Join(baseDir, file.Path)

		// Skip if exists and configured to skip
		if file.SkipIfExists {
			if _, err := os.Stat(fullPath); err == nil {
				continue
			}
		}

		content := file.Content
		if content == "" && file.TemplateName != "" {
			// Load template content if not inline
			var err error
			content, err = loadTemplate(file.TemplateName, template.Stack)
			if err != nil {
				return fmt.Errorf("failed to load template %s: %w", file.TemplateName, err)
			}
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
	}
	return nil
}

// loadTemplate loads a template file from the embedded filesystem.
func loadTemplate(name string, stack Stack) (string, error) {
	path := fmt.Sprintf("skills/%s/%s.md", stack, name)
	content, err := skillsFS.ReadFile(path)
	if err != nil {
		// Return empty string if template not found
		return "", nil
	}
	return string(content), nil
}

// getDefaultRegistry returns the default template registry.
func getDefaultRegistry() TemplateRegistry {
	return TemplateRegistry{
		StackReact:         reactTemplate(),
		StackVue:           vueTemplate(),
		StackAngular:       angularTemplate(),
		StackNextJS:        nextJSTemplate(),
		StackGoChi:         goChiTemplate(),
		StackGoEcho:        goEchoTemplate(),
		StackGoGin:         goGinTemplate(),
		StackPythonFastAPI: pythonFastAPITemplate(),
		StackPythonDjango:  pythonDjangoTemplate(),
		StackNodeExpress:   nodeExpressTemplate(),
	}
}

// toTitleCase converts a string to Title Case.
func toTitleCase(s string) string {
	return cases.Title(language.English, cases.NoLower).String(s)
}
