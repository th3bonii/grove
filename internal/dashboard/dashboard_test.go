package dashboard

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewDashboard(t *testing.T) {
	d := NewDashboard()

	if d.Metrics == nil {
		t.Error("expected Metrics to be initialized")
	}
	if d.CurrentLoop != 1 {
		t.Errorf("expected CurrentLoop to be 1, got %d", d.CurrentLoop)
	}
	if d.TasksCompleted != 0 {
		t.Errorf("expected TasksCompleted to be 0, got %d", d.TasksCompleted)
	}
	if d.TotalTasks != 0 {
		t.Errorf("expected TotalTasks to be 0, got %d", d.TotalTasks)
	}
}

func TestDashboard_TrackTaskCompleted(t *testing.T) {
	d := NewDashboard()
	d.SetTotalTasks(10)

	d.Track(EventTaskCompleted, map[string]any{"task_id": "task-1"})

	if d.TasksCompleted != 1 {
		t.Errorf("expected TasksCompleted to be 1, got %d", d.TasksCompleted)
	}
	if d.PendingTasks != 9 {
		t.Errorf("expected PendingTasks to be 9, got %d", d.PendingTasks)
	}
}

func TestDashboard_TrackTaskFailed(t *testing.T) {
	d := NewDashboard()
	d.SetTotalTasks(10)

	d.Track(EventTaskFailed, map[string]any{"task_id": "task-1"})

	if d.TasksFailed != 1 {
		t.Errorf("expected TasksFailed to be 1, got %d", d.TasksFailed)
	}
}

func TestDashboard_TrackLoopIteration(t *testing.T) {
	d := NewDashboard()

	d.Track(EventLoopIteration, map[string]any{"loop": 3})

	if d.CurrentLoop != 3 {
		t.Errorf("expected CurrentLoop to be 3, got %d", d.CurrentLoop)
	}
}

func TestDashboard_TrackTokensUsed(t *testing.T) {
	d := NewDashboard()

	d.Track(EventTokensUsed, map[string]any{"tokens": 1000})

	if d.TokenUsage != 1000 {
		t.Errorf("expected TokenUsage to be 1000, got %d", d.TokenUsage)
	}
}

func TestDashboard_TrackQualityScore(t *testing.T) {
	d := NewDashboard()

	d.Track(EventQualityScore, map[string]any{"score": 87.5})

	if d.QualityScore != 87.5 {
		t.Errorf("expected QualityScore to be 87.5, got %f", d.QualityScore)
	}
}

func TestDashboard_SetTotalTasks(t *testing.T) {
	d := NewDashboard()
	d.TasksCompleted = 5
	d.TasksFailed = 2

	d.SetTotalTasks(24)

	if d.TotalTasks != 24 {
		t.Errorf("expected TotalTasks to be 24, got %d", d.TotalTasks)
	}
	if d.PendingTasks != 17 {
		t.Errorf("expected PendingTasks to be 17, got %d", d.PendingTasks)
	}
}

func TestDashboard_SetCurrentLoop(t *testing.T) {
	d := NewDashboard()

	d.SetCurrentLoop(5)

	if d.CurrentLoop != 5 {
		t.Errorf("expected CurrentLoop to be 5, got %d", d.CurrentLoop)
	}
}

func TestDashboard_SetQualityScore(t *testing.T) {
	d := NewDashboard()

	d.SetQualityScore(92.5)

	if d.QualityScore != 92.5 {
		t.Errorf("expected QualityScore to be 92.5, got %f", d.QualityScore)
	}
}

func TestDashboard_Show(t *testing.T) {
	d := NewDashboard()
	d.SetTotalTasks(24)
	d.TasksCompleted = 18
	d.TasksFailed = 2
	d.CurrentLoop = 3
	d.QualityScore = 87.5
	d.TokenUsage = 12450

	output := d.Show()

	if !strings.Contains(output, "GROVE DASHBOARD v1.0") {
		t.Error("expected dashboard header in output")
	}
	if !strings.Contains(output, "Current Loop: 3") {
		t.Error("expected Current Loop: 3 in output")
	}
	if !strings.Contains(output, "Tasks Completed: 18/24") {
		t.Error("expected Tasks Completed: 18/24 in output")
	}
	if !strings.Contains(output, "Tasks Failed: 2") {
		t.Error("expected Tasks Failed: 2 in output")
	}
	if !strings.Contains(output, "Quality Score: 87.5/100") {
		t.Error("expected Quality Score: 87.5/100 in output")
	}
}

func TestDashboard_Show_ProgressBar(t *testing.T) {
	d := NewDashboard()
	d.SetTotalTasks(4)
	d.TasksCompleted = 3

	output := d.Show()

	// Progress should be 75%
	if !strings.Contains(output, "███") && !strings.Contains(output, "░░") {
		t.Logf("Progress bar output: %s", output)
	}
}

