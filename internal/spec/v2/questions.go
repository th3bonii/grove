// Package spec provides the GROVE Spec engine v2 for self-questioning components.
//
// This module implements the deep self-questioning system where the engine
// asks itself fundamental questions about each component to ensure quality
// and completeness. The system covers 6 core dimensions: WHY, HOW, WHERE,
// WHO, WHAT, and ALTERNATIVE.
package spec

import (
	"fmt"
	"strings"
)

// =============================================================================
// Question Categories
// =============================================================================

// QuestionCategory represents the type of self-question.
type QuestionCategory string

const (
	CategoryWhy         QuestionCategory = "why"
	CategoryHow         QuestionCategory = "how"
	CategoryWhere       QuestionCategory = "where"
	CategoryWho         QuestionCategory = "who"
	CategoryWhat        QuestionCategory = "what"
	CategoryAlternative QuestionCategory = "alternative"
)

// QuestionStatus represents the status of a question's answer.
type QuestionStatus string

const (
	StatusAnswered      QuestionStatus = "answered"
	StatusNeedsResearch QuestionStatus = "needs_research"
	StatusDeferred      QuestionStatus = "deferred"
	StatusNA            QuestionStatus = "n/a"
)

// QuestionSource represents where the answer came from.
type QuestionSource string

const (
	SourceInference    QuestionSource = "inference"
	SourceWeb          QuestionSource = "web"
	SourceMCP          QuestionSource = "mcp"
	SourceBestPractice QuestionSource = "best-practice"
	SourceContext      QuestionSource = "context"
)

// DeepQuestion represents a comprehensive self-question.
type DeepQuestion struct {
	ID         string           `json:"id"`
	Question   string           `json:"question"`
	Category   QuestionCategory `json:"category"`
	Answer     string           `json:"answer"`
	Status     QuestionStatus   `json:"status"`
	Source     QuestionSource   `json:"source"`
	Confidence float64          `json:"confidence"` // 0.0 - 1.0
	Tags       []string         `json:"tags"`
	Children   []DeepQuestion   `json:"children,omitempty"`
	NeedsWeb   bool             `json:"needs_web"`
	NeedsMCP   bool             `json:"needs_mcp"`
}

// ResearchTopic represents a topic that needs research.
type ResearchTopic struct {
	Question    string           `json:"question"`
	Category    QuestionCategory `json:"category"`
	Priority    string           `json:"priority"` // high, medium, low
	SearchQuery string           `json:"search_query"`
}

// =============================================================================
// Engine Extension for Deep Questions
// =============================================================================

// askDeepQuestions generates comprehensive self-questions for a component.
// This is the main entry point for the deep questioning system.
//
// The engine asks questions in this order:
// 1. WHY - Understanding user motivation and real objectives
// 2. WHAT - Understanding exactly what the component does
// 3. HOW - Understanding implementation approach
// 4. WHERE - Understanding context and dependencies
// 5. WHO - Understanding the user and their needs
// 6. ALTERNATIVE - Exploring alternatives and best practices
func (e *Engine) askDeepQuestions(comp Component) []DeepQuestion {
	questions := make([]DeepQuestion, 0)

	// Phase 1: WHY questions - understanding motivation
	questions = append(questions, e.askWhyQuestions(comp)...)

	// Phase 2: WHAT questions - understanding functionality
	questions = append(questions, e.askWhatQuestions(comp)...)

	// Phase 3: HOW questions - understanding implementation
	questions = append(questions, e.askHowQuestions(comp)...)

	// Phase 4: WHERE questions - understanding context
	questions = append(questions, e.askWhereQuestions(comp)...)

	// Phase 5: WHO questions - understanding users
	questions = append(questions, e.askWhoQuestions(comp)...)

	// Phase 6: ALTERNATIVE questions - exploring alternatives
	questions = append(questions, e.askAlternativeQuestions(comp)...)

	return questions
}

// =============================================================================
// WHY Questions - Understanding User Motivation
// =============================================================================

