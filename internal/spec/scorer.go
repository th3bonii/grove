package spec

import (
	"fmt"
	"strings"

	"github.com/Gentleman-Programming/grove/internal/types"
)

// Scorer evaluates content quality across 7 dimensions.
type Scorer struct {
	config *types.Config
}

// NewScorer creates a new Scorer instance.
func NewScorer(config *types.Config) *Scorer {
	return &Scorer{
		config: config,
	}
}

// ScoreContent evaluates content quality and returns a Score.
func (s *Scorer) ScoreContent(content string) *types.Score {
	if content == "" {
		return &types.Score{
			Overall:    0,
			Dimensions: make(map[string]float64),
			Breakdown:  make([]types.ScoreDimension, 0),
		}
	}

	// Calculate scores for each dimension
	dimensions := s.calculateDimensions(content)

	// Calculate weighted overall score
	overall := s.calculateWeightedOverall(dimensions)

	// Generate recommendations
	recommendations := s.generateRecommendations(dimensions)

	// Build dimensions map
	dimMap := make(map[string]float64)
	for _, d := range dimensions {
		dimMap[d.Name] = d.Score
	}

	return &types.Score{
		Overall:         overall,
		Dimensions:      dimMap,
		Breakdown:       dimensions,
		Recommendations: recommendations,
	}
}

// calculateDimensions calculates scores for all 7 dimensions.
func (s *Scorer) calculateDimensions(content string) []types.ScoreDimension {
	wordCount := countWords(content)
	lineCount := countLines(content)
	hasStructure := checkStructure(content)

	return []types.ScoreDimension{
		s.scoreCompleteness(content, wordCount, lineCount, hasStructure),
		s.scoreConsistency(content),
		s.scoreClarity(content, wordCount, lineCount),
		s.scoreTestability(content),
		s.scoreMaintainability(content, lineCount),
		s.scoreFeasibility(content),
		s.scoreTraceability(content, hasStructure),
	}
}

// scoreCompleteness evaluates if all required elements are present.
// Weight: 20%
func (s *Scorer) scoreCompleteness(content string, wordCount, lineCount int, hasStructure bool) types.ScoreDimension {
	maxScore := 10.0
	issues := []string{}
	score := maxScore

	// Check for essential sections
	requiredPatterns := []struct {
		pattern string
		name    string
		weight  float64
	}{
		{"##", "Section headers", 2.0},
		{"Requisit", "Requirements", 2.0},
		{"Escenario", "Scenarios", 1.5},
		{"GIVEN", "BDD format", 1.5},
		{"WHEN", "BDD format", 1.5},
		{"THEN", "BDD format", 1.5},
		{"Archivo", "File references", 1.5},
		{"src/", "Source files", 1.5},
	}

	for _, rp := range requiredPatterns {
		if !containsIgnoreCase(content, rp.pattern) {
			issues = append(issues, "Missing: "+rp.name)
			score -= rp.weight
		}
	}

	// Minimum word count check
	if wordCount < 100 {
		issues = append(issues, "Content too brief (less than 100 words)")
		score -= 1.0
	}

	if score < 0 {
		score = 0
	}

	return types.ScoreDimension{
		Name:     "Completeness",
		Score:    score,
		MaxScore: maxScore,
		Weight:   s.getWeight("completeness", 0.20),
		Details:  summarizeIssues(issues),
		Issues:   issues,
	}
}

// scoreConsistency evaluates consistency in terminology and formatting.
// Weight: 15%
func (s *Scorer) scoreConsistency(content string) types.ScoreDimension {
	maxScore := 10.0
	issues := []string{}
	score := maxScore

	// Check for inconsistent terminology
	inconsistencies := []struct {
		term1 string
		term2 string
	}{
		{"usuario", "cliente"},
		{"cliente", "usuario"},
		{"API", "api"},
		{"URL", "url"},
		{"JSON", "json"},
	}

	for _, inc := range inconsistencies {
		count1 := countOccurrencesIgnoreCase(content, inc.term1)
		count2 := countOccurrencesIgnoreCase(content, inc.term2)

		if count1 > 0 && count2 > 0 {
			// Both terms found - potential inconsistency
			if count1 != count2 {
				issues = append(issues, "Inconsistent: '"+inc.term1+"' vs '"+inc.term2+"'")
				score -= 1.0
			}
		}
	}

	// Check heading level consistency
	h1Count := strings.Count(content, "# ")
	h2Count := strings.Count(content, "## ")
	h3Count := strings.Count(content, "### ")

	if h1Count > 1 {
		issues = append(issues, "Multiple H1 headings found")
		score -= 0.5
	}

	if h3Count > 0 && h2Count == 0 {
		issues = append(issues, "H3 without H2 parent headings")
		score -= 0.5
	}

	if score < 0 {
		score = 0
	}

	return types.ScoreDimension{
		Name:     "Consistency",
		Score:    score,
		MaxScore: maxScore,
		Weight:   s.getWeight("consistency", 0.15),
		Details:  summarizeIssues(issues),
		Issues:   issues,
	}
}

