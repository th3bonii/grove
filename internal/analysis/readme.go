// Package analysis provides value-add analysis features for GROVE Spec.
//
// These features add real value to the specification process:
//   - Stack Analysis: Automatic technology stack recommendation
//   - Feasibility: Technical viability assessment
//   - Security: Vulnerability detection
//   - Tests: Test pattern generation (complementary to SDD)
//   - Time: Time estimation
//   - Competitor: Similar project analysis
//   - README: Automatic README generation
//   - GitHub: Issue creation from tasks
package analysis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// =============================================================================
// README Generator
// =============================================================================

// READMEGenerator generates README.md from specifications.
type READMEGenerator struct {
	projectDir string
}

// NewREADMEGenerator creates a new README generator.
func NewREADMEGenerator(projectDir string) *READMEGenerator {
	return &READMEGenerator{projectDir: projectDir}
}

// Generate creates a README.md from specs.
func (g *READMEGenerator) Generate(ctx context.Context) error {
	// Read SPEC.md for project info
	specContent := ""
	if content, err := os.ReadFile(filepath.Join(g.projectDir, "spec", "SPEC.md")); err == nil {
		specContent = string(content)
	}

	// Read DESIGN.md for tech stack
	designContent := ""
	if content, err := os.ReadFile(filepath.Join(g.projectDir, "spec", "DESIGN.md")); err == nil {
		designContent = string(content)
	}

	// Generate README
	readme := g.generateREADME(specContent, designContent)

	// Write README.md
	readmePath := filepath.Join(g.projectDir, "README.md")
	return os.WriteFile(readmePath, []byte(readme), 0644)
}

func (g *READMEGenerator) generateREADME(spec, design string) string {
	projectName := extractProjectName(spec)
	description := extractDescription(spec)

	return fmt.Sprintf(`# %s

%s

## 🚀 Quick Start

### Prerequisites

- Node.js 18+ (or your runtime)
- npm or yarn
- Git

### Installation

`+"```bash"+`
git clone https://github.com/your-username/%s.git
cd %s
npm install
`+"```"+`

### Development

`+"```bash"+`
npm run dev
`+"```"+`

### Production

`+"```bash"+`
npm run build
npm start
`+"```"+`

## 📋 Features

%s

## 🏗️ Architecture

See [DESIGN.md](spec/DESIGN.md) for detailed architecture.

## 📖 Documentation

- [Product Requirements](spec/SPEC.md)
- [Technical Design](spec/DESIGN.md)
- [Implementation Tasks](spec/TASKS.md)
- [User Flows](spec/FLOWS.md)

## 🧪 Testing

`+"```bash"+`
npm test
npm run test:e2e
`+"```"+`

## 📄 License

MIT

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request
`, projectName, description, projectName, projectName, extractFeatures(spec))
}

// =============================================================================
// GitHub Issues Creator
// =============================================================================

// GitHubIssuesCreator creates GitHub issues from TASKS.md.
type GitHubIssuesCreator struct {
	projectDir string
}

// NewGitHubIssuesCreator creates a new issues creator.
func NewGitHubIssuesCreator(projectDir string) *GitHubIssuesCreator {
	return &GitHubIssuesCreator{projectDir: projectDir}
}

// Issue represents a GitHub issue.
type Issue struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []string `json:"labels"`
}

// CreateIssues reads TASKS.md and creates issues.
func (c *GitHubIssuesCreator) CreateIssues(ctx context.Context) ([]Issue, error) {
	tasksPath := filepath.Join(c.projectDir, "spec", "TASKS.md")
	content, err := os.ReadFile(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TASKS.md: %w", err)
	}

	issues := c.parseTasksToIssues(string(content))
	return issues, nil
}

func (c *GitHubIssuesCreator) parseTasksToIssues(content string) []Issue {
	issues := make([]Issue, 0)
	lines := strings.Split(content, "\n")

	currentIssue := Issue{}
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "## ") {
			if currentIssue.Title != "" {
				issues = append(issues, currentIssue)
			}
			currentIssue = Issue{
				Title:  strings.TrimPrefix(line, "## "),
				Labels: []string{"enhancement"},
			}
		}

		if strings.HasPrefix(line, "- **Priority**: ") {
			priority := strings.TrimPrefix(line, "- **Priority**: ")
			currentIssue.Labels = append(currentIssue.Labels, "priority:"+priority)
		}

		if strings.HasPrefix(line, "- **Description**: ") {
			currentIssue.Body = strings.TrimPrefix(line, "- **Description**: ")
		}
	}

	if currentIssue.Title != "" {
		issues = append(issues, currentIssue)
	}

	return issues
}

