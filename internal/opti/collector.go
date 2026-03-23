package opti

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	groveerrors "github.com/Gentleman-Programming/grove/internal/errors"
)

// FileCandidate represents a candidate file for context collection.
type FileCandidate struct {
	Path      string  `json:"path"`       // Full file path
	Layer     int     `json:"layer"`      // Source layer (1-4)
	Score     float64 `json:"score"`      // Relevance score
	LayerName string  `json:"layer_name"` // Human-readable layer name
}

// ContextResult contains the collected context for prompt optimization.
type ContextResult struct {
	Files          []FileCandidate `json:"files"`           // Selected source files (max 3)
	AgentsContent  string          `json:"agents_content"`  // Relevant AGENTS.md content
	SpecContent    string          `json:"spec_content"`    // Relevant SPEC.md sections
	Skills         []string        `json:"skills"`          // Discovered relevant skills
	TotalTokens    int             `json:"total_tokens"`    // Estimated token count
	DependencyRefs []string        `json:"dependency_refs"` // Adjacent modules for cross-module changes
	LayerLog       []LayerEntry    `json:"layer_log"`       // Log of which layer each file came from
}

// LayerEntry logs which layer a file was selected from.
type LayerEntry struct {
	File      string `json:"file"`
	Layer     int    `json:"layer"`
	LayerName string `json:"layer_name"`
}

// Collector handles context collection using the 4-layer heuristic.
type Collector struct {
	projectRoot string
	agentsPath  string
	specPath    string
	skillsDir   string
}

// NewCollector creates a new ContextCollector.
func NewCollector(projectRoot string) *Collector {
	return &Collector{
		projectRoot: projectRoot,
		agentsPath:  filepath.Join(projectRoot, "AGENTS.md"),
		specPath:    filepath.Join(projectRoot, "SPEC.md"),
		skillsDir:   filepath.Join(projectRoot, ".opencode", "skills"),
	}
}

// Collect collects relevant context based on the classified intent.
// It uses the 4-layer heuristic to select up to 3 source files.
func (c *Collector) Collect(ctx context.Context, intent IntentClassification) (*ContextResult, error) {
	result := &ContextResult{
		Files:    make([]FileCandidate, 0),
		Skills:   make([]string, 0),
		LayerLog: make([]LayerEntry, 0),
	}

	// Layer 1: AGENTS.md explicit references
	agentsFiles, err := c.layer1AgentsReferences(intent)
	if err != nil {
		slog.Warn("failed to get agents.md references",
			slog.String("error", err.Error()),
			slog.String("intent", string(intent.Intent)))
	} else {
		for _, fc := range agentsFiles {
			fc.Layer = 1
			fc.LayerName = "AGENTS.md References"
			result.Files = append(result.Files, fc)
			result.LayerLog = append(result.LayerLog, LayerEntry{
				File:      fc.Path,
				Layer:     1,
				LayerName: "AGENTS.md References",
			})
		}
	}

	// Layer 2: Recent Git commits filtered by intent keywords
	if len(result.Files) < 3 {
		gitFiles, err := c.layer2GitHistory(intent)
		if err != nil {
			slog.Warn("failed to get git history references",
				slog.String("error", err.Error()),
				slog.String("intent", string(intent.Intent)))
		} else {
			for _, fc := range gitFiles {
				if c.fileNotSelected(result.Files, fc.Path) {
					fc.Layer = 2
					fc.LayerName = "Git History"
					result.Files = append(result.Files, fc)
					result.LayerLog = append(result.LayerLog, LayerEntry{
						File:      fc.Path,
						Layer:     2,
						LayerName: "Git History",
					})
				}
			}
		}
	}

	// Layer 3: Intent keyword path match
	if len(result.Files) < 3 {
		keywordFiles, err := c.layer3KeywordMatch(intent)
		if err != nil {
			slog.Warn("failed to get keyword path matches",
				slog.String("error", err.Error()),
				slog.String("intent", string(intent.Intent)))
		} else {
			for _, fc := range keywordFiles {
				if c.fileNotSelected(result.Files, fc.Path) {
					fc.Layer = 3
					fc.LayerName = "Keyword Path Match"
					result.Files = append(result.Files, fc)
					result.LayerLog = append(result.LayerLog, LayerEntry{
						File:      fc.Path,
						Layer:     3,
						LayerName: "Keyword Path Match",
					})
				}
			}
		}
	}

	// Layer 4: SPEC.md component references
	if len(result.Files) < 3 {
		specFiles, err := c.layer4SpecReferences(intent)
		if err != nil {
			slog.Warn("failed to get spec.md references",
				slog.String("error", err.Error()),
				slog.String("intent", string(intent.Intent)))
		} else {
			for _, fc := range specFiles {
				if c.fileNotSelected(result.Files, fc.Path) {
					fc.Layer = 4
					fc.LayerName = "SPEC.md References"
					result.Files = append(result.Files, fc)
					result.LayerLog = append(result.LayerLog, LayerEntry{
						File:      fc.Path,
						Layer:     4,
						LayerName: "SPEC.md References",
					})
				}
			}
		}
	}

	// Limit to 3 files
	if len(result.Files) > 3 {
		result.Files = result.Files[:3]
	}

	// Read AGENTS.md content
	if content, err := c.readFile(c.agentsPath); err != nil {
		slog.Warn("failed to read agents.md",
			slog.String("error", err.Error()),
			slog.String("path", c.agentsPath))
	} else {
		result.AgentsContent = c.filterAgentsContent(content, intent)
	}

	// Read SPEC.md content
	if content, err := c.readFile(c.specPath); err != nil {
		slog.Warn("failed to read spec.md",
			slog.String("error", err.Error()),
			slog.String("path", c.specPath))
	} else {
		result.SpecContent = c.filterSpecContent(content, intent)
	}

	// Discover relevant skills
	result.Skills = c.discoverSkills(intent)

	// Build dependency graph context for cross-module changes
	if c.isCrossModuleIntent(intent) {
		result.DependencyRefs = c.buildDependencyContext(result.Files)
	}

	return result, nil
}