// scoreClarity evaluates how clear and understandable the content is.
// Weight: 15%
func (s *Scorer) scoreClarity(content string, wordCount, lineCount int) types.ScoreDimension {
	maxScore := 10.0
	issues := []string{}
	score := maxScore

	// Check for very long lines
	longLines := 0
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if len(line) > 150 {
			longLines++
		}
	}

	if lineCount > 0 && longLines > lineCount/3 {
		issues = append(issues, "Many very long lines (wrap recommended)")
		score -= 1.0
	}

	// Check for passive voice indicators
	passiveCount := countOccurrencesIgnoreCase(content, "sera") +
		countOccurrencesIgnoreCase(content, "fueron") +
		countOccurrencesIgnoreCase(content, "ha sido")

	if passiveCount > wordCount/50 {
		issues = append(issues, "Heavy use of passive voice")
		score -= 0.5
	}

	// Check for vague language
	vagueTerms := []string{"etc", "y otros", "cosas", "algo", "posiblemente"}
	vagueCount := 0
	for _, term := range vagueTerms {
		vagueCount += countOccurrencesIgnoreCase(content, term)
	}

	if vagueCount > 3 {
		issues = append(issues, "Vague language detected")
		score -= 0.5
	}

	// Check for bullet points vs paragraphs
	bulletCount := strings.Count(content, "- ")
	if bulletCount < 3 && wordCount > 200 {
		issues = append(issues, "Consider using more bullet points")
		score -= 0.5
	}

	if score < 0 {
		score = 0
	}

	return types.ScoreDimension{
		Name:     "Clarity",
		Score:    score,
		MaxScore: maxScore,
		Weight:   s.getWeight("clarity", 0.15),
		Details:  summarizeIssues(issues),
		Issues:   issues,
	}
}

// scoreTestability evaluates how testable the specifications are.
// Weight: 15%
func (s *Scorer) scoreTestability(content string) types.ScoreDimension {
	maxScore := 10.0
	issues := []string{}
	score := maxScore

	// Check for scenario/case coverage
	scenarioCount := countOccurrencesIgnoreCase(content, "escenario") +
		countOccurrencesIgnoreCase(content, "scenario") +
		countOccurrencesIgnoreCase(content, "caso de prueba")

	if scenarioCount < 2 {
		issues = append(issues, "Few test scenarios defined")
		score -= 2.0
	}

	// Check for GIVEN-WHEN-THEN format
	givenCount := countOccurrencesIgnoreCase(content, "given")
	whenCount := countOccurrencesIgnoreCase(content, "when")
	thenCount := countOccurrencesIgnoreCase(content, "then")

	if givenCount > 0 || whenCount > 0 || thenCount > 0 {
		// BDD format used
		if givenCount == 0 || whenCount == 0 || thenCount == 0 {
			issues = append(issues, "Incomplete BDD format")
			score -= 1.0
		}
	}

	// Check for edge case mentions
	edgeCaseTerms := []string{"error", "fallo", "timeout", "vacio", "null", "limite", "nulo"}
	hasEdgeCases := false
	for _, term := range edgeCaseTerms {
		if containsIgnoreCase(content, term) {
			hasEdgeCases = true
			break
		}
	}

	if !hasEdgeCases {
		issues = append(issues, "No edge cases mentioned")
		score -= 1.0
	}

	// Check for measurable criteria
	measurableTerms := []string{"menor que", "mayor que", "igual a", "maximo", "minimo", "exactamente", "al menos", "no mas de"}
	hasMetrics := false
	for _, term := range measurableTerms {
		if containsIgnoreCase(content, term) {
			hasMetrics = true
			break
		}
	}

	if !hasMetrics {
		issues = append(issues, "No measurable criteria found")
		score -= 1.0
	}

	if score < 0 {
		score = 0
	}

	return types.ScoreDimension{
		Name:     "Testability",
		Score:    score,
		MaxScore: maxScore,
		Weight:   s.getWeight("testability", 0.15),
		Details:  summarizeIssues(issues),
		Issues:   issues,
	}
}

