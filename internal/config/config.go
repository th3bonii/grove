// Package config provides configuration management for the GROVE ecosystem.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	gerrors "github.com/Gentleman-Programming/grove/internal/errors"
)

// Default configuration values
const (
	DefaultProjectName        = "unnamed-project"
	DefaultMaxIterations      = 10
	DefaultQualityThreshold   = 0.75
	DefaultMaxRetries         = 3
	DefaultBackoffBaseSeconds = 2
	DefaultStateDirName       = ".grove-state"
	DefaultSpecDirName        = "spec"
	DefaultSkillsDirName      = ".opencode/skills"
	DefaultAgentsDirName      = ".opencode/agents"
)

// Config represents the main GROVE configuration.
type Config struct {
	// Project identification
	ProjectName string `json:"project_name" yaml:"project_name"`
	ProjectPath string `json:"project_path" yaml:"project_path"`

	// Spec configuration
	Spec SpecConfig `json:"spec" yaml:"spec"`

	// Loop configuration (Ralph Loop)
	Loop LoopConfig `json:"loop" yaml:"loop"`

	// Opti Prompt configuration
	Opti OptiConfig `json:"opti" yaml:"opti"`

	// Global settings
	MaxIterations    int                `json:"max_iterations" yaml:"max_iterations"`
	QualityThreshold float64            `json:"quality_threshold" yaml:"quality_threshold"`
	Verbose          bool               `json:"verbose" yaml:"verbose"`
	ScoringWeights   map[string]float64 `json:"scoring_weights,omitempty" yaml:"scoring_weights,omitempty"`
}

// SpecConfig contains configuration for GROVE Spec engine.
type SpecConfig struct {
	// Output settings
	OutputPath string `json:"output_path" yaml:"output_path"`
	SpecDir    string `json:"spec_dir" yaml:"spec_dir"`

	// Generation settings
	EnableSelfQuestioning bool `json:"enable_self_questioning" yaml:"enable_self_questioning"`
	MaxDepth              int  `json:"max_depth" yaml:"max_depth"`
	IncludeExamples       bool `json:"include_examples" yaml:"include_examples"`
	IncludeGlossary       bool `json:"include_glossary" yaml:"include_glossary"`

	// File generation
	GenerateSpec   bool `json:"generate_spec" yaml:"generate_spec"`
	GenerateDesign bool `json:"generate_design" yaml:"generate_design"`
	GenerateTasks  bool `json:"generate_tasks" yaml:"generate_tasks"`
	GenerateAgents bool `json:"generate_agents" yaml:"generate_agents"`
	GenerateSkills bool `json:"generate_skills" yaml:"generate_skills"`

	// Quality settings
	MinComponents    int     `json:"min_components" yaml:"min_components"`
	MinScenarios     int     `json:"min_scenarios" yaml:"min_scenarios"`
	QualityThreshold float64 `json:"quality_threshold" yaml:"quality_threshold"`
}

// LoopConfig contains configuration for GROVE Ralph Loop.
type LoopConfig struct {
	// State persistence
	StateDir    string `json:"state_dir" yaml:"state_dir"`
	StateFile   string `json:"state_file" yaml:"state_file"`
	LogFile     string `json:"log_file" yaml:"log_file"`
	MetricsFile string `json:"metrics_file" yaml:"metrics_file"`
	ReadyReport string `json:"ready_report" yaml:"ready_report"`

	// Execution settings
	AutoResume         bool `json:"auto_resume" yaml:"auto_resume"`
	ForceRestart       bool `json:"force_restart" yaml:"force_restart"`
	SkipValidation     bool `json:"skip_validation" yaml:"skip_validation"`
	SkipQualityGate    bool `json:"skip_quality_gate" yaml:"skip_quality_gate"`
	EnableReinvokeSpec bool `json:"enable_reinvoke_spec" yaml:"enable_reinvoke_spec"`

	// Retry settings
	MaxRetries         int `json:"max_retries" yaml:"max_retries"`
	BackoffBaseSeconds int `json:"backoff_base_seconds" yaml:"backoff_base_seconds"`

	// Agent settings
	EnableScopedAgents bool `json:"enable_scoped_agents" yaml:"enable_scoped_agents"`
	ContextWindowLimit int  `json:"context_window_limit" yaml:"context_window_limit"`

	// Quality gate
	QualityGateThreshold float64 `json:"quality_gate_threshold" yaml:"quality_gate_threshold"`
	DimensionThreshold   float64 `json:"dimension_threshold" yaml:"dimension_threshold"`
}

