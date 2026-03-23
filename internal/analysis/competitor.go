// Package analysis provides competitive analysis capabilities for Grove projects.
package analysis

import (
	"context"
	"strings"
	"time"
)

// CompetitorReport contains the full competitive analysis for a project.
type CompetitorReport struct {
	ProjectName         string            `json:"project_name"`
	GeneratedAt         time.Time         `json:"generated_at"`
	SimilarProjects     []SimilarProject  `json:"similar_projects"`
	MarketTrends        []MarketTrend     `json:"market_trends"`
	YourDifferentiation []Differentiation `json:"your_differentiation"`
	Recommendations     []Recommendation  `json:"recommendations"`
	RiskFactors         []RiskFactor      `json:"risk_factors"`
}

// SimilarProject represents a competitor project in the market.
type SimilarProject struct {
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	URL            string              `json:"url"`
	Stack          []string            `json:"stack"`
	Features       []Feature           `json:"features"`
	Architecture   ArchitecturePattern `json:"architecture"`
	Metrics        ProjectMetrics      `json:"metrics"`
	Strengths      []string            `json:"strengths"`
	Weaknesses     []string            `json:"weaknesses"`
	LastAnalyzedAt time.Time           `json:"last_analyzed_at"`
}

// Feature represents a specific feature implemented by a competitor.
type Feature struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Maturity    Maturity `json:"maturity"`
	Implemented bool     `json:"implemented"`
}

// Maturity represents how mature a feature is in the market.
type Maturity string

const (
	MaturityNascent   Maturity = "nascent"
	MaturityEmerging  Maturity = "emerging"
	MaturityMature    Maturity = "mature"
	MaturitySaturated Maturity = "saturated"
)

// ProjectMetrics contains popularity and engagement metrics.
type ProjectMetrics struct {
	Stars        int       `json:"stars"`
	Forks        int       `json:"forks"`
	Watchers     int       `json:"watchers"`
	OpenIssues   int       `json:"open_issues"`
	Contributors int       `json:"contributors"`
	LastCommit   time.Time `json:"last_commit"`
	LastRelease  time.Time `json:"last_release"`
	License      string    `json:"license"`
}

// ArchitecturePattern describes the architectural approach.
type ArchitecturePattern string

const (
	ArchMonolithic    ArchitecturePattern = "monolithic"
	ArchModular       ArchitecturePattern = "modular"
	ArchMicroservices ArchitecturePattern = "microservices"
	ArchServerless    ArchitecturePattern = "serverless"
	ArchEventDriven   ArchitecturePattern = "event-driven"
	ArchHexagonal     ArchitecturePattern = "hexagonal"
	ArchClean         ArchitecturePattern = "clean"
	ArchOnion         ArchitecturePattern = "onion"
	ArchDDD           ArchitecturePattern = "ddd"
	ArchPlugin        ArchitecturePattern = "plugin"
)

// MarketTrend represents an emerging market trend.
type MarketTrend struct {
	Trend        string    `json:"trend"`
	Description  string    `json:"description"`
	AdoptionRate string    `json:"adoption_rate"`
	FirstSeen    time.Time `json:"first_seen"`
	Competitors  []string  `json:"competitors"`
}

// Differentiation highlights how your project differs from competitors.
type Differentiation struct {
	Advantage string `json:"advantage"`
	Impact    string `json:"impact"`
	Evidence  string `json:"evidence"`
	Priority  int    `json:"priority"` // 1-5, 5 being highest
}

