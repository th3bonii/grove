// Package loop provides the GROVE Ready Report generation.
//
// GROVE-READY-REPORT generates a final report after loop execution,
// summarizing tasks completed, issues found, warnings, metrics, and
// the final production readiness decision.
package loop

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Gentleman-Programming/grove/internal/types"
)

// ReadyReportConfig contains configuration for report generation.
type ReadyReportConfig struct {
	// OutputPath is the path where the report will be saved.
	OutputPath string

	// IncludeMetrics includes execution metrics in the report.
	IncludeMetrics bool

	// IncludeSuggestions includes improvement suggestions.
	IncludeSuggestions bool
}

// DefaultReadyReportConfig returns the default configuration.
func DefaultReadyReportConfig() *ReadyReportConfig {
	return &ReadyReportConfig{
		IncludeMetrics:     true,
		IncludeSuggestions: true,
	}
}

// LoopMetrics contains execution metrics from the loop run.
type LoopMetrics struct {
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	TotalDuration  time.Duration `json:"total_duration"`
	TotalTasks     int           `json:"total_tasks"`
	CompletedTasks int           `json:"completed_tasks"`
	FailedTasks    int           `json:"failed_tasks"`
	Retries        int           `json:"retries"`
	TotalTokens    int64         `json:"total_tokens"`
}

// ReadyReportGenerator generates the GROVE-READY-REPORT.
type ReadyReportGenerator struct {
	config *ReadyReportConfig
}

// NewReadyReportGenerator creates a new ReadyReportGenerator.
func NewReadyReportGenerator(config *ReadyReportConfig) *ReadyReportGenerator {
	if config == nil {
		config = DefaultReadyReportConfig()
	}
	return &ReadyReportGenerator{
		config: config,
	}
}