// SelectFiles applies the 4-layer heuristic to select up to 3 relevant files.
// This is the core file selection algorithm.
func (c *Collector) SelectFiles(intent IntentClassification) ([]FileCandidate, error) {
	candidates := make([]FileCandidate, 0)
	var errs []error

	// Layer 1: AGENTS.md references (highest priority)
	layer1Files, err := c.layer1AgentsReferences(intent)
	if err != nil {
		slog.Warn("layer1 agents references failed",
			slog.String("error", err.Error()))
		errs = append(errs, err)
	} else {
		candidates = append(candidates, layer1Files...)
	}

	// Layer 2: Git history
	if len(candidates) < 3 {
		layer2Files, err := c.layer2GitHistory(intent)
		if err != nil {
			slog.Warn("layer2 git history failed",
				slog.String("error", err.Error()))
			errs = append(errs, err)
		} else {
			candidates = append(candidates, layer2Files...)
		}
	}

	// Layer 3: Keyword path match
	if len(candidates) < 3 {
		layer3Files, err := c.layer3KeywordMatch(intent)
		if err != nil {
			slog.Warn("layer3 keyword match failed",
				slog.String("error", err.Error()))
			errs = append(errs, err)
		} else {
			candidates = append(candidates, layer3Files...)
		}
	}

	// Layer 4: SPEC.md references (lowest priority)
	if len(candidates) < 3 {
		layer4Files, err := c.layer4SpecReferences(intent)
		if err != nil {
			slog.Warn("layer4 spec references failed",
				slog.String("error", err.Error()))
			errs = append(errs, err)
		} else {
			candidates = append(candidates, layer4Files...)
		}
	}

	// Remove duplicates and limit to 3
	seen := make(map[string]bool)
	unique := make([]FileCandidate, 0)
	for _, fc := range candidates {
		if !seen[fc.Path] {
			seen[fc.Path] = true
			unique = append(unique, fc)
			if len(unique) >= 3 {
				break
			}
		}
	}

	// Sort by layer (lower is better)
	sort.Slice(unique, func(i, j int) bool {
		if unique[i].Layer != unique[j].Layer {
			return unique[i].Layer < unique[j].Layer
		}
		return unique[i].Score > unique[j].Score
	})

	// Return partial result even if some layers failed
	if len(errs) > 0 && len(unique) == 0 {
		return unique, errors.Join(errs...)
	}

	return unique, nil
}

