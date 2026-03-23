package interfaces

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// MockSpecEngine is a mock implementation of SpecEngine.
type MockSpecEngine struct {
	GenerateSpecFunc func(ctx context.Context, input SpecInput) (*SpecOutput, error)
	UpdateSpecFunc   func(ctx context.Context, update UpdateInput) (*SpecOutput, error)
	ValidateSpecFunc func(ctx context.Context, spec SpecArtifact) (*ValidationResult, error)
	DecomposeFunc    func(ctx context.Context, req DecomposeInput) ([]Component, error)
}

func (m *MockSpecEngine) GenerateSpec(ctx context.Context, input SpecInput) (*SpecOutput, error) {
	if m.GenerateSpecFunc != nil {
		return m.GenerateSpecFunc(ctx, input)
	}
	return &SpecOutput{
		Spec:   &SpecArtifact{Phase: PhaseSpec, Content: "mock spec"},
		Design: &SpecArtifact{Phase: PhaseDesign, Content: "mock design"},
		Tasks:  []Task{},
	}, nil
}

func (m *MockSpecEngine) UpdateSpec(ctx context.Context, update UpdateInput) (*SpecOutput, error) {
	if m.UpdateSpecFunc != nil {
		return m.UpdateSpecFunc(ctx, update)
	}
	return &SpecOutput{
		Spec:   &SpecArtifact{Phase: PhaseSpec, Content: "updated spec"},
		Design: &SpecArtifact{Phase: PhaseDesign, Content: "updated design"},
		Tasks:  []Task{},
	}, nil
}

func (m *MockSpecEngine) ValidateSpec(ctx context.Context, spec SpecArtifact) (*ValidationResult, error) {
	if m.ValidateSpecFunc != nil {
		return m.ValidateSpecFunc(ctx, spec)
	}
	return &ValidationResult{
		Valid:        true,
		PassedChecks: []string{"structure", "completeness"},
	}, nil
}

func (m *MockSpecEngine) Decompose(ctx context.Context, req DecomposeInput) ([]Component, error) {
	if m.DecomposeFunc != nil {
		return m.DecomposeFunc(ctx, req)
	}
	return []Component{
		{Name: req.ComponentName, Type: "component", Justification: "mock decomposition"},
	}, nil
}

// MockSpecScorer is a mock implementation of SpecScorer.
type MockSpecScorer struct {
	ScoreSpecFunc      func(ctx context.Context, spec SpecArtifact) (*QualityScores, error)
	IsCompleteFunc     func(ctx context.Context, scores QualityScores) (bool, string, error)
	CalculateDeltaFunc func(ctx context.Context, oldSpec, newSpec SpecArtifact) (float64, error)
}

func (m *MockSpecScorer) ScoreSpec(ctx context.Context, spec SpecArtifact) (*QualityScores, error) {
	if m.ScoreSpecFunc != nil {
		return m.ScoreSpecFunc(ctx, spec)
	}
	return &QualityScores{
		FlowCoverage:          8.5,
		ComponentDepth:        8.0,
		LogicalConsistency:    9.0,
		Connectivity:          8.5,
		EdgeCases:             8.0,
		DecisionJustification: 9.0,
		AgentConsumability:    9.5,
	}, nil
}

func (m *MockSpecScorer) IsComplete(ctx context.Context, scores QualityScores) (bool, string, error) {
	if m.IsCompleteFunc != nil {
		return m.IsCompleteFunc(ctx, scores)
	}
	if scores.CompositeScore() >= 85 {
		return true, "normal", nil
	}
	return false, "below_threshold", nil
}

func (m *MockSpecScorer) CalculateDelta(ctx context.Context, oldSpec, newSpec SpecArtifact) (float64, error) {
	if m.CalculateDeltaFunc != nil {
		return m.CalculateDeltaFunc(ctx, oldSpec, newSpec)
	}
	return 5.0, nil
}

// MockLoopValidator is a mock implementation of LoopValidator.
type MockLoopValidator struct {
	ValidateAgentsFunc       func(ctx context.Context, projectPath string) (*AgentsValidationResult, error)
	ValidateSkillsFunc       func(ctx context.Context, projectPath string) (*SkillsValidationResult, error)
	ValidateSpecsFunc        func(ctx context.Context, projectPath string) (*SpecsValidationResult, error)
	ValidateStackFunc        func(ctx context.Context, design SpecArtifact) (*StackValidationResult, error)
	ValidateDependenciesFunc func(ctx context.Context, tasks []Task) (*DependenciesValidationResult, error)
	ValidateAllFunc          func(ctx context.Context, projectPath string) (*AllValidationsResult, error)
}

func (m *MockLoopValidator) ValidateAgents(ctx context.Context, projectPath string) (*AgentsValidationResult, error) {
	if m.ValidateAgentsFunc != nil {
		return m.ValidateAgentsFunc(ctx, projectPath)
	}
	return &AgentsValidationResult{
		RootValid:          true,
		ScopedValid:        true,
		SkillRegistryValid: true,
	}, nil
}

func (m *MockLoopValidator) ValidateSkills(ctx context.Context, projectPath string) (*SkillsValidationResult, error) {
	if m.ValidateSkillsFunc != nil {
		return m.ValidateSkillsFunc(ctx, projectPath)
	}
	return &SkillsValidationResult{Valid: true}, nil
}

func (m *MockLoopValidator) ValidateSpecs(ctx context.Context, projectPath string) (*SpecsValidationResult, error) {
	if m.ValidateSpecsFunc != nil {
		return m.ValidateSpecsFunc(ctx, projectPath)
	}
	return &SpecsValidationResult{
		SpecValid:   true,
		DesignValid: true,
		TasksValid:  true,
	}, nil
}

func (m *MockLoopValidator) ValidateStack(ctx context.Context, design SpecArtifact) (*StackValidationResult, error) {
	if m.ValidateStackFunc != nil {
		return m.ValidateStackFunc(ctx, design)
	}
	return &StackValidationResult{
		Valid:    true,
		Coherent: true,
		Declared: []string{"go", "yaml"},
		Detected: []string{"go", "yaml"},
	}, nil
}

func (m *MockLoopValidator) ValidateDependencies(ctx context.Context, tasks []Task) (*DependenciesValidationResult, error) {
	if m.ValidateDependenciesFunc != nil {
		return m.ValidateDependenciesFunc(ctx, tasks)
	}
	return &DependenciesValidationResult{Valid: true}, nil
}

func (m *MockLoopValidator) ValidateAll(ctx context.Context, projectPath string) (*AllValidationsResult, error) {
	if m.ValidateAllFunc != nil {
		return m.ValidateAllFunc(ctx, projectPath)
	}
	return &AllValidationsResult{
		Ready:      true,
		CanProceed: true,
	}, nil
}

// MockLoopOrchestrator is a mock implementation of LoopOrchestrator.
type MockLoopOrchestrator struct {
	StartFunc        func(ctx context.Context, config LoopConfig) error
	PauseFunc        func(ctx context.Context) error
	ResumeFunc       func(ctx context.Context) error
	StopFunc         func(ctx context.Context) error
	GetStateFunc     func(ctx context.Context) (*LoopState, error)
	ExecuteTaskFunc  func(ctx context.Context, task Task) (*TaskResult, error)
	ExecutePhaseFunc func(ctx context.Context, phase SpecPhase) (*PhaseResult, error)
}

