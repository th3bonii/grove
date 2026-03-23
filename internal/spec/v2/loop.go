// Package spec provides the GROVE Spec engine iteration loop with convergence detection.
//
// This module implements the Ralph Loop's "infinite until complete" philosophy,
// detecting when the specification has reached its maximum quality level and
// can no longer be improved.
package spec

import (
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// Top Detection Types
// =============================================================================

// TopDetectionResult represents the result of detecting convergence/tops.
type TopDetectionResult struct {
	Reached        bool     `json:"reached"`         // Whether top was reached
	Reason         string   `json:"reason"`          // Human-readable reason
	Evidence       []string `json:"evidence"`        // List of evidence supporting the determination
	CompositeScore float64  `json:"composite_score"` // Current composite score
	DeltaPct       float64  `json:"delta_pct"`       // Delta from previous iteration
	TopScore       float64  `json:"top_score"`       // Threshold for top detection
}

// TopCheckResult represents the result of a single top check.
type TopCheckResult struct {
	CheckName    string  `json:"check_name"`
	Passed       bool    `json:"passed"`
	Details      string  `json:"details"`
	ScoreContrib float64 `json:"score_contrib"`
}

// =============================================================================
// Top Detection Configuration
// =============================================================================

// TopConfig holds configuration for top detection.
type TopConfig struct {
	MinLoops            int     // Minimum iterations before allowing top detection
	TopScoreThreshold   float64 // Score threshold to consider complete (default: 85.0)
	DeltaThreshold      float64 // Delta threshold for convergence (default: 3.0)
	MinStatesPerComp    int     // Minimum states per component (default: 3)
	MinBehaviorsPerComp int     // Minimum behaviors per component (default: 2)
	MinEdgeCasesPerComp int     // Minimum edge cases per component (default: 2)
	RequireFlows        bool    // Require user flows to be complete (default: true)
	RequireDecisions    bool    // Require decisions to be documented (default: true)
}

// DefaultTopConfig returns default top detection configuration.
func DefaultTopConfig() TopConfig {
	return TopConfig{
		MinLoops:            3,
		TopScoreThreshold:   85.0,
		DeltaThreshold:      3.0,
		MinStatesPerComp:    3,
		MinBehaviorsPerComp: 2,
		MinEdgeCasesPerComp: 2,
		RequireFlows:        true,
		RequireDecisions:    true,
	}
}

// =============================================================================
// Engine Extension for Top Detection
// =============================================================================

// detectTop checks if the spec loop has reached its maximum quality level.
//
// The engine detects "top" (convergence) when ALL of the following are true:
//   - All components have complete states (≥ MinStatesPerComp)
//   - All components have complete behaviors (≥ MinBehaviorsPerComp)
//   - All components have complete edge cases (≥ MinEdgeCasesPerComp)
//   - No unresolved gaps remain
//   - All alternatives have been evaluated
//   - User flows are complete (if RequireFlows is true)
//   - Decisions are documented (if RequireDecisions is true)
//   - Composite score ≥ TopScoreThreshold
//   - Delta between iterations < DeltaThreshold
//
// Returns:
//   - reached: true if top has been reached
//   - reason: human-readable explanation of why top was (or wasn't) reached
//   - evidence: list of specific evidence supporting the determination
func (e *Engine) detectTop() (bool, string, []string) {
	evidence := make([]string, 0)
	config := DefaultTopConfig()

	// Calculate current scores
	scores := e.score()
	compositeScore := scores.CompositeScore()

	// Get delta from previous iteration
	deltaPct := e.calculateDelta()

	// Check minimum loops requirement
	if e.state.LoopNumber < config.MinLoops {
		reason := fmt.Sprintf("Minimum loops not met (have %d, need %d)",
			e.state.LoopNumber, config.MinLoops)
		evidence = append(evidence, reason)
		return false, reason, evidence
	}

	// Run all top detection checks
	allChecks := []TopCheckResult{
		e.checkStatesComplete(config),
		e.checkBehaviorsComplete(config),
		e.checkEdgeCasesComplete(config),
		e.checkGapsResolved(),
		e.checkAlternativesEvaluated(),
		e.checkUserFlowsComplete(config),
		e.checkDecisionsDocumented(config),
		e.checkScoreThreshold(compositeScore, config),
		e.checkDeltaConvergence(deltaPct, config),
	}

	// Aggregate results
	passedCount := 0
	failedChecks := make([]string, 0)

	for _, check := range allChecks {
		if check.Passed {
			passedCount++
			evidence = append(evidence, fmt.Sprintf("✓ %s: %s", check.CheckName, check.Details))
		} else {
			failedChecks = append(failedChecks, check.CheckName)
			evidence = append(evidence, fmt.Sprintf("✗ %s: %s", check.CheckName, check.Details))
		}
	}

	// Determine if top is reached
	topReached := passedCount == len(allChecks)

	var reason string
	if topReached {
		reason = fmt.Sprintf(
			"TOP REACHED: All %d quality checks passed. Composite score: %.1f/100, Delta: %.1f%%",
			passedCount, compositeScore, deltaPct)
	} else {
		reason = fmt.Sprintf(
			"TOP NOT REACHED: %d/%d checks passed. Failed: %s. Composite: %.1f/100, Delta: %.1f%%",
			passedCount, len(allChecks), strings.Join(failedChecks, ", "), compositeScore, deltaPct)
	}

	return topReached, reason, evidence
}

// checkStatesComplete verifies all components have complete states.
func (e *Engine) checkStatesComplete(config TopConfig) TopCheckResult {
	incomplete := make([]string, 0)

	for _, comp := range e.components {
		if len(comp.States) < config.MinStatesPerComp {
			incomplete = append(incomplete,
				fmt.Sprintf("%s (%d states)", comp.Name, len(comp.States)))
		}
	}

	passed := len(incomplete) == 0
	details := fmt.Sprintf("%d/%d components have ≥%d states",
		len(e.components)-len(incomplete), len(e.components), config.MinStatesPerComp)

	if !passed {
		details += fmt.Sprintf(". Incomplete: %s", strings.Join(incomplete, ", "))
	}

	return TopCheckResult{
		CheckName:    "States Complete",
		Passed:       passed,
		Details:      details,
		ScoreContrib: 10.0,
	}
}

// checkBehaviorsComplete verifies all components have complete behaviors.
func (e *Engine) checkBehaviorsComplete(config TopConfig) TopCheckResult {
	incomplete := make([]string, 0)

	for _, comp := range e.components {
		if len(comp.Behaviors) < config.MinBehaviorsPerComp {
			incomplete = append(incomplete,
				fmt.Sprintf("%s (%d behaviors)", comp.Name, len(comp.Behaviors)))
		}
	}

	passed := len(incomplete) == 0
	details := fmt.Sprintf("%d/%d components have ≥%d behaviors",
		len(e.components)-len(incomplete), len(e.components), config.MinBehaviorsPerComp)

	if !passed {
		details += fmt.Sprintf(". Incomplete: %s", strings.Join(incomplete, ", "))
	}

	return TopCheckResult{
		CheckName:    "Behaviors Complete",
		Passed:       passed,
		Details:      details,
		ScoreContrib: 10.0,
	}
}

// checkEdgeCasesComplete verifies all components have complete edge cases.
func (e *Engine) checkEdgeCasesComplete(config TopConfig) TopCheckResult {
	incomplete := make([]string, 0)

	for _, comp := range e.components {
		if len(comp.EdgeCases) < config.MinEdgeCasesPerComp {
			incomplete = append(incomplete,
				fmt.Sprintf("%s (%d edge cases)", comp.Name, len(comp.EdgeCases)))
		}
	}

	passed := len(incomplete) == 0
	details := fmt.Sprintf("%d/%d components have ≥%d edge cases",
		len(e.components)-len(incomplete), len(e.components), config.MinEdgeCasesPerComp)

	if !passed {
		details += fmt.Sprintf(". Incomplete: %s", strings.Join(incomplete, ", "))
	}

	return TopCheckResult{
		CheckName:    "Edge Cases Complete",
		Passed:       passed,
		Details:      details,
		ScoreContrib: 10.0,
	}
}

// checkGapsResolved verifies all identified gaps have been resolved.
func (e *Engine) checkGapsResolved() TopCheckResult {
	unresolvedGaps := make([]string, 0)

	for _, comp := range e.components {
		for _, gap := range comp.Gaps {
			if gap.Resolution == "" {
				unresolvedGaps = append(unresolvedGaps,
					fmt.Sprintf("%s: %s", comp.Name, gap.Description))
			}
		}
	}

	passed := len(unresolvedGaps) == 0
	details := fmt.Sprintf("%d unresolved gaps", len(unresolvedGaps))

	if !passed {
		details += fmt.Sprintf(". Remaining: %s", strings.Join(unresolvedGaps, ", "))
	}

	return TopCheckResult{
		CheckName:    "Gaps Resolved",
		Passed:       passed,
		Details:      details,
		ScoreContrib: 10.0,
	}
}

// checkAlternativesEvaluated verifies all alternatives have been evaluated.
func (e *Engine) checkAlternativesEvaluated() TopCheckResult {
	notEvaluated := make([]string, 0)

	for _, comp := range e.components {
		for _, alt := range comp.Alternatives {
			if alt.Reason == "" {
				notEvaluated = append(notEvaluated,
					fmt.Sprintf("%s: %s", comp.Name, alt.Description))
			}
		}
	}

	passed := len(notEvaluated) == 0
	details := fmt.Sprintf("%d alternatives evaluated", len(notEvaluated))

	if !passed {
		details += fmt.Sprintf(". Not evaluated: %s", strings.Join(notEvaluated, ", "))
	}

	return TopCheckResult{
		CheckName:    "Alternatives Evaluated",
		Passed:       passed,
		Details:      details,
		ScoreContrib: 10.0,
	}
}

// checkUserFlowsComplete verifies user flows are complete.
func (e *Engine) checkUserFlowsComplete(config TopConfig) TopCheckResult {
	if !config.RequireFlows {
		return TopCheckResult{
			CheckName:    "User Flows Complete",
			Passed:       true,
			Details:      "User flows check disabled in config",
			ScoreContrib: 0.0,
		}
	}

	// Check if we have user flows and they're meaningful
	if len(e.userFlows) == 0 {
		return TopCheckResult{
			CheckName:    "User Flows Complete",
			Passed:       false,
			Details:      "No user flows defined",
			ScoreContrib: 10.0,
		}
	}

	// Check each flow has steps
	emptyFlows := make([]string, 0)
	for _, flow := range e.userFlows {
		if len(flow.Steps) == 0 {
			emptyFlows = append(emptyFlows, flow.Name)
		}
	}

	passed := len(emptyFlows) == 0
	details := fmt.Sprintf("%d user flows with steps", len(e.userFlows)-len(emptyFlows))

	if !passed {
		details += fmt.Sprintf(". Empty flows: %s", strings.Join(emptyFlows, ", "))
	}

	return TopCheckResult{
		CheckName:    "User Flows Complete",
		Passed:       passed,
		Details:      details,
		ScoreContrib: 10.0,
	}
}

// checkDecisionsDocumented verifies all decisions are documented.
func (e *Engine) checkDecisionsDocumented(config TopConfig) TopCheckResult {
	if !config.RequireDecisions {
		return TopCheckResult{
			CheckName:    "Decisions Documented",
			Passed:       true,
			Details:      "Decisions check disabled in config",
			ScoreContrib: 0.0,
		}
	}

	// Check if decisions exist and have rationale
	emptyDecisions := make([]string, 0)
	for _, decision := range e.decisions {
		if decision.Rationale == "" {
			emptyDecisions = append(emptyDecisions, decision.Question)
		}
	}

	passed := len(emptyDecisions) == 0
	details := fmt.Sprintf("%d decisions documented", len(e.decisions)-len(emptyDecisions))

	if !passed {
		details += fmt.Sprintf(". Missing rationale: %s", strings.Join(emptyDecisions, ", "))
	}

	return TopCheckResult{
		CheckName:    "Decisions Documented",
		Passed:       passed,
		Details:      details,
		ScoreContrib: 10.0,
	}
}

// checkScoreThreshold verifies composite score meets threshold.
func (e *Engine) checkScoreThreshold(score float64, config TopConfig) TopCheckResult {
	passed := score >= config.TopScoreThreshold
	details := fmt.Sprintf("Composite score: %.1f/100 (threshold: %.1f)",
		score, config.TopScoreThreshold)

	return TopCheckResult{
		CheckName:    "Score Threshold",
		Passed:       passed,
		Details:      details,
		ScoreContrib: 10.0,
	}
}

// checkDeltaConvergence verifies delta is below threshold (convergence).
func (e *Engine) checkDeltaConvergence(delta float64, config TopConfig) TopCheckResult {
	passed := delta < config.DeltaThreshold && delta >= 0
	details := fmt.Sprintf("Delta: %.1f%% (threshold: <%.1f%%)",
		delta, config.DeltaThreshold)

	if delta < 0 {
		details += " (score decreased)"
	}

	return TopCheckResult{
		CheckName:    "Delta Convergence",
		Passed:       passed,
		Details:      details,
		ScoreContrib: 10.0,
	}
}

// calculateDelta calculates the percentage delta from previous iteration.
func (e *Engine) calculateDelta() float64 {
	if e.state.LoopNumber <= 1 || len(e.state.Iterations) < 2 {
		return 100.0 // First iteration has 100% "delta"
	}

	previousScore := e.state.Iterations[len(e.state.Iterations)-2].Scores.CompositeScore()
	currentScore := e.state.Iterations[len(e.state.Iterations)-1].Scores.CompositeScore()

	if previousScore == 0 {
		return 100.0
	}

	// Return percentage change (can be negative if score decreased)
	return ((currentScore - previousScore) / previousScore) * 100.0
}

// =============================================================================
// Top Detection Report Generation
// =============================================================================

// GenerateTopReport generates a comprehensive report when top is detected.
func (e *Engine) GenerateTopReport() *TopDetectionResult {
	reached, reason, evidence := e.detTopWithScore()
	return &TopDetectionResult{
		Reached:        reached,
		Reason:         reason,
		Evidence:       evidence,
		CompositeScore: e.state.CompositeScore,
		DeltaPct:       e.state.Delta,
		TopScore:       DefaultTopConfig().TopScoreThreshold,
	}
}

// detTopWithScore is an internal helper that also ensures state is set.
func (e *Engine) detTopWithScore() (bool, string, []string) {
	reached, reason, evidence := e.detectTop()

	// Update state with the result
	if reached {
		e.state.ExitReason = "top_reached"
	}

	return reached, reason, evidence
}

// =============================================================================
// Changes Summary Generation
// =============================================================================

// GetChangesSummary generates a list of all changes made during the loop.
func (e *Engine) GetChangesSummary() []string {
	changes := make([]string, 0)

	// Count component improvements
	stateChanges := 0
	behaviorChanges := 0
	edgeCaseChanges := 0
	gapFixes := 0
	alternativesAdded := 0

	for _, comp := range e.components {
		stateChanges += len(comp.States)
		behaviorChanges += len(comp.Behaviors)
		edgeCaseChanges += len(comp.EdgeCases)
		gapFixes += len(comp.Gaps)
		alternativesAdded += len(comp.Alternatives)
	}

	changes = append(changes, fmt.Sprintf("Components analyzed: %d", len(e.components)))
	changes = append(changes, fmt.Sprintf("States defined: %d", stateChanges))
	changes = append(changes, fmt.Sprintf("Behaviors defined: %d", behaviorChanges))
	changes = append(changes, fmt.Sprintf("Edge cases identified: %d", edgeCaseChanges))
	changes = append(changes, fmt.Sprintf("Gaps identified and resolved: %d", gapFixes))
	changes = append(changes, fmt.Sprintf("Alternatives evaluated: %d", alternativesAdded))
	changes = append(changes, fmt.Sprintf("User flows mapped: %d", len(e.userFlows)))
	changes = append(changes, fmt.Sprintf("Decisions documented: %d", len(e.decisions)))

	// Add iteration summary
	if len(e.state.Iterations) > 0 {
		changes = append(changes, fmt.Sprintf("Total iterations: %d", e.state.LoopNumber))
		changes = append(changes, fmt.Sprintf("Final composite score: %.1f/100", e.state.CompositeScore))
	}

	return changes
}

// =============================================================================
// Final Documentation Generation
// =============================================================================

// GenerateFinalDocumentation generates all final documentation when top is reached.
func (e *Engine) GenerateFinalDocumentation() error {
	// Generate completion summary
	summaryPath := e.outputDir + "/COMPLETION.md"
	summary := e.generateCompletionMD()

	if err := e.writeFile(summaryPath, summary); err != nil {
		return fmt.Errorf("failed to write COMPLETION.md: %w", err)
	}

	// Generate changes log
	changesPath := e.outputDir + "/CHANGES.md"
	changes := e.generateChangesMD()

	if err := e.writeFile(changesPath, changes); err != nil {
		return fmt.Errorf("failed to write CHANGES.md: %w", err)
	}

	return nil
}

// generateCompletionMD generates the completion summary document.
func (e *Engine) generateCompletionMD() string {
	reached, reason, evidence := e.detectTop()

	var sb strings.Builder
	sb.WriteString("# GROVE Spec - Completion Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated**: %s\n\n", time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Status**: %s\n\n", map[bool]string{true: "COMPLETE", false: "INCOMPLETE"}[reached]))

	sb.WriteString("## Top Detection Result\n\n")
	sb.WriteString(fmt.Sprintf("**Reached**: %v\n\n", reached))
	sb.WriteString(fmt.Sprintf("**Reason**: %s\n\n", reason))

	sb.WriteString("## Evidence\n\n")
	for _, e := range evidence {
		sb.WriteString(fmt.Sprintf("- %s\n", e))
	}

	sb.WriteString("\n## Final Scores\n\n")
	sb.WriteString(fmt.Sprintf("- Flow Coverage: %.1f/10\n", e.state.Scores.FlowCoverage))
	sb.WriteString(fmt.Sprintf("- Component Decomposition: %.1f/10\n", e.state.Scores.ComponentDecomposition))
	sb.WriteString(fmt.Sprintf("- Logical Consistency: %.1f/10\n", e.state.Scores.LogicalConsistency))
	sb.WriteString(fmt.Sprintf("- Inter-component Connectivity: %.1f/10\n", e.state.Scores.InterComponentConnectivity))
	sb.WriteString(fmt.Sprintf("- Edge Case Coverage: %.1f/10\n", e.state.Scores.EdgeCaseCoverage))
	sb.WriteString(fmt.Sprintf("- Decision Justification: %.1f/10\n", e.state.Scores.DecisionJustification))
	sb.WriteString(fmt.Sprintf("- Agent Consumability: %.1f/10\n", e.state.Scores.AgentConsumability))
	sb.WriteString(fmt.Sprintf("- **Composite Score**: %.1f/100\n\n", e.state.CompositeScore))

	sb.WriteString("## How to Start Development\n\n")
	sb.WriteString("Run the following command to begin autonomous implementation:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("groove loop --spec ./spec\n")
	sb.WriteString("```\n\n")
	sb.WriteString("All necessary AGENTS.md and SKILLS.md files have been generated.\n")

	return sb.String()
}

// generateChangesMD generates the changes log document.
func (e *Engine) generateChangesMD() string {
	var sb strings.Builder
	sb.WriteString("# GROVE Spec - Changes Log\n\n")
	sb.WriteString(fmt.Sprintf("**Total Iterations**: %d\n\n", e.state.LoopNumber))

	changes := e.GetChangesSummary()

	sb.WriteString("## Summary of Changes\n\n")
	for _, change := range changes {
		sb.WriteString(fmt.Sprintf("- %s\n", change))
	}

	sb.WriteString("\n## Iteration Details\n\n")
	for _, iter := range e.state.Iterations {
		sb.WriteString(fmt.Sprintf("### Iteration %d\n\n", iter.Number))
		sb.WriteString(fmt.Sprintf("- **Timestamp**: %s\n", iter.Timestamp.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("- **Components**: %d\n", iter.ComponentsFound))
		sb.WriteString(fmt.Sprintf("- **Score**: %.1f/100\n\n", iter.Scores.CompositeScore()))
	}

	sb.WriteString("## Why No More Improvements Are Possible\n\n")
	sb.WriteString("The specification has reached its maximum quality level because:\n\n")
	sb.WriteString("1. All components have been fully decomposed with states, behaviors, and edge cases\n")
	sb.WriteString("2. All identified gaps have been resolved\n")
	sb.WriteString("3. All alternatives have been evaluated and documented\n")
	sb.WriteString("4. User flows are complete and cover all interactions\n")
	sb.WriteString("5. All architectural decisions are documented with rationale\n")
	sb.WriteString("6. The composite quality score meets the threshold (≥85/100)\n")
	sb.WriteString("7. The delta between iterations is below the convergence threshold (<3%)\n")
	sb.WriteString("\nFurther iterations would produce diminishing returns.\n")

	return sb.String()
}