// layer1AgentsReferences scans AGENTS.md for explicit file references.
func (c *Collector) layer1AgentsReferences(intent IntentClassification) ([]FileCandidate, error) {
	var candidates []FileCandidate

	content, err := c.readFile(c.agentsPath)
	if err != nil {
		if c.fileExists(c.agentsPath) {
			return nil, groveerrors.NewFileError(c.agentsPath, "read agents references", err)
		}
		// File doesn't exist is not an error for this layer
		slog.Debug("agents.md not found, skipping layer 1",
			slog.String("path", c.agentsPath))
		return candidates, nil
	}

	// Find the scoped section for this intent
	sectionContent := c.filterAgentsContent(content, intent)

	// Extract @file references
	fileRefPattern := regexp.MustCompile(`@([^\s]+)`)
	matches := fileRefPattern.FindAllStringSubmatch(sectionContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			path := match[1]
			// Resolve relative paths
			if !filepath.IsAbs(path) {
				path = filepath.Join(c.projectRoot, path)
			}
			if c.fileExists(path) {
				candidates = append(candidates, FileCandidate{
					Path:  path,
					Layer: 1,
					Score: 1.0, // Layer 1 always has highest priority
				})
			} else {
				slog.Debug("referenced file not found",
					slog.String("path", path))
			}
		}
	}

	return candidates, nil
}

