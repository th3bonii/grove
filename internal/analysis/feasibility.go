package analysis

// FeasibilityAnalyzer analyzes technical feasibility.
type FeasibilityAnalyzer struct{}

// FeasibilityReport represents a feasibility analysis.
type FeasibilityReport struct {
	IsViable        bool     `json:"is_viable"`
	Complexity      string   `json:"complexity"` // low, medium, high, very-high
	TimeEstimate    string   `json:"time_estimate"`
	TeamSize        int      `json:"team_size"`
	BudgetEstimate  string   `json:"budget_estimate"`
	RiskLevel       string   `json:"risk_level"` // low, medium, high
	Requirements    []string `json:"requirements"`
	SuccessFactors  []string `json:"success_factors"`
	Blockers        []string `json:"blockers"`
	Recommendations []string `json:"recommendations"`
}

// NewFeasibilityAnalyzer creates a new feasibility analyzer.
func NewFeasibilityAnalyzer() *FeasibilityAnalyzer {
	return &FeasibilityAnalyzer{}
}

// Analyze performs feasibility analysis.
func (fa *FeasibilityAnalyzer) Analyze(components []string, techStack []string) *FeasibilityReport {
	report := &FeasibilityReport{
		Requirements:    make([]string, 0),
		SuccessFactors:  make([]string, 0),
		Blockers:        make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// Analyze complexity based on components
	numComponents := len(components)
	if numComponents <= 5 {
		report.Complexity = "low"
		report.TimeEstimate = "1-2 weeks"
		report.TeamSize = 1
	} else if numComponents <= 15 {
		report.Complexity = "medium"
		report.TimeEstimate = "3-4 weeks"
		report.TeamSize = 2
	} else if numComponents <= 30 {
		report.Complexity = "high"
		report.TimeEstimate = "6-8 weeks"
		report.TeamSize = 3
	} else {
		report.Complexity = "very-high"
		report.TimeEstimate = "3+ months"
		report.TeamSize = 5
	}

	// Determine viability
	report.IsViable = true
	if numComponents > 50 {
		report.IsViable = false
		report.Blockers = append(report.Blockers, "Too many components for single developer")
	}

	// Risk level
	switch report.Complexity {
	case "low":
		report.RiskLevel = "low"
	case "medium":
		report.RiskLevel = "medium"
	case "high":
		report.RiskLevel = "medium"
	default:
		report.RiskLevel = "high"
	}

	// Budget estimate
	switch report.Complexity {
	case "low":
		report.BudgetEstimate = "$1,000-3,000"
	case "medium":
		report.BudgetEstimate = "$3,000-10,000"
	case "high":
		report.BudgetEstimate = "$10,000-30,000"
	default:
		report.BudgetEstimate = "$30,000+"
	}

	// Generate requirements
	for _, comp := range components {
		report.Requirements = append(report.Requirements, "Implement "+comp)
	}

	// Generate success factors
	report.SuccessFactors = []string{
		"Clear requirements",
		"Defined tech stack",
		"Active development",
		"Regular testing",
	}

	// Generate recommendations
	report.Recommendations = []string{
		"Start with MVP",
		"Use existing libraries when possible",
		"Write tests from the beginning",
		"Deploy early and often",
	}

	return report
}
