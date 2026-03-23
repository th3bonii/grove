// Package spec provides the GROVE Spec engine for transforming raw ideas
// into complete, production-ready specifications.
//
// This file implements the AGENTS.md and SKILLS.md generation functionality.
package spec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// =============================================================================
// AGENTS.md Generation
// =============================================================================

// generateAgentsAndSkills is the main entry point for generating AGENTS.md and SKILLS.md.
// It scans for existing configuration in ~/.config/opencode/ and integrates rather
// than overwrites.
func (e *Engine) generateAgentsAndSkills() error {
	fmt.Println("  🤖 Generating AGENTS.md and SKILLS.md...")

	// Load existing skills from ~/.config/opencode/skills/
	existingSkills, err := e.loadExistingSkills()
	if err != nil {
		fmt.Printf("  ⚠ Warning: Could not load existing skills: %v\n", err)
	}

	// Detect tech stack for SKILLS.md generation
	detectedTechs := e.detectTechStack()

	// Generate root AGENTS.md
	if err := e.generateRootAgentsMD(existingSkills); err != nil {
		return fmt.Errorf("failed to generate root AGENTS.md: %w", err)
	}

	// Generate module-specific AGENTS.md
	if err := e.generateModuleAgentsMD(existingSkills); err != nil {
		return fmt.Errorf("failed to generate module AGENTS.md: %w", err)
	}

	// Generate SKILLS.md for detected technologies
	if err := e.generateSkillsMD(detectedTechs, existingSkills); err != nil {
		return fmt.Errorf("failed to generate SKILLS.md: %w", err)
	}

	fmt.Printf("  ✓ Generated AGENTS.md and %d SKILLS.md\n", len(detectedTechs))
	return nil
}

// loadExistingSkills loads skills from ~/.config/opencode/skills/
func (e *Engine) loadExistingSkills() (map[string]SkillInfo, error) {
	skills := make(map[string]SkillInfo)

	// Default skills location
	skillsDir := filepath.Join(os.Getenv("HOME"), ".config", "opencode", "skills")
	if homeDrive := os.Getenv("HOMEDRIVE"); homeDrive != "" {
		skillsDir = filepath.Join(homeDrive, os.Getenv("HOMEPATH"), ".config", "opencode", "skills")
	}

	// Also check ~/.config/opencode/AGENTS.md
	agentsPath := filepath.Join(filepath.Dir(skillsDir), "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		if content, err := os.ReadFile(agentsPath); err == nil {
			fmt.Printf("  ✓ Found existing AGENTS.md at %s\n", agentsPath)
			e.parseExistingAgents(string(content), skills)
		}
	}

	// Walk skills directory
	if err := filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, "SKILL.md") {
			return nil
		}

		skillName := filepath.Base(filepath.Dir(path))
		relPath, _ := filepath.Rel(skillsDir, path)

		skills[skillName] = SkillInfo{
			Name:        skillName,
			Path:        path,
			RelPath:     relPath,
			IsGlobal:    true,
			Description: e.extractSkillDescription(path),
			Triggers:    e.extractSkillTriggers(path),
		}

		return nil
	}); err != nil {
		return skills, err
	}

	fmt.Printf("  ✓ Loaded %d existing global skills\n", len(skills))
	return skills, nil
}

// SkillInfo represents a loaded skill.
type SkillInfo struct {
	Name        string
	Path        string
	RelPath     string
	IsGlobal    bool
	Description string
	Triggers    []string
	TechStack   string // frontend, backend, tools
}

// parseExistingAgents parses the existing AGENTS.md to extract skill info.
func (e *Engine) parseExistingAgents(content string, skills map[string]SkillInfo) {
	lines := strings.Split(content, "\n")
	var currentSection string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Track sections
		if strings.HasPrefix(line, "## ") {
			currentSection = strings.TrimPrefix(line, "## ")
		}

		// Parse skills from tables
		if strings.Contains(line, "|") && strings.Contains(line, "Skill") {
			continue // Header row
		}

		if currentSection == "Skills" && strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[1])
				if name != "" && name != "Skill" {
					if _, ok := skills[name]; !ok {
						skills[name] = SkillInfo{Name: name, IsGlobal: true}
					}
				}
			}
		}
	}
}