// OptiConfig contains configuration for GROVE Opti Prompt.
type OptiConfig struct {
	// Output settings
	OutputFormat string `json:"output_format" yaml:"output_format"` // "markdown", "plain"

	// Learning settings
	EnableExplanations bool `json:"enable_explanations" yaml:"enable_explain"`
	ShowTokenEstimate  bool `json:"show_token_estimate" yaml:"show_token_estimate"`
	TrackImprovement   bool `json:"track_improvement" yaml:"track_improvement"`

	// Context settings
	MaxContextFiles     int      `json:"max_context_files" yaml:"max_context_files"`
	ContextFilePatterns []string `json:"context_file_patterns,omitempty" yaml:"context_file_patterns,omitempty"`

	// Optimization settings
	IncludeFileReferences bool `json:"include_file_references" yaml:"include_file_references"`
	IncludeSkillCalls     bool `json:"include_skill_calls" yaml:"include_skill_calls"`
	BoundContextSize      int  `json:"bound_context_size" yaml:"bound_context_size"`
}

// Global configuration instance (thread-safe singleton)
var (
	cfg     *Config
	cfgOnce sync.Once
	cfgPath string
)

// LoadConfig loads configuration from the specified path.
// If path is empty, it looks for .grove/config.yaml in the current directory
// or falls back to default configuration.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get current directory: %w", err)
		}
		path = filepath.Join(cwd, ".grove", "config.yaml")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return default configuration if file doesn't exist
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// Apply defaults for missing fields
	config.applyDefaults()

	return &config, nil
}

// LoadGlobalConfig loads the global configuration (singleton pattern).
// Uses ~/.config/grove/config.yaml or falls back to defaults.
func LoadGlobalConfig() (*Config, error) {
	var loadErr error

	cfgOnce.Do(func() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			loadErr = fmt.Errorf("get home directory: %w", err)
			return
		}

		cfgPath = filepath.Join(homeDir, ".config", "grove", "config.yaml")
		cfg, loadErr = LoadConfig(cfgPath)
	})

	if loadErr != nil {
		return nil, loadErr
	}

	return cfg, nil
}

// GetConfig returns the global configuration instance.
// Panics if LoadGlobalConfig has not been called first.
func GetConfig() *Config {
	if cfg == nil {
		panic("configuration not loaded: call LoadGlobalConfig first")
	}
	return cfg
}

