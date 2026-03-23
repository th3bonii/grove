// Package analysis provides test pattern generation capabilities for Grove.
// It analyzes specs (SPEC.md, DESIGN.md, TASKS.md) and generates test patterns
// without implementing actual test code - that responsibility belongs to SDD.
//
// This module is designed to work alongside SDD (Spec-Driven Development):
// - GROVE Spec generates: test patterns, skeletons, criteria, structure (what to test)
// - SDD implements: real tests, execution, coverage reports (how to test)
package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TestPattern represents a single test pattern with its metadata
type TestPattern struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Component      string   `json:"component"`
	Preconditions  []string `json:"preconditions"`
	Steps          []string `json:"steps"`
	ExpectedResult string   `json:"expected_result"`
	TestType       string   `json:"test_type"` // unit, integration, e2e
	Priority       string   `json:"priority"`  // high, medium, low
	Tags           []string `json:"tags"`
}

// AcceptanceCriteria represents acceptance criteria derived from specs
type AcceptanceCriteria struct {
	ID          string `json:"id"`
	Criteria    string `json:"criteria"`
	Source      string `json:"source"` // which spec file
	Component   string `json:"component"`
	TestPattern string `json:"test_pattern"`
	VerifiedBy  string `json:"verified_by"` // test name that verifies this
}

// TestPatterns is the main output structure containing all generated patterns
type TestPatterns struct {
	UnitTestPatterns        []TestPattern        `json:"unit_test_patterns"`
	IntegrationTestPatterns []TestPattern        `json:"integration_test_patterns"`
	E2ETestPatterns         []TestPattern        `json:"e2e_test_patterns"`
	AcceptanceCriteria      []AcceptanceCriteria `json:"acceptance_criteria"`
	Metadata                TestMetadata         `json:"metadata"`
}

// TestMetadata contains information about the analysis
type TestMetadata struct {
	SpecFilesAnalyzed []string `json:"spec_files_analyzed"`
	ComponentsFound   []string `json:"components_found"`
	GeneratedAt       string   `json:"generated_at"`
	GoVersion         string   `json:"go_version"`
}

// SpecReader reads and parses specification files
type SpecReader struct {
	specPath string
}

// NewSpecReader creates a new SpecReader for the given specification path
func NewSpecReader(specPath string) *SpecReader {
	return &SpecReader{specPath: specPath}
}

// SpecContent holds parsed content from specification files
type SpecContent struct {
	SpecFile     string
	Content      string
	Components   []string
	Features     []string
	Requirements []string
}

// ReadSpecFiles reads all specification files from the given directory
func (sr *SpecReader) ReadSpecFiles() (map[string]*SpecContent, error) {
	specFiles := map[string]*SpecContent{
		"SPEC.md":   nil,
		"DESIGN.md": nil,
		"TASKS.md":  nil,
	}

	for filename := range specFiles {
		path := filepath.Join(sr.specPath, filename)
		content, err := os.ReadFile(path)
		if err != nil {
			// Skip if file doesn't exist
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading %s: %w", filename, err)
		}

		specFiles[filename] = sr.parseSpecFile(filename, string(content))
	}

	return specFiles, nil
}

// parseSpecFile parses a specification file and extracts components and features
func (sr *SpecReader) parseSpecFile(filename, content string) *SpecContent {
	sc := &SpecContent{
		SpecFile: filename,
		Content:  content,
	}

	// Extract components (capitalized words, often in headers)
	componentPattern := regexp.MustCompile(`(?m)^##?\s+(?:Component|Module|Service|Package):\s*(\w+)`)
	matches := componentPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			sc.Components = append(sc.Components, match[1])
		}
	}

	// Extract features (words after "Feature:" or "Functionality:")
	featurePattern := regexp.MustCompile(`(?mi)^[-*]\s*(?:Feature|Functionality):\s*(.+)$`)
	matches = featurePattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			sc.Features = append(sc.Features, strings.TrimSpace(match[1]))
		}
	}

	// Extract requirements (lines starting with "REQ" or "Requirement")
	reqPattern := regexp.MustCompile(`(?mi)^[-*]?\s*(?:REQ|Requirement)\s*[:\-]?\s*(.+)$`)
	matches = reqPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			sc.Requirements = append(sc.Requirements, strings.TrimSpace(match[1]))
		}
	}

	return sc
}

// TestPatternGenerator generates test patterns from specification content
type TestPatternGenerator struct {
	specs map[string]*SpecContent
}

// NewTestPatternGenerator creates a new generator with the parsed specs
func NewTestPatternGenerator(specs map[string]*SpecContent) *TestPatternGenerator {
	return &TestPatternGenerator{specs: specs}
}