func (m *MockLoopOrchestrator) Start(ctx context.Context, config LoopConfig) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx, config)
	}
	return nil
}

func (m *MockLoopOrchestrator) Pause(ctx context.Context) error {
	if m.PauseFunc != nil {
		return m.PauseFunc(ctx)
	}
	return nil
}

func (m *MockLoopOrchestrator) Resume(ctx context.Context) error {
	if m.ResumeFunc != nil {
		return m.ResumeFunc(ctx)
	}
	return nil
}

func (m *MockLoopOrchestrator) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	return nil
}

func (m *MockLoopOrchestrator) GetState(ctx context.Context) (*LoopState, error) {
	if m.GetStateFunc != nil {
		return m.GetStateFunc(ctx)
	}
	return &LoopState{LoopNumber: 1}, nil
}

func (m *MockLoopOrchestrator) ExecuteTask(ctx context.Context, task Task) (*TaskResult, error) {
	if m.ExecuteTaskFunc != nil {
		return m.ExecuteTaskFunc(ctx, task)
	}
	return &TaskResult{TaskID: task.ID, Success: true, Duration: time.Second}, nil
}

func (m *MockLoopOrchestrator) ExecutePhase(ctx context.Context, phase SpecPhase) (*PhaseResult, error) {
	if m.ExecutePhaseFunc != nil {
		return m.ExecutePhaseFunc(ctx, phase)
	}
	return &PhaseResult{Phase: phase, TotalTasks: 5, Completed: 5}, nil
}

// MockIntentClassifier is a mock implementation of IntentClassifier.
type MockIntentClassifier struct {
	ClassifyIntentFunc  func(ctx context.Context, input string) (*IntentClassification, error)
	ExtractEntitiesFunc func(ctx context.Context, input string) ([]Entity, error)
	DetectAmbiguityFunc func(ctx context.Context, input string) (*AmbiguityReport, error)
}

func (m *MockIntentClassifier) ClassifyIntent(ctx context.Context, input string) (*IntentClassification, error) {
	if m.ClassifyIntentFunc != nil {
		return m.ClassifyIntentFunc(ctx, input)
	}
	return &IntentClassification{
		PrimaryIntent: IntentCreate,
		Confidence:    0.95,
	}, nil
}

func (m *MockIntentClassifier) ExtractEntities(ctx context.Context, input string) ([]Entity, error) {
	if m.ExtractEntitiesFunc != nil {
		return m.ExtractEntitiesFunc(ctx, input)
	}
	return []Entity{}, nil
}

func (m *MockIntentClassifier) DetectAmbiguity(ctx context.Context, input string) (*AmbiguityReport, error) {
	if m.DetectAmbiguityFunc != nil {
		return m.DetectAmbiguityFunc(ctx, input)
	}
	return &AmbiguityReport{Ambiguous: false}, nil
}

// MockContextCollector is a mock implementation of ContextCollector.
type MockContextCollector struct {
	CollectTaskContextFunc   func(ctx context.Context, task Task) (*TaskContext, error)
	CollectSpecContextFunc   func(ctx context.Context, scope string) (*SpecContext, error)
	CollectAgentsContextFunc func(ctx context.Context, module string) (*AgentsContext, error)
	CollectSkillsContextFunc func(ctx context.Context, triggers []string) ([]SkillContext, error)
	BoundContextFunc         func(ctx context.Context, context Context, maxTokens int) (*BoundedContext, error)
}

func (m *MockContextCollector) CollectTaskContext(ctx context.Context, task Task) (*TaskContext, error) {
	if m.CollectTaskContextFunc != nil {
		return m.CollectTaskContextFunc(ctx, task)
	}
	return &TaskContext{Task: task}, nil
}

func (m *MockContextCollector) CollectSpecContext(ctx context.Context, scope string) (*SpecContext, error) {
	if m.CollectSpecContextFunc != nil {
		return m.CollectSpecContextFunc(ctx, scope)
	}
	return &SpecContext{RelevantSections: []SpecSection{}}, nil
}

func (m *MockContextCollector) CollectAgentsContext(ctx context.Context, module string) (*AgentsContext, error) {
	if m.CollectAgentsContextFunc != nil {
		return m.CollectAgentsContextFunc(ctx, module)
	}
	return &AgentsContext{
		RootAgents: &RootAgentsData{
			ProjectContext: "mock project",
			SkillRegistry:  map[string]string{},
		},
	}, nil
}

func (m *MockContextCollector) CollectSkillsContext(ctx context.Context, triggers []string) ([]SkillContext, error) {
	if m.CollectSkillsContextFunc != nil {
		return m.CollectSkillsContextFunc(ctx, triggers)
	}
	return []SkillContext{}, nil
}

func (m *MockContextCollector) BoundContext(ctx context.Context, context Context, maxTokens int) (*BoundedContext, error) {
	if m.BoundContextFunc != nil {
		return m.BoundContextFunc(ctx, context, maxTokens)
	}
	return &BoundedContext{Content: "bounded context", Tokens: 100, Truncated: false}, nil
}

// MockStateManager is a mock implementation of StateManager.
type MockStateManager struct {
	SaveLoopStateFunc func(ctx context.Context, state *LoopState) error
	LoadLoopStateFunc func(ctx context.Context, projectPath string) (*LoopState, error)
	SaveSpecStateFunc func(ctx context.Context, state *SpecLoopState) error
	LoadSpecStateFunc func(ctx context.Context, projectPath string) (*SpecLoopState, error)
	ClearStateFunc    func(ctx context.Context, projectPath string) error
	AtomicWriteFunc   func(ctx context.Context, path string, data interface{}) error
}

func (m *MockStateManager) SaveLoopState(ctx context.Context, state *LoopState) error {
	if m.SaveLoopStateFunc != nil {
		return m.SaveLoopStateFunc(ctx, state)
	}
	return nil
}

func (m *MockStateManager) LoadLoopState(ctx context.Context, projectPath string) (*LoopState, error) {
	if m.LoadLoopStateFunc != nil {
		return m.LoadLoopStateFunc(ctx, projectPath)
	}
	return &LoopState{LoopNumber: 1}, nil
}

func (m *MockStateManager) SaveSpecState(ctx context.Context, state *SpecLoopState) error {
	if m.SaveSpecStateFunc != nil {
		return m.SaveSpecStateFunc(ctx, state)
	}
	return nil
}

func (m *MockStateManager) LoadSpecState(ctx context.Context, projectPath string) (*SpecLoopState, error) {
	if m.LoadSpecStateFunc != nil {
		return m.LoadSpecStateFunc(ctx, projectPath)
	}
	return &SpecLoopState{LoopNumber: 1}, nil
}

func (m *MockStateManager) ClearState(ctx context.Context, projectPath string) error {
	if m.ClearStateFunc != nil {
		return m.ClearStateFunc(ctx, projectPath)
	}
	return nil
}