// DefaultConfig returns a configuration with default values.
func DefaultConfig() *Config {
	return &Config{
		ProjectName:      DefaultProjectName,
		ProjectPath:      "",
		MaxIterations:    DefaultMaxIterations,
		QualityThreshold: DefaultQualityThreshold,
		Verbose:          false,
		ScoringWeights:   defaultScoringWeights(),
		Spec: SpecConfig{
			OutputPath:            "spec",
			SpecDir:               DefaultSpecDirName,
			EnableSelfQuestioning: true,
			MaxDepth:              5,
			IncludeExamples:       true,
			IncludeGlossary:       true,
			GenerateSpec:          true,
			GenerateDesign:        true,
			GenerateTasks:         true,
			GenerateAgents:        true,
			GenerateSkills:        false,
			MinComponents:         3,
			MinScenarios:          5,
			QualityThreshold:      DefaultQualityThreshold,
		},
		Loop: LoopConfig{
			StateDir:             DefaultStateDirName,
			StateFile:            "GROVE-LOOP-STATE.json",
			LogFile:              "GROVE-LOOP-LOG.md",
			MetricsFile:          "GROVE-LOOP-METRICS.json",
			ReadyReport:          "GROVE-READY-REPORT.md",
			AutoResume:           true,
			ForceRestart:         false,
			SkipValidation:       false,
			SkipQualityGate:      false,
			EnableReinvokeSpec:   true,
			MaxRetries:           DefaultMaxRetries,
			BackoffBaseSeconds:   DefaultBackoffBaseSeconds,
			EnableScopedAgents:   true,
			ContextWindowLimit:   100000,
			QualityGateThreshold: 0.70,
			DimensionThreshold:   0.80,
		},
		Opti: OptiConfig{
			OutputFormat:          "markdown",
			EnableExplanations:    true,
			ShowTokenEstimate:     true,
			TrackImprovement:      true,
			MaxContextFiles:       10,
			ContextFilePatterns:   []string{"*.go", "*.ts", "*.tsx", "*.md"},
			IncludeFileReferences: true,
			IncludeSkillCalls:     true,
			BoundContextSize:      50000,
		},
	}
}