// GenerateReadyReport generates the final GROVE-READY-REPORT.
func (r *ReadyReportGenerator) GenerateReadyReport(tasks []Task, metrics *LoopMetrics) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("# GROVE-READY-REPORT\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	// Executive Summary
	sb.WriteString("## Executive Summary\n\n")

	completedCount := 0
	failedCount := 0
	var failedTaskIDs []string

	for _, task := range tasks {
		if task.Completed {
			completedCount++
		} else {
			failedCount++
			failedTaskIDs = append(failedTaskIDs, task.ID)
		}
	}

	totalTasks := len(tasks)
	successRate := 0.0
	if totalTasks > 0 {
		successRate = float64(completedCount) / float64(totalTasks)
	}

	// Determine overall status
	overallStatus := r.determineStatus(successRate, failedCount, tasks)

	sb.WriteString(fmt.Sprintf("- **Total Tasks**: %d\n", totalTasks))
	sb.WriteString(fmt.Sprintf("- **Completed**: %d\n", completedCount))
	sb.WriteString(fmt.Sprintf("- **Failed**: %d\n", failedCount))
	sb.WriteString(fmt.Sprintf("- **Success Rate**: %.1f%%\n\n", successRate*100))

	// Status Badge
	sb.WriteString(fmt.Sprintf("### Status: %s\n\n", overallStatus))

	// Tasks Summary
	sb.WriteString("## Tasks Summary\n\n")
	sb.WriteString("| Task ID | Title | Phase | Status |\n")
	sb.WriteString("|---------|-------|-------|--------|\n")

	for _, task := range tasks {
		status := "✅ Completed"
		if !task.Completed {
			status = "❌ Failed"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			task.ID, task.Title, task.Phase, status))
	}
	sb.WriteString("\n")

	// Issues Found
	if len(failedTaskIDs) > 0 {
		sb.WriteString("## Issues Found\n\n")
		for _, taskID := range failedTaskIDs {
			// Find the task to get details
			for _, task := range tasks {
				if task.ID == taskID {
					sb.WriteString(fmt.Sprintf("### Task: %s - %s\n", taskID, task.Title))
					sb.WriteString(fmt.Sprintf("- Phase: %s\n", task.Phase))
					sb.WriteString("\n")
					break
				}
			}
		}
	}

	// Warnings (incomplete tasks that weren't marked as failed)
	var pendingWarnings []string
	for _, task := range tasks {
		if !task.Completed && !contains(failedTaskIDs, task.ID) {
			pendingWarnings = append(pendingWarnings, fmt.Sprintf("- Task %s (%s) was not completed", task.ID, task.Title))
		}
	}

	if len(pendingWarnings) > 0 {
		sb.WriteString("## Warnings\n\n")
		for _, w := range pendingWarnings {
			sb.WriteString(w + "\n")
		}
		sb.WriteString("\n")
	}

	// Metrics (if enabled)
	if r.config.IncludeMetrics && metrics != nil {
		sb.WriteString("## Loop Metrics\n\n")
		sb.WriteString(fmt.Sprintf("- **Start Time**: %s\n", metrics.StartTime.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("- **End Time**: %s\n", metrics.EndTime.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("- **Total Duration**: %s\n", metrics.TotalDuration.String()))
		sb.WriteString(fmt.Sprintf("- **Total Tasks**: %d\n", metrics.TotalTasks))
		sb.WriteString(fmt.Sprintf("- **Completed Tasks**: %d\n", metrics.CompletedTasks))
		sb.WriteString(fmt.Sprintf("- **Failed Tasks**: %d\n", metrics.FailedTasks))
		sb.WriteString(fmt.Sprintf("- **Retries**: %d\n", metrics.Retries))
		if metrics.TotalTokens > 0 {
			sb.WriteString(fmt.Sprintf("- **Total Tokens**: %d\n", metrics.TotalTokens))
		}
		sb.WriteString("\n")
	}

	// Final Decision
	sb.WriteString("## Final Decision\n\n")
	if overallStatus == "PRODUCTION READY ✓" {
		sb.WriteString("**PRODUCTION READY ✓**\n\n")
		sb.WriteString("The implementation has completed successfully and all tasks have been verified.\n")
		sb.WriteString("The project is ready for production deployment.\n")
	} else {
		sb.WriteString("**Issues Pending**\n\n")
		sb.WriteString("The following issues need to be resolved before production deployment:\n\n")
		for _, taskID := range failedTaskIDs {
			sb.WriteString(fmt.Sprintf("- Task %s failed\n", taskID))
		}
		if len(pendingWarnings) > 0 {
			sb.WriteString("\nAdditionally, the following tasks were not completed:\n")
			for _, w := range pendingWarnings {
				sb.WriteString(w + "\n")
			}
		}
	}
	sb.WriteString("\n")

	// Recommendations (if enabled)
	if r.config.IncludeSuggestions && overallStatus != "PRODUCTION READY ✓" {
		sb.WriteString("## Recommendations\n\n")
		sb.WriteString("To achieve production readiness:\n")
		sb.WriteString("1. Review and fix the failed tasks\n")
		sb.WriteString("2. Run verification on completed tasks\n")
		sb.WriteString("3. Ensure all acceptance criteria are met\n")
		if len(pendingWarnings) > 0 {
			sb.WriteString("4. Complete or remove incomplete tasks\n")
		}
		sb.WriteString("\n")
	}

	// Save to file if output path is configured
	if r.config.OutputPath != "" {
		if err := os.WriteFile(r.config.OutputPath, []byte(sb.String()), 0644); err != nil {
			return sb.String(), fmt.Errorf("failed to write report to %s: %w", r.config.OutputPath, err)
		}
	}

	return sb.String(), nil
}

// determineStatus determines the overall production readiness status.
func (r *ReadyReportGenerator) determineStatus(successRate float64, failedCount int, tasks []Task) string {
	// If no tasks were executed
	if len(tasks) == 0 {
		return "NO TASKS EXECUTED"
	}

	// If any task failed, not ready
	if failedCount > 0 {
		return "Issues Pending: " + fmt.Sprintf("%d task(s) failed", failedCount)
	}

	// If success rate is below 80%, not ready
	if successRate < 0.8 {
		return "INCOMPLETE"
	}

	// If success rate is 100%, production ready
	if successRate == 1.0 {
		return "PRODUCTION READY ✓"
	}

	// Otherwise, ready with warnings
	return "PRODUCTION READY ✓ (with warnings)"
}

// contains checks if a string is in a slice.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GenerateReadyReport is a convenience function that creates a generator and generates the report.
func GenerateReadyReport(tasks []Task, metrics *LoopMetrics) (string, error) {
	config := DefaultReadyReportConfig()
	generator := NewReadyReportGenerator(config)
	return generator.GenerateReadyReport(tasks, metrics)
}

// GenerateReadyReportToFile generates the report and saves it to a file.
func GenerateReadyReportToFile(tasks []Task, metrics *LoopMetrics, outputPath string) (string, error) {
	config := &ReadyReportConfig{
		OutputPath:         outputPath,
		IncludeMetrics:     true,
		IncludeSuggestions: true,
	}
	generator := NewReadyReportGenerator(config)
	return generator.GenerateReadyReport(tasks, metrics)
}

// VerifyReportToTask converts a types.VerifyReport to a Task for reporting purposes.
func VerifyReportToTask(report *types.VerifyReport) Task {
	return Task{
		ID:          report.TaskID,
		Title:       fmt.Sprintf("Verification for %s", report.TaskID),
		Phase:       "verify",
		Completed:   report.Status == types.VerifyStatusPassed,
		Description: report.Message,
	}
}