// extractSkillDescription extracts description from a SKILL.md file.
func (e *Engine) extractSkillDescription(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ") {
			// Remove # and ## prefix
			desc := strings.TrimPrefix(line, "## ")
			desc = strings.TrimPrefix(desc, "# ")
			return desc
		}
		if strings.HasPrefix(line, "**Description**:") {
			return strings.TrimPrefix(line, "**Description**:")
		}
	}
	return ""
}

// extractSkillTriggers extracts trigger patterns from a SKILL.md file.
func (e *Engine) extractSkillTriggers(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var triggers []string
	lines := strings.Split(string(content), "\n")
	var inTriggerSection bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "trigger") {
			inTriggerSection = true
			continue
		}
		if inTriggerSection && strings.HasPrefix(line, "- ") {
			triggers = append(triggers, strings.TrimPrefix(line, "- "))
		}
		if inTriggerSection && strings.HasPrefix(line, "## ") {
			break
		}
	}

	return triggers
}

// detectTechStack detects the technology stack from the project.
func (e *Engine) detectTechStack() map[string][]TechItem {
	techs := make(map[string][]TechItem)

	// Check for package.json (Node.js/Frontend)
	if pkgPath := filepath.Join(e.projectDir, "package.json"); exists(pkgPath) {
		techs["frontend"] = append(techs["frontend"], TechItem{Name: "nodejs", Purpose: "package manager"})
		e.detectFrontendFromPackage(pkgPath, techs)
	}

	// Check for go.mod (Go)
	if goPath := filepath.Join(e.projectDir, "go.mod"); exists(goPath) {
		techs["backend"] = append(techs["backend"], TechItem{Name: "go", Purpose: "backend"})
	}

	// Check for requirements.txt or pyproject.toml (Python)
	if reqPath := filepath.Join(e.projectDir, "requirements.txt"); exists(reqPath) {
		techs["backend"] = append(techs["backend"], TechItem{Name: "python", Purpose: "backend"})
	}

	// Check for docker-compose.yml or Dockerfile
	if dockerPath := filepath.Join(e.projectDir, "docker-compose.yml"); exists(dockerPath) {
		techs["tools"] = append(techs["tools"], TechItem{Name: "docker", Purpose: "containerization"})
	}

	// Check for .github/workflows
	if wfPath := filepath.Join(e.projectDir, ".github", "workflows"); exists(wfPath) {
		techs["tools"] = append(techs["tools"], TechItem{Name: "github-actions", Purpose: "CI/CD"})
	}

	// Check for tsconfig.json
	if tsPath := filepath.Join(e.projectDir, "tsconfig.json"); exists(tsPath) {
		techs["frontend"] = append(techs["frontend"], TechItem{Name: "typescript", Purpose: "language"})
	}

	// Use already detected tech stack if available
	if len(e.techStack.Frontend) > 0 {
		techs["frontend"] = e.techStack.Frontend
	}
	if len(e.techStack.Backend) > 0 {
		techs["backend"] = e.techStack.Backend
	}
	if len(e.techStack.Tools) > 0 {
		techs["tools"] = e.techStack.Tools
	}

	return techs
}