// GetProjectDir returns the project directory path.
// If projectPath is not set, uses the current working directory.
func GetProjectDir(config *Config) string {
	if config.ProjectPath != "" {
		return config.ProjectPath
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

// GetSkillsDir returns the skills directory path for the project.
// Combines project directory with the configured skills directory.
func GetSkillsDir(config *Config) string {
	projectDir := GetProjectDir(config)
	skillsDir := config.Spec.OutputPath
	if skillsDir == "" {
		skillsDir = DefaultSkillsDirName
	}

	// If it's an absolute path, return as-is
	if filepath.IsAbs(skillsDir) {
		return skillsDir
	}

	return filepath.Join(projectDir, skillsDir)
}

// GetSpecDir returns the specification directory path.
func GetSpecDir(config *Config) string {
	projectDir := GetProjectDir(config)
	specDir := config.Spec.SpecDir
	if specDir == "" {
		specDir = DefaultSpecDirName
	}

	if filepath.IsAbs(specDir) {
		return specDir
	}

	return filepath.Join(projectDir, specDir)
}

// GetLoopStateDir returns the Ralph Loop state directory path.
func GetLoopStateDir(config *Config) string {
	projectDir := GetProjectDir(config)
	stateDir := config.Loop.StateDir
	if stateDir == "" {
		stateDir = DefaultStateDirName
	}

	if filepath.IsAbs(stateDir) {
		return stateDir
	}

	return filepath.Join(projectDir, stateDir)
}

// EnsureStateDir ensures the state directory exists.
// Creates the directory if it doesn't exist with default permissions.
func EnsureStateDir(config *Config) error {
	stateDir := GetLoopStateDir(config)

	info, err := os.Stat(stateDir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("state path exists but is not a directory: %s", stateDir)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("check state directory: %w", err)
	}

	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	return nil
}

// EnsureSpecDir ensures the specification directory exists.
func EnsureSpecDir(config *Config) error {
	specDir := GetSpecDir(config)

	info, err := os.Stat(specDir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("spec path exists but is not a directory: %s", specDir)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("check spec directory: %w", err)
	}

	if err := os.MkdirAll(specDir, 0o755); err != nil {
		return fmt.Errorf("create spec directory: %w", err)
	}

	return nil
}

// GetLoopStatePath returns the full path to the loop state file.
func GetLoopStatePath(config *Config) string {
	return filepath.Join(GetLoopStateDir(config), config.Loop.StateFile)
}

// GetLoopLogPath returns the full path to the loop log file.
func GetLoopLogPath(config *Config) string {
	return filepath.Join(GetLoopStateDir(config), config.Loop.LogFile)
}

// GetLoopMetricsPath returns the full path to the metrics file.
func GetLoopMetricsPath(config *Config) string {
	return filepath.Join(GetLoopStateDir(config), config.Loop.MetricsFile)
}

// GetReadyReportPath returns the full path to the production readiness report.
func GetReadyReportPath(config *Config) string {
	return filepath.Join(GetLoopStateDir(config), config.Loop.ReadyReport)
}

// SaveConfig saves the configuration to the specified path.
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration and returns an error if any field is invalid.
// Returns nil if the configuration is valid.
func (c *Config) Validate() error {
	// Validate ProjectName - must not be empty
	if c.ProjectName == "" {
		return gerrors.NewValidationError("ProjectName", "required", errors.New("project name cannot be empty")).
			WithMessage("ProjectName is required")
	}

	// Validate MaxIterations - must be greater than 0
	if c.MaxIterations <= 0 {
		return gerrors.NewValidationError("MaxIterations", "positive", errors.New("must be greater than 0")).
			WithValue(c.MaxIterations)
	}

	// Validate QualityThreshold - must be between 0.0 and 1.0
	if c.QualityThreshold < 0.0 || c.QualityThreshold > 1.0 {
		return gerrors.NewValidationError("QualityThreshold", "range[0.0,1.0]", errors.New("must be between 0.0 and 1.0")).
			WithValue(c.QualityThreshold)
	}

	// Validate Spec config
	if err := c.validateSpecConfig(); err != nil {
		return err
	}

	// Validate Loop config
	if err := c.validateLoopConfig(); err != nil {
		return err
	}

	// Validate Opti config
	if err := c.validateOptiConfig(); err != nil {
		return err
	}

	return nil
}

// validateSpecConfig validates the SpecConfig section.
func (c *Config) validateSpecConfig() error {
	// Validate MaxDepth - must be greater than 0
	if c.Spec.MaxDepth <= 0 {
		return gerrors.NewValidationError("Spec.MaxDepth", "positive", errors.New("must be greater than 0")).
			WithValue(c.Spec.MaxDepth)
	}

	// Validate QualityThreshold - must be between 0.0 and 1.0
	if c.Spec.QualityThreshold < 0.0 || c.Spec.QualityThreshold > 1.0 {
		return gerrors.NewValidationError("Spec.QualityThreshold", "range[0.0,1.0]", errors.New("must be between 0.0 and 1.0")).
			WithValue(c.Spec.QualityThreshold)
	}

	return nil
}

// validateLoopConfig validates the LoopConfig section.
func (c *Config) validateLoopConfig() error {
	// Validate MaxRetries - must be >= 0
	if c.Loop.MaxRetries < 0 {
		return gerrors.NewValidationError("Loop.MaxRetries", "non-negative", errors.New("must be greater than or equal to 0")).
			WithValue(c.Loop.MaxRetries)
	}

	// Validate BackoffBaseSeconds - must be > 0
	if c.Loop.BackoffBaseSeconds <= 0 {
		return gerrors.NewValidationError("Loop.BackoffBaseSeconds", "positive", errors.New("must be greater than 0")).
			WithValue(c.Loop.BackoffBaseSeconds)
	}

	// Validate QualityGateThreshold - must be between 0.0 and 1.0
	if c.Loop.QualityGateThreshold < 0.0 || c.Loop.QualityGateThreshold > 1.0 {
		return gerrors.NewValidationError("Loop.QualityGateThreshold", "range[0.0,1.0]", errors.New("must be between 0.0 and 1.0")).
			WithValue(c.Loop.QualityGateThreshold)
	}

	// Validate DimensionThreshold - must be between 0.0 and 1.0
	if c.Loop.DimensionThreshold < 0.0 || c.Loop.DimensionThreshold > 1.0 {
		return gerrors.NewValidationError("Loop.DimensionThreshold", "range[0.0,1.0]", errors.New("must be between 0.0 and 1.0")).
			WithValue(c.Loop.DimensionThreshold)
	}

	return nil
}

// validateOptiConfig validates the OptiConfig section.
func (c *Config) validateOptiConfig() error {
	// Validate OutputFormat - must be 'markdown' or 'plain'
	validFormats := map[string]bool{"markdown": true, "plain": true}
	if !validFormats[c.Opti.OutputFormat] {
		return gerrors.NewValidationError("Opti.OutputFormat", "enum[markdown,plain]", errors.New("invalid output format")).
			WithValue(c.Opti.OutputFormat).
			WithMessage("must be 'markdown' or 'plain'")
	}

	// Validate MaxContextFiles - must be greater than 0
	if c.Opti.MaxContextFiles <= 0 {
		return gerrors.NewValidationError("Opti.MaxContextFiles", "positive", errors.New("must be greater than 0")).
			WithValue(c.Opti.MaxContextFiles)
	}

	return nil
}

// ValidateAll returns all validation errors as a slice.
// This is useful when you want to collect all errors instead of failing on the first one.
func (c *Config) ValidateAll() []error {
	var errs []error

	// Validate ProjectName - must not be empty
	if c.ProjectName == "" {
		errs = append(errs, gerrors.NewValidationError("ProjectName", "required", errors.New("project name cannot be empty")).
			WithMessage("ProjectName is required"))
	}

	// Validate MaxIterations - must be greater than 0
	if c.MaxIterations <= 0 {
		errs = append(errs, gerrors.NewValidationError("MaxIterations", "positive", errors.New("must be greater than 0")).
			WithValue(c.MaxIterations))
	}

	// Validate QualityThreshold - must be between 0.0 and 1.0
	if c.QualityThreshold < 0.0 || c.QualityThreshold > 1.0 {
		errs = append(errs, gerrors.NewValidationError("QualityThreshold", "range[0.0,1.0]", errors.New("must be between 0.0 and 1.0")).
			WithValue(c.QualityThreshold))
	}

	// Validate Spec config
	if c.Spec.MaxDepth <= 0 {
		errs = append(errs, gerrors.NewValidationError("Spec.MaxDepth", "positive", errors.New("must be greater than 0")).
			WithValue(c.Spec.MaxDepth))
	}

	if c.Spec.QualityThreshold < 0.0 || c.Spec.QualityThreshold > 1.0 {
		errs = append(errs, gerrors.NewValidationError("Spec.QualityThreshold", "range[0.0,1.0]", errors.New("must be between 0.0 and 1.0")).
			WithValue(c.Spec.QualityThreshold))
	}

	// Validate Loop config
	if c.Loop.MaxRetries < 0 {
		errs = append(errs, gerrors.NewValidationError("Loop.MaxRetries", "non-negative", errors.New("must be greater than or equal to 0")).
			WithValue(c.Loop.MaxRetries))
	}

	if c.Loop.BackoffBaseSeconds <= 0 {
		errs = append(errs, gerrors.NewValidationError("Loop.BackoffBaseSeconds", "positive", errors.New("must be greater than 0")).
			WithValue(c.Loop.BackoffBaseSeconds))
	}

	if c.Loop.QualityGateThreshold < 0.0 || c.Loop.QualityGateThreshold > 1.0 {
		errs = append(errs, gerrors.NewValidationError("Loop.QualityGateThreshold", "range[0.0,1.0]", errors.New("must be between 0.0 and 1.0")).
			WithValue(c.Loop.QualityGateThreshold))
	}

	if c.Loop.DimensionThreshold < 0.0 || c.Loop.DimensionThreshold > 1.0 {
		errs = append(errs, gerrors.NewValidationError("Loop.DimensionThreshold", "range[0.0,1.0]", errors.New("must be between 0.0 and 1.0")).
			WithValue(c.Loop.DimensionThreshold))
	}

	// Validate Opti config
	validFormats := map[string]bool{"markdown": true, "plain": true}
	if !validFormats[c.Opti.OutputFormat] {
		errs = append(errs, gerrors.NewValidationError("Opti.OutputFormat", "enum[markdown,plain]", errors.New("invalid output format")).
			WithValue(c.Opti.OutputFormat).
			WithMessage("must be 'markdown' or 'plain'"))
	}

	if c.Opti.MaxContextFiles <= 0 {
		errs = append(errs, gerrors.NewValidationError("Opti.MaxContextFiles", "positive", errors.New("must be greater than 0")).
			WithValue(c.Opti.MaxContextFiles))
	}

	return errs
}

// applyDefaults sets default values for unset fields.
func (c *Config) applyDefaults() {
	if c.ProjectName == "" {
		c.ProjectName = DefaultProjectName
	}

	if c.MaxIterations == 0 {
		c.MaxIterations = DefaultMaxIterations
	}

	if c.QualityThreshold == 0 {
		c.QualityThreshold = DefaultQualityThreshold
	}

	if c.ScoringWeights == nil {
		c.ScoringWeights = defaultScoringWeights()
	}

	// Apply defaults to Spec config
	if c.Spec.OutputPath == "" {
		c.Spec.OutputPath = "spec"
	}

	if c.Spec.SpecDir == "" {
		c.Spec.SpecDir = DefaultSpecDirName
	}

	if c.Spec.MaxDepth == 0 {
		c.Spec.MaxDepth = 5
	}

	if c.Spec.MinComponents == 0 {
		c.Spec.MinComponents = 3
	}

	if c.Spec.MinScenarios == 0 {
		c.Spec.MinScenarios = 5
	}

	if c.Spec.QualityThreshold == 0 {
		c.Spec.QualityThreshold = DefaultQualityThreshold
	}

	// Apply defaults to Loop config
	if c.Loop.StateDir == "" {
		c.Loop.StateDir = DefaultStateDirName
	}

	if c.Loop.StateFile == "" {
		c.Loop.StateFile = "GROVE-LOOP-STATE.json"
	}

	if c.Loop.LogFile == "" {
		c.Loop.LogFile = "GROVE-LOOP-LOG.md"
	}

	if c.Loop.MetricsFile == "" {
		c.Loop.MetricsFile = "GROVE-LOOP-METRICS.json"
	}

	if c.Loop.ReadyReport == "" {
		c.Loop.ReadyReport = "GROVE-READY-REPORT.md"
	}

	if c.Loop.MaxRetries == 0 {
		c.Loop.MaxRetries = DefaultMaxRetries
	}

	if c.Loop.BackoffBaseSeconds == 0 {
		c.Loop.BackoffBaseSeconds = DefaultBackoffBaseSeconds
	}

	if c.Loop.ContextWindowLimit == 0 {
		c.Loop.ContextWindowLimit = 100000
	}

	if c.Loop.QualityGateThreshold == 0 {
		c.Loop.QualityGateThreshold = 0.70
	}

	if c.Loop.DimensionThreshold == 0 {
		c.Loop.DimensionThreshold = 0.80
	}

	// Apply defaults to Opti config
	if c.Opti.OutputFormat == "" {
		c.Opti.OutputFormat = "markdown"
	}

	if c.Opti.MaxContextFiles == 0 {
		c.Opti.MaxContextFiles = 10
	}

	if c.Opti.BoundContextSize == 0 {
		c.Opti.BoundContextSize = 50000
	}

	if c.Opti.ContextFilePatterns == nil {
		c.Opti.ContextFilePatterns = []string{"*.go", "*.ts", "*.tsx", "*.md"}
	}
}

// defaultScoringWeights returns the default scoring weights for quality assessment.
func defaultScoringWeights() map[string]float64 {
	return map[string]float64{
		"flow_coverage":          0.20,
		"edge_cases":             0.15,
		"component_depth":        0.15,
		"logical_consistency":    0.20,
		"connectivity":           0.10,
		"decision_justification": 0.10,
		"agent_consumability":    0.10,
	}
}

// =============================================================================
// Config Watcher - Hot Reload Support
// =============================================================================

// DefaultConfigWatchInterval is the default polling interval for config watching.
const DefaultConfigWatchInterval = 5 * time.Second

// ConfigWatcher monitors configuration file for changes and triggers callbacks.
type ConfigWatcher struct {
	path     string
	config   *Config
	onChange func(*Config)
	stop     chan struct{}
	done     chan struct{}
	interval time.Duration
	mu       sync.RWMutex
}

// NewConfigWatcher creates a new configuration watcher for the specified path.
func NewConfigWatcher(path string) *ConfigWatcher {
	return &ConfigWatcher{
		path:     path,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		interval: DefaultConfigWatchInterval,
	}
}

// NewConfigWatcherWithInterval creates a new configuration watcher with custom interval.
func NewConfigWatcherWithInterval(path string, interval time.Duration) *ConfigWatcher {
	return &ConfigWatcher{
		path:     path,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		interval: interval,
	}
}

// Watch starts monitoring the configuration file for changes.
// The onChange callback is invoked whenever the config file is modified.
// Returns an error if the watcher cannot be started.
func (w *ConfigWatcher) Watch(onChange func(*Config)) error {
	if onChange == nil {
		return fmt.Errorf("onChange callback cannot be nil")
	}

	w.mu.Lock()
	w.onChange = onChange
	w.mu.Unlock()

	go w.watchLoop()

	return nil
}

// watchLoop is the internal polling loop that monitors file changes.
func (w *ConfigWatcher) watchLoop() {
	defer close(w.done)

	// Get initial modification time
	lastMod, err := w.getLastModTime()
	if err != nil {
		// File might not exist yet, that's okay
		lastMod = time.Time{}
	}

	for {
		select {
		case <-w.stop:
			return
		case <-time.After(w.interval):
			currentMod, err := w.getLastModTime()
			if err != nil {
				// File might have been deleted or not created yet
				continue
			}

			// Check if file was modified
			if !lastMod.IsZero() && !currentMod.IsZero() && currentMod.After(lastMod) {
				w.reloadAndNotify()
				lastMod = currentMod
			} else if lastMod.IsZero() && !currentMod.IsZero() {
				// File was just created
				lastMod = currentMod
			}
		}
	}
}

// getLastModTime returns the last modification time of the config file.
func (w *ConfigWatcher) getLastModTime() (time.Time, error) {
	info, err := os.Stat(w.path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// reloadAndNotify reloads the configuration and notifies the callback.
func (w *ConfigWatcher) reloadAndNotify() {
	newConfig, err := LoadConfig(w.path)
	if err != nil {
		// Log error but don't crash the watcher
		fmt.Printf("config: failed to reload: %v\n", err)
		return
	}

	w.mu.Lock()
	callback := w.onChange
	w.mu.Unlock()

	if callback != nil {
		callback(newConfig)
	}
}

// Reload manually reloads the configuration from disk.
// Returns the new configuration or an error if reload fails.
func (w *ConfigWatcher) Reload() (*Config, error) {
	return LoadConfig(w.path)
}

// Stop stops the configuration watcher.
// After calling Stop, the watcher will no longer monitor for changes.
func (w *ConfigWatcher) Stop() error {
	select {
	case <-w.stop:
		// Already stopped
		return nil
	default:
		close(w.stop)
	}

	// Wait for watchLoop to finish
	<-w.done
	return nil
}

// Interval returns the current polling interval.
func (w *ConfigWatcher) Interval() time.Duration {
	return w.interval
}

// SetInterval changes the polling interval (only effective before Watch is called).
func (w *ConfigWatcher) SetInterval(interval time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.interval = interval
}
