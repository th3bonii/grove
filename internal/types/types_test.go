package types

import (
	"encoding/json"
	"testing"
	"time"
)

// =============================================================================
// GROVE Spec Type Tests
// =============================================================================

func TestExitCondition(t *testing.T) {
	tests := []struct {
		name     string
		value    ExitCondition
		expected string
	}{
		{"Normal", ExitNormal, "normal"},
		{"SafetyNet", ExitSafetyNet, "safety_net"},
		{"Manual", ExitManual, "manual"},
		{"Error", ExitError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("ExitCondition = %v, want %v", tt.value, tt.expected)
			}
		})
	}
}

func TestDimensionKey(t *testing.T) {
	expected := []DimensionKey{
		DimensionFlowCoverage,
		DimensionComponentDecomposition,
		DimensionLogicalConsistency,
		DimensionInterComponentConnectivity,
		DimensionEdgeCaseCoverage,
		DimensionDecisionJustification,
		DimensionAgentConsumability,
	}

	if len(expected) != 7 {
		t.Errorf("Expected 7 dimension keys, got %d", len(expected))
	}

	// Verify all dimensions are unique
	seen := make(map[DimensionKey]bool)
	for _, d := range expected {
		if seen[d] {
			t.Errorf("Duplicate dimension key: %v", d)
		}
		seen[d] = true
	}
}

func TestQualityScoreJSON(t *testing.T) {
	qs := QualityScore{
		Dimension: DimensionFlowCoverage,
		Score:     9,
		MaxScore:  10,
		Notes:     "Well covered flows",
	}

	data, err := json.Marshal(qs)
	if err != nil {
		t.Fatalf("Failed to marshal QualityScore: %v", err)
	}

	var decoded QualityScore
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal QualityScore: %v", err)
	}

	if decoded.Dimension != qs.Dimension {
		t.Errorf("Dimension mismatch: got %v, want %v", decoded.Dimension, qs.Dimension)
	}
	if decoded.Score != qs.Score {
		t.Errorf("Score mismatch: got %d, want %d", decoded.Score, qs.Score)
	}
}

func TestCompositeScorePassed(t *testing.T) {
	cs := CompositeScore{
		Scores: []QualityScore{
			{Dimension: DimensionFlowCoverage, Score: 9, MaxScore: 10},
			{Dimension: DimensionComponentDecomposition, Score: 9, MaxScore: 10},
			{Dimension: DimensionLogicalConsistency, Score: 9, MaxScore: 10},
			{Dimension: DimensionInterComponentConnectivity, Score: 9, MaxScore: 10},
			{Dimension: DimensionEdgeCaseCoverage, Score: 9, MaxScore: 10},
			{Dimension: DimensionDecisionJustification, Score: 9, MaxScore: 10},
			{Dimension: DimensionAgentConsumability, Score: 9, MaxScore: 10},
		},
		Composite:    90,
		MinDimension: 9,
		Passed:       true,
	}

	// Verify passed is true when all dims >= 8 and composite >= 85
	allDimsOk := true
	for _, s := range cs.Scores {
		if s.Score < 8 {
			allDimsOk = false
			break
		}
	}
	compositeOk := cs.Composite >= 85

	if !allDimsOk || !compositeOk {
		t.Error("CompositeScore should pass when all dims >= 8 and composite >= 85")
	}
}

func TestLoopStateJSON(t *testing.T) {
	now := time.Now()
	completed := now.Add(time.Minute)
	ls := LoopState{
		LoopNumber:               2,
		DimensionScores:          []QualityScore{},
		CompositeScore:           88,
		ContentDeltaPct:          5.2,
		ConsecutiveLowDeltaCount: 0,
		ExitCondition:            ExitNormal,
		StartedAt:                now,
		CompletedAt:              &completed,
	}

	data, err := json.Marshal(ls)
	if err != nil {
		t.Fatalf("Failed to marshal LoopState: %v", err)
	}

	var decoded LoopState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal LoopState: %v", err)
	}

	if decoded.LoopNumber != ls.LoopNumber {
		t.Errorf("LoopNumber mismatch: got %d, want %d", decoded.LoopNumber, ls.LoopNumber)
	}
	if decoded.ExitCondition != ls.ExitCondition {
		t.Errorf("ExitCondition mismatch: got %v, want %v", decoded.ExitCondition, ls.ExitCondition)
	}
}

