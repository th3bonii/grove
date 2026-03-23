package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test main config defaults
	if cfg.ProjectName != DefaultProjectName {
		t.Errorf("expected ProjectName %q, got %q", DefaultProjectName, cfg.ProjectName)
	}

	if cfg.MaxIterations != DefaultMaxIterations {
		t.Errorf("expected MaxIterations %d, got %d", DefaultMaxIterations, cfg.MaxIterations)
	}

	if cfg.QualityThreshold != DefaultQualityThreshold {
		t.Errorf("expected QualityThreshold %f, got %f", DefaultQualityThreshold, cfg.QualityThreshold)
	}

	// Test Spec config defaults
	if cfg.Spec.OutputPath != "spec" {
		t.Errorf("expected Spec.OutputPath 'spec', got %q", cfg.Spec.OutputPath)
	}

	if cfg.Spec.SpecDir != DefaultSpecDirName {
		t.Errorf("expected Spec.SpecDir %q, got %q", DefaultSpecDirName, cfg.Spec.SpecDir)
	}

	if !cfg.Spec.EnableSelfQuestioning {
		t.Error("expected Spec.EnableSelfQuestioning to be true")
	}

	if cfg.Spec.MaxDepth != 5 {
		t.Errorf("expected Spec.MaxDepth 5, got %d", cfg.Spec.MaxDepth)
	}

	// Test Loop config defaults
	if cfg.Loop.StateDir != DefaultStateDirName {
		t.Errorf("expected Loop.StateDir %q, got %q", DefaultStateDirName, cfg.Loop.StateDir)
	}

	if cfg.Loop.StateFile != "GROVE-LOOP-STATE.json" {
		t.Errorf("expected Loop.StateFile 'GROVE-LOOP-STATE.json', got %q", cfg.Loop.StateFile)
	}

	if cfg.Loop.MaxRetries != DefaultMaxRetries {
		t.Errorf("expected Loop.MaxRetries %d, got %d", DefaultMaxRetries, cfg.Loop.MaxRetries)
	}

	if cfg.Loop.QualityGateThreshold != 0.70 {
		t.Errorf("expected Loop.QualityGateThreshold 0.70, got %f", cfg.Loop.QualityGateThreshold)
	}

	// Test Opti config defaults
	if cfg.Opti.OutputFormat != "markdown" {
		t.Errorf("expected Opti.OutputFormat 'markdown', got %q", cfg.Opti.OutputFormat)
	}

	if !cfg.Opti.EnableExplanations {
		t.Error("expected Opti.EnableExplanations to be true")
	}

	if cfg.Opti.MaxContextFiles != 10 {
		t.Errorf("expected Opti.MaxContextFiles 10, got %d", cfg.Opti.MaxContextFiles)
	}

	// Test scoring weights
	if cfg.ScoringWeights == nil {
		t.Fatal("expected ScoringWeights to be non-nil")
	}

	expectedWeights := map[string]float64{
		"flow_coverage":          0.20,
		"edge_cases":             0.15,
		"component_depth":        0.15,
		"logical_consistency":    0.20,
		"connectivity":           0.10,
		"decision_justification": 0.10,
		"agent_consumability":    0.10,
	}

	for name, expected := range expectedWeights {
		if got, ok := cfg.ScoringWeights[name]; !ok {
			t.Errorf("expected scoring weight %q to exist", name)
		} else if got != expected {
			t.Errorf("expected scoring weight %q to be %f, got %f", name, expected, got)
		}
	}
}

func TestGetProjectDir(t *testing.T) {
	cfg := &Config{
		ProjectName: "test-project",
		ProjectPath: "/custom/path",
	}

	dir := GetProjectDir(cfg)
	if dir != "/custom/path" {
		t.Errorf("expected project dir '/custom/path', got %q", dir)
	}
}

func TestGetProjectDirFallback(t *testing.T) {
	cfg := &Config{
		ProjectName: "test-project",
		ProjectPath: "", // Empty, should use cwd
	}

	dir := GetProjectDir(cfg)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	if dir != cwd {
		t.Errorf("expected project dir to be cwd %q, got %q", cwd, dir)
	}
}

func TestGetSkillsDir(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/project",
		Spec: SpecConfig{
			OutputPath: "skills",
		},
	}

	dir := GetSkillsDir(cfg)
	expected := filepath.Join("/project", "skills")
	if dir != expected {
		t.Errorf("expected skills dir %q, got %q", expected, dir)
	}
}

func TestGetSkillsDirAbsolute(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/project",
		Spec: SpecConfig{
			OutputPath: "/absolute/skills",
		},
	}

	dir := GetSkillsDir(cfg)
	if dir != "/absolute/skills" {
		t.Errorf("expected absolute skills dir '/absolute/skills', got %q", dir)
	}
}

