package opti

import (
	"context"
	"fmt"
	"strings"
)

// PromptElementType represents the type of optimization element.
type PromptElementType string

const (
	ElementFileReference   PromptElementType = "file-reference"
	ElementScopeBoundary   PromptElementType = "scope-boundary"
	ElementSkillInvocation PromptElementType = "skill-invocation"
	ElementSuccessCriteria PromptElementType = "success-criteria"
	ElementPlanMode        PromptElementType = "plan-mode"
	ElementOutOfScope      PromptElementType = "out-of-scope-boundary"
	ElementWarning         PromptElementType = "warning"
)

// PromptElement represents a single element in the optimized prompt.
type PromptElement struct {
	Type        PromptElementType `json:"type"`        // Element type
	Content     string            `json:"content"`     // The actual content
	Explanation string            `json:"explanation"` // WHY explanation
	LineNumber  int               `json:"line_number"` // Position in prompt
}

// OptimizedPrompt represents the final optimized prompt ready for OpenCode.
type OptimizedPrompt struct {
	Original    string          `json:"original"`     // Original user input
	Optimized   string          `json:"optimized"`    // Final optimized prompt
	Elements    []PromptElement `json:"elements"`     // Individual elements with explanations
	TokenCount  int             `json:"token_count"`  // Estimated token count
	TokenBudget int             `json:"token_budget"` // Budget limit (default 2000)
	SkillsUsed  []string        `json:"skills_used"`  // Skills referenced
	Warnings    []string        `json:"warnings"`     // Any warnings or issues
}

// Optimizer generates optimized prompts from user input and context.
type Optimizer struct {
	tokenBudget int
}

// NewOptimizer creates a new PromptOptimizer.
func NewOptimizer(tokenBudget int) *Optimizer {
	if tokenBudget <= 0 {
		tokenBudget = 2000 // Default token budget
	}
	return &Optimizer{
		tokenBudget: tokenBudget,
	}
}

// Optimize generates an optimized prompt from user input and collected context.
func (o *Optimizer) Optimize(ctx context.Context, input string, classification IntentClassification, context *ContextResult) (*OptimizedPrompt, error) {
	result := &OptimizedPrompt{
		Original:    input,
		Elements:    make([]PromptElement, 0),
		TokenBudget: o.tokenBudget,
		SkillsUsed:  context.Skills,
		Warnings:    make([]string, 0),
	}

	var promptParts []string
	lineNum := 1

	// 1. Add file references (@file paths)
	fileRefs := o.buildFileReferences(context)
	for _, ref := range fileRefs {
		result.Elements = append(result.Elements, PromptElement{
			Type:       ElementFileReference,
			Content:    ref,
			LineNumber: lineNum,
		})
		promptParts = append(promptParts, ref)
		lineNum++
	}

	// 2. Add the core request
	coreRequest := o.buildCoreRequest(input, classification)
	result.Elements = append(result.Elements, PromptElement{
		Type:       ElementFileReference, // Primary action type
		Content:    coreRequest,
		LineNumber: lineNum,
	})
	promptParts = append(promptParts, coreRequest)
	lineNum++

	// 3. Add skill invocations
	if len(context.Skills) > 0 {
		skillInvocation := o.buildSkillInvocation(context.Skills)
		result.Elements = append(result.Elements, PromptElement{
			Type:       ElementSkillInvocation,
			Content:    skillInvocation,
			LineNumber: lineNum,
		})
		promptParts = append(promptParts, skillInvocation)
		lineNum++
	}

	// 4. Add success criteria
	successCriteria := o.buildSuccessCriteria(classification, context)
	result.Elements = append(result.Elements, PromptElement{
		Type:        ElementSuccessCriteria,
		Content:     successCriteria,
		Explanation: "Success criteria define when the task is complete, preventing scope creep and unclear expectations.",
		LineNumber:  lineNum,
	})
	promptParts = append(promptParts, successCriteria)
	lineNum++

	// 5. Add scope boundaries
	scopeBoundary := o.buildScopeBoundary(classification, context)
	result.Elements = append(result.Elements, PromptElement{
		Type:        ElementScopeBoundary,
		Content:     scopeBoundary,
		Explanation: "Scope boundaries prevent the agent from modifying unrelated files or making unintended changes.",
		LineNumber:  lineNum,
	})
	promptParts = append(promptParts, scopeBoundary)
	lineNum++

	// 6. Add out-of-scope explicitly
	outOfScope := o.buildOutOfScope(classification, context)
	result.Elements = append(result.Elements, PromptElement{
		Type:        ElementOutOfScope,
		Content:     outOfScope,
		Explanation: "Explicitly stating what NOT to change reduces the risk of accidental modifications.",
		LineNumber:  lineNum,
	})
	promptParts = append(promptParts, outOfScope)
	lineNum++

	// 7. Add Plan mode recommendation for risky/large changes
	if o.shouldRecommendPlanMode(classification, context) {
		planRecommendation := o.buildPlanModeRecommendation()
		result.Elements = append(result.Elements, PromptElement{
			Type:        ElementPlanMode,
			Content:     planRecommendation,
			Explanation: "Plan mode is recommended for complex changes to ensure all implications are considered before execution.",
			LineNumber:  lineNum,
		})
		promptParts = append(promptParts, planRecommendation)
		lineNum++
	}

	// 8. Add dependency context for cross-module changes
	if len(context.DependencyRefs) > 0 {
		depContext := o.buildDependencyContext(context.DependencyRefs)
		result.Elements = append(result.Elements, PromptElement{
			Type:        ElementScopeBoundary,
			Content:     depContext,
			Explanation: "This file imports/is imported by the primary target — changes may propagate.",
			LineNumber:  lineNum,
		})
		promptParts = append(promptParts, depContext)
		lineNum++
	}

	// 9. Check AGENTS.md for additional instructions
	if context.AgentsContent != "" {
		agentsNote := o.buildAgentsNote()
		result.Elements = append(result.Elements, PromptElement{
			Type:        ElementWarning,
			Content:     agentsNote,
			Explanation: "Always follow the project's AGENTS.md conventions and constraints.",
			LineNumber:  lineNum,
		})
		promptParts = append(promptParts, agentsNote)
		lineNum++
	}

	// Assemble the optimized prompt
	result.Optimized = strings.Join(promptParts, "\n\n")

	// Calculate token count (rough estimate: 1 token ≈ 4 chars)
	result.TokenCount = len(result.Optimized) / 4

	// Add warnings if approaching token limit
	if result.TokenCount > int(float64(o.tokenBudget)*0.9) {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"Token count (%d) is approaching budget limit (%d)",
			result.TokenCount, o.tokenBudget,
		))
	}

	// Add warning if no files were selected
	if len(fileRefs) == 0 {
		result.Warnings = append(result.Warnings, "No explicit file references could be determined")
	}

	return result, nil
}