// GenerateIssuesMarkdown creates a markdown file with issues.
func (c *GitHubIssuesCreator) GenerateIssuesMarkdown(ctx context.Context) error {
	issues, err := c.CreateIssues(ctx)
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("# GitHub Issues\n\n")
	sb.WriteString("Auto-generated from TASKS.md\n\n")

	for i, issue := range issues {
		sb.WriteString(fmt.Sprintf("## Issue %d: %s\n\n", i+1, issue.Title))
		sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", issue.Body))
		sb.WriteString(fmt.Sprintf("**Labels:** %s\n\n", strings.Join(issue.Labels, ", ")))
		sb.WriteString("---\n\n")
	}

	issuesPath := filepath.Join(c.projectDir, "spec", "GITHUB-issues.md")
	return os.WriteFile(issuesPath, []byte(sb.String()), 0644)
}

// =============================================================================
// Wireframe Generator (Text-based)
// =============================================================================

// WireframeGenerator generates text-based wireframes.
type WireframeGenerator struct {
	projectDir string
}

// NewWireframeGenerator creates a new wireframe generator.
func NewWireframeGenerator(projectDir string) *WireframeGenerator {
	return &WireframeGenerator{projectDir: projectDir}
}

// Generate creates wireframe documentation.
func (g *WireframeGenerator) Generate(ctx context.Context) error {
	// Read SPEC.md for component info
	specPath := filepath.Join(g.projectDir, "spec", "SPEC.md")
	content, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read SPEC.md: %w", err)
	}

	wireframes := g.generateWireframes(string(content))

	wireframesPath := filepath.Join(g.projectDir, "spec", "WIREFRAMES.md")
	return os.WriteFile(wireframesPath, []byte(wireframes), 0644)
}

func (g *WireframeGenerator) generateWireframes(spec string) string {
	var sb strings.Builder

	sb.WriteString("# Wireframes\n\n")
	sb.WriteString("Text-based wireframe representations.\n\n")

	// Generate common wireframes
	sb.WriteString("## Desktop Layout\n\n")
	sb.WriteString("```\n")
	sb.WriteString("┌─────────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│ Header                                                      │\n")
	sb.WriteString("├─────────────────────────────────────────────────────────────┤\n")
	sb.WriteString("│ Sidebar │ Main Content Area                                │\n")
	sb.WriteString("│         │                                                  │\n")
	sb.WriteString("│         │                                                  │\n")
	sb.WriteString("│         │                                                  │\n")
	sb.WriteString("├─────────────────────────────────────────────────────────────┤\n")
	sb.WriteString("│ Footer                                                      │\n")
	sb.WriteString("└─────────────────────────────────────────────────────────────┘\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Mobile Layout\n\n")
	sb.WriteString("```\n")
	sb.WriteString("┌─────────────────────┐\n")
	sb.WriteString("│ Header              │\n")
	sb.WriteString("├─────────────────────┤\n")
	sb.WriteString("│ Main Content        │\n")
	sb.WriteString("│                     │\n")
	sb.WriteString("│                     │\n")
	sb.WriteString("├─────────────────────┤\n")
	sb.WriteString("│ Nav │ Home │ Profile│\n")
	sb.WriteString("└─────────────────────┘\n")
	sb.WriteString("```\n\n")

	return sb.String()
}

// =============================================================================
// Helper Functions
// =============================================================================

func extractProjectName(spec string) string {
	lines := strings.Split(spec, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "My Project"
}

func extractDescription(spec string) string {
	lines := strings.Split(spec, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && len(line) > 20 {
			return line
		}
	}
	return "A great application"
}

func extractFeatures(spec string) string {
	var features []string
	lines := strings.Split(spec, "\n")

	inFeatures := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(strings.ToLower(line), "feature") ||
			strings.Contains(strings.ToLower(line), "requisito") {
			inFeatures = true
			continue
		}

		if inFeatures && strings.HasPrefix(line, "- ") {
			features = append(features, line)
		}

		if inFeatures && strings.HasPrefix(line, "#") {
			break
		}
	}

	if len(features) == 0 {
		return "- Feature 1\n- Feature 2\n- Feature 3"
	}

	return strings.Join(features, "\n")
}