func (m *MockStateManager) AtomicWrite(ctx context.Context, path string, data interface{}) error {
	if m.AtomicWriteFunc != nil {
		return m.AtomicWriteFunc(ctx, path, data)
	}
	return nil
}

// MockAuditLogger is a mock implementation of AuditLogger.
type MockAuditLogger struct {
	LogTaskFunc       func(ctx context.Context, event TaskLogEvent) error
	LogLoopFunc       func(ctx context.Context, event LoopLogEvent) error
	LogErrorFunc      func(ctx context.Context, event ErrorLogEvent) error
	LogSpecChangeFunc func(ctx context.Context, event SpecChangeEvent) error
	LogDecisionFunc   func(ctx context.Context, event DecisionLogEvent) error
	GetAuditTrailFunc func(ctx context.Context, projectPath string, since time.Time) ([]AuditEntry, error)
	ExportMetricsFunc func(ctx context.Context, projectPath string) (*BuildMetrics, error)
}

func (m *MockAuditLogger) LogTask(ctx context.Context, event TaskLogEvent) error {
	if m.LogTaskFunc != nil {
		return m.LogTaskFunc(ctx, event)
	}
	return nil
}

func (m *MockAuditLogger) LogLoop(ctx context.Context, event LoopLogEvent) error {
	if m.LogLoopFunc != nil {
		return m.LogLoopFunc(ctx, event)
	}
	return nil
}

func (m *MockAuditLogger) LogError(ctx context.Context, event ErrorLogEvent) error {
	if m.LogErrorFunc != nil {
		return m.LogErrorFunc(ctx, event)
	}
	return nil
}

func (m *MockAuditLogger) LogSpecChange(ctx context.Context, event SpecChangeEvent) error {
	if m.LogSpecChangeFunc != nil {
		return m.LogSpecChangeFunc(ctx, event)
	}
	return nil
}

func (m *MockAuditLogger) LogDecision(ctx context.Context, event DecisionLogEvent) error {
	if m.LogDecisionFunc != nil {
		return m.LogDecisionFunc(ctx, event)
	}
	return nil
}

func (m *MockAuditLogger) GetAuditTrail(ctx context.Context, projectPath string, since time.Time) ([]AuditEntry, error) {
	if m.GetAuditTrailFunc != nil {
		return m.GetAuditTrailFunc(ctx, projectPath, since)
	}
	return []AuditEntry{}, nil
}

func (m *MockAuditLogger) ExportMetrics(ctx context.Context, projectPath string) (*BuildMetrics, error) {
	if m.ExportMetricsFunc != nil {
		return m.ExportMetricsFunc(ctx, projectPath)
	}
	return &BuildMetrics{ProjectID: "test"}, nil
}

// MockVerifyReporter is a mock implementation of VerifyReporter.
type MockVerifyReporter struct {
	GenerateReportFunc func(ctx context.Context, task Task, spec *SpecContext, result VerifyInput) (*VerifyReport, error)
	CompareResultsFunc func(ctx context.Context, old, new VerifyReport) (*VerifyComparison, error)
}

func (m *MockVerifyReporter) GenerateReport(ctx context.Context, task Task, spec *SpecContext, result VerifyInput) (*VerifyReport, error) {
	if m.GenerateReportFunc != nil {
		return m.GenerateReportFunc(ctx, task, spec, result)
	}
	return &VerifyReport{
		TaskID:      task.ID,
		Verdict:     VerdictPass,
		PassedCount: 5,
		FailedCount: 0,
	}, nil
}

func (m *MockVerifyReporter) CompareResults(ctx context.Context, old, new VerifyReport) (*VerifyComparison, error) {
	if m.CompareResultsFunc != nil {
		return m.CompareResultsFunc(ctx, old, new)
	}
	return &VerifyComparison{Improved: false, Regressed: false}, nil
}

// MockWebSearchCache is a mock implementation of WebSearchCache.
type MockWebSearchCache struct {
	GetFunc       func(ctx context.Context, query string) (*CachedResult, error)
	SetFunc       func(ctx context.Context, query string, result *CachedResult) error
	IsExpiredFunc func(ctx context.Context, query string, maxAge time.Duration) (bool, error)
	ClearFunc     func(ctx context.Context) error
	SaveFunc      func(ctx context.Context) error
	LoadFunc      func(ctx context.Context) error
}

func (m *MockWebSearchCache) Get(ctx context.Context, query string) (*CachedResult, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, query)
	}
	return nil, nil
}

func (m *MockWebSearchCache) Set(ctx context.Context, query string, result *CachedResult) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, query, result)
	}
	return nil
}

func (m *MockWebSearchCache) IsExpired(ctx context.Context, query string, maxAge time.Duration) (bool, error) {
	if m.IsExpiredFunc != nil {
		return m.IsExpiredFunc(ctx, query, maxAge)
	}
	return false, nil
}

func (m *MockWebSearchCache) Clear(ctx context.Context) error {
	if m.ClearFunc != nil {
		return m.ClearFunc(ctx)
	}
	return nil
}

func (m *MockWebSearchCache) Save(ctx context.Context) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx)
	}
	return nil
}

func (m *MockWebSearchCache) Load(ctx context.Context) error {
	if m.LoadFunc != nil {
		return m.LoadFunc(ctx)
	}
	return nil
}

// MockAgentSpawner is a mock implementation of AgentSpawner.
type MockAgentSpawner struct {
	SpawnImplementationFunc func(ctx context.Context, config AgentConfig) (*AgentSession, error)
	SpawnVerificationFunc   func(ctx context.Context, config AgentConfig) (*AgentSession, error)
	TerminateFunc           func(ctx context.Context, sessionID string) error
	GetSessionStatusFunc    func(ctx context.Context, sessionID string) (*AgentSessionStatus, error)
}

func (m *MockAgentSpawner) SpawnImplementation(ctx context.Context, config AgentConfig) (*AgentSession, error) {
	if m.SpawnImplementationFunc != nil {
		return m.SpawnImplementationFunc(ctx, config)
	}
	return &AgentSession{SessionID: "session-1", TaskID: config.TaskID}, nil
}

func (m *MockAgentSpawner) SpawnVerification(ctx context.Context, config AgentConfig) (*AgentSession, error) {
	if m.SpawnVerificationFunc != nil {
		return m.SpawnVerificationFunc(ctx, config)
	}
	return &AgentSession{SessionID: "session-2", TaskID: config.TaskID}, nil
}

func (m *MockAgentSpawner) Terminate(ctx context.Context, sessionID string) error {
	if m.TerminateFunc != nil {
		return m.TerminateFunc(ctx, sessionID)
	}
	return nil
}

func (m *MockAgentSpawner) GetSessionStatus(ctx context.Context, sessionID string) (*AgentSessionStatus, error) {
	if m.GetSessionStatusFunc != nil {
		return m.GetSessionStatusFunc(ctx, sessionID)
	}
	return &AgentSessionStatus{SessionID: sessionID, Status: "completed"}, nil
}