// scoreMaintainability evaluates how maintainable the code/design would be.
// Weight: 15%
func (s *Scorer) scoreMaintainability(content string, lineCount int) types.ScoreDimension {
	maxScore := 10.0
	issues := []string{}
	score := maxScore

	// Check for documentation of interfaces
	interfaceCount := countOccurrencesIgnoreCase(content, "interfaz") +
		countOccurrencesIgnoreCase(content, "interface") +
		countOccurrencesIgnoreCase(content, "API")

	if interfaceCount < 1 {
		issues = append(issues, "No interfaces/APIs documented")
		score -= 1.5
	}

	// Check for dependency documentation
	depTerms := []string{"depende de", "depends on", "requiere", "requires"}
	hasDeps := false
	for _, term := range depTerms {
		if containsIgnoreCase(content, term) {
			hasDeps = true
			break
		}
	}

	if !hasDeps {
		issues = append(issues, "No dependencies documented")
		score -= 1.0
	}

	// Check for versioning strategy
	versionTerms := []string{"version", "v1", "v2", "semver"}
	hasVersioning := false
	for _, term := range versionTerms {
		if containsIgnoreCase(content, term) {
			hasVersioning = true
			break
		}
	}

	if !hasVersioning {
		issues = append(issues, "No versioning strategy")
		score -= 1.0
	}

	// Check for change management
	changeTerms := []string{"rollback", "migracion", "migration", "breaking"}
	hasChangeMgmt := false
	for _, term := range changeTerms {
		if containsIgnoreCase(content, term) {
			hasChangeMgmt = true
			break
		}
	}

	if !hasChangeMgmt {
		issues = append(issues, "No change management strategy")
		score -= 0.5
	}

	if score < 0 {
		score = 0
	}

	return types.ScoreDimension{
		Name:     "Maintainability",
		Score:    score,
		MaxScore: maxScore,
		Weight:   s.getWeight("maintainability", 0.15),
		Details:  summarizeIssues(issues),
		Issues:   issues,
	}
}

// scoreFeasibility evaluates the feasibility of implementation.
// Weight: 10%
func (s *Scorer) scoreFeasibility(content string) types.ScoreDimension {
	maxScore := 10.0
	issues := []string{}
	score := maxScore

	// Check for resource estimates
	estimateTerms := []string{"horas", "hours", "dias", "days", "semanas", "weeks", "story points"}
	hasEstimates := false
	for _, term := range estimateTerms {
		if containsIgnoreCase(content, term) {
			hasEstimates = true
			break
		}
	}

	if !hasEstimates {
		issues = append(issues, "No effort estimates provided")
		score -= 2.0
	}

	// Check for scope clarity
	scopeTerms := []string{"alcance", "scope", "dentro de", "inside", "fuera de", "outside"}
	hasScope := false
	for _, term := range scopeTerms {
		if containsIgnoreCase(content, term) {
			hasScope = true
			break
		}
	}

	if !hasScope {
		issues = append(issues, "Scope not clearly defined")
		score -= 1.5
	}

	// Check for risk identification
	riskTerms := []string{"riesgo", "risk", "blocker", "dependencia critica"}
	hasRisks := false
	for _, term := range riskTerms {
		if containsIgnoreCase(content, term) {
			hasRisks = true
			break
		}
	}

	if !hasRisks {
		issues = append(issues, "No risks identified")
		score -= 1.0
	}

	// Check for acceptance criteria
	acceptanceTerms := []string{"criterio de aceptacion", "acceptance criteria", "definition of done"}
	hasAcceptance := false
	for _, term := range acceptanceTerms {
		if containsIgnoreCase(content, term) {
			hasAcceptance = true
			break
		}
	}

	if !hasAcceptance {
		issues = append(issues, "No acceptance criteria")
		score -= 1.5
	}

	if score < 0 {
		score = 0
	}

	return types.ScoreDimension{
		Name:     "Feasibility",
		Score:    score,
		MaxScore: maxScore,
		Weight:   s.getWeight("feasibility", 0.10),
		Details:  summarizeIssues(issues),
		Issues:   issues,
	}
}