// GenerateTestPatterns generates all test patterns from the specs
func (tpg *TestPatternGenerator) GenerateTestPatterns() *TestPatterns {
	tp := &TestPatterns{
		UnitTestPatterns:        make([]TestPattern, 0),
		IntegrationTestPatterns: make([]TestPattern, 0),
		E2ETestPatterns:         make([]TestPattern, 0),
		AcceptanceCriteria:      make([]AcceptanceCriteria, 0),
		Metadata: TestMetadata{
			SpecFilesAnalyzed: make([]string, 0),
			ComponentsFound:   make([]string, 0),
			GeneratedAt:       "2026-03-23",
			GoVersion:         "1.23",
		},
	}

	// Collect all components and specs analyzed
	componentSet := make(map[string]bool)
	for filename, spec := range tpg.specs {
		if spec != nil {
			tp.Metadata.SpecFilesAnalyzed = append(tp.Metadata.SpecFilesAnalyzed, filename)
			for _, comp := range spec.Components {
				if !componentSet[comp] {
					componentSet[comp] = true
					tp.Metadata.ComponentsFound = append(tp.Metadata.ComponentsFound, comp)
				}
			}
		}
	}

	// Generate patterns for each spec
	for filename, spec := range tpg.specs {
		if spec == nil {
			continue
		}

		// Generate unit test patterns
		for _, comp := range spec.Components {
			unitPattern := tpg.generateUnitTestPattern(comp, spec)
			tp.UnitTestPatterns = append(tp.UnitTestPatterns, unitPattern)
		}

		// Generate integration test patterns
		for _, feature := range spec.Features {
			intPattern := tpg.generateIntegrationTestPattern(feature, spec)
			tp.IntegrationTestPatterns = append(tp.IntegrationTestPatterns, intPattern)
		}

		// Generate E2E test patterns
		for _, req := range spec.Requirements {
			e2ePattern := tpg.generateE2ETestPattern(req, spec)
			tp.E2ETestPatterns = append(tp.E2ETestPatterns, e2ePattern)
		}

		// Generate acceptance criteria
		for i, req := range spec.Requirements {
			criteria := tpg.generateAcceptanceCriteria(req, filename, i)
			tp.AcceptanceCriteria = append(tp.AcceptanceCriteria, criteria)
		}
	}

	return tp
}

// generateUnitTestPattern creates a unit test pattern for a component
func (tpg *TestPatternGenerator) generateUnitTestPattern(component string, spec *SpecContent) TestPattern {
	preconditions := []string{
		fmt.Sprintf("Component '%s' is initialized", component),
		"Dependencies are mocked or stubbed",
		"Test environment is configured",
	}

	steps := []string{
		"1. Set up test fixtures for the component",
		"2. Prepare mock dependencies",
		"3. Invoke the method/function under test",
		"4. Assert expected behavior",
		"5. Verify no side effects",
	}

	return TestPattern{
		Name:           fmt.Sprintf("Test%sUnit", component),
		Description:    fmt.Sprintf("Unit tests for %s component functionality", component),
		Component:      component,
		Preconditions:  preconditions,
		Steps:          steps,
		ExpectedResult: "All assertions pass, component behaves as specified",
		TestType:       "unit",
		Priority:       "high",
		Tags:           []string{component, "unit", "core"},
	}
}

// generateIntegrationTestPattern creates an integration test pattern for a feature
func (tpg *TestPatternGenerator) generateIntegrationTestPattern(feature string, spec *SpecContent) TestPattern {
	preconditions := []string{
		fmt.Sprintf("Feature '%s' is implemented", feature),
		"Database/services are available",
		"Integration test environment is configured",
	}

	steps := []string{
		"1. Set up integration test environment",
		"2. Initialize required services",
		"3. Execute feature workflow",
		"4. Verify data consistency across services",
		"5. Clean up test resources",
	}

	return TestPattern{
		Name:           fmt.Sprintf("Test%sIntegration", normalizeName(feature)),
		Description:    fmt.Sprintf("Integration tests for %s feature", feature),
		Component:      extractComponent(feature),
		Preconditions:  preconditions,
		Steps:          steps,
		ExpectedResult: "All integration points work correctly, data flows properly",
		TestType:       "integration",
		Priority:       "high",
		Tags:           []string{feature, "integration", "workflow"},
	}
}

// generateE2ETestPattern creates an E2E test pattern for a requirement
func (tpg *TestPatternGenerator) generateE2ETestPattern(requirement string, spec *SpecContent) TestPattern {
	preconditions := []string{
		"Application is fully deployed",
		"All services are running",
		"Test user accounts are available",
	}

	steps := []string{
		"1. Navigate to starting point in application",
		"2. Perform user actions as specified in requirement",
		"3. Verify end-to-end flow completion",
		"4. Check final state matches expected outcome",
		"5. Verify no regressions in other features",
	}

	return TestPattern{
		Name:           fmt.Sprintf("Test%sE2E", normalizeName(requirement)),
		Description:    fmt.Sprintf("End-to-end test for: %s", requirement),
		Component:      "full-system",
		Preconditions:  preconditions,
		Steps:          steps,
		ExpectedResult: "Complete user workflow succeeds as specified",
		TestType:       "e2e",
		Priority:       "medium",
		Tags:           []string{"e2e", "user-flow", "regression"},
	}
}