// MockDocumentGenerator is a mock implementation of DocumentGenerator.
type MockDocumentGenerator struct {
	GenerateSpecDocFunc   func(ctx context.Context, spec *SpecOutput) (*GeneratedDoc, error)
	GenerateDesignDocFunc func(ctx context.Context, spec *SpecOutput) (*GeneratedDoc, error)
	GenerateTasksDocFunc  func(ctx context.Context, tasks []Task) (*GeneratedDoc, error)
	GenerateAgentsDocFunc func(ctx context.Context, config AgentsDocConfig) (*GeneratedDoc, error)
	GenerateSkillDocFunc  func(ctx context.Context, config SkillDocConfig) (*GeneratedDoc, error)
}

func (m *MockDocumentGenerator) GenerateSpecDoc(ctx context.Context, spec *SpecOutput) (*GeneratedDoc, error) {
	if m.GenerateSpecDocFunc != nil {
		return m.GenerateSpecDocFunc(ctx, spec)
	}
	return &GeneratedDoc{Path: "SPEC.md", Content: "# Spec", Hash: "abc123"}, nil
}

func (m *MockDocumentGenerator) GenerateDesignDoc(ctx context.Context, spec *SpecOutput) (*GeneratedDoc, error) {
	if m.GenerateDesignDocFunc != nil {
		return m.GenerateDesignDocFunc(ctx, spec)
	}
	return &GeneratedDoc{Path: "DESIGN.md", Content: "# Design", Hash: "def456"}, nil
}

func (m *MockDocumentGenerator) GenerateTasksDoc(ctx context.Context, tasks []Task) (*GeneratedDoc, error) {
	if m.GenerateTasksDocFunc != nil {
		return m.GenerateTasksDocFunc(ctx, tasks)
	}
	return &GeneratedDoc{Path: "TASKS.md", Content: "# Tasks", Hash: "ghi789"}, nil
}

func (m *MockDocumentGenerator) GenerateAgentsDoc(ctx context.Context, config AgentsDocConfig) (*GeneratedDoc, error) {
	if m.GenerateAgentsDocFunc != nil {
		return m.GenerateAgentsDocFunc(ctx, config)
	}
	return &GeneratedDoc{Path: "AGENTS.md", Content: "# Agents", Hash: "jkl012"}, nil
}

func (m *MockDocumentGenerator) GenerateSkillDoc(ctx context.Context, config SkillDocConfig) (*GeneratedDoc, error) {
	if m.GenerateSkillDocFunc != nil {
		return m.GenerateSkillDocFunc(ctx, config)
	}
	return &GeneratedDoc{Path: config.SkillPath + "/SKILL.md", Content: "# Skill", Hash: "mno345"}, nil
}

// MockQualityGate is a mock implementation of QualityGate.
type MockQualityGate struct {
	EvaluateFunc         func(ctx context.Context, specs []SpecArtifact) (*QualityGateResult, error)
	ShouldInvokeSpecFunc func(ctx context.Context, result *QualityGateResult) (bool, *FeedbackPayload, error)
	GetThresholdsFunc    func(ctx context.Context) (*QualityThresholds, error)
}

func (m *MockQualityGate) Evaluate(ctx context.Context, specs []SpecArtifact) (*QualityGateResult, error) {
	if m.EvaluateFunc != nil {
		return m.EvaluateFunc(ctx, specs)
	}
	return &QualityGateResult{
		Passed:       true,
		OverallScore: 90.0,
		Scores:       QualityScores{FlowCoverage: 9.0},
	}, nil
}

func (m *MockQualityGate) ShouldInvokeSpec(ctx context.Context, result *QualityGateResult) (bool, *FeedbackPayload, error) {
	if m.ShouldInvokeSpecFunc != nil {
		return m.ShouldInvokeSpecFunc(ctx, result)
	}
	return false, nil, nil
}

func (m *MockQualityGate) GetThresholds(ctx context.Context) (*QualityThresholds, error) {
	if m.GetThresholdsFunc != nil {
		return m.GetThresholdsFunc(ctx)
	}
	return &QualityThresholds{MinimumDimensionScore: 8, MinimumCompositeScore: 85}, nil
}

// MockEngramClient is a mock implementation of EngramClient.
type MockEngramClient struct {
	SaveFunc       func(ctx context.Context, observation EngramObservation) error
	SearchFunc     func(ctx context.Context, query string, options *SearchOptions) ([]EngramObservation, error)
	GetFunc        func(ctx context.Context, id int) (*EngramObservation, error)
	UpdateFunc     func(ctx context.Context, id int, observation EngramObservation) error
	DeleteFunc     func(ctx context.Context, id int) error
	GetSessionFunc func(ctx context.Context) (*SessionContext, error)
}

func (m *MockEngramClient) Save(ctx context.Context, observation EngramObservation) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, observation)
	}
	return nil
}

func (m *MockEngramClient) Search(ctx context.Context, query string, options *SearchOptions) ([]EngramObservation, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query, options)
	}
	return []EngramObservation{}, nil
}

func (m *MockEngramClient) Get(ctx context.Context, id int) (*EngramObservation, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return &EngramObservation{ID: id, Title: "test"}, nil
}

func (m *MockEngramClient) Update(ctx context.Context, id int, observation EngramObservation) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, observation)
	}
	return nil
}

func (m *MockEngramClient) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockEngramClient) GetSession(ctx context.Context) (*SessionContext, error) {
	if m.GetSessionFunc != nil {
		return m.GetSessionFunc(ctx)
	}
	return &SessionContext{ID: "session-1", Project: "test"}, nil
}

// MockProductionReadinessChecker is a mock implementation of ProductionReadinessChecker.
type MockProductionReadinessChecker struct {
	CheckFunc               func(ctx context.Context, projectPath string) (*ReadinessReport, error)
	CheckUserFlowsFunc      func(ctx context.Context, projectPath string) (*FlowCheckResult, error)
	CheckImportsFunc        func(ctx context.Context, projectPath string) (*ImportCheckResult, error)
	CheckDependenciesFunc   func(ctx context.Context, projectPath string) (*DependencyCheckResult, error)
	CheckSpecComplianceFunc func(ctx context.Context, projectPath string, specs []SpecArtifact) (*SpecComplianceResult, error)
}

func (m *MockProductionReadinessChecker) Check(ctx context.Context, projectPath string) (*ReadinessReport, error) {
	if m.CheckFunc != nil {
		return m.CheckFunc(ctx, projectPath)
	}
	return &ReadinessReport{Ready: true, TotalChecks: 10, PassedChecks: 10}, nil
}

func (m *MockProductionReadinessChecker) CheckUserFlows(ctx context.Context, projectPath string) (*FlowCheckResult, error) {
	if m.CheckUserFlowsFunc != nil {
		return m.CheckUserFlowsFunc(ctx, projectPath)
	}
	return &FlowCheckResult{Valid: true, Flows: []FlowStatus{}}, nil
}

func (m *MockProductionReadinessChecker) CheckImports(ctx context.Context, projectPath string) (*ImportCheckResult, error) {
	if m.CheckImportsFunc != nil {
		return m.CheckImportsFunc(ctx, projectPath)
	}
	return &ImportCheckResult{Valid: true}, nil
}

func (m *MockProductionReadinessChecker) CheckDependencies(ctx context.Context, projectPath string) (*DependencyCheckResult, error) {
	if m.CheckDependenciesFunc != nil {
		return m.CheckDependenciesFunc(ctx, projectPath)
	}
	return &DependencyCheckResult{Valid: true}, nil
}

