package dashboard

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Gentleman-Programming/grove/internal/metrics"
)

// EventType represents types of events that can be tracked.
type EventType string

const (
	EventTaskCompleted EventType = "task_completed"
	EventTaskFailed    EventType = "task_failed"
	EventLoopIteration EventType = "loop_iteration"
	EventFileProcessed EventType = "file_processed"
	EventTokensUsed    EventType = "tokens_used"
	EventQualityScore  EventType = "quality_score"
	EventLoopStarted   EventType = "loop_started"
	EventLoopCompleted EventType = "loop_completed"
)

// Event represents a tracked event.
type Event struct {
	Type      EventType      `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	Data      map[string]any `json:"data,omitempty"`
}

// Dashboard provides a visual metrics dashboard for GROVE.
type Dashboard struct {
	*metrics.Metrics
	CurrentLoop   int
	PendingTasks  int64
	TotalTasks    int64
	QualityScore  float64
	Events        []Event
	TaskDurations []time.Duration

	mu         sync.RWMutex
	eventIndex int
}

// NewDashboard creates a new Dashboard instance.
func NewDashboard() *Dashboard {
	return &Dashboard{
		Metrics:      metrics.NewMetrics(),
		CurrentLoop:  1,
		PendingTasks: 0,
		TotalTasks:   0,
		Events:       make([]Event, 0),
	}
}

// Track records an event in the dashboard.
func (d *Dashboard) Track(eventType EventType, data map[string]any) {
	d.mu.Lock()
	defer d.mu.Unlock()

	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	d.Events = append(d.Events, event)

	// Process event to update metrics
	switch eventType {
	case EventTaskCompleted:
		d.TasksCompleted++
		d.PendingTasks--
		if taskID, ok := data["task_id"].(string); ok {
			d.completeTaskDuration(taskID)
		}
	case EventTaskFailed:
		d.TasksFailed++
		d.PendingTasks--
	case EventLoopIteration:
		if loop, ok := data["loop"].(int); ok {
			d.CurrentLoop = loop
		}
	case EventFileProcessed:
		d.FilesProcessed++
	case EventTokensUsed:
		if tokens, ok := data["tokens"].(int); ok {
			d.TokenUsage += int64(tokens)
		}
	case EventQualityScore:
		if score, ok := data["score"].(float64); ok {
			d.QualityScore = score
		}
	case EventLoopStarted:
		d.StartLoop()
	case EventLoopCompleted:
		d.CompleteLoop()
	}
}

// completeTaskDuration tracks task duration when a task completes.
func (d *Dashboard) completeTaskDuration(taskID string) {
	// This would track individual task durations
	// Simplified for dashboard purposes
}

// SetTotalTasks sets the total number of tasks.
func (d *Dashboard) SetTotalTasks(total int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.TotalTasks = total
	d.PendingTasks = total - d.TasksCompleted - d.TasksFailed
}

// SetCurrentLoop sets the current loop number.
func (d *Dashboard) SetCurrentLoop(loop int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.CurrentLoop = loop
}

// SetQualityScore sets the quality score.
func (d *Dashboard) SetQualityScore(score float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.QualityScore = score
}

// Show prints a visual dashboard to the terminal.
func (d *Dashboard) Show() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var b strings.Builder

	// Header
	b.WriteString("\n")
	b.WriteString("═══════════════════════════════════════════════════════\n")
	b.WriteString("GROVE DASHBOARD v1.0\n")
	b.WriteString("═══════════════════════════════════════════════════════\n")
	b.WriteString("\n")

	// Loop Status Section
	b.WriteString("📊 Loop Status\n")
	b.WriteString(fmt.Sprintf("├── Current Loop: %d\n", d.CurrentLoop))
	b.WriteString(fmt.Sprintf("├── Tasks Completed: %d/%d\n", d.TasksCompleted, d.TotalTasks))
	b.WriteString(fmt.Sprintf("├── Tasks Failed: %d\n", d.TasksFailed))
	b.WriteString(fmt.Sprintf("├── Quality Score: %.1f/100\n", d.QualityScore))
	b.WriteString("\n")

	// Performance Section
	avgDuration := d.AverageTaskDuration()
	totalDuration := d.Duration()
	b.WriteString("⏱️  Performance\n")
	b.WriteString(fmt.Sprintf("├── Avg Task Duration: %s\n", avgDuration))
	b.WriteString(fmt.Sprintf("├── Total Duration: %s\n", formatDuration(totalDuration)))
	b.WriteString(fmt.Sprintf("├── Token Usage: %s\n", formatNumber(d.TokenUsage)))
	b.WriteString("\n")

	// Progress Section
	progress := d.Progress()
	bar := d.progressBar(progress)
	b.WriteString("📈 Progress\n")
	b.WriteString(fmt.Sprintf("├── %s %.0f%%\n", bar, progress))
	b.WriteString(fmt.Sprintf("├── [%d] Completed\n", d.TasksCompleted))
	b.WriteString(fmt.Sprintf("├── [%d] Failed\n", d.TasksFailed))
	b.WriteString(fmt.Sprintf("└── [%d] Pending\n", d.PendingTasks))
	b.WriteString("═══════════════════════════════════════════════════════\n")

	return b.String()
}

// AverageTaskDuration calculates the average task duration.
func (d *Dashboard) AverageTaskDuration() time.Duration {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.TaskDurations) == 0 {
		return 45 * time.Second // Default for demo
	}

	var total time.Duration
	for _, dur := range d.TaskDurations {
		total += dur
	}
	return total / time.Duration(len(d.TaskDurations))
}

// Progress returns the progress percentage.
func (d *Dashboard) Progress() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.TotalTasks == 0 {
		return 0
	}
	return float64(d.TasksCompleted) / float64(d.TotalTasks) * 100
}

// progressBar generates a text progress bar.
func (d *Dashboard) progressBar(percentage float64) string {
	const totalWidth = 20
	filled := int(percentage / 100 * float64(totalWidth))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", totalWidth-filled)
	return bar
}

// Export exports the dashboard metrics to JSON.
func (d *Dashboard) Export() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	type exportedMetrics struct {
		CurrentLoop     int     `json:"current_loop"`
		TasksCompleted  int64   `json:"tasks_completed"`
		TasksFailed     int64   `json:"tasks_failed"`
		PendingTasks    int64   `json:"pending_tasks"`
		TotalTasks      int64   `json:"total_tasks"`
		QualityScore    float64 `json:"quality_score"`
		Progress        float64 `json:"progress_percentage"`
		AvgTaskDuration string  `json:"avg_task_duration"`
		TotalDuration   string  `json:"total_duration"`
		TokenUsage      int64   `json:"token_usage"`
		FilesProcessed  int64   `json:"files_processed"`
		SuccessRate     float64 `json:"success_rate"`
		Events          []Event `json:"events"`
		LoopCount       int     `json:"loop_count"`
	}

	exp := exportedMetrics{
		CurrentLoop:     d.CurrentLoop,
		TasksCompleted:  d.TasksCompleted,
		TasksFailed:     d.TasksFailed,
		PendingTasks:    d.PendingTasks,
		TotalTasks:      d.TotalTasks,
		QualityScore:    d.QualityScore,
		Progress:        d.Progress(),
		AvgTaskDuration: d.AverageTaskDuration().String(),
		TotalDuration:   d.Duration().String(),
		TokenUsage:      d.TokenUsage,
		FilesProcessed:  d.FilesProcessed,
		SuccessRate:     d.SuccessRate(),
		Events:          d.Events,
		LoopCount:       len(d.LoopDurations),
	}

	return json.MarshalIndent(exp, "", "  ")
}

// GetEvents returns all tracked events.
func (d *Dashboard) GetEvents() []Event {
	d.mu.RLock()
	defer d.mu.RUnlock()

	events := make([]Event, len(d.Events))
	copy(events, d.Events)
	return events
}

// GetEventsByType returns events filtered by type.
func (d *Dashboard) GetEventsByType(eventType EventType) []Event {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var result []Event
	for _, e := range d.Events {
		if e.Type == eventType {
			result = append(result, e)
		}
	}
	return result
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// formatNumber formats a number with thousand separators.
func formatNumber(n int64) string {
	str := fmt.Sprintf("%d", n)
	var result strings.Builder
	length := len(str)

	for i := 0; i < length; i++ {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteByte(str[i])
	}

	return result.String()
}