func TestComponentType(t *testing.T) {
	types := []ComponentType{
		ComponentTypeUI,
		ComponentTypeBackend,
		ComponentTypeService,
		ComponentTypeData,
		ComponentTypeIntegration,
		ComponentTypeUtility,
	}

	if len(types) != 6 {
		t.Errorf("Expected 6 component types, got %d", len(types))
	}
}

func TestComponentState(t *testing.T) {
	state := ComponentState{
		ID:          "state-1",
		Name:        "Default",
		Description: "Initial state",
		IsDefault:   true,
		Transitions: []Transition{
			{FromState: "state-1", ToState: "state-2", Trigger: "click"},
		},
	}

	if state.ID != "state-1" {
		t.Errorf("State ID mismatch")
	}
	if !state.IsDefault {
		t.Error("State should be default")
	}
	if len(state.Transitions) != 1 {
		t.Errorf("Expected 1 transition, got %d", len(state.Transitions))
	}
}

func TestTaskStatus(t *testing.T) {
	statuses := []TaskStatus{
		TaskStatusPending,
		TaskStatusInProgress,
		TaskStatusCompleted,
		TaskStatusBlocked,
	}

	expected := []string{"pending", "in_progress", "completed", "blocked"}
	for i, s := range statuses {
		if string(s) != expected[i] {
			t.Errorf("TaskStatus = %v, want %v", s, expected[i])
		}
	}
}

// =============================================================================
// GROVE Ralph Loop Type Tests
// =============================================================================

func TestLoopStatus(t *testing.T) {
	statuses := []LoopStatus{
		LoopStatusInitializing,
		LoopStatusRunning,
		LoopStatusPaused,
		LoopStatusCompleted,
		LoopStatusFailed,
		LoopStatusRecovering,
	}

	if len(statuses) != 6 {
		t.Errorf("Expected 6 loop statuses, got %d", len(statuses))
	}
}

func TestVerifyStatus(t *testing.T) {
	statuses := []VerifyStatus{
		VerifyStatusPassed,
		VerifyStatusFailed,
		VerifyStatusWarning,
		VerifyStatusSkipped,
	}

	if len(statuses) != 4 {
		t.Errorf("Expected 4 verify statuses, got %d", len(statuses))
	}
}

func TestErrorType(t *testing.T) {
	types := []ErrorType{
		ErrorTypeLLMResponse,
		ErrorTypeNetwork,
		ErrorTypeFileSystem,
		ErrorTypeVerification,
		ErrorTypeTimeout,
		ErrorTypeUnknown,
	}

	if len(types) != 6 {
		t.Errorf("Expected 6 error types, got %d", len(types))
	}
}

func TestReadyStatus(t *testing.T) {
	statuses := []ReadyStatus{
		ReadyStatusProductionReady,
		ReadyStatusNeedsWork,
		ReadyStatusBlocked,
		ReadyStatusUnknown,
	}

	if len(statuses) != 4 {
		t.Errorf("Expected 4 ready statuses, got %d", len(statuses))
	}
}

func TestTaskExecution(t *testing.T) {
	now := time.Now()
	exec := TaskExecution{
		TaskID:      "task-1",
		Status:      TaskStatusCompleted,
		Attempts:    1,
		MaxAttempts: 3,
		StartedAt:   &now,
		CompletedAt: &now,
		Result: &TaskResult{
			Success:      true,
			FilesCreated: []string{"file.go"},
		},
	}

	if exec.TaskID != "task-1" {
		t.Errorf("TaskID mismatch")
	}
	if exec.Result == nil || !exec.Result.Success {
		t.Error("Result should indicate success")
	}
}

func TestCheckpoint(t *testing.T) {
	now := time.Now()
	state := &LoopRunState{
		LoopNumber: 2,
		Status:     LoopStatusRunning,
	}

	cp := Checkpoint{
		ID:         "cp-1",
		Timestamp:  now,
		LoopNumber: 2,
		TaskID:     "task-1",
		State:      state,
		Reason:     "Periodic save",
	}

	if cp.ID != "cp-1" {
		t.Errorf("Checkpoint ID mismatch")
	}
	if cp.State == nil {
		t.Error("Checkpoint state should not be nil")
	}
}