func TestGetSpecDir(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/project",
		Spec: SpecConfig{
			SpecDir: "docs/spec",
		},
	}

	dir := GetSpecDir(cfg)
	expected := filepath.Join("/project", "docs/spec")
	if dir != expected {
		t.Errorf("expected spec dir %q, got %q", expected, dir)
	}
}

func TestGetLoopStateDir(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/project",
		Loop: LoopConfig{
			StateDir: ".grove-state",
		},
	}

	dir := GetLoopStateDir(cfg)
	expected := filepath.Join("/project", ".grove-state")
	if dir != expected {
		t.Errorf("expected state dir %q, got %q", expected, dir)
	}
}

func TestEnsureStateDir(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state")

	cfg := &Config{
		ProjectPath: statePath,
		Loop: LoopConfig{
			StateDir: "nested/state",
		},
	}

	err := EnsureStateDir(cfg)
	if err != nil {
		t.Fatalf("EnsureStateDir failed: %v", err)
	}

	// Verify directory was created
	expectedPath := filepath.Join(statePath, "nested", "state")
	info, err := os.Stat(expectedPath)
	if err != nil {
		t.Fatalf("state directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("created path is not a directory")
	}

	// On Windows, permissions may differ - just verify directory exists
	t.Logf("Created directory with permissions: %o", info.Mode().Perm())
}

func TestEnsureStateDirAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state", "dir")

	// Create the directory first
	if err := os.MkdirAll(statePath, 0o755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	cfg := &Config{
		ProjectPath: tmpDir,
		Loop: LoopConfig{
			StateDir: "state/dir",
		},
	}

	err := EnsureStateDir(cfg)
	if err != nil {
		t.Errorf("EnsureStateDir should not fail for existing directory: %v", err)
	}
}

func TestEnsureStateDirNotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file")

	// Create a file instead of directory
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := &Config{
		ProjectPath: tmpDir,
		Loop: LoopConfig{
			StateDir: "file",
		},
	}

	err := EnsureStateDir(cfg)
	if err == nil {
		t.Error("EnsureStateDir should fail when path exists but is not a directory")
	}
}

func TestEnsureSpecDir(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		ProjectPath: tmpDir,
		Spec: SpecConfig{
			SpecDir: "spec-docs",
		},
	}

	err := EnsureSpecDir(cfg)
	if err != nil {
		t.Fatalf("EnsureSpecDir failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "spec-docs")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("spec directory was not created")
	}
}

func TestGetLoopStatePath(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/project",
		Loop: LoopConfig{
			StateDir:  ".grove-state",
			StateFile: "state.json",
		},
	}

	path := GetLoopStatePath(cfg)
	expected := filepath.Join("/project", ".grove-state", "state.json")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestGetLoopLogPath(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/project",
		Loop: LoopConfig{
			StateDir: ".grove-state",
			LogFile:  "log.md",
		},
	}

	path := GetLoopLogPath(cfg)
	expected := filepath.Join("/project", ".grove-state", "log.md")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestGetLoopMetricsPath(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/project",
		Loop: LoopConfig{
			StateDir:    ".grove-state",
			MetricsFile: "metrics.json",
		},
	}

	path := GetLoopMetricsPath(cfg)
	expected := filepath.Join("/project", ".grove-state", "metrics.json")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestGetReadyReportPath(t *testing.T) {
	cfg := &Config{
		ProjectPath: "/project",
		Loop: LoopConfig{
			StateDir:    ".grove-state",
			ReadyReport: "ready.md",
		},
	}

	path := GetReadyReportPath(cfg)
	expected := filepath.Join("/project", ".grove-state", "ready.md")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig should not fail for non-existent file, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig should return default config for non-existent file")
	}

	if cfg.ProjectName != DefaultProjectName {
		t.Errorf("expected default ProjectName %q, got %q", DefaultProjectName, cfg.ProjectName)
	}
}

func TestLoadConfigValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
project_name: "test-config"
project_path: "/test/path"
max_iterations: 5
quality_threshold: 0.80
verbose: true
spec:
  output_path: "custom-spec"
  enable_self_questioning: false
  max_depth: 10
loop:
  state_dir: "custom-state"
  max_retries: 5