// buildFileReferences creates @file path references from selected files.
func (o *Optimizer) buildFileReferences(ctx *ContextResult) []string {
	var refs []string
	seen := make(map[string]bool)

	for _, file := range ctx.Files {
		if file.Path == "" || seen[file.Path] {
			continue
		}
		seen[file.Path] = true

		// Get relative path for cleaner reference
		relPath := file.Path
		if strings.Contains(file.Path, "/src/") {
			idx := strings.Index(file.Path, "/src/")
			relPath = "@." + file.Path[idx+4:]
		} else if strings.Contains(file.Path, "\\src\\") {
			idx := strings.Index(file.Path, "\\src\\")
			relPath = "@." + file.Path[idx+5:]
		} else {
			relPath = "@" + file.Path
		}

		refs = append(refs, relPath)
	}

	return refs
}

// buildCoreRequest creates the core task description.
func (o *Optimizer) buildCoreRequest(input string, classification IntentClassification) string {
	// Capitalize and punctuate if needed
	request := strings.TrimSpace(input)
	if !strings.HasSuffix(request, ".") && !strings.HasSuffix(request, "!") && !strings.HasSuffix(request, "?") {
		request = request + "."
	}

	// Add intent-specific framing if helpful
	switch classification.Intent {
	case IntentFeatureAddition:
		if !strings.Contains(strings.ToLower(request), "add") {
			request = "Add: " + request
		}
	case IntentBugFix:
		if !strings.Contains(strings.ToLower(request), "fix") {
			request = "Fix: " + request
		}
	case IntentRefactor:
		if !strings.Contains(strings.ToLower(request), "refactor") {
			request = "Refactor: " + request
		}
	}

	return request
}

// buildSkillInvocation creates skill() calls for discovered skills.
func (o *Optimizer) buildSkillInvocation(skills []string) string {
	if len(skills) == 0 {
		return ""
	}

	var calls []string
	for _, skill := range skills {
		calls = append(calls, fmt.Sprintf("skill({ name: '%s' })", skill))
	}

	return strings.Join(calls, "\n")
}