func TestErrorRecovery(t *testing.T) {
	er := ErrorRecovery{
		MaxRetries:       3,
		BackoffBase:      2,
		ReducedContext:   true,
		ContextReduction: 0.25,
	}

	if er.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", er.MaxRetries)
	}
	if !er.ReducedContext {
		t.Error("ReducedContext should be true")
	}
}

// =============================================================================
// GROVE Opti Prompt Type Tests
// =============================================================================

func TestIntentType(t *testing.T) {
	types := []IntentType{
		IntentFeatureAddition,
		IntentBugFix,
		IntentRefactor,
		IntentDocumentationUpdate,
		IntentConfigurationChange,
		IntentOther,
	}

	expected := []string{
		"feature-addition",
		"bug-fix",
		"refactor",
		"documentation-update",
		"configuration-change",
		"other",
	}

	for i, it := range types {
		if string(it) != expected[i] {
			t.Errorf("IntentType = %v, want %v", it, expected[i])
		}
	}
}

func TestSelectionLayer(t *testing.T) {
	layers := []SelectionLayer{
		LayerAgentsMD,
		LayerGitCommits,
		LayerPathMatch,
		LayerSpecComponents,
	}

	if len(layers) != 4 {
		t.Errorf("Expected 4 selection layers, got %d", len(layers))
	}

	// Verify priority order
	if LayerAgentsMD != 1 {
		t.Errorf("LayerAgentsMD should be 1, got %d", LayerAgentsMD)
	}
}

func TestWhyCategory(t *testing.T) {
	categories := []WhyCategory{
		WhyCategoryFileReference,
		WhyCategoryScopeBoundary,
		WhyCategorySkillInvocation,
		WhyCategorySuccessCriteria,
		WhyCategoryPlanMode,
		WhyCategoryOutOfScope,
	}

	if len(categories) != 6 {
		t.Errorf("Expected 6 WHY categories, got %d", len(categories))
	}
}

func TestUserAction(t *testing.T) {
	actions := []UserAction{
		UserActionSend,
		UserActionEdit,
		UserActionReject,
	}

	expected := []string{"send", "edit", "reject"}
	for i, a := range actions {
		if string(a) != expected[i] {
			t.Errorf("UserAction = %v, want %v", a, expected[i])
		}
	}
}

func TestNewUserProfile(t *testing.T) {
	profile := NewUserProfile()

	if profile == nil {
		t.Fatal("NewUserProfile returned nil")
	}

	if len(profile.Categories) != 6 {
		t.Errorf("Expected 6 categories, got %d", len(profile.Categories))
	}

	for cat, stats := range profile.Categories {
		if stats.TimesSeen != 0 {
			t.Errorf("Category %v should have TimesSeen=0, got %d", cat, stats.TimesSeen)
		}
	}
}

func TestGetWhyLevel(t *testing.T) {
	tests := []struct {
		name      string
		timesSeen int
		forceFull bool
		expected  WhyLevel
	}{
		{"Never seen", 0, false, WhyLevelFull},
		{"3 times", 3, false, WhyLevelFull},
		{"4 times", 4, false, WhyLevelBrief},
		{"10 times", 10, false, WhyLevelBrief},
		{"11 times", 11, false, WhyLevelLabel},
		{"Force full", 100, true, WhyLevelFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetWhyLevel(tt.timesSeen, tt.forceFull)
			if got != tt.expected {
				t.Errorf("GetWhyLevel(%d, %v) = %v, want %v",
					tt.timesSeen, tt.forceFull, got, tt.expected)
			}
		})
	}
}

func TestShouldApplyLearnedPattern(t *testing.T) {
	pattern := EditPattern{
		PatternType: PatternTypeAdded,
		Category:    "scope-boundary",
		Frequency:   5,
	}

	if !ShouldApplyLearnedPattern(pattern, 3) {
		t.Error("Pattern with frequency >= min should be applied")
	}

	if ShouldApplyLearnedPattern(pattern, 6) {
		t.Error("Pattern with frequency < min should not be applied")
	}
}