func (m *MockProductionReadinessChecker) CheckSpecCompliance(ctx context.Context, projectPath string, specs []SpecArtifact) (*SpecComplianceResult, error) {
	if m.CheckSpecComplianceFunc != nil {
		return m.CheckSpecComplianceFunc(ctx, projectPath, specs)
	}
	return &SpecComplianceResult{Compliant: true}, nil
}

// ============================================================================
// Interface Compliance Tests
// ============================================================================

func TestSpecEngineInterface(t *testing.T) {
	var _ SpecEngine = (*MockSpecEngine)(nil)

	engine := &MockSpecEngine{}
	ctx := context.Background()

	// Test GenerateSpec
	spec, err := engine.GenerateSpec(ctx, SpecInput{ProjectPath: "/test"})
	if err != nil {
		t.Errorf("GenerateSpec failed: %v", err)
	}
	if spec == nil {
		t.Error("GenerateSpec returned nil")
	}

	// Test UpdateSpec
	updated, err := engine.UpdateSpec(ctx, UpdateInput{})
	if err != nil {
		t.Errorf("UpdateSpec failed: %v", err)
	}
	if updated == nil {
		t.Error("UpdateSpec returned nil")
	}

	// Test ValidateSpec
	validated, err := engine.ValidateSpec(ctx, SpecArtifact{})
	if err != nil {
		t.Errorf("ValidateSpec failed: %v", err)
	}
	if validated == nil {
		t.Error("ValidateSpec returned nil")
	}

	// Test Decompose
	components, err := engine.Decompose(ctx, DecomposeInput{ComponentName: "test"})
	if err != nil {
		t.Errorf("Decompose failed: %v", err)
	}
	if len(components) == 0 {
		t.Error("Decompose returned no components")
	}
}

func TestSpecScorerInterface(t *testing.T) {
	var _ SpecScorer = (*MockSpecScorer)(nil)

	scorer := &MockSpecScorer{}
	ctx := context.Background()

	// Test ScoreSpec
	scores, err := scorer.ScoreSpec(ctx, SpecArtifact{})
	if err != nil {
		t.Errorf("ScoreSpec failed: %v", err)
	}
	if scores == nil {
		t.Error("ScoreSpec returned nil")
	}

	// Test IsComplete
	complete, reason, err := scorer.IsComplete(ctx, QualityScores{})
	if err != nil {
		t.Errorf("IsComplete failed: %v", err)
	}
	if complete && reason != "normal" {
		t.Errorf("IsComplete returned unexpected result: %v, %v", complete, reason)
	}

	// Test CalculateDelta
	delta, err := scorer.CalculateDelta(ctx, SpecArtifact{}, SpecArtifact{})
	if err != nil {
		t.Errorf("CalculateDelta failed: %v", err)
	}
	if delta < 0 {
		t.Error("CalculateDelta returned negative delta")
	}
}

func TestLoopValidatorInterface(t *testing.T) {
	var _ LoopValidator = (*MockLoopValidator)(nil)

	validator := &MockLoopValidator{}
	ctx := context.Background()

	// Test ValidateAgents
	result, err := validator.ValidateAgents(ctx, "/test")
	if err != nil {
		t.Errorf("ValidateAgents failed: %v", err)
	}
	if result == nil {
		t.Error("ValidateAgents returned nil")
	}

	// Test ValidateSkills
	skillsResult, err := validator.ValidateSkills(ctx, "/test")
	if err != nil {
		t.Errorf("ValidateSkills failed: %v", err)
	}
	if skillsResult == nil {
		t.Error("ValidateSkills returned nil")
	}

	// Test ValidateSpecs
	specsResult, err := validator.ValidateSpecs(ctx, "/test")
	if err != nil {
		t.Errorf("ValidateSpecs failed: %v", err)
	}
	if specsResult == nil {
		t.Error("ValidateSpecs returned nil")
	}

	// Test ValidateStack
	stackResult, err := validator.ValidateStack(ctx, SpecArtifact{})
	if err != nil {
		t.Errorf("ValidateStack failed: %v", err)
	}
	if stackResult == nil {
		t.Error("ValidateStack returned nil")
	}

	// Test ValidateDependencies
	depsResult, err := validator.ValidateDependencies(ctx, []Task{})
	if err != nil {
		t.Errorf("ValidateDependencies failed: %v", err)
	}
	if depsResult == nil {
		t.Error("ValidateDependencies returned nil")
	}

	// Test ValidateAll
	allResult, err := validator.ValidateAll(ctx, "/test")
	if err != nil {
		t.Errorf("ValidateAll failed: %v", err)
	}
	if allResult == nil {
		t.Error("ValidateAll returned nil")
	}
}

func TestLoopOrchestratorInterface(t *testing.T) {
	var _ LoopOrchestrator = (*MockLoopOrchestrator)(nil)

	orchestrator := &MockLoopOrchestrator{}
	ctx := context.Background()

	// Test Start
	err := orchestrator.Start(ctx, LoopConfig{})
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}

	// Test Pause
	err = orchestrator.Pause(ctx)
	if err != nil {
		t.Errorf("Pause failed: %v", err)
	}

	// Test Resume
	err = orchestrator.Resume(ctx)
	if err != nil {
		t.Errorf("Resume failed: %v", err)
	}

	// Test Stop
	err = orchestrator.Stop(ctx)
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	// Test GetState
	state, err := orchestrator.GetState(ctx)
	if err != nil {
		t.Errorf("GetState failed: %v", err)
	}
	if state == nil {
		t.Error("GetState returned nil")
	}

	// Test ExecuteTask
	result, err := orchestrator.ExecuteTask(ctx, Task{ID: "test-1"})
	if err != nil {
		t.Errorf("ExecuteTask failed: %v", err)
	}
	if result == nil {
		t.Error("ExecuteTask returned nil")
	}

	// Test ExecutePhase
	phaseResult, err := orchestrator.ExecutePhase(ctx, PhaseSpec)
	if err != nil {
		t.Errorf("ExecutePhase failed: %v", err)
	}
	if phaseResult == nil {
		t.Error("ExecutePhase returned nil")
	}
}

func TestIntentClassifierInterface(t *testing.T) {
	var _ IntentClassifier = (*MockIntentClassifier)(nil)

	classifier := &MockIntentClassifier{}
	ctx := context.Background()

	// Test ClassifyIntent
	classification, err := classifier.ClassifyIntent(ctx, "create a new feature")
	if err != nil {
		t.Errorf("ClassifyIntent failed: %v", err)
	}
	if classification == nil {
		t.Error("ClassifyIntent returned nil")
	}

	// Test ExtractEntities
	entities, err := classifier.ExtractEntities(ctx, "create a new feature")
	if err != nil {
		t.Errorf("ExtractEntities failed: %v", err)
	}
	if entities == nil {
		t.Error("ExtractEntities returned nil")
	}

	// Test DetectAmbiguity
	ambiguity, err := classifier.DetectAmbiguity(ctx, "create something")
	if err != nil {
		t.Errorf("DetectAmbiguity failed: %v", err)
	}
	if ambiguity == nil {
		t.Error("DetectAmbiguity returned nil")
	}
}

