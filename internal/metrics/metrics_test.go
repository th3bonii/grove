package metrics

import (
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	if m == nil {
		t.Fatal("NewMetrics() returned nil")
	}

	if m.TasksCompleted != 0 {
		t.Errorf("expected TasksCompleted to be 0, got %d", m.TasksCompleted)
	}

	if m.TasksFailed != 0 {
		t.Errorf("expected TasksFailed to be 0, got %d", m.TasksFailed)
	}

	if m.TokenUsage != 0 {
		t.Errorf("expected TokenUsage to be 0, got %d", m.TokenUsage)
	}

	if m.FilesProcessed != 0 {
		t.Errorf("expected FilesProcessed to be 0, got %d", m.FilesProcessed)
	}

	if m.taskTimers == nil {
		t.Error("expected taskTimers to be initialized")
	}
}

func TestStartAndCompleteTask(t *testing.T) {
	tests := []struct {
		name    string
		taskID  string
		success bool
	}{
		{"successful task", "task-1", true},
		{"failed task", "task-2", false},
		{"another successful task", "task-3", true},
	}

	m := NewMetrics()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.StartTask(tt.taskID)
			time.Sleep(time.Millisecond) // Ensure some time passes
			m.CompleteTask(tt.taskID, tt.success)
		})
	}

	if m.TasksCompleted != 2 {
		t.Errorf("expected TasksCompleted to be 2, got %d", m.TasksCompleted)
	}

	if m.TasksFailed != 1 {
		t.Errorf("expected TasksFailed to be 1, got %d", m.TasksFailed)
	}
}

func TestCompleteTaskWithoutStart(t *testing.T) {
	m := NewMetrics()

	// Complete a task that was never started
	m.CompleteTask("unknown-task", true)

	if m.TasksCompleted != 0 {
		t.Errorf("expected TasksCompleted to be 0, got %d", m.TasksCompleted)
	}
}

func TestStartLoopAndCompleteLoop(t *testing.T) {
	m := NewMetrics()

	m.StartLoop()
	time.Sleep(10 * time.Millisecond)
	m.CompleteLoop()

	if len(m.LoopDurations) != 1 {
		t.Fatalf("expected 1 loop duration, got %d", len(m.LoopDurations))
	}

	if m.LoopDurations[0] < 10*time.Millisecond {
		t.Errorf("expected loop duration >= 10ms, got %s", m.LoopDurations[0])
	}
}

func TestMultipleLoopIterations(t *testing.T) {
	m := NewMetrics()

	for i := 0; i < 5; i++ {
		m.StartLoop()
		time.Sleep(5 * time.Millisecond)
		m.CompleteLoop()
	}

	if len(m.LoopDurations) != 5 {
		t.Errorf("expected 5 loop durations, got %d", len(m.LoopDurations))
	}
}

func TestAddTokens(t *testing.T) {
	m := NewMetrics()

	m.AddTokens(100)
	m.AddTokens(250)
	m.AddTokens(50)

	if m.TokenUsage != 400 {
		t.Errorf("expected TokenUsage to be 400, got %d", m.TokenUsage)
	}
}

func TestAddFiles(t *testing.T) {
	m := NewMetrics()

	m.AddFiles(5)
	m.AddFiles(3)
	m.AddFiles(2)

	if m.FilesProcessed != 10 {
		t.Errorf("expected FilesProcessed to be 10, got %d", m.FilesProcessed)
	}
}

func TestDuration(t *testing.T) {
	m := NewMetrics()

	m.StartLoop()
	time.Sleep(20 * time.Millisecond)
	m.CompleteLoop()

	duration := m.Duration()
	if duration < 20*time.Millisecond {
		t.Errorf("expected Duration >= 20ms, got %s", duration)
	}
}

func TestDurationWithoutStart(t *testing.T) {
	m := NewMetrics()

	// Duration without StartLoop should still work (uses time.Since)
	duration := m.Duration()
	if duration < 0 {
		t.Errorf("expected Duration >= 0, got %s", duration)
	}
}