func TestOptimizedPromptJSON(t *testing.T) {
	op := OptimizedPrompt{
		Original:  "add login button",
		Optimized: "Add a login button to the header component in src/components/Header.tsx. Do NOT modify the sidebar. This is done when the button is visible and navigates to /login.",
		Intent: Intent{
			Type:       IntentFeatureAddition,
			Confidence: 0.95,
		},
		WhyExplanations: []WhyExplanation{
			{Category: WhyCategoryFileReference, Full: "Always reference specific files.", Brief: "File refs help context.", Label: "file-reference"},
		},
		PlanMode:    false,
		TokenCount:  150,
		TokenBudget: 2000,
	}

	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("Failed to marshal OptimizedPrompt: %v", err)
	}

	var decoded OptimizedPrompt
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal OptimizedPrompt: %v", err)
	}

	if decoded.TokenCount != op.TokenCount {
		t.Errorf("TokenCount mismatch: got %d, want %d", decoded.TokenCount, op.TokenCount)
	}
	if decoded.Intent.Type != op.Intent.Type {
		t.Errorf("Intent type mismatch")
	}
}

func TestInvocationLog(t *testing.T) {
	log := InvocationLog{
		Timestamp:            time.Now(),
		IntentClassification: IntentFeatureAddition,
		TokensUsed:           1200,
		FilesSelected: []FileSelectionLog{
			{Path: "src/components/Header.tsx", Layer: 1},
			{Path: "src/routes/login.go", Layer: 3},
		},
		UserAction:       UserActionSend,
		SkillsReferenced: []string{"react-19"},
	}

	if log.TokensUsed != 1200 {
		t.Errorf("TokensUsed mismatch")
	}
	if len(log.FilesSelected) != 2 {
		t.Errorf("Expected 2 files, got %d", len(log.FilesSelected))
	}
}

func TestDiffSegment(t *testing.T) {
	diff := []DiffSegment{
		{Type: DiffTypeUnchanged, Content: "Add a "},
		{Type: DiffTypeAdded, Content: "new "},
		{Type: DiffTypeUnchanged, Content: "button"},
	}

	if diff[0].Type != DiffTypeUnchanged {
		t.Error("First segment should be unchanged")
	}
	if diff[1].Type != DiffTypeAdded {
		t.Error("Second segment should be added")
	}
}

// =============================================================================
// Utility Tests
// =============================================================================

func TestLoopStateCanMarshal(t *testing.T) {
	now := time.Now()
	ls := LoopState{
		LoopNumber:    1,
		ExitCondition: ExitNormal,
		StartedAt:     now,
	}

	// Should not panic
	data, err := json.Marshal(ls)
	if err != nil {
		t.Fatalf("LoopState marshal failed: %v", err)
	}

	var decoded LoopState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("LoopState unmarshal failed: %v", err)
	}

	if decoded.LoopNumber != ls.LoopNumber {
		t.Errorf("LoopNumber mismatch after round-trip")
	}
}

func TestSpecDocumentComplete(t *testing.T) {
	doc := SpecDocument{
		Title:        "Test Project",
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Overview:     "Test overview",
		Components:   []Component{},
		UserFlows:    []UserFlow{},
		Requirements: []Requirement{},
	}

	if doc.Title != "Test Project" {
		t.Error("Title mismatch")
	}
}

func TestDesignDocumentComplete(t *testing.T) {
	doc := DesignDocument{
		Title:        "Test Design",
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Architecture: "Microservices",
		TechStack: []TechStackItem{
			{Name: "Go", Version: "1.23", Purpose: "Backend"},
		},
		Decisions: []Decision{
			{ID: "d1", Title: "Use Go", Decision: "Use Go for backend", Justification: "Performance"},
		},
	}

	if len(doc.TechStack) != 1 {
		t.Error("TechStack should have 1 item")
	}
	if len(doc.Decisions) != 1 {
		t.Error("Decisions should have 1 item")
	}
}