func TestContextCollectorInterface(t *testing.T) {
	var _ ContextCollector = (*MockContextCollector)(nil)

	collector := &MockContextCollector{}
	ctx := context.Background()

	// Test CollectTaskContext
	taskCtx, err := collector.CollectTaskContext(ctx, Task{ID: "test-1"})
	if err != nil {
		t.Errorf("CollectTaskContext failed: %v", err)
	}
	if taskCtx == nil {
		t.Error("CollectTaskContext returned nil")
	}

	// Test CollectSpecContext
	specCtx, err := collector.CollectSpecContext(ctx, "spec")
	if err != nil {
		t.Errorf("CollectSpecContext failed: %v", err)
	}
	if specCtx == nil {
		t.Error("CollectSpecContext returned nil")
	}

	// Test CollectAgentsContext
	agentsCtx, err := collector.CollectAgentsContext(ctx, "module")
	if err != nil {
		t.Errorf("CollectAgentsContext failed: %v", err)
	}
	if agentsCtx == nil {
		t.Error("CollectAgentsContext returned nil")
	}

	// Test CollectSkillsContext
	skillsCtx, err := collector.CollectSkillsContext(ctx, []string{"skill1"})
	if err != nil {
		t.Errorf("CollectSkillsContext failed: %v", err)
	}
	if skillsCtx == nil {
		t.Error("CollectSkillsContext returned nil")
	}

	// Test BoundContext
	bounded, err := collector.BoundContext(ctx, nil, 1000)
	if err != nil {
		t.Errorf("BoundContext failed: %v", err)
	}
	if bounded == nil {
		t.Error("BoundContext returned nil")
	}
}

func TestStateManagerInterface(t *testing.T) {
	var _ StateManager = (*MockStateManager)(nil)

	manager := &MockStateManager{}
	ctx := context.Background()

	// Test SaveLoopState
	state := &LoopState{LoopNumber: 1}
	err := manager.SaveLoopState(ctx, state)
	if err != nil {
		t.Errorf("SaveLoopState failed: %v", err)
	}

	// Test LoadLoopState
	loaded, err := manager.LoadLoopState(ctx, "/test")
	if err != nil {
		t.Errorf("LoadLoopState failed: %v", err)
	}
	if loaded == nil {
		t.Error("LoadLoopState returned nil")
	}

	// Test SaveSpecState
	specState := &SpecLoopState{LoopNumber: 1}
	err = manager.SaveSpecState(ctx, specState)
	if err != nil {
		t.Errorf("SaveSpecState failed: %v", err)
	}

	// Test LoadSpecState
	specLoaded, err := manager.LoadSpecState(ctx, "/test")
	if err != nil {
		t.Errorf("LoadSpecState failed: %v", err)
	}
	if specLoaded == nil {
		t.Error("LoadSpecState returned nil")
	}

	// Test ClearState
	err = manager.ClearState(ctx, "/test")
	if err != nil {
		t.Errorf("ClearState failed: %v", err)
	}

	// Test AtomicWrite
	err = manager.AtomicWrite(ctx, "/test/file.json", state)
	if err != nil {
		t.Errorf("AtomicWrite failed: %v", err)
	}
}

func TestAuditLoggerInterface(t *testing.T) {
	var _ AuditLogger = (*MockAuditLogger)(nil)

	logger := &MockAuditLogger{}
	ctx := context.Background()

	// Test LogTask
	err := logger.LogTask(ctx, TaskLogEvent{TaskID: "test-1", Action: "started"})
	if err != nil {
		t.Errorf("LogTask failed: %v", err)
	}

	// Test LogLoop
	err = logger.LogLoop(ctx, LoopLogEvent{LoopNumber: 1, Action: "started"})
	if err != nil {
		t.Errorf("LogLoop failed: %v", err)
	}

	// Test LogError
	err = logger.LogError(ctx, ErrorLogEvent{Error: "test error", Type: "test"})
	if err != nil {
		t.Errorf("LogError failed: %v", err)
	}

	// Test LogSpecChange
	err = logger.LogSpecChange(ctx, SpecChangeEvent{ChangeType: "added", Location: "test"})
	if err != nil {
		t.Errorf("LogSpecChange failed: %v", err)
	}

	// Test LogDecision
	err = logger.LogDecision(ctx, DecisionLogEvent{DecisionID: "dec-1", Decision: "test decision"})
	if err != nil {
		t.Errorf("LogDecision failed: %v", err)
	}

	// Test GetAuditTrail
	trail, err := logger.GetAuditTrail(ctx, "/test", time.Now().Add(-time.Hour))
	if err != nil {
		t.Errorf("GetAuditTrail failed: %v", err)
	}
	if trail == nil {
		t.Error("GetAuditTrail returned nil")
	}

	// Test ExportMetrics
	metrics, err := logger.ExportMetrics(ctx, "/test")
	if err != nil {
		t.Errorf("ExportMetrics failed: %v", err)
	}
	if metrics == nil {
		t.Error("ExportMetrics returned nil")
	}
}

func TestVerifyReporterInterface(t *testing.T) {
	var _ VerifyReporter = (*MockVerifyReporter)(nil)

	reporter := &MockVerifyReporter{}
	ctx := context.Background()

	// Test GenerateReport
	report, err := reporter.GenerateReport(ctx, Task{ID: "test-1"}, nil, VerifyInput{})
	if err != nil {
		t.Errorf("GenerateReport failed: %v", err)
	}
	if report == nil {
		t.Error("GenerateReport returned nil")
	}

	// Test CompareResults
	comparison, err := reporter.CompareResults(ctx, VerifyReport{}, VerifyReport{})
	if err != nil {
		t.Errorf("CompareResults failed: %v", err)
	}
	if comparison == nil {
		t.Error("CompareResults returned nil")
	}
}

func TestWebSearchCacheInterface(t *testing.T) {
	var _ WebSearchCache = (*MockWebSearchCache)(nil)

	cache := &MockWebSearchCache{}
	ctx := context.Background()

	// Test Get
	_, err := cache.Get(ctx, "test query")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	// result can be nil for cache miss

	// Test Set
	err = cache.Set(ctx, "test query", &CachedResult{Query: "test"})
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// Test IsExpired
	expired, err := cache.IsExpired(ctx, "test query", time.Hour)
	if err != nil {
		t.Errorf("IsExpired failed: %v", err)
	}
	if expired {
		t.Error("IsExpired should return false for new entry")
	}

	// Test Clear
	err = cache.Clear(ctx)
	if err != nil {
		t.Errorf("Clear failed: %v", err)
	}

	// Test Save
	err = cache.Save(ctx)
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Test Load
	err = cache.Load(ctx)
	if err != nil {
		t.Errorf("Load failed: %v", err)
	}
}