// askWhyQuestions generates WHY questions for understanding motivation.
func (e *Engine) askWhyQuestions(comp Component) []DeepQuestion {
	questions := []DeepQuestion{
		{
			ID:         fmt.Sprintf("%s-why-1", comp.Name),
			Question:   fmt.Sprintf("Why does the user want '%s'?", comp.Name),
			Category:   CategoryWhy,
			Answer:     inferWhyUserWants(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.8,
			Tags:       []string{"motivation", "user-goal"},
		},
		{
			ID:         fmt.Sprintf("%s-why-2", comp.Name),
			Question:   fmt.Sprintf("What is the real objective behind '%s'?", comp.Name),
			Category:   CategoryWhy,
			Answer:     inferRealObjective(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.7,
			Tags:       []string{"objective", "business-value"},
		},
		{
			ID:         fmt.Sprintf("%s-why-3", comp.Name),
			Question:   fmt.Sprintf("Is '%s' the best way to achieve this objective?", comp.Name),
			Category:   CategoryWhy,
			Answer:     "",
			Status:     StatusNeedsResearch,
			Source:     SourceInference,
			Confidence: 0.0,
			Tags:       []string{"alternatives", "validation"},
			NeedsWeb:   true,
		},
		{
			ID:         fmt.Sprintf("%s-why-4", comp.Name),
			Question:   fmt.Sprintf("What problem does '%s' solve?", comp.Name),
			Category:   CategoryWhy,
			Answer:     inferProblemSolved(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.75,
			Tags:       []string{"problem", "pain-point"},
		},
		{
			ID:         fmt.Sprintf("%s-why-5", comp.Name),
			Question:   fmt.Sprintf("What would happen if '%s' didn't exist?", comp.Name),
			Category:   CategoryWhy,
			Answer:     inferImpactWithout(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.65,
			Tags:       []string{"impact", "necessity"},
		},
	}

	return questions
}

// =============================================================================
// WHAT Questions - Understanding Functionality
// =============================================================================

// askWhatQuestions generates WHAT questions for understanding functionality.
func (e *Engine) askWhatQuestions(comp Component) []DeepQuestion {
	questions := []DeepQuestion{
		{
			ID:         fmt.Sprintf("%s-what-1", comp.Name),
			Question:   fmt.Sprintf("What exactly does '%s' do?", comp.Name),
			Category:   CategoryWhat,
			Answer:     comp.Description,
			Status:     StatusAnswered,
			Source:     SourceContext,
			Confidence: 1.0,
			Tags:       []string{"functionality", "core-feature"},
		},
		{
			ID:         fmt.Sprintf("%s-what-2", comp.Name),
			Question:   fmt.Sprintf("What data does '%s' need as input?", comp.Name),
			Category:   CategoryWhat,
			Answer:     inferInputData(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.7,
			Tags:       []string{"input", "data"},
		},
		{
			ID:         fmt.Sprintf("%s-what-3", comp.Name),
			Question:   fmt.Sprintf("What does '%s' return as output?", comp.Name),
			Category:   CategoryWhat,
			Answer:     inferOutputData(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.7,
			Tags:       []string{"output", "result"},
		},
		{
			ID:         fmt.Sprintf("%s-what-4", comp.Name),
			Question:   fmt.Sprintf("What are the key states of '%s'?", comp.Name),
			Category:   CategoryWhat,
			Answer:     formatStates(comp.States),
			Status:     StatusAnswered,
			Source:     SourceContext,
			Confidence: 1.0,
			Tags:       []string{"states", "lifecycle"},
		},
		{
			ID:         fmt.Sprintf("%s-what-5", comp.Name),
			Question:   fmt.Sprintf("What behaviors/actions does '%s' perform?", comp.Name),
			Category:   CategoryWhat,
			Answer:     formatBehaviors(comp.Behaviors),
			Status:     StatusAnswered,
			Source:     SourceContext,
			Confidence: 1.0,
			Tags:       []string{"behaviors", "actions"},
		},
		{
			ID:         fmt.Sprintf("%s-what-6", comp.Name),
			Question:   fmt.Sprintf("What edge cases must '%s' handle?", comp.Name),
			Category:   CategoryWhat,
			Answer:     formatEdgeCases(comp.EdgeCases),
			Status:     StatusAnswered,
			Source:     SourceContext,
			Confidence: 0.9,
			Tags:       []string{"edge-cases", "error-handling"},
		},
	}

	return questions
}

// =============================================================================
// HOW Questions - Understanding Implementation
// =============================================================================

// askHowQuestions generates HOW questions for understanding implementation.
func (e *Engine) askHowQuestions(comp Component) []DeepQuestion {
	questions := []DeepQuestion{
		{
			ID:         fmt.Sprintf("%s-how-1", comp.Name),
			Question:   fmt.Sprintf("How should '%s' be implemented?", comp.Name),
			Category:   CategoryHow,
			Answer:     "",
			Status:     StatusNeedsResearch,
			Source:     SourceWeb,
			Confidence: 0.0,
			Tags:       []string{"implementation", "approach"},
			NeedsWeb:   true,
		},
		{
			ID:         fmt.Sprintf("%s-how-2", comp.Name),
			Question:   fmt.Sprintf("How should '%s' handle error states?", comp.Name),
			Category:   CategoryHow,
			Answer:     inferErrorHandling(comp),
			Status:     StatusAnswered,
			Source:     SourceBestPractice,
			Confidence: 0.85,
			Tags:       []string{"error-handling", "resilience"},
		},
		{
			ID:         fmt.Sprintf("%s-how-3", comp.Name),
			Question:   fmt.Sprintf("How should '%s' handle loading states?", comp.Name),
			Category:   CategoryHow,
			Answer:     inferLoadingHandling(comp),
			Status:     StatusAnswered,
			Source:     SourceBestPractice,
			Confidence: 0.8,
			Tags:       []string{"loading", "ux"},
		},
		{
			ID:         fmt.Sprintf("%s-how-4", comp.Name),
			Question:   fmt.Sprintf("How does '%s' interact with other components?", comp.Name),
			Category:   CategoryHow,
			Answer:     formatDependencies(comp.Dependencies),
			Status:     StatusAnswered,
			Source:     SourceContext,
			Confidence: 0.9,
			Tags:       []string{"integration", "dependencies"},
		},
		{
			ID:         fmt.Sprintf("%s-how-5", comp.Name),
			Question:   fmt.Sprintf("How should '%s' be tested?", comp.Name),
			Category:   CategoryHow,
			Answer:     "",
			Status:     StatusNeedsResearch,
			Source:     SourceWeb,
			Confidence: 0.0,
			Tags:       []string{"testing", "quality"},
			NeedsWeb:   true,
		},
		{
			ID:         fmt.Sprintf("%s-how-6", comp.Name),
			Question:   fmt.Sprintf("How should '%s' be optimized for performance?", comp.Name),
			Category:   CategoryHow,
			Answer:     "",
			Status:     StatusNeedsResearch,
			Source:     SourceWeb,
			Confidence: 0.0,
			Tags:       []string{"performance", "optimization"},
			NeedsWeb:   true,
		},
	}

	return questions
}

// =============================================================================
// WHERE Questions - Understanding Context
// =============================================================================

// askWhereQuestions generates WHERE questions for understanding context.
func (e *Engine) askWhereQuestions(comp Component) []DeepQuestion {
	questions := []DeepQuestion{
		{
			ID:         fmt.Sprintf("%s-where-1", comp.Name),
			Question:   fmt.Sprintf("Where will '%s' be used?", comp.Name),
			Category:   CategoryWhere,
			Answer:     inferUsageContext(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.7,
			Tags:       []string{"context", "location"},
		},
		{
			ID:         fmt.Sprintf("%s-where-2", comp.Name),
			Question:   fmt.Sprintf("In what context does '%s' operate?", comp.Name),
			Category:   CategoryWhere,
			Answer:     inferOperatingContext(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.7,
			Tags:       []string{"environment", "context"},
		},
		{
			ID:         fmt.Sprintf("%s-where-3", comp.Name),
			Question:   fmt.Sprintf("What other parts of the system are affected by '%s'?", comp.Name),
			Category:   CategoryWhere,
			Answer:     formatDependencies(comp.Dependencies),
			Status:     StatusAnswered,
			Source:     SourceContext,
			Confidence: 0.9,
			Tags:       []string{"impact", "side-effects"},
		},
		{
			ID:         fmt.Sprintf("%s-where-4", comp.Name),
			Question:   fmt.Sprintf("Where does the data for '%s' come from?", comp.Name),
			Category:   CategoryWhere,
			Answer:     inferDataSource(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.6,
			Tags:       []string{"data-source", "origin"},
		},
		{
			ID:         fmt.Sprintf("%s-where-5", comp.Name),
			Question:   fmt.Sprintf("Where should '%s' be placed in the architecture?", comp.Name),
			Category:   CategoryWhere,
			Answer:     inferArchitecturePlacement(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.65,
			Tags:       []string{"architecture", "placement"},
		},
	}

	return questions
}

// =============================================================================
// WHO Questions - Understanding Users
// =============================================================================

// askWhoQuestions generates WHO questions for understanding users.
func (e *Engine) askWhoQuestions(comp Component) []DeepQuestion {
	questions := []DeepQuestion{
		{
			ID:         fmt.Sprintf("%s-who-1", comp.Name),
			Question:   fmt.Sprintf("Who will use '%s'?", comp.Name),
			Category:   CategoryWho,
			Answer:     inferTargetUser(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.7,
			Tags:       []string{"user", "audience"},
		},
		{
			ID:         fmt.Sprintf("%s-who-2", comp.Name),
			Question:   fmt.Sprintf("What level of expertise do users of '%s' have?", comp.Name),
			Category:   CategoryWho,
			Answer:     inferUserExpertise(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.6,
			Tags:       []string{"expertise", "skill-level"},
		},
		{
			ID:         fmt.Sprintf("%s-who-3", comp.Name),
			Question:   fmt.Sprintf("What do users expect from '%s'?", comp.Name),
			Category:   CategoryWho,
			Answer:     inferUserExpectations(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.65,
			Tags:       []string{"expectations", "user-experience"},
		},
		{
			ID:         fmt.Sprintf("%s-who-4", comp.Name),
			Question:   fmt.Sprintf("How will '%s' be discovered by users?", comp.Name),
			Category:   CategoryWho,
			Answer:     inferDiscoveryMethod(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.6,
			Tags:       []string{"discovery", "onboarding"},
		},
		{
			ID:         fmt.Sprintf("%s-who-5", comp.Name),
			Question:   fmt.Sprintf("What language/localization is needed for '%s'?", comp.Name),
			Category:   CategoryWho,
			Answer:     inferLocalization(comp),
			Status:     StatusAnswered,
			Source:     SourceInference,
			Confidence: 0.7,
			Tags:       []string{"i18n", "localization"},
		},
	}

	return questions
}

// =============================================================================
// ALTERNATIVE Questions - Exploring Alternatives
// =============================================================================

// askAlternativeQuestions generates ALTERNATIVE questions for exploring alternatives.
func (e *Engine) askAlternativeQuestions(comp Component) []DeepQuestion {
	questions := []DeepQuestion{
		{
			ID:         fmt.Sprintf("%s-alt-1", comp.Name),
			Question:   fmt.Sprintf("Is there a better way to implement '%s'?", comp.Name),
			Category:   CategoryAlternative,
			Answer:     "",
			Status:     StatusNeedsResearch,
			Source:     SourceWeb,
			Confidence: 0.0,
			Tags:       []string{"alternatives", "comparison"},
			NeedsWeb:   true,
		},
		{
			ID:         fmt.Sprintf("%s-alt-2", comp.Name),
			Question:   fmt.Sprintf("What are the best practices for '%s'?", comp.Name),
			Category:   CategoryAlternative,
			Answer:     "",
			Status:     StatusNeedsResearch,
			Source:     SourceWeb,
			Confidence: 0.0,
			Tags:       []string{"best-practices", "standards"},
			NeedsWeb:   true,
		},
		{
			ID:         fmt.Sprintf("%s-alt-3", comp.Name),
			Question:   "What do competitors/industry leaders do for similar features?",
			Category:   CategoryAlternative,
			Answer:     "",
			Status:     StatusNeedsResearch,
			Source:     SourceWeb,
			Confidence: 0.0,
			Tags:       []string{"competition", "industry-standard"},
			NeedsWeb:   true,
		},
		{
			ID:         fmt.Sprintf("%s-alt-4", comp.Name),
			Question:   "What are the pros/cons of current approach vs alternatives?",
			Category:   CategoryAlternative,
			Answer:     "",
			Status:     StatusNeedsResearch,
			Source:     SourceMCP,
			Confidence: 0.0,
			Tags:       []string{"tradeoffs", "analysis"},
			NeedsMCP:   true,
		},
		{
			ID:         fmt.Sprintf("%s-alt-5", comp.Name),
			Question:   fmt.Sprintf("Should '%s' be built in-house or use a third-party solution?", comp.Name),
			Category:   CategoryAlternative,
			Answer:     "",
			Status:     StatusNeedsResearch,
			Source:     SourceInference,
			Confidence: 0.5,
			Tags:       []string{"build-vs-buy", "outsourcing"},
		},
	}

	return questions
}

// =============================================================================
// Research Topic Extraction
// =============================================================================

// getResearchTopics extracts topics that need web/MCP research.
func (e *Engine) getResearchTopics(questions []DeepQuestion) []ResearchTopic {
	topics := make([]ResearchTopic, 0)

	for _, q := range questions {
		if q.Status == StatusNeedsResearch {
			priority := "medium"
			if q.Category == CategoryAlternative {
				priority = "high"
			} else if q.Category == CategoryHow {
				priority = "medium"
			}

			searchQuery := q.Question
			if q.NeedsWeb {
				searchQuery = fmt.Sprintf("%s best practices implementation", compNameFromID(q.ID))
			}

			topics = append(topics, ResearchTopic{
				Question:    q.Question,
				Category:    q.Category,
				Priority:    priority,
				SearchQuery: searchQuery,
			})
		}
	}

	return topics
}

// =============================================================================
// Inference Helpers
// =============================================================================

func compNameFromID(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-2], "-")
	}
	return id
}

func inferWhyUserWants(comp Component) string {
	lower := strings.ToLower(comp.Name)
	if strings.Contains(lower, "login") || strings.Contains(lower, "auth") {
		return "To authenticate users and secure access to the application"
	}
	if strings.Contains(lower, "search") {
		return "To find content or data quickly and efficiently"
	}
	if strings.Contains(lower, "form") {
		return "To collect and validate user input"
	}
	if strings.Contains(lower, "dashboard") || strings.Contains(lower, "analytics") {
		return "To visualize data and gain insights"
	}
	if strings.Contains(lower, "notification") {
		return "To keep users informed of important events"
	}
	return fmt.Sprintf("To enable %s functionality as part of the user interface", strings.ToLower(comp.Name))
}

func inferRealObjective(comp Component) string {
	lower := strings.ToLower(comp.Name)
	if strings.Contains(lower, "button") {
		return "To provide a clear call-to-action for user interactions"
	}
	if strings.Contains(lower, "input") {
		return "To capture user data with proper validation"
	}
	if strings.Contains(lower, "modal") {
		return "To present focused information or actions without leaving context"
	}
	return "To deliver the specified feature with optimal user experience"
}

func inferProblemSolved(comp Component) string {
	return fmt.Sprintf("Solves the problem of implementing %s functionality", strings.ToLower(comp.Name))
}

func inferImpactWithout(comp Component) string {
	return fmt.Sprintf("Without '%s', users would lack %s capability", comp.Name, strings.ToLower(comp.Name))
}

func inferInputData(comp Component) string {
	lower := strings.ToLower(comp.Name)
	if strings.Contains(lower, "form") {
		return "User input data, validation rules, submission handler"
	}
	if strings.Contains(lower, "search") {
		return "Search query, filters, pagination parameters"
	}
	if strings.Contains(lower, "button") || strings.Contains(lower, "click") {
		return "Click event, action handler, state"
	}
	return "Props, state data, user interactions"
}

func inferOutputData(comp Component) string {
	lower := strings.ToLower(comp.Name)
	if strings.Contains(lower, "form") {
		return "Validated form data, submission status, error messages"
	}
	if strings.Contains(lower, "search") {
		return "Search results, pagination info, loading state"
	}
	if strings.Contains(lower, "button") {
		return "Action trigger, state update, callback execution"
	}
	return "Rendered UI, state updates, event emissions"
}

func inferErrorHandling(comp Component) string {
	return "Display clear error messages, provide recovery actions, log errors, maintain stable state"
}

func inferLoadingHandling(comp Component) string {
	return "Show skeleton/spinner, disable interactions during loading, handle timeout gracefully"
}

func inferUsageContext(comp Component) string {
	lower := strings.ToLower(comp.Name)
	if strings.Contains(lower, "modal") {
		return "Overlays, dialogs, focused interactions"
	}
	if strings.Contains(lower, "sidebar") || strings.Contains(lower, "nav") {
		return "Navigation areas, sidebars, headers"
	}
	if strings.Contains(lower, "form") {
		return "Data entry sections, settings pages, registration flows"
	}
	return "User interface as specified in requirements"
}

func inferOperatingContext(comp Component) string {
	lower := strings.ToLower(comp.Name)
	if strings.Contains(lower, "service") || strings.Contains(lower, "api") {
		return "Backend, server-side, API layer"
	}
	if strings.Contains(lower, "ui") || strings.Contains(lower, "button") {
		return "Client-side, browser environment"
	}
	return "Application runtime environment"
}

func inferDataSource(comp Component) string {
	lower := strings.ToLower(comp.Name)
	if strings.Contains(lower, "user") {
		return "User input, auth system, profile data"
	}
	if strings.Contains(lower, "config") {
		return "Configuration files, environment variables, database"
	}
	return "Props, state, API calls, local storage"
}

func inferArchitecturePlacement(comp Component) string {
	switch comp.Type {
	case "ui":
		return "Presentation layer, component tree"
	case "service":
		return "Business logic layer, service layer"
	case "data":
		return "Data layer, repository pattern"
	case "integration":
		return "Integration layer, API gateway"
	default:
		return "Feature module, domain layer"
	}
}

func inferTargetUser(comp Component) string {
	return "End users of the application"
}

func inferUserExpertise(comp Component) string {
	lower := strings.ToLower(comp.Name)
	if strings.Contains(lower, "admin") || strings.Contains(lower, "config") {
		return "Technical users, administrators, power users"
	}
	if strings.Contains(lower, "onboarding") || strings.Contains(lower, "tutorial") {
		return "New users, novice users"
	}
	return "General users with standard technical literacy"
}

func inferUserExpectations(comp Component) string {
	return "Responsive, intuitive, accessible, consistent with platform conventions"
}

func inferDiscoveryMethod(comp Component) string {
	lower := strings.ToLower(comp.Name)
	if strings.Contains(lower, "setting") || strings.Contains(lower, "config") {
		return "Settings menu, preferences panel"
	}
	if strings.Contains(lower, "help") || strings.Contains(lower, "faq") {
		return "Help section, documentation, search"
	}
	return "Natural placement in user flow, navigation, search"
}

func inferLocalization(comp Component) string {
	return "Multi-language support required, RTL consideration for international users"
}

// =============================================================================
// Formatters
// =============================================================================

func formatStates(states []ComponentState) string {
	if len(states) == 0 {
		return "No states defined"
	}
	names := make([]string, len(states))
	for i, s := range states {
		names[i] = s.Name
	}
	return strings.Join(names, ", ")
}

func formatBehaviors(behaviors []Behavior) string {
	if len(behaviors) == 0 {
		return "No behaviors defined"
	}
	names := make([]string, len(behaviors))
	for i, b := range behaviors {
		names[i] = b.Name
	}
	return strings.Join(names, ", ")
}

func formatEdgeCases(edgeCases []EdgeCase) string {
	if len(edgeCases) == 0 {
		return "No edge cases defined"
	}
	scenarios := make([]string, len(edgeCases))
	for i, e := range edgeCases {
		scenarios[i] = fmt.Sprintf("%s (%s)", e.Scenario, e.Severity)
	}
	return strings.Join(scenarios, ", ")
}

func formatDependencies(deps []string) string {
	if len(deps) == 0 {
		return "No explicit dependencies"
	}
	return strings.Join(deps, ", ")
}