// generateAcceptanceCriteria creates acceptance criteria from a requirement
func (tpg *TestPatternGenerator) generateAcceptanceCriteria(requirement, source string, index int) AcceptanceCriteria {
	return AcceptanceCriteria{
		ID:          fmt.Sprintf("AC-%s-%d", strings.ToUpper(normalizeName(requirement)), index+1),
		Criteria:    requirement,
		Source:      source,
		Component:   extractComponent(requirement),
		TestPattern: fmt.Sprintf("Test%sE2E", normalizeName(requirement)),
		VerifiedBy:  "Manual + automated E2E tests",
	}
}

// normalizeName converts a string to a valid Go identifier
func normalizeName(s string) string {
	// Remove special characters and capitalize words
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	s = re.ReplaceAllString(s, "")
	words := strings.Fields(s)
	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}
	return strings.Join(words, "")
}

// extractComponent extracts the component name from a feature or requirement
func extractComponent(s string) string {
	// Common patterns: "Feature X for Y", "Y component", etc.
	patterns := []string{
		`(?:feature|functionality)\s+(\w+)`,
		`(\w+)\s+(?:component|module|service)`,
		`(\w+)\s+(?:function|method|handler)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(strings.ToLower(s))
		if len(matches) > 1 {
			return strings.Title(matches[1])
		}
	}

	// Default to first word
	words := strings.Fields(s)
	if len(words) > 0 {
		return strings.Title(words[0])
	}

	return "Unknown"
}

// GenerateTestPatternsFromPath generates test patterns from a specification directory
func GenerateTestPatternsFromPath(specPath string) (*TestPatterns, error) {
	reader := NewSpecReader(specPath)
	specs, err := reader.ReadSpecFiles()
	if err != nil {
		return nil, fmt.Errorf("reading spec files: %w", err)
	}

	generator := NewTestPatternGenerator(specs)
	return generator.GenerateTestPatterns(), nil
}

// ValidateTestPatterns validates the generated test patterns
func ValidateTestPatterns(tp *TestPatterns) []string {
	errors := make([]string, 0)

	// Validate unit tests
	for i, test := range tp.UnitTestPatterns {
		if test.Name == "" {
			errors = append(errors, fmt.Sprintf("unit test %d: missing name", i))
		}
		if len(test.Steps) == 0 {
			errors = append(errors, fmt.Sprintf("unit test %s: missing steps", test.Name))
		}
	}

	// Validate integration tests
	for i, test := range tp.IntegrationTestPatterns {
		if test.Name == "" {
			errors = append(errors, fmt.Sprintf("integration test %d: missing name", i))
		}
		if len(test.Steps) == 0 {
			errors = append(errors, fmt.Sprintf("integration test %s: missing steps", test.Name))
		}
	}

	// Validate E2E tests
	for i, test := range tp.E2ETestPatterns {
		if test.Name == "" {
			errors = append(errors, fmt.Sprintf("e2e test %d: missing name", i))
		}
		if len(test.Steps) == 0 {
			errors = append(errors, fmt.Sprintf("e2e test %s: missing steps", test.Name))
		}
	}

	// Validate acceptance criteria
	for i, ac := range tp.AcceptanceCriteria {
		if ac.ID == "" {
			errors = append(errors, fmt.Sprintf("acceptance criteria %d: missing ID", i))
		}
		if ac.Criteria == "" {
			errors = append(errors, fmt.Sprintf("acceptance criteria %s: missing criteria", ac.ID))
		}
	}

	return errors
}

// GetTestSummary returns a human-readable summary of test patterns
func GetTestSummary(tp *TestPatterns) string {
	var sb strings.Builder

	sb.WriteString("=== Test Pattern Summary ===\n\n")
	sb.WriteString(fmt.Sprintf("Unit Tests: %d\n", len(tp.UnitTestPatterns)))
	sb.WriteString(fmt.Sprintf("Integration Tests: %d\n", len(tp.IntegrationTestPatterns)))
	sb.WriteString(fmt.Sprintf("E2E Tests: %d\n", len(tp.E2ETestPatterns)))
	sb.WriteString(fmt.Sprintf("Acceptance Criteria: %d\n\n", len(tp.AcceptanceCriteria)))

	sb.WriteString("Components Found:\n")
	for _, comp := range tp.Metadata.ComponentsFound {
		sb.WriteString(fmt.Sprintf("  - %s\n", comp))
	}

	sb.WriteString("\nSpec Files Analyzed:\n")
	for _, file := range tp.Metadata.SpecFilesAnalyzed {
		sb.WriteString(fmt.Sprintf("  - %s\n", file))
	}

	return sb.String()
}