// layer2GitHistory retrieves recent files from git history matching intent keywords.
func (c *Collector) layer2GitHistory(intent IntentClassification) ([]FileCandidate, error) {
	var candidates []FileCandidate

	// Check if git is available
	cmd := exec.Command("git", "log", "--name-only", "-n", "20", "--pretty=format:")
	cmd.Dir = c.projectRoot

	output, err := cmd.Output()
	if err != nil {
		// Git not available is not an error, just skip this layer
		slog.Debug("git not available for layer 2",
			slog.String("project_root", c.projectRoot),
			slog.String("error", err.Error()))
		return candidates, nil
	}

	lines := strings.Split(string(output), "\n")
	recentFiles := make([]string, 0, 20)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "Merge") {
			// Simple heuristic: if line looks like a file path
			if strings.Contains(line, "/") || strings.HasSuffix(line, ".go") ||
				strings.HasSuffix(line, ".ts") || strings.HasSuffix(line, ".tsx") ||
				strings.HasSuffix(line, ".js") || strings.HasSuffix(line, ".jsx") {
				recentFiles = append(recentFiles, line)
			}
		}
	}

	// Score files by keyword match
	for _, file := range recentFiles {
		score := c.scoreFileByKeywords(file, intent.Keywords)
		if score > 0 {
			candidates = append(candidates, FileCandidate{
				Path:  filepath.Join(c.projectRoot, file),
				Layer: 2,
				Score: score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	return candidates, nil
}

// layer3KeywordMatch scores all files by path keyword overlap.
func (c *Collector) layer3KeywordMatch(intent IntentClassification) ([]FileCandidate, error) {
	var candidates []FileCandidate

	// Walk the project tree
	err := filepath.Walk(c.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Debug("error accessing path during walk",
				slog.String("path", path),
				slog.String("error", err.Error()))
			return nil // Skip inaccessible paths
		}

		// Skip hidden directories, node_modules, .git, etc.
		base := filepath.Base(path)
		if len(base) > 0 && (base[0] == '.' || base == "node_modules" || base == "vendor") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only consider source files
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if !c.isSourceFile(ext) {
			return nil
		}

		score := c.scoreFileByKeywords(path, intent.Keywords)
		if score > 0 {
			candidates = append(candidates, FileCandidate{
				Path:  path,
				Layer: 3,
				Score: score,
			})
		}

		return nil
	})

	if err != nil {
		return candidates, groveerrors.NewOptiError("keyword_match", "walk project tree", err)
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Return top candidates
	if len(candidates) > 10 {
		candidates = candidates[:10]
	}

	return candidates, nil
}

// layer4SpecReferences scans SPEC.md for component references.
func (c *Collector) layer4SpecReferences(intent IntentClassification) ([]FileCandidate, error) {
	var candidates []FileCandidate

	content, err := c.readFile(c.specPath)
	if err != nil {
		if c.fileExists(c.specPath) {
			return nil, groveerrors.NewFileError(c.specPath, "read spec references", err)
		}
		// File doesn't exist is not an error for this layer
		slog.Debug("spec.md not found, skipping layer 4",
			slog.String("path", c.specPath))
		return candidates, nil
	}

	// Find relevant sections
	sectionContent := c.filterSpecContent(content, intent)

	// Extract component/module names from SPEC.md
	// Look for patterns like "## ComponentName", "- ComponentName", "Component: Name"
	componentPatterns := []string{
		`##\s+([A-Z][a-zA-Z0-9]+)`,
		`###\s+([A-Z][a-zA-Z0-9]+)`,
		`-\s+\*\*([A-Z][a-zA-Z0-9]+)\*\*`,
		`([A-Z][a-zA-Z0-9]+)\s+component`,
		`component\s+([A-Z][a-zA-Z0-9]+)`,
	}

	var componentNames []string
	seen := make(map[string]bool)

	for _, pattern := range componentPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindAllStringSubmatch(sectionContent, -1)
		for _, match := range matches {
			if len(match) > 1 && !seen[strings.ToLower(match[1])] {
				seen[strings.ToLower(match[1])] = true
				componentNames = append(componentNames, match[1])
			}
		}
	}

	// Search for files matching these components
	for _, compName := range componentNames {
		matchingFiles := c.findFilesByComponentName(compName)
		for _, file := range matchingFiles {
			candidates = append(candidates, FileCandidate{
				Path:  file,
				Layer: 4,
				Score: 0.5, // Lower base score for layer 4
			})
		}
	}

	return candidates, nil
}

// scoreFileByKeywords calculates a relevance score based on keyword overlap.
func (c *Collector) scoreFileByKeywords(path string, keywords []string) float64 {
	if len(keywords) == 0 {
		return 0
	}

	// Get relative path from project root
	relPath, err := filepath.Rel(c.projectRoot, path)
	if err != nil {
		return 0
	}

	// Normalize path for matching
	normalizedPath := strings.ToLower(relPath)
	// Split on various separators
	pathParts := regexp.MustCompile(`[/\\._-]`).Split(normalizedPath, -1)

	matchCount := 0
	for _, keyword := range keywords {
		lowerKeyword := strings.ToLower(keyword)
		for _, part := range pathParts {
			if part == lowerKeyword || strings.Contains(part, lowerKeyword) {
				matchCount++
				break
			}
		}
	}

	if matchCount == 0 {
		return 0
	}

	return float64(matchCount) / float64(len(keywords))
}

// findFilesByComponentName searches for files matching a component name.
func (c *Collector) findFilesByComponentName(componentName string) []string {
	var matches []string

	lowerName := strings.ToLower(componentName)

	err := filepath.Walk(c.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		baseName := strings.ToLower(info.Name())
		if strings.Contains(baseName, lowerName) || strings.HasPrefix(baseName, lowerName) {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		return matches
	}

	return matches
}

// filterAgentsContent extracts the relevant section from AGENTS.md.
func (c *Collector) filterAgentsContent(content string, intent IntentClassification) string {
	// Simple implementation: return content up to a reasonable size
	// A full implementation would find the scoped section
	lines := strings.Split(content, "\n")
	var filtered []string
	charCount := 0
	maxChars := 3000

	intentKeywords := map[string]string{
		"feature-addition":     "feature|ui|component|addition",
		"bug-fix":              "bug|fix|issue|error",
		"refactor":             "refactor|restructure",
		"documentation-update": "doc|document|comment",
		"configuration-change": "config|setting|env",
	}

	keyword := intentKeywords[string(intent.Intent)]
	if keyword == "" {
		keyword = "general|default"
	}

	inSection := false
	sectionPattern := regexp.MustCompile(`(?i)` + keyword)

	for _, line := range lines {
		// Look for section headers
		if strings.HasPrefix(strings.TrimSpace(line), "##") {
			inSection = sectionPattern.MatchString(line)
		}

		if inSection {
			filtered = append(filtered, line)
			charCount += len(line)
			if charCount > maxChars {
				break
			}
		}
	}

	return strings.Join(filtered, "\n")
}

// filterSpecContent extracts relevant sections from SPEC.md.
func (c *Collector) filterSpecContent(content string, intent IntentClassification) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	charCount := 0
	maxChars := 2000

	// Find sections matching intent keywords
	for _, keyword := range intent.Keywords {
		keywordPattern := regexp.MustCompile(`(?i)` + keyword)
		if keywordPattern.MatchString(content) {
			// Find matching sections
			for i, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "##") &&
					keywordPattern.MatchString(line) {
					// Include this section and next 20 lines
					end := i + 20
					if end > len(lines) {
						end = len(lines)
					}
					for j := i; j < end; j++ {
						filtered = append(filtered, lines[j])
						charCount += len(lines[j])
						if charCount > maxChars {
							return strings.Join(filtered, "\n")
						}
					}
				}
			}
		}
	}

	// Fallback: return first part of spec
	if len(filtered) == 0 {
		for i, line := range lines {
			if i > 50 { // About 2000 chars
				break
			}
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}

// discoverSkills finds relevant skills based on intent.
func (c *Collector) discoverSkills(intent IntentClassification) []string {
	var skills []string

	// Check project skills
	if c.dirExists(c.skillsDir) {
		err := filepath.Walk(c.skillsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() || info.Name() != "SKILL.md" {
				return nil
			}

			// Check if skill name/description matches intent
			content, err := c.readFile(path)
			if err != nil {
				return nil
			}

			for _, keyword := range intent.Keywords {
				if strings.Contains(strings.ToLower(content), strings.ToLower(keyword)) {
					// Extract skill name from path
					skillName := filepath.Base(filepath.Dir(path))
					skills = append(skills, skillName)
					break
				}
			}

			return nil
		})

		if err == nil {
			// Deduplicate
			seen := make(map[string]bool)
			var unique []string
			for _, s := range skills {
				if !seen[s] {
					seen[s] = true
					unique = append(unique, s)
				}
			}
			skills = unique
		}
	}

	return skills
}