// Recommendation provides actionable advice based on competitive analysis.
type Recommendation struct {
	Type        RecommendationType `json:"type"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Priority    Priority           `json:"priority"`
	Effort      Effort             `json:"effort"`
	LearnFrom   []string           `json:"learn_from"` // competitor names
	Avoid       []string           `json:"avoid"`
}

// RecommendationType categorizes recommendations.
type RecommendationType string

const (
	RecommendationLearn         RecommendationType = "learn"
	RecommendationAvoid         RecommendationType = "avoid"
	RecommendationAdopt         RecommendationType = "adopt"
	RecommendationDifferentiate RecommendationType = "differentiate"
	RecommendationMonitor       RecommendationType = "monitor"
)

// Priority levels for recommendations.
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

// Effort estimates implementation effort.
type Effort string

const (
	EffortMinimal Effort = "minimal"
	EffortLow     Effort = "low"
	EffortMedium  Effort = "medium"
	EffortHigh    Effort = "high"
	EffortMassive Effort = "massive"
)

// RiskFactor identifies potential market risks.
type RiskFactor struct {
	Risk        string `json:"risk"`
	Description string `json:"description"`
	Likelihood  string `json:"likelihood"`
	Impact      string `json:"impact"`
	Mitigation  string `json:"mitigation"`
}

// Researcher defines the interface for web research capabilities.
type Researcher interface {
	SearchProjects(ctx context.Context, query string) ([]ResearchResult, error)
	AnalyzeRepository(ctx context.Context, url string) (*RepositoryAnalysis, error)
}

// ResearchResult represents a research finding from web search.
type ResearchResult struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Stars       int       `json:"stars"`
	Language    string    `json:"language"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RepositoryAnalysis contains detailed repository information.
type RepositoryAnalysis struct {
	Readme        string   `json:"readme"`
	Technologies  []string `json:"technologies"`
	Architecture  []string `json:"architecture"`
	Dependencies  []string `json:"dependencies"`
	Contributors  int      `json:"contributors"`
	RecentChanges []string `json:"recent_changes"`
}

// CompetitorAnalyzer provides competitive analysis functionality.
type CompetitorAnalyzer struct {
	researcher    Researcher
	projectName   string
	projectDomain string
	projectStack  []string
}

// AnalyzerOption configures the CompetitorAnalyzer.
type AnalyzerOption func(*CompetitorAnalyzer)

// WithResearcher sets the researcher implementation.
func WithResearcher(r Researcher) AnalyzerOption {
	return func(a *CompetitorAnalyzer) {
		a.researcher = r
	}
}

// WithProjectDomain sets the domain/niche for the project.
func WithProjectDomain(domain string) AnalyzerOption {
	return func(a *CompetitorAnalyzer) {
		a.projectDomain = domain
	}
}

// WithProjectStack sets the technology stack for the project.
func WithProjectStack(stack ...string) AnalyzerOption {
	return func(a *CompetitorAnalyzer) {
		a.projectStack = stack
	}
}

// NewCompetitorAnalyzer creates a new competitive analyzer.
func NewCompetitorAnalyzer(projectName string, opts ...AnalyzerOption) *CompetitorAnalyzer {
	analyzer := &CompetitorAnalyzer{
		projectName: projectName,
	}

	for _, opt := range opts {
		opt(analyzer)
	}

	return analyzer
}

// Analyze performs a full competitive analysis.
// It searches for similar projects, analyzes their implementations,
// and generates recommendations.
func (a *CompetitorAnalyzer) Analyze(ctx context.Context) (*CompetitorReport, error) {
	report := &CompetitorReport{
		ProjectName: a.projectName,
		GeneratedAt: time.Now(),
	}

	// Step 1: Search for similar projects via web research
	similarProjects, err := a.searchSimilarProjects(ctx)
	if err != nil {
		return nil, err
	}
	report.SimilarProjects = similarProjects

	// Step 2: Analyze each project's stack and architecture
	analyzedProjects, err := a.analyzeProjects(ctx, similarProjects)
	if err != nil {
		return nil, err
	}
	report.SimilarProjects = analyzedProjects

	// Step 3: Identify market trends
	report.MarketTrends = a.identifyMarketTrends(analyzedProjects)

	// Step 4: Generate differentiation analysis
	report.YourDifferentiation = a.generateDifferentiation(analyzedProjects)

	// Step 5: Generate recommendations
	report.Recommendations = a.generateRecommendations(analyzedProjects)

	// Step 6: Identify risk factors
	report.RiskFactors = a.identifyRiskFactors(analyzedProjects)

	return report, nil
}

// searchSimilarProjects searches for similar projects using the web researcher.
func (a *CompetitorAnalyzer) searchSimilarProjects(ctx context.Context) ([]SimilarProject, error) {
	if a.researcher == nil {
		return nil, nil
	}

	// Build search query from project domain and stack
	query := a.buildSearchQuery()
	results, err := a.researcher.SearchProjects(ctx, query)
	if err != nil {
		return nil, err
	}

	// Convert research results to similar projects
	projects := make([]SimilarProject, 0, len(results))
	for _, r := range results {
		project := SimilarProject{
			Name:           r.Title,
			Description:    r.Description,
			URL:            r.URL,
			LastAnalyzedAt: time.Now(),
		}
		if !r.UpdatedAt.IsZero() {
			project.Metrics.LastCommit = r.UpdatedAt
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// buildSearchQuery constructs a search query from project attributes.
func (a *CompetitorAnalyzer) buildSearchQuery() string {
	query := a.projectName
	if a.projectDomain != "" {
		query += " " + a.projectDomain
	}
	if len(a.projectStack) > 0 {
		query += " " + a.projectStack[0]
	}
	return query
}

// analyzeProjects performs deep analysis on found projects.
func (a *CompetitorAnalyzer) analyzeProjects(ctx context.Context, projects []SimilarProject) ([]SimilarProject, error) {
	if a.researcher == nil {
		return projects, nil
	}

	for i := range projects {
		analysis, err := a.researcher.AnalyzeRepository(ctx, projects[i].URL)
		if err != nil {
			continue // Skip failed analyses
		}

		projects[i].Stack = analysis.Technologies
		projects[i].Architecture = a.inferArchitecture(analysis.Architecture)
	}

	return projects, nil
}

// inferArchitecture guesses architecture pattern from analysis data.
func (a *CompetitorAnalyzer) inferArchitecture(archHints []string) ArchitecturePattern {
	for _, hint := range archHints {
		switch hint {
		case "monolith":
			return ArchMonolithic
		case "modular", "modules":
			return ArchModular
		case "microservices", "micro-services":
			return ArchMicroservices
		case "serverless", "lambda", "functions":
			return ArchServerless
		case "event-driven", "events", "cqrs":
			return ArchEventDriven
		case "hexagonal", "ports-and-adapters", "onion":
			return ArchHexagonal
		case "clean", "clean-architecture":
			return ArchClean
		case "ddd", "domain-driven":
			return ArchDDD
		case "plugin", "plugins", "extensible":
			return ArchPlugin
		}
	}
	return ArchModular // Default assumption
}

// identifyMarketTrends analyzes projects to find market trends.
func (a *CompetitorAnalyzer) identifyMarketTrends(projects []SimilarProject) []MarketTrend {
	trendMap := make(map[string]*MarketTrend)

	// Analyze feature patterns across projects
	featureCounts := make(map[string]int)
	for _, p := range projects {
		for _, f := range p.Features {
			featureCounts[f.Name]++
		}
	}

	// Convert common features to trends
	for feature, count := range featureCounts {
		if count >= len(projects)/2 {
			trend := &MarketTrend{
				Trend:        feature,
				Description:  "Adopted by majority of competitors",
				AdoptionRate: "high",
				FirstSeen:    time.Now().AddDate(-1, 0, 0), // Approximation
			}
			trendMap[feature] = trend
		}
	}

	// Analyze stack patterns
	stackCounts := make(map[string]int)
	for _, p := range projects {
		for _, tech := range p.Stack {
			stackCounts[tech]++
		}
	}

	// Convert common stacks to trends
	for stack, count := range stackCounts {
		if count >= len(projects)/2 {
			trend := &MarketTrend{
				Trend:        stack + " ecosystem",
				Description:  "Dominant technology choice",
				AdoptionRate: "high",
				FirstSeen:    time.Now().AddDate(-2, 0, 0),
			}
			trendMap[stack] = trend
		}
	}

	trends := make([]MarketTrend, 0, len(trendMap))
	for _, t := range trendMap {
		trends = append(trends, *t)
	}

	return trends
}

// generateDifferentiation identifies how your project differs from competitors.
func (a *CompetitorAnalyzer) generateDifferentiation(projects []SimilarProject) []Differentiation {
	var differentiations []Differentiation

	// Analyze what competitors are NOT doing
	commonStacks := a.getCommonStacks(projects)

	// Find gaps in the market
	differentiations = append(differentiations, Differentiation{
		Advantage: "Unique positioning",
		Impact:    "High",
		Evidence:  "No direct competitor offers this combination",
		Priority:  5,
	})

	// Differentiation from stack choices
	ownStacks := make(map[string]bool)
	for _, s := range a.projectStack {
		ownStacks[s] = true
	}

	for _, cs := range commonStacks {
		if !ownStacks[cs] {
			differentiations = append(differentiations, Differentiation{
				Advantage: "Different tech stack: " + cs,
				Impact:    "Medium",
				Evidence:  "Competitors use different approach",
				Priority:  3,
			})
		}
	}

	return differentiations
}

// getCommonFeatures returns features present in most projects.
func (a *CompetitorAnalyzer) getCommonFeatures(projects []SimilarProject) []string {
	if len(projects) == 0 {
		return nil
	}

	featureCount := make(map[string]int)
	for _, p := range projects {
		for _, f := range p.Features {
			featureCount[f.Name]++
		}
	}

	threshold := len(projects) / 2
	var common []string
	for f, count := range featureCount {
		if count >= threshold {
			common = append(common, f)
		}
	}

	return common
}

// getCommonStacks returns stacks used by most projects.
func (a *CompetitorAnalyzer) getCommonStacks(projects []SimilarProject) []string {
	if len(projects) == 0 {
		return nil
	}

	stackCount := make(map[string]int)
	for _, p := range projects {
		for _, s := range p.Stack {
			stackCount[s]++
		}
	}

	threshold := len(projects) / 2
	var common []string
	for s, count := range stackCount {
		if count >= threshold {
			common = append(common, s)
		}
	}

	return common
}

// generateRecommendations creates actionable recommendations based on analysis.
func (a *CompetitorAnalyzer) generateRecommendations(projects []SimilarProject) []Recommendation {
	var recommendations []Recommendation

	// Learn from high-performing competitors
	highPerformers := a.filterHighPerformers(projects)
	if len(highPerformers) > 0 {
		recommendations = append(recommendations, Recommendation{
			Type:        RecommendationLearn,
			Title:       "Learn from successful competitors",
			Description: "Study " + highPerformers[0].Name + " and similar high-performing projects",
			Priority:    PriorityHigh,
			Effort:      EffortLow,
			LearnFrom:   a.extractNames(highPerformers),
		})
	}

	// Common patterns to adopt
	commonFeatures := a.getCommonFeatures(projects)
	if len(commonFeatures) > 0 {
		recommendations = append(recommendations, Recommendation{
			Type:        RecommendationAdopt,
			Title:       "Adopt industry-standard features",
			Description: "Implement features that are now table stakes: " + strings.Join(commonFeatures, ", "),
			Priority:    PriorityMedium,
			Effort:      EffortMedium,
		})
	}

	// Areas to avoid based on competitor failures
	recommendations = append(recommendations, Recommendation{
		Type:        RecommendationAvoid,
		Title:       "Avoid common failure patterns",
		Description: "Don't repeat mistakes that caused competitor weaknesses",
		Priority:    PriorityHigh,
		Effort:      EffortMinimal,
		Avoid:       a.getCommonWeaknesses(projects),
	})

	// Differentiate strategy
	recommendations = append(recommendations, Recommendation{
		Type:        RecommendationDifferentiate,
		Title:       "Focus on differentiation",
		Description: "Invest in unique value that competitors lack",
		Priority:    PriorityCritical,
		Effort:      EffortHigh,
	})

	// Technology monitoring
	recommendations = append(recommendations, Recommendation{
		Type:        RecommendationMonitor,
		Title:       "Monitor emerging technologies",
		Description: "Watch for new technologies gaining adoption",
		Priority:    PriorityLow,
		Effort:      EffortMinimal,
	})

	return recommendations
}

// filterHighPerformers returns projects with strong metrics.
func (a *CompetitorAnalyzer) filterHighPerformers(projects []SimilarProject) []SimilarProject {
	var performers []SimilarProject
	for _, p := range projects {
		if p.Metrics.Stars > 1000 || p.Metrics.Forks > 100 {
			performers = append(performers, p)
		}
	}
	return performers
}

// extractNames returns project names from a slice.
func (a *CompetitorAnalyzer) extractNames(projects []SimilarProject) []string {
	names := make([]string, len(projects))
	for i, p := range projects {
		names[i] = p.Name
	}
	return names
}

// getCommonWeaknesses identifies weaknesses mentioned across competitors.
func (a *CompetitorAnalyzer) getCommonWeaknesses(projects []SimilarProject) []string {
	weaknessCount := make(map[string]int)
	for _, p := range projects {
		for _, w := range p.Weaknesses {
			weaknessCount[w]++
		}
	}

	var common []string
	for w, count := range weaknessCount {
		if count >= len(projects)/3 {
			common = append(common, w)
		}
	}

	return common
}

// identifyRiskFactors identifies potential risks from competitive landscape.
func (a *CompetitorAnalyzer) identifyRiskFactors(projects []SimilarProject) []RiskFactor {
	var risks []RiskFactor

	// Check for market saturation
	if len(projects) > 10 {
		risks = append(risks, RiskFactor{
			Risk:        "Market Saturation",
			Description: "Many competitors exist in this space",
			Likelihood:  "high",
			Impact:      "medium",
			Mitigation:  "Focus on differentiation and niche positioning",
		})
	}

	// Check for stale competitors (potential for disruption)
	staleCount := 0
	for _, p := range projects {
		if time.Since(p.Metrics.LastCommit) > 365*24*time.Hour {
			staleCount++
		}
	}
	if staleCount > len(projects)/2 {
		risks = append(risks, RiskFactor{
			Risk:        "Competitor Stagnation",
			Description: "Many competitors are not actively maintained",
			Likelihood:  "medium",
			Impact:      "high",
			Mitigation:  "Actively maintain and update the project",
		})
	}

	// Check for well-funded competitors
	highStars := 0
	for _, p := range projects {
		if p.Metrics.Stars > 5000 {
			highStars++
		}
	}
	if highStars > 2 {
		risks = append(risks, RiskFactor{
			Risk:        "Established Players",
			Description: "Well-established competitors with large communities",
			Likelihood:  "high",
			Impact:      "high",
			Mitigation:  "Find underserved niches and build strong community",
		})
	}

	return risks
}

// AnalyzeStack analyzes a given technology stack against competitors.
func (a *CompetitorAnalyzer) AnalyzeStack(stack []string) []StackAnalysis {
	var analyses []StackAnalysis

	stackCounts := make(map[string]int)
	projects := a.findProjectsWithStack(stack)
	for range projects {
		for _, s := range stack {
			stackCounts[s]++
		}
	}

	for _, tech := range stack {
		analysis := StackAnalysis{
			Technology:   tech,
			UsageCount:   stackCounts[tech],
			Popularity:   a.assessTechPopularity(tech),
			Trend:        a.assessTechTrend(tech),
			Alternatives: a.findTechAlternatives(tech),
		}
		analyses = append(analyses, analysis)
	}

	return analyses
}

// findProjectsWithStack returns projects using a specific stack.
func (a *CompetitorAnalyzer) findProjectsWithStack(stack []string) []SimilarProject {
	// This would normally search through analyzed projects
	return nil
}

// assessTechPopularity evaluates technology popularity.
func (a *CompetitorAnalyzer) assessTechPopularity(tech string) string {
	popularTechs := map[string]bool{
		"Go": true, "Rust": true, "TypeScript": true, "React": true,
		"Python": true, "Node.js": true, "PostgreSQL": true,
	}

	if popularTechs[tech] {
		return "high"
	}
	return "medium"
}

// assessTechTrend evaluates technology trend direction.
func (a *CompetitorAnalyzer) assessTechTrend(tech string) string {
	growingTechs := map[string]bool{
		"Rust": true, "TypeScript": true, "Go": true, "WASM": true,
		"Bun": true, "HTMX": true,
	}

	if growingTechs[tech] {
		return "growing"
	}
	return "stable"
}

// findTechAlternatives suggests alternative technologies.
func (a *CompetitorAnalyzer) findTechAlternatives(tech string) []string {
	alternatives := map[string][]string{
		"JavaScript": {"TypeScript", "Dart", "WASM"},
		"MongoDB":    {"PostgreSQL", "MySQL", "SQLite"},
		"Docker":     {"Podman", "containerd", "runc"},
	}

	if alts, ok := alternatives[tech]; ok {
		return alts
	}
	return nil
}

// StackAnalysis contains analysis of a specific technology.
type StackAnalysis struct {
	Technology   string   `json:"technology"`
	UsageCount   int      `json:"usage_count"`
	Popularity   string   `json:"popularity"`
	Trend        string   `json:"trend"`
	Alternatives []string `json:"alternatives"`
}
