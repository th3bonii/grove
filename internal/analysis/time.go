package analysis

import (
	"fmt"
	"strings"
)

// Phase represents a project phase in the development lifecycle
type Phase string

const (
	PhaseSetup        Phase = "Setup & Configuration"
	PhaseCoreFeatures Phase = "Core Features"
	PhaseIntegration  Phase = "Integration"
	PhaseTesting      Phase = "Testing"
	PhasePolish       Phase = "Polish & Bug Fixes"
	PhaseDeployment   Phase = "Deployment"
)

// TechStack represents the technology stack characteristics
type TechStack struct {
	Name           string
	ComponentCount int
	Complexity     string // low, medium, high, very-high
}

// complexityToValue converts complexity string to numeric value.
func complexityToValue(c string) int {
	switch c {
	case "low":
		return 1
	case "medium":
		return 2
	case "high":
		return 3
	case "very-high":
		return 4
	default:
		return 2 // Default to medium
	}
}

// PhaseBreakdown represents time breakdown for a single phase
type PhaseBreakdown struct {
	Phase      Phase   `json:"phase"`
	Weeks      float64 `json:"weeks"`
	Percentage float64 `json:"percentage"`
}

// VelocityAssumptions contains assumptions about team velocity
type VelocityAssumptions struct {
	HoursPerWeek      float64 `json:"hoursPerWeek"`
	ComponentsPerWeek float64 `json:"componentsPerWeek"`
	MeetingsOverhead  float64 `json:"meetingsOverhead"` // as decimal (0.2 = 20%)
}

// TimeEstimate contains the full time estimation result
type TimeEstimate struct {
	TotalWeeks          string              `json:"totalWeeks"`
	TotalHours          float64             `json:"totalHours"`
	Breakdown           []PhaseBreakdown    `json:"breakdown"`
	TeamSize            int                 `json:"teamSize"`
	VelocityAssumptions VelocityAssumptions `json:"velocityAssumptions"`
	ComponentsAnalyzed  int                 `json:"componentsAnalyzed"`
}

// DefaultVelocityAssumptions returns standard velocity assumptions
func DefaultVelocityAssumptions() VelocityAssumptions {
	return VelocityAssumptions{
		HoursPerWeek:      32, // accounting for meetings, breaks, context switching
		ComponentsPerWeek: 2,  // 2 medium complexity components per developer per week
		MeetingsOverhead:  0.2,
	}
}

// CalculateTimeEstimate estimates project duration based on inputs
func CalculateTimeEstimate(
	components int,
	componentComplexities []string,
	teamSize int,
	stacks []TechStack,
) TimeEstimate {
	if teamSize <= 0 {
		teamSize = 1
	}

	// Calculate weighted complexity score
	totalComplexity := 0
	for _, c := range componentComplexities {
		totalComplexity += complexityToValue(c)
	}
	if len(componentComplexities) == 0 {
		totalComplexity = components * complexityToValue("medium")
	}

	// Add stack complexities
	for _, stack := range stacks {
		totalComplexity += stack.ComponentCount * complexityToValue(stack.Complexity)
	}

	// Base hours calculation
	baseHours := float64(totalComplexity) * 20 // 20 hours per complexity point

	// Stack-specific multipliers
	stackMultiplier := 1.0
	for _, stack := range stacks {
		switch strings.ToLower(stack.Name) {
		case "go", "python":
			stackMultiplier += 0.1
		case "react", "angular", "vue":
			stackMultiplier += 0.15
		case "kubernetes", "terraform":
			stackMultiplier += 0.2
		case "graphql":
			stackMultiplier += 0.15
		case "ml", "ai":
			stackMultiplier += 0.3
		}
	}

	totalHours := baseHours * stackMultiplier

	// Team efficiency (diminishing returns for larger teams)
	teamEfficiency := 1.0
	if teamSize > 1 {
		teamEfficiency = 1.0 / (1.0 + float64(teamSize-1)*0.15)
	}

	totalHours /= teamEfficiency

	// Phase breakdown
	breakdown := calculatePhaseBreakdown(totalHours)

	// Calculate total weeks
	hoursPerWeek := DefaultVelocityAssumptions().HoursPerWeek * float64(teamSize)
	totalWeeks := totalHours / hoursPerWeek

	return TimeEstimate{
		TotalWeeks:          formatWeeks(totalWeeks),
		TotalHours:          totalHours,
		Breakdown:           breakdown,
		TeamSize:            teamSize,
		VelocityAssumptions: DefaultVelocityAssumptions(),
		ComponentsAnalyzed:  components,
	}
}

// calculatePhaseBreakdown returns time distribution across phases
func calculatePhaseBreakdown(totalHours float64) []PhaseBreakdown {
	// Standard phase distribution percentages
	phasePercentages := map[Phase]float64{
		PhaseSetup:        0.10, // 10%
		PhaseCoreFeatures: 0.35, // 35%
		PhaseIntegration:  0.20, // 20%
		PhaseTesting:      0.15, // 15%
		PhasePolish:       0.12, // 12%
		PhaseDeployment:   0.08, // 8%
	}

	breakdown := make([]PhaseBreakdown, 0, len(phasePercentages))

	phases := []Phase{
		PhaseSetup,
		PhaseCoreFeatures,
		PhaseIntegration,
		PhaseTesting,
		PhasePolish,
		PhaseDeployment,
	}

	for _, phase := range phases {
		percentage := phasePercentages[phase]
		hours := totalHours * percentage
		weeks := hours / (DefaultVelocityAssumptions().HoursPerWeek) // per developer

		breakdown = append(breakdown, PhaseBreakdown{
			Phase:      phase,
			Weeks:      roundToHalf(weeks),
			Percentage: percentage * 100,
		})
	}

	return breakdown
}

// formatWeeks formats weeks as a human-readable string
func formatWeeks(weeks float64) string {
	if weeks < 1 {
		return fmt.Sprintf("%.0f days", weeks*5)
	}

	wholeWeeks := int(weeks)
	partialWeek := weeks - float64(wholeWeeks)

	if partialWeek < 0.25 {
		return fmt.Sprintf("%d week%s", wholeWeeks, plural(wholeWeeks))
	} else if partialWeek < 0.75 {
		return fmt.Sprintf("%d.5 week%s", wholeWeeks, plural(wholeWeeks))
	} else {
		return fmt.Sprintf("%d week%s", wholeWeeks+1, plural(wholeWeeks+1))
	}
}

// roundToHalf rounds a float to nearest 0.5
func roundToHalf(f float64) float64 {
	return float64(int(f*2)) / 2
}

// plural returns "s" if n is not 1
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// EstimateFromProjectSpec creates an estimate from a simplified project specification
func EstimateFromProjectSpec(
	totalComponents int,
	avgComplexity string,
	teamSize int,
	stackNames []string,
) TimeEstimate {
	// Generate complexity list based on average
	complexities := make([]string, totalComponents)
	for i := range complexities {
		complexities[i] = avgComplexity
	}

	// Build tech stacks
	stacks := make([]TechStack, 0, len(stackNames))
	for _, name := range stackNames {
		stacks = append(stacks, TechStack{
			Name:       name,
			Complexity: "medium",
		})
	}

	return CalculateTimeEstimate(totalComponents, complexities, teamSize, stacks)
}