opti:
  output_format: "plain"
  enable_explanations: false
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.ProjectName != "test-config" {
		t.Errorf("expected ProjectName 'test-config', got %q", cfg.ProjectName)
	}

	if cfg.ProjectPath != "/test/path" {
		t.Errorf("expected ProjectPath '/test/path', got %q", cfg.ProjectPath)
	}

	if cfg.MaxIterations != 5 {
		t.Errorf("expected MaxIterations 5, got %d", cfg.MaxIterations)
	}

	if cfg.QualityThreshold != 0.80 {
		t.Errorf("expected QualityThreshold 0.80, got %f", cfg.QualityThreshold)
	}

	if !cfg.Verbose {
		t.Error("expected Verbose to be true")
	}

	if cfg.Spec.OutputPath != "custom-spec" {
		t.Errorf("expected Spec.OutputPath 'custom-spec', got %q", cfg.Spec.OutputPath)
	}

	if cfg.Spec.EnableSelfQuestioning {
		t.Error("expected Spec.EnableSelfQuestioning to be false")
	}

	if cfg.Loop.MaxRetries != 5 {
		t.Errorf("expected Loop.MaxRetries 5, got %d", cfg.Loop.MaxRetries)
	}

	if cfg.Opti.OutputFormat != "plain" {
		t.Errorf("expected Opti.OutputFormat 'plain', got %q", cfg.Opti.OutputFormat)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		ProjectName: "save-test",
		ProjectPath: "/save/path",
		Spec: SpecConfig{
			OutputPath: "saved-spec",
		},
	}

	err := SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Verify content is valid YAML
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}

	if len(data) == 0 {
		t.Error("saved config is empty")
	}

	// Load and verify
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loaded.ProjectName != "save-test" {
		t.Errorf("expected ProjectName 'save-test', got %q", loaded.ProjectName)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				ProjectName:      "valid",
				MaxIterations:    5,
				QualityThreshold: 0.8,
				Spec: SpecConfig{
					MaxDepth:         5,
					QualityThreshold: 0.7,
				},
				Loop: LoopConfig{
					MaxRetries:           3,
					QualityGateThreshold: 0.6,
					BackoffBaseSeconds:   2, // Required for validation
					DimensionThreshold:   0.7,
				},
				Opti: OptiConfig{
					OutputFormat:    "markdown",
					MaxContextFiles: 10,
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			config: &Config{
				ProjectName:      "",
				MaxIterations:    5,
				QualityThreshold: 0.8,
			},
			wantErr: true,
		},
		{
			name: "invalid max iterations",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    0,
				QualityThreshold: 0.8,
			},
			wantErr: true,
		},
		{
			name: "invalid quality threshold",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: 1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid opti format",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: 0.8,
				Opti: OptiConfig{
					OutputFormat: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "multiple errors",
			config: &Config{
				ProjectName:      "",
				MaxIterations:    0,
				QualityThreshold: 1.5,
				Spec: SpecConfig{
					MaxDepth: 0,
				},
				Opti: OptiConfig{
					OutputFormat:    "invalid",
					MaxContextFiles: 0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() expected no error, got %v", err)
			}
		})
	}
}

func TestConfigApplyDefaults(t *testing.T) {
	cfg := &Config{}

	// Apply defaults by unmarshaling empty YAML
	data := []byte("{}")
	if err := yaml.Unmarshal(data, cfg); err != nil {
		t.Fatalf("failed to unmarshal empty config: %v", err)
	}

	cfg.applyDefaults()

	// Verify defaults were applied
	if cfg.ProjectName != DefaultProjectName {
		t.Errorf("expected ProjectName %q, got %q", DefaultProjectName, cfg.ProjectName)
	}

	if cfg.MaxIterations != DefaultMaxIterations {
		t.Errorf("expected MaxIterations %d, got %d", DefaultMaxIterations, cfg.MaxIterations)
	}

	if cfg.Spec.OutputPath != "spec" {
		t.Errorf("expected Spec.OutputPath 'spec', got %q", cfg.Spec.OutputPath)
	}

	if cfg.Loop.StateDir != DefaultStateDirName {
		t.Errorf("expected Loop.StateDir %q, got %q", DefaultStateDirName, cfg.Loop.StateDir)
	}

	if cfg.Opti.OutputFormat != "markdown" {
		t.Errorf("expected Opti.OutputFormat 'markdown', got %q", cfg.Opti.OutputFormat)
	}
}

func TestDefaultScoringWeights(t *testing.T) {
	weights := defaultScoringWeights()

	if weights == nil {
		t.Fatal("defaultScoringWeights returned nil")
	}

	// Verify all weights sum to 1.0
	var sum float64
	for _, w := range weights {
		sum += w
	}

	// Use small epsilon for float comparison
	const epsilon = 0.001
	if sum < 1.0-epsilon || sum > 1.0+epsilon {
		t.Errorf("scoring weights sum to %f, expected 1.0", sum)
	}
}