// buildDependencyContext builds dependency graph context for cross-module changes.
func (c *Collector) buildDependencyContext(files []FileCandidate) []string {
	var refs []string
	maxTokens := 500 // Budget for dependency context

	for _, file := range files {
		if maxTokens <= 0 {
			break
		}

		// Read only file headers (first 20 tokens of imports)
		imports := c.extractImports(file.Path)
		for _, imp := range imports {
			if maxTokens <= 0 {
				break
			}
			refs = append(refs, imp)
			maxTokens -= 10 // Rough token estimate per import
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, r := range refs {
		if !seen[r] {
			seen[r] = true
			unique = append(unique, r)
		}
	}

	return unique
}

// extractImports extracts import/require statements from file headers only.
func (c *Collector) extractImports(filePath string) []string {
	var imports []string

	content, err := c.readFile(filePath)
	if err != nil {
		return imports
	}

	lines := strings.Split(content, "\n")
	tokenCount := 0
	maxTokens := 20

	// Common import patterns
	importPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^import\s+"([^"]+)"`),
		regexp.MustCompile(`^import\s+'([^']+)'`),
		regexp.MustCompile(`^import\s+\(`),
		regexp.MustCompile(`^require\(['"]([^'"]+)['"]\)`),
		regexp.MustCompile(`from\s+['"]([^'"]+)['"]`),
	}

	inBlock := false
	for _, line := range lines {
		tokenCount += len(strings.Fields(line))
		if tokenCount > maxTokens {
			break
		}

		trimmed := strings.TrimSpace(line)

		// Handle import blocks
		if strings.HasPrefix(trimmed, "import") && strings.Contains(trimmed, "(") {
			inBlock = true
			continue
		}
		if inBlock {
			if strings.Contains(trimmed, ")") {
				inBlock = false
			}
			imports = append(imports, trimmed)
			continue
		}

		// Match single line imports
		for _, pattern := range importPatterns {
			matches := pattern.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				imports = append(imports, matches[1])
				break
			}
		}
	}

	return imports
}

// isCrossModuleIntent determines if the intent involves cross-module changes.
func (c *Collector) isCrossModuleIntent(intent IntentClassification) bool {
	crossModuleKeywords := []string{
		"connect", "integrate", "link", "dependency",
		"bridge", "interface", "couple", "refactor",
	}

	if intent.Intent == IntentRefactor || intent.Intent == IntentBugFix {
		return true
	}

	for _, keyword := range crossModuleKeywords {
		for _, kw := range intent.Keywords {
			if strings.Contains(strings.ToLower(kw), keyword) {
				return true
			}
		}
	}

	return false
}

// Helper methods

func (c *Collector) readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (c *Collector) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (c *Collector) dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (c *Collector) fileNotSelected(files []FileCandidate, path string) bool {
	for _, f := range files {
		if f.Path == path {
			return false
		}
	}
	return true
}

func (c *Collector) isSourceFile(ext string) bool {
	sourceExtensions := map[string]bool{
		".go": true, ".ts": true, ".tsx": true,
		".js": true, ".jsx": true, ".py": true,
		".rs": true, ".java": true, ".cs": true,
	}
	return sourceExtensions[ext]
}