func TestAgentSpawnerInterface(t *testing.T) {
	var _ AgentSpawner = (*MockAgentSpawner)(nil)

	spawner := &MockAgentSpawner{}
	ctx := context.Background()

	// Test SpawnImplementation
	session, err := spawner.SpawnImplementation(ctx, AgentConfig{TaskID: "test-1"})
	if err != nil {
		t.Errorf("SpawnImplementation failed: %v", err)
	}
	if session == nil {
		t.Error("SpawnImplementation returned nil")
	}

	// Test SpawnVerification
	verifySession, err := spawner.SpawnVerification(ctx, AgentConfig{TaskID: "test-1"})
	if err != nil {
		t.Errorf("SpawnVerification failed: %v", err)
	}
	if verifySession == nil {
		t.Error("SpawnVerification returned nil")
	}

	// Test Terminate
	err = spawner.Terminate(ctx, "session-1")
	if err != nil {
		t.Errorf("Terminate failed: %v", err)
	}

	// Test GetSessionStatus
	status, err := spawner.GetSessionStatus(ctx, "session-1")
	if err != nil {
		t.Errorf("GetSessionStatus failed: %v", err)
	}
	if status == nil {
		t.Error("GetSessionStatus returned nil")
	}
}

func TestDocumentGeneratorInterface(t *testing.T) {
	var _ DocumentGenerator = (*MockDocumentGenerator)(nil)

	generator := &MockDocumentGenerator{}
	ctx := context.Background()

	// Test GenerateSpecDoc
	doc, err := generator.GenerateSpecDoc(ctx, &SpecOutput{})
	if err != nil {
		t.Errorf("GenerateSpecDoc failed: %v", err)
	}
	if doc == nil {
		t.Error("GenerateSpecDoc returned nil")
	}

	// Test GenerateDesignDoc
	designDoc, err := generator.GenerateDesignDoc(ctx, &SpecOutput{})
	if err != nil {
		t.Errorf("GenerateDesignDoc failed: %v", err)
	}
	if designDoc == nil {
		t.Error("GenerateDesignDoc returned nil")
	}

	// Test GenerateTasksDoc
	tasksDoc, err := generator.GenerateTasksDoc(ctx, []Task{})
	if err != nil {
		t.Errorf("GenerateTasksDoc failed: %v", err)
	}
	if tasksDoc == nil {
		t.Error("GenerateTasksDoc returned nil")
	}

	// Test GenerateAgentsDoc
	agentsDoc, err := generator.GenerateAgentsDoc(ctx, AgentsDocConfig{})
	if err != nil {
		t.Errorf("GenerateAgentsDoc failed: %v", err)
	}
	if agentsDoc == nil {
		t.Error("GenerateAgentsDoc returned nil")
	}

	// Test GenerateSkillDoc
	skillDoc, err := generator.GenerateSkillDoc(ctx, SkillDocConfig{SkillName: "test", SkillPath: "/test"})
	if err != nil {
		t.Errorf("GenerateSkillDoc failed: %v", err)
	}
	if skillDoc == nil {
		t.Error("GenerateSkillDoc returned nil")
	}
}

func TestQualityGateInterface(t *testing.T) {
	var _ QualityGate = (*MockQualityGate)(nil)

	gate := &MockQualityGate{}
	ctx := context.Background()

	// Test Evaluate
	result, err := gate.Evaluate(ctx, []SpecArtifact{})
	if err != nil {
		t.Errorf("Evaluate failed: %v", err)
	}
	if result == nil {
		t.Error("Evaluate returned nil")
	}

	// Test ShouldInvokeSpec
	should, payload, err := gate.ShouldInvokeSpec(ctx, &QualityGateResult{})
	if err != nil {
		t.Errorf("ShouldInvokeSpec failed: %v", err)
	}
	if should && payload != nil {
		t.Error("ShouldInvokeSpec returned unexpected payload")
	}

	// Test GetThresholds
	thresholds, err := gate.GetThresholds(ctx)
	if err != nil {
		t.Errorf("GetThresholds failed: %v", err)
	}
	if thresholds == nil {
		t.Error("GetThresholds returned nil")
	}
}

func TestEngramClientInterface(t *testing.T) {
	var _ EngramClient = (*MockEngramClient)(nil)

	client := &MockEngramClient{}
	ctx := context.Background()

	// Test Save
	err := client.Save(ctx, EngramObservation{Title: "test", Type: "decision"})
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Test Search
	results, err := client.Search(ctx, "test", nil)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}
	if results == nil {
		t.Error("Search returned nil")
	}

	// Test Get
	obs, err := client.Get(ctx, 1)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if obs == nil {
		t.Error("Get returned nil")
	}

	// Test Update
	err = client.Update(ctx, 1, EngramObservation{Title: "updated"})
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// Test Delete
	err = client.Delete(ctx, 1)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Test GetSession
	session, err := client.GetSession(ctx)
	if err != nil {
		t.Errorf("GetSession failed: %v", err)
	}
	if session == nil {
		t.Error("GetSession returned nil")
	}
}

func TestProductionReadinessCheckerInterface(t *testing.T) {
	var _ ProductionReadinessChecker = (*MockProductionReadinessChecker)(nil)

	checker := &MockProductionReadinessChecker{}
	ctx := context.Background()

	// Test Check
	report, err := checker.Check(ctx, "/test")
	if err != nil {
		t.Errorf("Check failed: %v", err)
	}
	if report == nil {
		t.Error("Check returned nil")
	}

	// Test CheckUserFlows
	flowsResult, err := checker.CheckUserFlows(ctx, "/test")
	if err != nil {
		t.Errorf("CheckUserFlows failed: %v", err)
	}
	if flowsResult == nil {
		t.Error("CheckUserFlows returned nil")
	}

	// Test CheckImports
	importsResult, err := checker.CheckImports(ctx, "/test")
	if err != nil {
		t.Errorf("CheckImports failed: %v", err)
	}
	if importsResult == nil {
		t.Error("CheckImports returned nil")
	}

	// Test CheckDependencies
	depsResult, err := checker.CheckDependencies(ctx, "/test")
	if err != nil {
		t.Errorf("CheckDependencies failed: %v", err)
	}
	if depsResult == nil {
		t.Error("CheckDependencies returned nil")
	}

	// Test CheckSpecCompliance
	complianceResult, err := checker.CheckSpecCompliance(ctx, "/test", []SpecArtifact{})
	if err != nil {
		t.Errorf("CheckSpecCompliance failed: %v", err)
	}
	if complianceResult == nil {
		t.Error("CheckSpecCompliance returned nil")
	}
}

// ============================================================================
// Type and Helper Tests
// ============================================================================

func TestQualityScoresCompositeScore(t *testing.T) {
	scores := QualityScores{
		FlowCoverage:          8.0,
		ComponentDepth:        8.0,
		LogicalConsistency:    8.0,
		Connectivity:          8.0,
		EdgeCases:             8.0,
		DecisionJustification: 8.0,
		AgentConsumability:    8.0,
	}

	composite := scores.CompositeScore()
	expected := 8.0 * 0.20 * 5 // FlowCoverage + ComponentDepth + LogicalConsistency + Connectivity + EdgeCases
	expected += 8.0 * 0.10 * 2 // DecisionJustification + AgentConsumability

	if composite != expected {
		t.Errorf("CompositeScore() = %v, want %v", composite, expected)
	}
}