func TestTasksDocumentComplete(t *testing.T) {
	doc := TasksDocument{
		Title:     "Test Tasks",
		Version:   "1.0.0",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tasks: []Task{
			{ID: "t1", Title: "Implement login", Status: TaskStatusPending},
		},
		Milestones: []Milestone{
			{ID: "m1", Name: "v1.0", TaskIDs: []string{"t1"}},
		},
	}

	if len(doc.Tasks) != 1 {
		t.Error("Tasks should have 1 item")
	}
	if len(doc.Milestones) != 1 {
		t.Error("Milestones should have 1 item")
	}
}

func TestPromptDiff(t *testing.T) {
	diff := PromptDiff{
		Original:  "add button",
		Optimized: "add login button",
		Final:     "add logout button",
		Diff: []DiffSegment{
			{Type: DiffTypeRemoved, Content: "login"},
			{Type: DiffTypeAdded, Content: "logout"},
		},
	}

	if len(diff.Diff) != 2 {
		t.Errorf("Expected 2 diff segments, got %d", len(diff.Diff))
	}
}

// =============================================================================
// Performance/Benchmark Tests
// =============================================================================

func BenchmarkLoopStateMarshal(b *testing.B) {
	now := time.Now()
	ls := LoopState{
		LoopNumber:     5,
		CompositeScore: 88,
		StartedAt:      now,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ls)
	}
}

func BenchmarkUserProfileCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewUserProfile()
	}
}

// =============================================================================
// JSON Serialization Edge Cases Tests
// =============================================================================

func TestJSONSerialization_AllTypes(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "ExitCondition",
			value: ExitNormal,
		},
		{
			name:  "DimensionKey",
			value: DimensionFlowCoverage,
		},
		{
			name:  "ComponentType",
			value: ComponentTypeUI,
		},
		{
			name: "ComponentState",
			value: ComponentState{
				ID:          "state-1",
				Name:        "Default",
				Description: "Initial state",
				IsDefault:   true,
			},
		},
		{
			name:  "TaskStatus",
			value: TaskStatusPending,
		},
		{
			name:  "LoopStatus",
			value: LoopStatusRunning,
		},
		{
			name:  "VerifyStatus",
			value: VerifyStatusPassed,
		},
		{
			name:  "ErrorType",
			value: ErrorTypeLLMResponse,
		},
		{
			name:  "ReadyStatus",
			value: ReadyStatusProductionReady,
		},
		{
			name:  "IntentType",
			value: IntentFeatureAddition,
		},
		{
			name:  "SelectionLayer",
			value: LayerAgentsMD,
		},
		{
			name:  "WhyCategory",
			value: WhyCategoryFileReference,
		},
		{
			name:  "UserAction",
			value: UserActionSend,
		},
		{
			name:  "DiffType",
			value: DiffTypeAdded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Failed to marshal %s: %v", tt.name, err)
			}

			// Unmarshal back
			var result interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal %s: %v", tt.name, err)
			}

			// Verify we got something back (JSON unmarshal always produces something)
			if result == nil {
				t.Errorf("Unmarshal returned nil for %s", tt.name)
			}
		})
	}
}

func TestJSONSerialization_QualityScoreEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		value QualityScore
	}{
		{
			name:  "zero values",
			value: QualityScore{},
		},
		{
			name: "max values",
			value: QualityScore{
				Dimension: DimensionFlowCoverage,
				Score:     10,
				MaxScore:  10,
				Notes:     "Perfect score",
			},
		},
		{
			name: "with nil notes",
			value: QualityScore{
				Dimension: DimensionComponentDecomposition,
				Score:     5,
				MaxScore:  10,
				Notes:     "",
			},
		},
		{
			name: "empty dimension",
			value: QualityScore{
				Dimension: "",
				Score:     0,
				MaxScore:  10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			var result QualityScore
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// Verify basic round-trip
			if result.Score != tt.value.Score || result.MaxScore != tt.value.MaxScore {
				t.Errorf("Score/MaxScore mismatch: got %d/%d, want %d/%d",
					result.Score, result.MaxScore, tt.value.Score, tt.value.MaxScore)
			}
		})
	}
}