// detectFrontendFromPackage detects frontend frameworks from package.json.
func (e *Engine) detectFrontendFromPackage(pkgPath string, techs map[string][]TechItem) {
	content, err := os.ReadFile(pkgPath)
	if err != nil {
		return
	}

	lower := strings.ToLower(string(content))

	if strings.Contains(lower, "react") {
		techs["frontend"] = append(techs["frontend"], TechItem{Name: "react", Purpose: "framework"})
	}
	if strings.Contains(lower, "vue") {
		techs["frontend"] = append(techs["frontend"], TechItem{Name: "vue", Purpose: "framework"})
	}
	if strings.Contains(lower, "angular") {
		techs["frontend"] = append(techs["frontend"], TechItem{Name: "angular", Purpose: "framework"})
	}
	if strings.Contains(lower, "next") {
		techs["frontend"] = append(techs["frontend"], TechItem{Name: "nextjs", Purpose: "framework"})
	}
	if strings.Contains(lower, "svelte") {
		techs["frontend"] = append(techs["frontend"], TechItem{Name: "svelte", Purpose: "framework"})
	}
	if strings.Contains(lower, "express") {
		techs["backend"] = append(techs["backend"], TechItem{Name: "express", Purpose: "framework"})
	}
	if strings.Contains(lower, "fastapi") {
		techs["backend"] = append(techs["backend"], TechItem{Name: "fastapi", Purpose: "framework"})
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// =============================================================================
// Root AGENTS.md Generation
// =============================================================================

// generateRootAgentsMD generates the root AGENTS.md file.
func (e *Engine) generateRootAgentsMD(existingSkills map[string]SkillInfo) error {
	// Check if AGENTS.md already exists
	agentsPath := filepath.Join(e.projectDir, "AGENTS.md")

	var content string
	if _, err := os.Stat(agentsPath); err == nil {
		// Merge with existing
		existing, _ := os.ReadFile(agentsPath)
		content = e.mergeAgentsMD(string(existing), existingSkills)
		fmt.Println("  ✓ Merged with existing AGENTS.md")
	} else {
		// Generate new
		content = e.generateNewAgentsMD(existingSkills)
	}

	e.writeFile(agentsPath, content)
	return nil
}

// generateNewAgentsMD generates a new root AGENTS.md.
func (e *Engine) generateNewAgentsMD(existingSkills map[string]SkillInfo) string {
	var sb strings.Builder

	sb.WriteString("# AGENTS.md — GROVE Agent Configuration\n\n")
	sb.WriteString("## Scope\n\n")
	sb.WriteString("This file configures the AI agents for the GROVE project.\n")
	sb.WriteString("Generated by `grove-spec` from raw ideas to production-ready specifications.\n\n")

	sb.WriteString("## Available Skills\n\n")
	sb.WriteString("| Skill | Description | Trigger |\n")
	sb.WriteString("|-------|-------------|---------|\n")

	// Add global skills from ~/.config/opencode/skills/
	for name, info := range existingSkills {
		triggers := strings.Join(info.Triggers, ", ")
		if triggers == "" {
			triggers = fmt.Sprintf("When working on %s", name)
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", name, info.Description, triggers))
	}

	// Add SDD skills
	sb.WriteString("| `sdd-init` | Initialize SDD context | `/sdd-init` |\n")
	sb.WriteString("| `sdd-explore` | Explore and investigate | `/sdd-explore <topic>` |\n")
	sb.WriteString("| `sdd-propose` | Create change proposal | `/sdd-new <name>` |\n")
	sb.WriteString("| `sdd-spec` | Write specifications | After proposal |\n")
	sb.WriteString("| `sdd-design` | Technical design | After spec |\n")
	sb.WriteString("| `sdd-tasks` | Task breakdown | After design |\n")
	sb.WriteString("| `sdd-apply` | Implementation | `/sdd-apply` |\n")
	sb.WriteString("| `sdd-verify` | Verification | `/sdd-verify` |\n")
	sb.WriteString("| `sdd-archive` | Archive change | `/sdd-archive` |\n")

	sb.WriteString("\n## Auto-Invocation\n\n")
	sb.WriteString("Skills are automatically loaded based on context:\n\n")
	sb.WriteString("| Context | Skill(s) Loaded |\n")
	sb.WriteString("|---------|----------------|\n")
	sb.WriteString("| Go tests, Bubbletea TUI | `go-testing` |\n")
	sb.WriteString("| Creating new AI skills | `skill-creator` |\n")
	sb.WriteString("| SDD workflow | Phase-specific skill |\n")
	sb.WriteString("| React/Next.js projects | `react-19`, `nextjs-15` |\n")
	sb.WriteString("| Angular projects | `angular/core`, `angular/forms` |\n")

	sb.WriteString("\n## Module-Specific Agents\n\n")
	sb.WriteString("Each module has its own AGENTS.md:\n\n")
	sb.WriteString("- `grove/internal/spec/AGENTS.md` — Spec engine\n")
	sb.WriteString("- `grove/internal/loop/AGENTS.md` — Loop orchestrator\n")
	sb.WriteString("- `grove/internal/opti/AGENTS.md` — Optimizer\n")
	sb.WriteString("- `grove/internal/sdd/AGENTS.md` — SDD client\n")
	sb.WriteString("- `grove/internal/engram/AGENTS.md` — Engram integration\n")

	sb.WriteString("\n## Conventions\n\n")
	sb.WriteString("### Writing Skills\n\n")
	sb.WriteString("1. Skills go in `.opencode/skills/<skill-name>/SKILL.md`\n")
	sb.WriteString("2. Include `## Trigger` section with context patterns\n")
	sb.WriteString("3. Include `## Instructions` with detailed workflow\n")
	sb.WriteString("4. Follow the skill template in `.opencode/SKILL_TEMPLATE.md`\n\n")

	sb.WriteString("### Agent Behavior\n\n")
	sb.WriteString("- NEVER add 'Co-Authored-By' or AI attribution to commits\n")
	sb.WriteString("- Use conventional commits format only\n")
	sb.WriteString("- When asking questions, STOP and wait for response\n")
	sb.WriteString("- Verify technical claims before stating them\n")

	return sb.String()
}

// mergeAgentsMD merges new skills with existing AGENTS.md.
func (e *Engine) mergeAgentsMD(existing string, newSkills map[string]SkillInfo) string {
	lines := strings.Split(existing, "\n")
	var inSkillsTable bool
	var result []string

	for _, line := range lines {
		// Detect skills table
		if strings.Contains(line, "| Skill | Description") {
			inSkillsTable = true
			result = append(result, line)
			continue
		}

		// End of skills table
		if inSkillsTable && strings.HasPrefix(strings.TrimSpace(line), "|") &&
			!strings.Contains(line, "|") {
			inSkillsTable = false
		}

		// Skip if already has the skill
		if inSkillsTable && strings.HasPrefix(strings.TrimSpace(line), "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				skillName := strings.TrimSpace(parts[1])
				if _, exists := newSkills[skillName]; !exists {
					result = append(result, line)
				}
			}
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// =============================================================================
// Module-Specific AGENTS.md Generation
// =============================================================================

// ModuleInfo represents a module in the project.
type ModuleInfo struct {
	Name        string
	Path        string
	AGENTSPath  string
	Skills      []string
	Patterns    []string
	Description string
}

// generateModuleAgentsMD generates AGENTS.md for each module.
func (e *Engine) generateModuleAgentsMD(globalSkills map[string]SkillInfo) error {
	// Define modules in the project
	modules := []ModuleInfo{
		{
			Name:        "spec",
			Path:        "grove/internal/spec",
			Description: "GROVE Spec engine for idea-to-specification",
			Skills:      []string{"sdd-spec", "sdd-design", "ralph-loop"},
			Patterns:    []string{"Iteration loop", "Quality scoring", "Component decomposition"},
		},
		{
			Name:        "loop",
			Path:        "grove/internal/loop",
			Description: "Orchestrates autonomous documentation-to-code loop",
			Skills:      []string{"sdd-apply", "sdd-verify"},
			Patterns:    []string{"Checkpoint system", "Error recovery", "Continuous mode"},
		},
		{
			Name:        "opti",
			Path:        "grove/internal/opti",
			Description: "Optimizer for prompt and code quality",
			Skills:      []string{"sdd-explore"},
			Patterns:    []string{"Token optimization", "Quality metrics"},
		},
		{
			Name:        "sdd",
			Path:        "grove/internal/sdd",
			Description: "SDD client for Spec-Driven Development",
			Skills:      []string{"sdd-init", "sdd-tasks", "sdd-archive"},
			Patterns:    []string{"Artifact management", "Phase orchestration"},
		},
		{
			Name:        "engram",
			Path:        "grove/internal/engram",
			Description: "Persistent memory integration",
			Skills:      []string{},
			Patterns:    []string{"Memory save/load", "Context recovery"},
		},
	}

	for _, module := range modules {
		if err := e.generateModuleAGENTS(&module, globalSkills); err != nil {
			return fmt.Errorf("failed to generate AGENTS.md for %s: %w", module.Name, err)
		}
	}

	return nil
}

// generateModuleAGENTS generates AGENTS.md for a specific module.
func (e *Engine) generateModuleAGENTS(module *ModuleInfo, globalSkills map[string]SkillInfo) error {
	modulePath := filepath.Join(e.projectDir, module.Path)
	agentsPath := filepath.Join(modulePath, "AGENTS.md")

	// Create directory if needed
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		return err
	}

	var content string
	if _, err := os.Stat(agentsPath); err == nil {
		// Merge with existing
		existing, _ := os.ReadFile(agentsPath)
		content = e.mergeModuleAgents(string(existing), module, globalSkills)
	} else {
		// Generate new
		content = e.generateModuleAgentsContent(module, globalSkills)
	}

	e.writeFile(agentsPath, content)
	return nil
}

// generateModuleAgentsContent generates content for module AGENTS.md.
func (e *Engine) generateModuleAgentsContent(module *ModuleInfo, globalSkills map[string]SkillInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s — Agent Configuration\n\n", strings.Title(module.Name)))
	sb.WriteString(fmt.Sprintf("## Description\n\n%s\n\n", module.Description))

	sb.WriteString("## Module Skills\n\n")
	sb.WriteString("| Skill | Purpose |\n")
	sb.WriteString("|-------|---------|\n")
	for _, skill := range module.Skills {
		sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", skill, e.getSkillPurpose(skill)))
	}

	sb.WriteString("\n## Implementation Patterns\n\n")
	for _, pattern := range module.Patterns {
		sb.WriteString(fmt.Sprintf("- **%s**\n", pattern))
	}

	sb.WriteString("\n## Integration Points\n\n")
	sb.WriteString("This module integrates with:\n\n")
	sb.WriteString("- Root AGENTS.md (global skills)\n")
	sb.WriteString("- SDD workflow phases\n")

	sb.WriteString("\n## Conventions\n\n")
	sb.WriteString("### Code Standards\n\n")
	sb.WriteString("- Use Go 1.21+ idioms\n")
	sb.WriteString("- Run `go fmt` before commits\n")
	sb.WriteString("- Add tests for new functionality\n")
	sb.WriteString("- Document exported functions\n\n")

	return sb.String()
}

// mergeModuleAgents merges content with existing module AGENTS.md.
func (e *Engine) mergeModuleAgents(existing string, module *ModuleInfo, globalSkills map[string]SkillInfo) string {
	// Simple merge: append new patterns if not present
	if !strings.Contains(existing, strings.Join(module.Patterns, ",")) {
		existing += "\n## New Patterns\n\n"
		for _, pattern := range module.Patterns {
			existing += fmt.Sprintf("- %s\n", pattern)
		}
	}
	return existing
}

// getSkillPurpose returns the purpose of a skill.
func (e *Engine) getSkillPurpose(skill string) string {
	purposes := map[string]string{
		"sdd-spec":    "Write specifications with requirements and scenarios",
		"sdd-design":  "Create technical design with architecture decisions",
		"sdd-apply":   "Implement tasks following specs and design",
		"sdd-verify":  "Validate implementation against specs",
		"sdd-init":    "Initialize SDD context in project",
		"sdd-tasks":   "Break down changes into task checklists",
		"sdd-archive": "Sync delta specs to main and archive",
		"sdd-explore": "Explore and investigate ideas",
		"ralph-loop":  "Autonomous documentation-to-code loop",
	}
	if p, ok := purposes[skill]; ok {
		return p
	}
	return "Custom skill"
}

// =============================================================================
// SKILLS.md Generation
// =============================================================================

// generateSkillsMD generates SKILLS.md for each detected technology.
func (e *Engine) generateSkillsMD(techStack map[string][]TechItem, existingSkills map[string]SkillInfo) error {
	skillsDir := filepath.Join(e.projectDir, ".opencode", "skills")
	os.MkdirAll(skillsDir, 0755)

	// Map technology to skill name
	techToSkill := map[string]string{
		"react":          "react-19",
		"vue":            "vue-3",
		"angular":        "angular",
		"nextjs":         "nextjs-15",
		"svelte":         "svelte",
		"nodejs":         "nodejs",
		"go":             "go",
		"python":         "python",
		"fastapi":        "django-drf",
		"express":        "nodejs",
		"docker":         "docker",
		"github-actions": "github-pr",
	}

	generated := make(map[string]bool)

	for category, items := range techStack {
		for _, item := range items {
			skillName := techToSkill[item.Name]
			if skillName == "" {
				continue
			}

			if generated[skillName] {
				continue
			}

			// Check if skill exists in global
			if _, exists := existingSkills[skillName]; exists {
				fmt.Printf("  ✓ Skill %s exists globally, linking\n", skillName)
				e.linkGlobalSkill(skillsDir, skillName, existingSkills[skillName])
				generated[skillName] = true
				continue
			}

			// Generate skill if not exists
			if err := e.generateTechSkill(skillsDir, item, category); err != nil {
				fmt.Printf("  ⚠ Warning: Could not generate skill for %s: %v\n", item.Name, err)
			}
			generated[skillName] = true
		}
	}

	return nil
}

// generateTechSkill generates a SKILL.md for a technology.
func (e *Engine) generateTechSkill(skillsDir string, tech TechItem, category string) error {
	skillDir := filepath.Join(skillsDir, tech.Name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")

	// Check if already exists
	if _, err := os.Stat(skillPath); err == nil {
		return nil // Already exists
	}

	content := e.generateSkillContent(tech, category)
	e.writeFile(skillPath, content)
	return nil
}

// generateSkillContent generates content for a technology SKILL.md.
func (e *Engine) generateSkillContent(tech TechItem, category string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s Skill\n\n", strings.Title(tech.Name)))
	sb.WriteString(fmt.Sprintf("**Purpose**: %s\n\n", tech.Purpose))
	sb.WriteString("## Trigger\n\n")
	sb.WriteString(fmt.Sprintf("When working on %s projects with %s\n\n", category, tech.Name))

	sb.WriteString("## Instructions\n\n")
	sb.WriteString("### Best Practices\n\n")

	// Add best practices based on technology
	switch tech.Name {
	case "react":
		sb.WriteString("- Use React 19 with Server Components\n")
		sb.WriteString("- Prefer `use client` directive for interactive components\n")
		sb.WriteString("- Use `use` hook for data fetching\n")
		sb.WriteString("- Follow container-presentational pattern\n")
	case "angular":
		sb.WriteString("- Use standalone components\n")
		sb.WriteString("- Follow Signal-based reactivity\n")
		sb.WriteString("- Use functional interceptors and guards\n")
		sb.WriteString("- Implement onPush change detection\n")
	case "go":
		sb.WriteString("- Use Go 1.21+ idioms\n")
		sb.WriteString("- Follow standard project layout\n")
		sb.WriteString("- Use context for cancellation\n")
		sb.WriteString("- Implement proper error handling\n")
	case "typescript":
		sb.WriteString("- Enable strict mode\n")
		sb.WriteString("- Use type inference when possible\n")
		sb.WriteString("- Prefer interfaces over types for objects\n")
	}

	sb.WriteString("\n## Patterns\n\n")
	sb.WriteString("- **Setup**: Initialize project with recommended tooling\n")
	sb.WriteString("- **Structure**: Organize by feature/domain\n")
	sb.WriteString("- **Testing**: Add unit and integration tests\n")

	return sb.String()
}

// linkGlobalSkill creates a symlink or reference to global skill.
func (e *Engine) linkGlobalSkill(skillsDir string, skillName string, info SkillInfo) error {
	linkDir := filepath.Join(skillsDir, skillName)

	// Check if already linked
	if _, err := os.Stat(linkDir); err == nil {
		return nil
	}

	// Create reference file instead of symlink (Windows compatibility)
	refContent := fmt.Sprintf("# %s Skill\n\n", skillName)
	refContent += "**Note**: This skill is defined globally.\n\n"
	refContent += fmt.Sprintf("Location: `%s`\n\n", info.Path)
	refContent += "## Instructions\n\n"
	refContent += "Follow the global skill definition.\n"

	e.writeFile(filepath.Join(linkDir, "SKILL.md"), refContent)
	return nil
}

// =============================================================================
// Legacy compatibility
// =============================================================================

// generateAgentsAndSkillsOld is the old signature for compatibility
func (e *Engine) generateAgentsAndSkillsOld(ctx context.Context) error {
	return e.generateAgentsAndSkills()
}
