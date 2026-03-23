// Package loop provides the quality gate for documentation scoring.
//
// Quality Gate evaluates documentation completeness, coherence, and format
// before executing the Ralph Loop to ensure the implementation has a solid
// foundation.
package loop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// QualityGateConfig contains configuration for quality gate scoring.
type QualityGateConfig struct {
	// Threshold is the minimum score (0.0-1.0) required to pass the quality gate.
	Threshold float64

	// AgentsPath is the path to the AGENTS.md file (optional).
	AgentsPath string

	// EnableAutoEscalation invokes grove-spec when score is below threshold.
	EnableAutoEscalation bool
}

// DefaultQualityGateConfig returns the default configuration.
func DefaultQualityGateConfig() *QualityGateConfig {
	return &QualityGateConfig{
		Threshold:            0.7,
		EnableAutoEscalation: true,
	}
}

// QualityGateResult contains the result of quality gate evaluation.
type QualityGateResult struct {
	Passed           bool     `json:"passed"`
	Score            float64  `json:"score"`
	Threshold        float64  `json:"threshold"`
	MissingFiles     []string `json:"missing_files"`
	WeakDimensions   []string `json:"weak_dimensions"`
	Recommendations  []string `json:"recommendations"`
	ShouldEscalate   bool     `json:"should_escalate"`
	EscalationReason string   `json:"escalation_reason,omitempty"`
}

// DimensionScore represents the score for a single quality dimension.
type DimensionScore struct {
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	MaxScore    float64 `json:"max_score"`
	Description string  `json:"description"`
}

// QualityScorer evaluates documentation quality.
type QualityScorer struct {
	config *QualityGateConfig
}

// NewQualityScorer creates a new QualityScorer.
func NewQualityScorer(config *QualityGateConfig) *QualityScorer {
	if config == nil {
		config = DefaultQualityGateConfig()
	}
	return &QualityScorer{
		config: config,
	}
}

// ScoreDocumentation evaluates the quality of documentation at the given path.
// It checks for completeness, coherence, and format of SPEC.md, DESIGN.md, TASKS.md, and AGENTS.md.
func (q *QualityScorer) ScoreDocumentation(docsPath string) (*QualityGateResult, error) {
	result := &QualityGateResult{
		Passed:    true,
		Score:     0.0,
		Threshold: q.config.Threshold,
	}

	// Check required files
	requiredFiles := []string{
		"SPEC.md",
		"DESIGN.md",
		"TASKS.md",
	}

	var foundFiles []string
	var missingFiles []string

	for _, file := range requiredFiles {
		path := filepath.Join(docsPath, file)
		if _, err := os.Stat(path); err == nil {
			foundFiles = append(foundFiles, file)
		} else {
			missingFiles = append(missingFiles, file)
		}
	}

	result.MissingFiles = missingFiles

	// If AGENTS.md is provided, check it too
	if q.config.AgentsPath != "" {
		if _, err := os.Stat(q.config.AgentsPath); err == nil {
			foundFiles = append(foundFiles, "AGENTS.md")
		}
	}

	// Score each dimension
	dimensionScores := q.scoreDimensions(docsPath, foundFiles)

	// Calculate overall score
	var totalScore, maxScore float64
	for _, dim := range dimensionScores {
		totalScore += dim.Score
		maxScore += dim.MaxScore
	}

	if maxScore > 0 {
		result.Score = totalScore / maxScore
	}

	// Check if passed threshold
	if result.Score < q.config.Threshold {
		result.Passed = false
		result.ShouldEscalate = q.config.EnableAutoEscalation
		result.EscalationReason = fmt.Sprintf("Documentation score (%.2f) is below threshold (%.2f)", result.Score, q.config.Threshold)
	}

	// Identify weak dimensions
	for _, dim := range dimensionScores {
		if dim.Score < dim.MaxScore*0.6 {
			result.WeakDimensions = append(result.WeakDimensions, dim.Name)
		}
	}

	// Generate recommendations
	result.Recommendations = q.generateRecommendations(result, dimensionScores)

	return result, nil
}

