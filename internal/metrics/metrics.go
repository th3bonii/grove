package metrics

import (
	"fmt"
	"sync"
	"time"
)

// Metrics holds performance metrics for the Ralph Loop.
type Metrics struct {
	TasksCompleted int64
	TasksFailed    int64
	LoopDurations  []time.Duration
	TokenUsage     int64
	FilesProcessed int64
	StartTime      time.Time
	EndTime        time.Time

	mu         sync.RWMutex
	taskTimers map[string]time.Time
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		taskTimers: make(map[string]time.Time),
	}
}

// StartTask records the start time for a task.
func (m *Metrics) StartTask(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskTimers[taskID] = time.Now()
}

// CompleteTask records task completion and updates counters.
func (m *Metrics) CompleteTask(taskID string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.taskTimers[taskID]; exists {
		if success {
			m.TasksCompleted++
		} else {
			m.TasksFailed++
		}
		delete(m.taskTimers, taskID)
	}
}

// StartLoop marks the beginning of a loop iteration.
func (m *Metrics) StartLoop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StartTime = time.Now()
}

// CompleteLoop marks the end of a loop iteration and records duration.
func (m *Metrics) CompleteLoop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EndTime = time.Now()
	if !m.StartTime.IsZero() {
		m.LoopDurations = append(m.LoopDurations, m.EndTime.Sub(m.StartTime))
	}
}

// AddTokens increments the token usage counter.
func (m *Metrics) AddTokens(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TokenUsage += int64(n)
}

// AddFiles increments the files processed counter.
func (m *Metrics) AddFiles(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FilesProcessed += int64(n)
}

// Duration returns the total elapsed time from StartTime to EndTime.
func (m *Metrics) Duration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.EndTime.IsZero() {
		return time.Since(m.StartTime)
	}
	return m.EndTime.Sub(m.StartTime)
}

// AverageLoopDuration returns the average loop duration.
func (m *Metrics) AverageLoopDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.LoopDurations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range m.LoopDurations {
		total += d
	}
	return total / time.Duration(len(m.LoopDurations))
}

// TotalTasks returns the total number of tasks attempted.
func (m *Metrics) TotalTasks() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.TasksCompleted + m.TasksFailed
}

// SuccessRate returns the success rate as a percentage.
func (m *Metrics) SuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.TasksCompleted + m.TasksFailed
	if total == 0 {
		return 0
	}
	return float64(m.TasksCompleted) / float64(total) * 100
}

// Report generates a markdown-formatted metrics report.
func (m *Metrics) Report() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report := fmt.Sprintf(`## Ralph Loop Metrics Report

### Task Statistics
- **Tasks Completed:** %d
- **Tasks Failed:** %d
- **Total Tasks:** %d
- **Success Rate:** %.2f%%

### Performance
- **Total Duration:** %s
- **Average Loop Duration:** %s
- **Loop Iterations:** %d

### Resource Usage
- **Token Usage:** %d
- **Files Processed:** %d

### Loop Duration History
`, m.TasksCompleted, m.TasksFailed, m.TotalTasks(), m.SuccessRate(),
		m.Duration().Round(time.Millisecond),
		m.AverageLoopDuration().Round(time.Millisecond),
		len(m.LoopDurations),
		m.TokenUsage,
		m.FilesProcessed)

	if len(m.LoopDurations) == 0 {
		report += "> No loop iterations recorded\n"
	} else {
		for i, d := range m.LoopDurations {
			report += fmt.Sprintf("- Iteration %d: %s\n", i+1, d.Round(time.Millisecond))
		}
	}

	return report
}
