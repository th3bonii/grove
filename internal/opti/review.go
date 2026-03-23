package opti

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReviewResult represents the outcome of a review interaction.
type ReviewResult struct {
	Original  string           `json:"original"`         // Original user input
	Optimized *OptimizedPrompt `json:"optimized"`        // The optimized prompt
	Action    string           `json:"action"`           // send, edit, or reject
	Edited    string           `json:"edited,omitempty"` // User's edited version if they chose to edit
}

// PresentForReview displays the optimized prompt and waits for user action.
// In CLI mode, it shows a side-by-side comparison and waits for input.
// Returns a ReviewResult with the user's chosen action.
func PresentForReview(optimized *OptimizedPrompt) *ReviewResult {
	result := &ReviewResult{
		Original:  optimized.Original,
		Optimized: optimized,
	}

	// Display the review interface
	displayReviewUI(optimized)

	// Read user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n[Accept] [Edit] [Cancel]: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "accept", "a", "":
		result.Action = "send"
	case "edit", "e":
		result.Action = "edit"
		fmt.Println("\nEnter your edited prompt (press Ctrl+D to finish):")
		var edited strings.Builder
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			edited.WriteString(scanner.Text())
			edited.WriteString("\n")
		}
		result.Edited = strings.TrimSpace(edited.String())
		if result.Edited == "" {
			result.Edited = optimized.Optimized
		}
	case "cancel", "c", "reject", "r":
		result.Action = "reject"
	default:
		// Default to edit for unknown input
		result.Action = "edit"
		result.Edited = optimized.Optimized
	}

	return result
}

// displayReviewUI renders the review interface to stdout.
func displayReviewUI(optimized *OptimizedPrompt) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("GROVE OPTIMIZER - Prompt Review")
	fmt.Println(strings.Repeat("=", 70))

	// Original input
	fmt.Println("\n📝 ORIGINAL:")
	fmt.Println("─" + strings.Repeat("─", 50))
	fmt.Println(truncateString(optimized.Original, 200))

	// Optimized prompt
	fmt.Println("\n✨ OPTIMIZED:")
	fmt.Println("─" + strings.Repeat("─", 50))
	fmt.Println(optimized.Optimized)

	// Elements breakdown
	fmt.Println("\n📋 ELEMENTS BREAKDOWN:")
	fmt.Println("─" + strings.Repeat("─", 50))
	for _, elem := range optimized.Elements {
		fmt.Printf("  [%s] %s\n", elem.Type, truncateString(elem.Content, 60))
		if elem.Explanation != "" {
			fmt.Printf("      └─ %s\n", truncateString(elem.Explanation, 80))
		}
	}

	// Warnings
	if len(optimized.Warnings) > 0 {
		fmt.Println("\n⚠️  WARNINGS:")
		fmt.Println("─" + strings.Repeat("─", 50))
		for _, w := range optimized.Warnings {
			fmt.Printf("  • %s\n", w)
		}
	}

	// Stats
	fmt.Println("\n📊 STATS:")
	fmt.Println("─" + strings.Repeat("─", 50))
	fmt.Printf("  Tokens: %d / %d budget\n", optimized.TokenCount, optimized.TokenBudget)
	fmt.Printf("  Elements: %d\n", len(optimized.Elements))
	if len(optimized.SkillsUsed) > 0 {
		fmt.Printf("  Skills: %s\n", strings.Join(optimized.SkillsUsed, ", "))
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
}

// truncateString truncates a string to maxLen, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > 1 {
		s = strings.Join(lines[:2], "...")
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ReviewWithExplainer is a review function that uses the Explainer for adaptive explanations.
func ReviewWithExplainer(explainer *Explainer, optimized *OptimizedPrompt) *ReviewResult {
	result := &ReviewResult{
		Original:  optimized.Original,
		Optimized: optimized,
	}

	// Display with adaptive explanations
	displayReviewUIWithExplainer(explainer, optimized)

	// Read user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n[Accept] [Edit] [Cancel]: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "accept", "a", "":
		result.Action = "send"
	case "edit", "e":
		result.Action = "edit"
		fmt.Println("\nEnter your edited prompt (press Enter for default):")
		edited, _ := reader.ReadString('\n')
		edited = strings.TrimSpace(edited)
		if edited == "" {
			result.Edited = optimized.Optimized
		} else {
			result.Edited = edited
		}
	case "cancel", "c", "reject", "r":
		result.Action = "reject"
	default:
		result.Action = "edit"
		result.Edited = optimized.Optimized
	}

	return result
}

// displayReviewUIWithExplainer renders the review interface with adaptive explanations.
func displayReviewUIWithExplainer(explainer *Explainer, optimized *OptimizedPrompt) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("GROVE OPTIMIZER - Adaptive Prompt Review")
	fmt.Println(strings.Repeat("=", 70))

	// Original input
	fmt.Println("\n📝 ORIGINAL:")
	fmt.Println("─" + strings.Repeat("─", 50))
	fmt.Println(truncateString(optimized.Original, 200))

	// Optimized prompt
	fmt.Println("\n✨ OPTIMIZED:")
	fmt.Println("─" + strings.Repeat("─", 50))
	fmt.Println(optimized.Optimized)

	// Elements breakdown with ADAPTIVE explanations
	fmt.Println("\n📋 ADAPTIVE EXPLANATIONS:")
	fmt.Println("─" + strings.Repeat("─", 50))
	fmt.Println("(Explanation depth adapts to your experience level)")
	fmt.Println()

	for _, elem := range optimized.Elements {
		category := string(elem.Type)
		profile := explainer.getCategoryProfile(category)

		// Use adaptive explanation
		explanation := explainer.GenerateAdapted(category, profile)

		// Show element
		contentPreview := truncateString(elem.Content, 50)
		if len(contentPreview) < len(elem.Content) {
			contentPreview += "..."
		}

		// Determine indicator based on times_seen
		indicator := "📖" // Full explanation
		if profile.TimesSeen > 3 {
			indicator = "📝" // Short reminder
		}
		if profile.TimesSeen > 10 {
			indicator = "⚡" // Expert (label only)
		}

		fmt.Printf("  %s [%s] %s\n", indicator, category, contentPreview)
		fmt.Printf("     └─ %s\n", explanation)
		fmt.Printf("       (seen %d times)\n", profile.TimesSeen)
	}

	// Warnings
	if len(optimized.Warnings) > 0 {
		fmt.Println("\n⚠️  WARNINGS:")
		fmt.Println("─" + strings.Repeat("─", 50))
		for _, w := range optimized.Warnings {
			fmt.Printf("  • %s\n", w)
		}
	}

	// Legend
	fmt.Println("\n📚 LEGEND:")
	fmt.Println("  📖 Full explanation (new pattern)")
	fmt.Println("  📝 Short reminder (familiar)")
	fmt.Println("  ⚡ Label only (expert)")

	fmt.Println("\n" + strings.Repeat("=", 70))
}