func TestAverageLoopDuration(t *testing.T) {
	m := NewMetrics()

	// Add durations directly for testing
	m.mu.Lock()
	m.LoopDurations = []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
	}
	m.mu.Unlock()

	avg := m.AverageLoopDuration()
	expected := 200 * time.Millisecond

	if avg != expected {
		t.Errorf("expected AverageLoopDuration to be %s, got %s", expected, avg)
	}
}

func TestAverageLoopDurationEmpty(t *testing.T) {
	m := NewMetrics()

	avg := m.AverageLoopDuration()
	if avg != 0 {
		t.Errorf("expected AverageLoopDuration to be 0 for empty loops, got %s", avg)
	}
}

func TestTotalTasks(t *testing.T) {
	m := NewMetrics()

	m.StartTask("task-1")
	m.CompleteTask("task-1", true)

	m.StartTask("task-2")
	m.CompleteTask("task-2", false)

	m.StartTask("task-3")
	m.CompleteTask("task-3", true)

	if m.TotalTasks() != 3 {
		t.Errorf("expected TotalTasks to be 3, got %d", m.TotalTasks())
	}
}

func TestSuccessRate(t *testing.T) {
	tests := []struct {
		name         string
		completed    int64
		failed       int64
		expectedRate float64
	}{
		{"all success", 10, 0, 100.0},
		{"all failed", 0, 10, 0.0},
		{"half success", 5, 5, 50.0},
		{"partial", 7, 3, 70.0},
		{"no tasks", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetrics()
			m.TasksCompleted = tt.completed
			m.TasksFailed = tt.failed

			rate := m.SuccessRate()
			if rate != tt.expectedRate {
				t.Errorf("expected SuccessRate to be %.2f, got %.2f", tt.expectedRate, rate)
			}
		})
	}
}

func TestReport(t *testing.T) {
	m := NewMetrics()
	m.StartTask("task-1")
	m.CompleteTask("task-1", true)
	m.StartTask("task-2")
	m.CompleteTask("task-2", false)

	m.StartLoop()
	time.Sleep(5 * time.Millisecond)
	m.CompleteLoop()

	m.AddTokens(500)
	m.AddFiles(10)

	report := m.Report()

	// Check that report contains expected sections
	expectedSections := []string{
		"## Ralph Loop Metrics Report",
		"**Tasks Completed:** 1",
		"**Tasks Failed:** 1",
		"**Token Usage:** 500",
		"**Files Processed:** 10",
		"### Loop Duration History",
	}

	for _, section := range expectedSections {
		if !contains(report, section) {
			t.Errorf("report missing section: %s", section)
		}
	}
}

func TestReportEmptyMetrics(t *testing.T) {
	m := NewMetrics()
	report := m.Report()

	if !contains(report, "## Ralph Loop Metrics Report") {
		t.Error("report should contain header")
	}

	if !contains(report, "**Tasks Completed:** 0") {
		t.Error("report should contain zero completed tasks")
	}

	if !contains(report, "> No loop iterations recorded") {
		t.Error("report should indicate no loop iterations")
	}
}

func TestConcurrentAccess(t *testing.T) {
	m := NewMetrics()
	done := make(chan bool)

	// Run concurrent operations
	for i := 0; i < 10; i++ {
		go func(id int) {
			taskID := "task-" + string(rune('0'+id))
			m.StartTask(taskID)
			m.CompleteTask(taskID, id%2 == 0)
			m.AddTokens(id * 100)
			m.AddFiles(id)
			m.StartLoop()
			m.CompleteLoop()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify counters are consistent
	if m.TasksCompleted+m.TasksFailed != 10 {
		t.Errorf("expected 10 total tasks, got %d", m.TasksCompleted+m.TasksFailed)
	}
}

func TestTaskTimerClearedOnComplete(t *testing.T) {
	m := NewMetrics()

	m.StartTask("task-1")
	m.StartTask("task-1") // Start same task again

	m.CompleteTask("task-1", true)

	m.mu.RLock()
	_, exists := m.taskTimers["task-1"]
	m.mu.RUnlock()

	if exists {
		t.Error("task timer should be cleared after CompleteTask")
	}
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