func TestTaskStatusValues(t *testing.T) {
	statuses := []TaskStatus{
		TaskStatusPending,
		TaskStatusRunning,
		TaskStatusCompleted,
		TaskStatusFailed,
		TaskStatusDeferred,
		TaskStatusSkipped,
	}

	expected := []string{"pending", "running", "completed", "failed", "deferred", "skipped"}

	for i, status := range statuses {
		if string(status) != expected[i] {
			t.Errorf("TaskStatus = %v, want %v", status, expected[i])
		}
	}
}

func TestSpecPhaseValues(t *testing.T) {
	phases := []SpecPhase{
		PhaseExplore,
		PhaseProposal,
		PhaseSpec,
		PhaseDesign,
		PhaseTasks,
		PhaseApply,
		PhaseVerify,
		PhaseArchive,
	}

	expected := []string{"explore", "proposal", "spec", "design", "tasks", "apply", "verify", "archive"}

	for i, phase := range phases {
		if string(phase) != expected[i] {
			t.Errorf("SpecPhase = %v, want %v", phase, expected[i])
		}
	}
}

func TestExitConditionValues(t *testing.T) {
	conditions := []ExitCondition{
		ExitNormal,
		ExitSafetyNet,
		ExitManual,
		ExitError,
	}

	expected := []string{"normal", "safety_net", "manual", "error"}

	for i, cond := range conditions {
		if string(cond) != expected[i] {
			t.Errorf("ExitCondition = %v, want %v", cond, expected[i])
		}
	}
}

func TestIntentTypeValues(t *testing.T) {
	intents := []IntentType{
		IntentCreate,
		IntentUpdate,
		IntentDelete,
		IntentQuery,
		IntentClarify,
		IntentContinue,
		IntentStop,
		IntentHelp,
	}

	expected := []string{"create", "update", "delete", "query", "clarify", "continue", "stop", "help"}

	for i, intent := range intents {
		if string(intent) != expected[i] {
			t.Errorf("IntentType = %v, want %v", intent, expected[i])
		}
	}
}

func TestVerdictValues(t *testing.T) {
	if string(VerdictPass) != "PASS" {
		t.Errorf("VerdictPass = %v, want PASS", VerdictPass)
	}
	if string(VerdictFail) != "FAIL" {
		t.Errorf("VerdictFail = %v, want FAIL", VerdictFail)
	}
}

func TestMockWithCustomFunctions(t *testing.T) {
	// Test SpecEngine with custom function
	customCalled := false
	engine := &MockSpecEngine{
		GenerateSpecFunc: func(ctx context.Context, input SpecInput) (*SpecOutput, error) {
			customCalled = true
			return &SpecOutput{
				Spec:  &SpecArtifact{Phase: PhaseSpec, Content: "custom spec"},
				Tasks: []Task{{ID: "custom-task"}},
			}, nil
		},
	}

	result, err := engine.GenerateSpec(context.Background(), SpecInput{ProjectPath: "/custom"})
	if err != nil {
		t.Errorf("Custom GenerateSpec failed: %v", err)
	}
	if !customCalled {
		t.Error("Custom function was not called")
	}
	if len(result.Tasks) != 1 || result.Tasks[0].ID != "custom-task" {
		t.Error("Custom function did not return expected result")
	}

	// Test LoopOrchestrator with custom state
	orchestrator := &MockLoopOrchestrator{
		GetStateFunc: func(ctx context.Context) (*LoopState, error) {
			return &LoopState{
				LoopNumber:     5,
				CurrentTaskID:  "task-42",
				CompletedTasks: []string{"task-1", "task-2"},
			}, nil
		},
	}

	state, err := orchestrator.GetState(context.Background())
	if err != nil {
		t.Errorf("Custom GetState failed: %v", err)
	}
	if state.LoopNumber != 5 {
		t.Errorf("LoopNumber = %v, want 5", state.LoopNumber)
	}
	if state.CurrentTaskID != "task-42" {
		t.Errorf("CurrentTaskID = %v, want task-42", state.CurrentTaskID)
	}
}

func TestMockAgentSpawnerWithVerification(t *testing.T) {
	spawner := &MockAgentSpawner{
		SpawnVerificationFunc: func(ctx context.Context, config AgentConfig) (*AgentSession, error) {
			if config.TaskType != "verification" {
				t.Errorf("TaskType = %v, want verification", config.TaskType)
			}
			return &AgentSession{
				SessionID:   "verify-session-1",
				TaskID:      config.TaskID,
				StartedAt:   time.Now(),
				ContextUsed: 500,
			}, nil
		},
	}

	session, err := spawner.SpawnVerification(context.Background(), AgentConfig{
		TaskID:   "test-1",
		TaskType: "verification",
	})
	if err != nil {
		t.Errorf("SpawnVerification failed: %v", err)
	}
	if session.SessionID != "verify-session-1" {
		t.Errorf("SessionID = %v, want verify-session-1", session.SessionID)
	}
}

func TestMockAuditLoggerWithErrorEvents(t *testing.T) {
	logger := &MockAuditLogger{
		LogErrorFunc: func(ctx context.Context, event ErrorLogEvent) error {
			if event.Type != "llm_failure" {
				t.Errorf("Error type = %v, want llm_failure", event.Type)
			}
			if !event.Retriable {
				t.Error("Expected error to be retriable")
			}
			return nil
		},
	}

	err := logger.LogError(context.Background(), ErrorLogEvent{
		TaskID:     "task-1",
		LoopNumber: 2,
		Error:      "LLM returned empty response",
		Type:       "llm_failure",
		Retriable:  true,
	})
	if err != nil {
		t.Errorf("LogError failed: %v", err)
	}
}

func TestMockQualityGateWithFeedback(t *testing.T) {
	gate := &MockQualityGate{
		ShouldInvokeSpecFunc: func(ctx context.Context, result *QualityGateResult) (bool, *FeedbackPayload, error) {
			if !result.Passed && result.OverallScore < 85 {
				return true, &FeedbackPayload{
					Trigger:      "quality-gate-failure",
					LoopNumber:   1,
					QualityScore: result.OverallScore,
					Missing:      []string{"error handling", "edge cases"},
				}, nil
			}
			return false, nil, nil
		},
	}

	// Test case where spec should be invoked
	shouldInvoke, payload, err := gate.ShouldInvokeSpec(context.Background(), &QualityGateResult{
		Passed:       false,
		OverallScore: 70,
	})
	if err != nil {
		t.Errorf("ShouldInvokeSpec failed: %v", err)
	}
	if !shouldInvoke {
		t.Error("Expected spec to be invoked for low score")
	}
	if payload == nil {
		t.Error("Expected feedback payload")
	}
	if payload.Trigger != "quality-gate-failure" {
		t.Errorf("Trigger = %v, want quality-gate-failure", payload.Trigger)
	}

	// Test case where spec should not be invoked
	shouldNotInvoke, _, err := gate.ShouldInvokeSpec(context.Background(), &QualityGateResult{
		Passed:       true,
		OverallScore: 90,
	})
	if err != nil {
		t.Errorf("ShouldInvokeSpec failed: %v", err)
	}
	if shouldNotInvoke {
		t.Error("Expected spec not to be invoked for high score")
	}
}