// buildSuccessCriteria generates success criteria from context.
func (o *Optimizer) buildSuccessCriteria(classification IntentClassification, ctx *ContextResult) string {
	// Derive from intent type
	var criteria string

	switch classification.Intent {
	case IntentFeatureAddition:
		criteria = "This is done when the new feature is implemented and functional"
	case IntentBugFix:
		criteria = "This is done when the bug is fixed and does not recur"
	case IntentRefactor:
		criteria = "This is done when the code is restructured and all tests pass"
	case IntentDocumentationUpdate:
		criteria = "This is done when the documentation is accurate and complete"
	case IntentConfigurationChange:
		criteria = "This is done when the configuration is updated and verified"
	default:
		criteria = "This is done when the requested changes are complete and verified"
	}

	// Add specific file references if available
	if len(ctx.Files) > 0 {
		firstFile := ctx.Files[0].Path
		if strings.Contains(firstFile, "/src/") {
			idx := strings.Index(firstFile, "/src/")
			criteria = fmt.Sprintf("This is done when %s is updated and working correctly", "@."+firstFile[idx+4:])
		}
	}

	// Add domain-specific detail
	if classification.Domain != "" {
		criteria = fmt.Sprintf("This is done when %s is working correctly in the %s module",
			strings.ToLower(classification.Domain), classification.Domain)
	}

	return criteria
}

// buildScopeBoundary creates scope boundaries from context.
func (o *Optimizer) buildScopeBoundary(classification IntentClassification, ctx *ContextResult) string {
	if len(ctx.Files) == 0 {
		return "Scope: Only modify the files directly related to this task"
	}

	var files []string
	for _, f := range ctx.Files {
		// Extract just the filename or last path component
		base := f.Path
		if idx := strings.LastIndex(base, "/"); idx >= 0 {
			base = base[idx+1:]
		}
		if idx := strings.LastIndex(base, "\\"); idx >= 0 {
			base = base[idx+1:]
		}
		files = append(files, base)
	}

	return fmt.Sprintf("Scope: Focus on %s", strings.Join(files, ", "))
}

// buildOutOfScope creates explicit out-of-scope boundaries.
func (o *Optimizer) buildOutOfScope(classification IntentClassification, ctx *ContextResult) string {
	// Common patterns to exclude based on intent
	var exclusions []string

	switch classification.Intent {
	case IntentFeatureAddition:
		exclusions = []string{"Do NOT modify unrelated components", "Do NOT add unrelated features"}
	case IntentBugFix:
		exclusions = []string{"Do NOT refactor working code", "Do NOT change unrelated functionality"}
	case IntentRefactor:
		exclusions = []string{"Do NOT change functionality", "Do NOT add new features", "Do NOT fix bugs"}
	case IntentDocumentationUpdate:
		exclusions = []string{"Do NOT modify code", "Do NOT change behavior"}
	case IntentConfigurationChange:
		exclusions = []string{"Do NOT modify application logic", "Do NOT change unrelated configs"}
	}

	// Add specific file exclusions based on selected files
	if len(ctx.Files) > 0 {
		// Common files to exclude based on common patterns
		exclusions = append(exclusions, "Do NOT modify navigation or layout components unless explicitly required")
	}

	return strings.Join(exclusions, "\n")
}

// shouldRecommendPlanMode determines if Plan mode should be recommended.
func (o *Optimizer) shouldRecommendPlanMode(classification IntentClassification, ctx *ContextResult) bool {
	// More than 3 files affected
	if len(ctx.Files) > 3 {
		return true
	}

	// Core logic or critical files
	criticalPatterns := []string{"auth", "security", "core", "main", "index", "app"}
	for _, file := range ctx.Files {
		lowerPath := strings.ToLower(file.Path)
		for _, pattern := range criticalPatterns {
			if strings.Contains(lowerPath, pattern) {
				return true
			}
		}
	}

	// Certain intent types
	if classification.Intent == IntentRefactor || classification.Intent == IntentBugFix {
		if classification.Confidence < 0.7 {
			return true // Uncertain about scope
		}
	}

	return false
}

// buildPlanModeRecommendation creates the Plan mode recommendation.
func (o *Optimizer) buildPlanModeRecommendation() string {
	return "Consider running in Plan mode first to review the proposed changes before execution"
}

// buildDependencyContext creates context about adjacent modules.
func (o *Optimizer) buildDependencyContext(depRefs []string) string {
	if len(depRefs) == 0 {
		return ""
	}

	var notes []string
	for _, ref := range depRefs {
		notes = append(notes, fmt.Sprintf("@%s — this file imports/is imported by the primary target", ref))
	}

	return strings.Join(notes, "\n")
}

// buildAgentsNote adds a reference to AGENTS.md conventions.
func (o *Optimizer) buildAgentsNote() string {
	return "Follow the conventions and constraints defined in AGENTS.md"
}

// CalculateTokenCount estimates the token count for a string.
func (o *Optimizer) CalculateTokenCount(text string) int {
	// Simple estimation: ~4 characters per token on average
	return len(text) / 4
}
