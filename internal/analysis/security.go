package analysis

// SecurityAnalyzer analyzes security of an architecture.
type SecurityAnalyzer struct{}

// SecurityReport represents a security analysis report.
type SecurityReport struct {
	Critical        []SecurityIssue     `json:"critical"`
	Warnings        []SecurityIssue     `json:"warnings"`
	Good            []SecurityGood      `json:"good"`
	Checklist       []SecurityCheckItem `json:"checklist"`
	Recommendations []string            `json:"recommendations"`
}

// SecurityIssue represents a security issue.
type SecurityIssue struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Fix         string `json:"fix"`
}

// SecurityGood represents something done well.
type SecurityGood struct {
	Category    string `json:"category"`
	Description string `json:"description"`
}

// SecurityCheckItem represents a checklist item.
type SecurityCheckItem struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Status      bool   `json:"status"`
}

// NewSecurityAnalyzer creates a new security analyzer.
func NewSecurityAnalyzer() *SecurityAnalyzer {
	return &SecurityAnalyzer{}
}

// Analyze performs security analysis.
func (sa *SecurityAnalyzer) Analyze(components []string) *SecurityReport {
	report := &SecurityReport{
		Critical:        make([]SecurityIssue, 0),
		Warnings:        make([]SecurityIssue, 0),
		Good:            make([]SecurityGood, 0),
		Checklist:       make([]SecurityCheckItem, 0),
		Recommendations: make([]string, 0),
	}

	// Check for common issues
	for _, comp := range components {
		lower := toLower(comp)

		// Check for authentication
		if contains(lower, "auth") || contains(lower, "login") {
			report.Good = append(report.Good, SecurityGood{
				Category:    "authentication",
				Description: "Authentication component detected",
			})
		}

		// Check for API endpoints
		if contains(lower, "api") || contains(lower, "endpoint") {
			report.Warnings = append(report.Warnings, SecurityIssue{
				Category:    "api_security",
				Description: "API endpoints detected",
				Fix:         "Add rate limiting and input validation",
			})
		}

		// Check for database
		if contains(lower, "database") || contains(lower, "db") {
			report.Warnings = append(report.Warnings, SecurityIssue{
				Category:    "data_protection",
				Description: "Database access detected",
				Fix:         "Use parameterized queries, encrypt sensitive data",
			})
		}
	}

	// Generate checklist
	report.Checklist = []SecurityCheckItem{
		{Category: "Authentication", Description: "User authentication implemented", Status: false},
		{Category: "Authorization", Description: "Role-based access control", Status: false},
		{Category: "Input Validation", Description: "All inputs validated", Status: false},
		{Category: "HTTPS", Description: "All traffic over HTTPS", Status: false},
		{Category: "Rate Limiting", Description: "API rate limiting enabled", Status: false},
		{Category: "Logging", Description: "Security events logged", Status: false},
	}

	// Generate recommendations
	report.Recommendations = []string{
		"Implement authentication for all protected routes",
		"Add input validation for all user inputs",
		"Use HTTPS everywhere",
		"Implement rate limiting for API endpoints",
		"Log security-relevant events",
	}

	return report
}

// Helper functions
func toLower(s string) string {
	result := ""
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			result += string(c + 32)
		} else {
			result += string(c)
		}
	}
	return result
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