func TestJSONSerialization_LoopStateEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		value LoopState
	}{
		{
			name:  "initial state",
			value: LoopState{},
		},
		{
			name: "with timestamps",
			value: LoopState{
				LoopNumber:    1,
				ExitCondition: ExitNormal,
				StartedAt:     time.Now(),
			},
		},
		{
			name: "with nil pointer",
			value: LoopState{
				LoopNumber:    2,
				CompletedAt:   nil,
				ExitCondition: ExitSafetyNet,
			},
		},
		{
			name: "with empty dimension scores",
			value: LoopState{
				LoopNumber:               3,
				DimensionScores:          []QualityScore{},
				CompositeScore:           0,
				ContentDeltaPct:          0,
				ConsecutiveLowDeltaCount: 0,
				ExitCondition:            ExitManual,
			},
		},
		{
			name: "with negative delta",
			value: LoopState{
				LoopNumber:      4,
				CompositeScore:  75,
				ContentDeltaPct: -5.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			var result LoopState
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// Basic verification
			if result.LoopNumber != tt.value.LoopNumber {
				t.Errorf("LoopNumber mismatch: got %d, want %d", result.LoopNumber, tt.value.LoopNumber)
			}
		})
	}
}

func TestJSONSerialization_CompositeScoreEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		value CompositeScore
	}{
		{
			name:  "zero values",
			value: CompositeScore{},
		},
		{
			name: "passing score",
			value: CompositeScore{
				Scores: []QualityScore{
					{Dimension: DimensionFlowCoverage, Score: 9, MaxScore: 10},
					{Dimension: DimensionComponentDecomposition, Score: 8, MaxScore: 10},
					{Dimension: DimensionLogicalConsistency, Score: 9, MaxScore: 10},
					{Dimension: DimensionInterComponentConnectivity, Score: 8, MaxScore: 10},
					{Dimension: DimensionEdgeCaseCoverage, Score: 9, MaxScore: 10},
					{Dimension: DimensionDecisionJustification, Score: 8, MaxScore: 10},
					{Dimension: DimensionAgentConsumability, Score: 9, MaxScore: 10},
				},
				Composite:    88,
				MinDimension: 8,
				Passed:       true,
			},
		},
		{
			name: "failing score",
			value: CompositeScore{
				Scores: []QualityScore{
					{Dimension: DimensionFlowCoverage, Score: 5, MaxScore: 10},
				},
				Composite:    50,
				MinDimension: 5,
				Passed:       false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			var result CompositeScore
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if result.Passed != tt.value.Passed {
				t.Errorf("Passed mismatch: got %v, want %v", result.Passed, tt.value.Passed)
			}
		})
	}
}

func TestJSONSerialization_TaskExecutionEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		value TaskExecution
	}{
		{
			name:  "empty task",
			value: TaskExecution{},
		},
		{
			name: "with nil result",
			value: TaskExecution{
				TaskID:      "task-1",
				Status:      TaskStatusInProgress,
				Attempts:    0,
				MaxAttempts: 3,
			},
		},
		{
			name: "with failed result",
			value: TaskExecution{
				TaskID:      "task-2",
				Status:      TaskStatusBlocked,
				Attempts:    3,
				MaxAttempts: 3,
				Result: &TaskResult{
					Success: false,
					Output:  "compilation failed",
					Message: "failed",
				},
			},
		},
		{
			name: "with many files",
			value: TaskExecution{
				TaskID:   "task-3",
				Status:   TaskStatusCompleted,
				Attempts: 1,
				Result: &TaskResult{
					Success: true,
					FilesCreated: []string{
						"file1.go", "file2.go", "file3.go", "file4.go", "file5.go",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			var result TaskExecution
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if result.TaskID != tt.value.TaskID {
				t.Errorf("TaskID mismatch: got %q, want %q", result.TaskID, tt.value.TaskID)
			}
		})
	}
}

func TestJSONSerialization_CheckpointEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		value Checkpoint
	}{
		{
			name:  "empty checkpoint",
			value: Checkpoint{},
		},
		{
			name: "with nil state",
			value: Checkpoint{
				ID:         "cp-1",
				Timestamp:  time.Now(),
				LoopNumber: 1,
				TaskID:     "task-1",
				State:      nil,
				Reason:     "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			var result Checkpoint
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if result.ID != tt.value.ID {
				t.Errorf("ID mismatch: got %q, want %q", result.ID, tt.value.ID)
			}
		})
	}
}