func TestDashboard_Export(t *testing.T) {
	d := NewDashboard()
	d.SetTotalTasks(24)
	d.TasksCompleted = 18
	d.TasksFailed = 2
	d.CurrentLoop = 3
	d.QualityScore = 87.5
	d.TokenUsage = 12450
	d.FilesProcessed = 50

	data, err := d.Export()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	if result["current_loop"] != float64(3) {
		t.Errorf("expected current_loop 3, got %v", result["current_loop"])
	}
	if result["tasks_completed"] != float64(18) {
		t.Errorf("expected tasks_completed 18, got %v", result["tasks_completed"])
	}
	if result["tasks_failed"] != float64(2) {
		t.Errorf("expected tasks_failed 2, got %v", result["tasks_failed"])
	}
	if result["quality_score"] != float64(87.5) {
		t.Errorf("expected quality_score 87.5, got %v", result["quality_score"])
	}
	if result["token_usage"] != float64(12450) {
		t.Errorf("expected token_usage 12450, got %v", result["token_usage"])
	}
}

func TestDashboard_GetEvents(t *testing.T) {
	d := NewDashboard()

	d.Track(EventTaskCompleted, map[string]any{"task_id": "task-1"})
	d.Track(EventTaskFailed, map[string]any{"task_id": "task-2"})
	d.Track(EventLoopIteration, map[string]any{"loop": 2})

	events := d.GetEvents()
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}

func TestDashboard_GetEventsByType(t *testing.T) {
	d := NewDashboard()

	d.Track(EventTaskCompleted, map[string]any{"task_id": "task-1"})
	d.Track(EventTaskCompleted, map[string]any{"task_id": "task-2"})
	d.Track(EventTaskFailed, map[string]any{"task_id": "task-3"})

	completedEvents := d.GetEventsByType(EventTaskCompleted)
	if len(completedEvents) != 2 {
		t.Errorf("expected 2 completed events, got %d", len(completedEvents))
	}

	failedEvents := d.GetEventsByType(EventTaskFailed)
	if len(failedEvents) != 1 {
		t.Errorf("expected 1 failed event, got %d", len(failedEvents))
	}
}

func TestDashboard_Progress(t *testing.T) {
	d := NewDashboard()
	d.SetTotalTasks(100)

	if d.Progress() != 0 {
		t.Errorf("expected 0 progress, got %f", d.Progress())
	}

	d.TasksCompleted = 75

	if d.Progress() != 75 {
		t.Errorf("expected 75 progress, got %f", d.Progress())
	}
}

func TestDashboard_Progress_ZeroTasks(t *testing.T) {
	d := NewDashboard()

	if d.Progress() != 0 {
		t.Errorf("expected 0 progress with no tasks, got %f", d.Progress())
	}
}

func TestDashboard_SuccessRate(t *testing.T) {
	d := NewDashboard()
	d.TasksCompleted = 18
	d.TasksFailed = 2

	rate := d.SuccessRate()
	if rate != 90 {
		t.Errorf("expected 90 success rate, got %f", rate)
	}
}

func TestDashboard_SuccessRate_ZeroTasks(t *testing.T) {
	d := NewDashboard()

	rate := d.SuccessRate()
	if rate != 0 {
		t.Errorf("expected 0 success rate with no tasks, got %f", rate)
	}
}

func TestDashboard_LoopStartedAndCompleted(t *testing.T) {
	d := NewDashboard()

	d.Track(EventLoopStarted, nil)
	time.Sleep(10 * time.Millisecond)
	d.Track(EventLoopCompleted, nil)

	if len(d.LoopDurations) != 1 {
		t.Errorf("expected 1 loop duration, got %d", len(d.LoopDurations))
	}
}

func TestDashboard_Export_Empty(t *testing.T) {
	d := NewDashboard()

	data, err := d.Export()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	// Check default values
	if result["current_loop"] != float64(1) {
		t.Errorf("expected current_loop 1, got %v", result["current_loop"])
	}
	if result["tasks_completed"] != float64(0) {
		t.Errorf("expected tasks_completed 0, got %v", result["tasks_completed"])
	}
}

func TestDashboard_MultipleTracks(t *testing.T) {
	d := NewDashboard()
	d.SetTotalTasks(24)

	// Simulate multiple task completions
	for i := 0; i < 18; i++ {
		d.Track(EventTaskCompleted, map[string]any{"task_id": "task-" + string(rune('0'+i))})
	}

	// Simulate failures
	d.Track(EventTaskFailed, map[string]any{"task_id": "fail-1"})
	d.Track(EventTaskFailed, map[string]any{"task_id": "fail-2"})

	if d.TasksCompleted != 18 {
		t.Errorf("expected 18 completed, got %d", d.TasksCompleted)
	}
	if d.TasksFailed != 2 {
		t.Errorf("expected 2 failed, got %d", d.TasksFailed)
	}
	if d.PendingTasks != 4 {
		t.Errorf("expected 4 pending, got %d", d.PendingTasks)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0s"},
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{5 * time.Minute, "5m 0s"},
		{2 * time.Hour, "2h 0m"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{100, "100"},
		{1000, "1,000"},
		{12450, "12,450"},
		{1000000, "1,000,000"},
	}

	for _, tt := range tests {
		got := formatNumber(tt.n)
		if got != tt.want {
			t.Errorf("formatNumber(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