// =============================================================================
// Edge Cases Tests
// =============================================================================

func TestValidate_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "negative max iterations",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    -1,
				QualityThreshold: 0.8,
			},
			wantErr: true,
		},
		{
			name: "negative max depth",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: 0.8,
				Spec: SpecConfig{
					MaxDepth: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "negative max retries",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: 0.8,
				Loop: LoopConfig{
					MaxRetries: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "negative quality gate threshold",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: 0.8,
				Loop: LoopConfig{
					QualityGateThreshold: -0.5,
				},
			},
			wantErr: true,
		},
		{
			name: "negative spec quality threshold",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: 0.8,
				Spec: SpecConfig{
					QualityThreshold: -0.1,
				},
			},
			wantErr: true,
		},
		{
			name: "zero max context files",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: 0.8,
				Opti: OptiConfig{
					MaxContextFiles: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "empty project name",
			config: &Config{
				ProjectName:      "",
				MaxIterations:    5,
				QualityThreshold: 0.8,
			},
			wantErr: true,
		},
		{
			name: "whitespace project name",
			config: &Config{
				ProjectName:      "   ",
				MaxIterations:    5,
				QualityThreshold: 0.8,
			},
			wantErr: true,
		},
		{
			name: "quality threshold below zero",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: -0.1,
			},
			wantErr: true,
		},
		{
			name: "quality threshold above 1",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: 1.5,
			},
			wantErr: true,
		},
		{
			name: "empty opti output format",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: 0.8,
				Opti: OptiConfig{
					OutputFormat: "",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid opti output format",
			config: &Config{
				ProjectName:      "test",
				MaxIterations:    5,
				QualityThreshold: 0.8,
				Opti: OptiConfig{
					OutputFormat: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() expected no error, got %v", err)
			}
		})
	}
}

func TestValidateAll_MultipleErrors(t *testing.T) {
	cfg := &Config{
		ProjectName:      "",
		MaxIterations:    -5,
		QualityThreshold: 2.0,
		Spec: SpecConfig{
			MaxDepth: -10,
		},
		Loop: LoopConfig{
			MaxRetries:           -3,
			QualityGateThreshold: -0.1,
		},
		Opti: OptiConfig{
			OutputFormat:    "invalid",
			MaxContextFiles: -1,
		},
	}

	errs := cfg.ValidateAll()

	// Should have multiple errors
	if len(errs) == 0 {
		t.Error("ValidateAll() should return errors for invalid config")
	}

	t.Logf("ValidateAll() returned %d errors:", len(errs))
	for _, err := range errs {
		t.Logf("  - %v", err)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write invalid YAML
	invalidYAML := `
project_name: "test"
max_iterations: 5
quality_threshold: 0.8
spec:
  output_path: "spec"
invalid_yaml_field: [1, 2, 3
  missing_bracket: true
`

	if err := os.WriteFile(configPath, []byte(invalidYAML), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Should fail to parse
	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig should fail for invalid YAML")
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write empty file
	if err := os.WriteFile(configPath, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed for empty file: %v", err)
	}

	// Should return default config
	if cfg.ProjectName != DefaultProjectName {
		t.Errorf("expected default ProjectName %q, got %q", DefaultProjectName, cfg.ProjectName)
	}
}

func TestLoadConfig_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write truly invalid YAML (not just unusual)
	malformedYAML := `project_name: !!python/object/apply:os.system ["rm -rf /"]`

	if err := os.WriteFile(configPath, []byte(malformedYAML), 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Should handle gracefully (yaml.v3 is safe)
	_, err := LoadConfig(configPath)
	if err != nil {
		t.Logf("LoadConfig returned error for malformed YAML: %v", err)
	}
}

func TestGetSkillsDir_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		wantEmpty bool
	}{
		{
			name: "empty project path",
			cfg: &Config{
				ProjectPath: "",
				Spec: SpecConfig{
					OutputPath: "spec",
				},
			},
			wantEmpty: true, // Empty path returns empty
		},
		{
			name: "dot path",
			cfg: &Config{
				ProjectPath: ".",
				Spec: SpecConfig{
					OutputPath: "spec",
				},
			},
			wantEmpty: false,
		},
		{
			name: "double dot path",
			cfg: &Config{
				ProjectPath: "..",
				Spec: SpecConfig{
					OutputPath: "spec",
				},
			},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := GetSkillsDir(tt.cfg)
			if tt.wantEmpty && dir != "" {
				t.Errorf("Expected empty for %s, got %q", tt.name, dir)
			}
			if !tt.wantEmpty && dir == "" {
				t.Errorf("Expected non-empty for %s", tt.name)
			}
		})
	}
}