// scoreTraceability evaluates traceability between requirements and tasks.
// Weight: 10%
func (s *Scorer) scoreTraceability(content string, hasStructure bool) types.ScoreDimension {
	maxScore := 10.0
	issues := []string{}
	score := maxScore

	// Check for requirement IDs
	reqIdPattern := []string{"REQ-", "R-", "requisito-", "requirement-"}
	hasReqIds := false
	for _, pattern := range reqIdPattern {
		if containsIgnoreCase(content, pattern) {
			hasReqIds = true
			break
		}
	}

	if !hasReqIds {
		issues = append(issues, "No requirement IDs found")
		score -= 2.0
	}

	// Check for task IDs
	taskIdPattern := []string{"TASK-", "T-", "tarea-", "task-"}
	hasTaskIds := false
	for _, pattern := range taskIdPattern {
		if containsIgnoreCase(content, pattern) {
			hasTaskIds = true
			break
		}
	}

	if !hasTaskIds {
		issues = append(issues, "No task IDs found")
		score -= 1.5
	}

	// Check for links/references between sections
	linkTerms := []string{"ver tambien", "see also", "ref:", "relacionado con", "related to"}
	hasLinks := false
	for _, term := range linkTerms {
		if containsIgnoreCase(content, term) {
			hasLinks = true
			break
		}
	}

	if !hasLinks {
		issues = append(issues, "No cross-references between sections")
		score -= 1.5
	}

	// Check for priority assignment
	priorityTerms := []string{"critico", "critical", "high", "medium", "low", "prioridad"}
	hasPriority := false
	for _, term := range priorityTerms {
		if containsIgnoreCase(content, term) {
			hasPriority = true
			break
		}
	}

	if !hasPriority {
		issues = append(issues, "No priorities assigned")
		score -= 1.0
	}

	if score < 0 {
		score = 0
	}

	return types.ScoreDimension{
		Name:     "Traceability",
		Score:    score,
		MaxScore: maxScore,
		Weight:   s.getWeight("traceability", 0.10),
		Details:  summarizeIssues(issues),
		Issues:   issues,
	}
}

// Helper methods

func (s *Scorer) getWeight(name string, defaultWeight float64) float64 {
	if s.config != nil && s.config.ScoringWeights != nil {
		if weight, ok := s.config.ScoringWeights[name]; ok {
			return weight
		}
	}
	return defaultWeight
}

func (s *Scorer) calculateWeightedOverall(dimensions []types.ScoreDimension) float64 {
	var total float64

	for _, dim := range dimensions {
		// Normalize score to 0-1 range
		normalizedScore := dim.Score / dim.MaxScore
		total += normalizedScore * dim.Weight
	}

	// Convert back to 0-10 scale
	return total * 10
}

func (s *Scorer) generateRecommendations(dimensions []types.ScoreDimension) []string {
	var recommendations []string

	for _, dim := range dimensions {
		ratio := dim.Score / dim.MaxScore
		if ratio < 0.7 {
			recommendations = append(recommendations,
				fmt.Sprintf("Improve %s: %s", dim.Name, dim.Details))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Content quality is good")
	}

	return recommendations
}

// Utility functions

func countWords(content string) int {
	words := strings.Fields(content)
	return len(words)
}

func countLines(content string) int {
	if content == "" {
		return 0
	}
	return strings.Count(content, "\n") + 1
}

func checkStructure(content string) bool {
	hasHeaders := strings.Contains(content, "##")
	hasLists := strings.Contains(content, "- ")
	return hasHeaders && hasLists
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func countOccurrences(s, substr string) int {
	return strings.Count(s, substr)
}

func countOccurrencesIgnoreCase(s, substr string) int {
	return strings.Count(strings.ToLower(s), strings.ToLower(substr))
}

func summarizeIssues(issues []string) string {
	if len(issues) == 0 {
		return "No issues found"
	}
	if len(issues) <= 2 {
		return strings.Join(issues, "; ")
	}
	return fmt.Sprintf("%d issues: %s...", len(issues), strings.Join(issues[:2], ", "))
}