// scoreDimensions evaluates each quality dimension.
func (q *QualityScorer) scoreDimensions(docsPath string, foundFiles []string) []DimensionScore {
	var scores []DimensionScore

	// 1. Completeness (40% of total)
	completenessScore := float64(len(foundFiles)) / 4.0 // 4 files: SPEC, DESIGN, TASKS, AGENTS
	if completenessScore > 1.0 {
		completenessScore = 1.0
	}
	scores = append(scores, DimensionScore{
		Name:        "completeness",
		Score:       completenessScore * 100,
		MaxScore:    100,
		Description: fmt.Sprintf("Found %d of 4 required files", len(foundFiles)),
	})

	// 2. SPEC.md quality
	specScore := q.scoreSpecFile(filepath.Join(docsPath, "SPEC.md"))
	scores = append(scores, DimensionScore{
		Name:        "spec_quality",
		Score:       specScore,
		MaxScore:    100,
		Description: "SPEC.md structure and content quality",
	})

	// 3. DESIGN.md quality
	designScore := q.scoreDesignFile(filepath.Join(docsPath, "DESIGN.md"))
	scores = append(scores, DimensionScore{
		Name:        "design_quality",
		Score:       designScore,
		MaxScore:    100,
		Description: "DESIGN.md architecture and decisions",
	})

	// 4. TASKS.md quality
	tasksScore := q.scoreTasksFile(filepath.Join(docsPath, "TASKS.md"))
	scores = append(scores, DimensionScore{
		Name:        "tasks_quality",
		Score:       tasksScore,
		MaxScore:    100,
		Description: "TASKS.md task breakdown quality",
	})

	return scores
}

// scoreSpecFile evaluates the quality of SPEC.md.
func (q *QualityScorer) scoreSpecFile(path string) float64 {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0.0
	}

	score := 0.0

	// Check for required sections (20 points each)
	contentStr := string(content)
	sections := map[string]bool{
		"overview":       false,
		"requirements":   false,
		"user stories":   false,
		"acceptance":     false,
		"non-functional": false,
	}

	for section := range sections {
		if strings.Contains(strings.ToLower(contentStr), section) {
			score += 20
		}
	}

	// Check minimum length (10 points)
	if len(contentStr) > 500 {
		score += 10
	}

	return score
}

// scoreDesignFile evaluates the quality of DESIGN.md.
func (q *QualityScorer) scoreDesignFile(path string) float64 {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0.0
	}

	score := 0.0
	contentStr := string(content)

	// Check for required sections (25 points each)
	sections := map[string]bool{
		"architecture": false,
		"data model":   false,
		"api":          false,
		"components":   false,
	}

	for section := range sections {
		if strings.Contains(strings.ToLower(contentStr), section) {
			score += 25
		}
	}

	// Check minimum length
	if len(contentStr) > 300 {
		score += 10
	}

	return score
}

// scoreTasksFile evaluates the quality of TASKS.md.
func (q *QualityScorer) scoreTasksFile(path string) float64 {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0.0
	}

	score := 0.0
	contentStr := string(content)

	// Check for task format (checkboxes)
	if strings.Contains(contentStr, "- [ ]") || strings.Contains(contentStr, "- [x]") {
		score += 40
	}

	// Check for phases/groups
	if strings.Contains(strings.ToLower(contentStr), "phase") {
		score += 30
	}

	// Check minimum task count
	taskCount := strings.Count(contentStr, "- [")
	if taskCount >= 3 {
		score += 30
	}

	return score
}

// generateRecommendations generates recommendations based on the evaluation.
func (q *QualityScorer) generateRecommendations(result *QualityGateResult, dimensions []DimensionScore) []string {
	var recommendations []string

	if len(result.MissingFiles) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Missing required files: %s", strings.Join(result.MissingFiles, ", ")))
	}

	for _, dim := range dimensions {
		if dim.Score < dim.MaxScore*0.5 {
			recommendations = append(recommendations, fmt.Sprintf("Improve %s: %s", dim.Name, dim.Description))
		}
	}

	if result.ShouldEscalate {
		recommendations = append(recommendations, "Consider running grove-spec to improve documentation quality")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Documentation is ready for implementation")
	}

	return recommendations
}

// ScoreDocumentation is a convenience function that creates a scorer and evaluates documentation.
func ScoreDocumentation(specPath, agentsPath string) (*QualityGateResult, error) {
	config := &QualityGateConfig{
		Threshold:            0.7,
		AgentsPath:           agentsPath,
		EnableAutoEscalation: true,
	}

	scorer := NewQualityScorer(config)
	return scorer.ScoreDocumentation(specPath)
}
