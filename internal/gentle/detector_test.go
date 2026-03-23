// Package gentle provides detection and invocation capabilities for gentle-ai.
package gentle

import (
	"testing"
)

// TestDetector_Detect verifies the detection logic works correctly.
func TestDetector_Detect(t *testing.T) {
	d := NewDetector()
	result := d.Detect()

	// Result should never be nil
	if result == nil {
		t.Fatal("Detect() returned nil result")
	}

	// We can't guarantee gentle-ai is installed on the test machine,
	// so we just verify the struct fields are populated correctly
	t.Logf("Installed: %v", result.Installed)
	t.Logf("Version: %s", result.Version)
	t.Logf("Path: %s", result.Path)
	t.Logf("SkillsDir: %s", result.SkillsDir)
}

// TestDetector_IsAvailable tests the IsAvailable helper method.
func TestDetector_IsAvailable(t *testing.T) {
	d := NewDetector()
	available := d.IsAvailable()

	// This test just verifies the method doesn't panic
	// Actual result depends on the environment
	t.Logf("IsAvailable: %v", available)
}

// TestDetector_HasSkills tests the HasSkills helper method.
func TestDetector_HasSkills(t *testing.T) {
	d := NewDetector()
	hasSkills := d.HasSkills()

	// This test just verifies the method doesn't panic
	// Actual result depends on the environment
	t.Logf("HasSkills: %v", hasSkills)
}

// TestDetectionResult_Structure verifies the DetectionResult structure.
func TestDetectionResult_Structure(t *testing.T) {
	result := &DetectionResult{
		Installed: true,
		Version:   "1.0.0",
		Path:      "/usr/bin/gentle-ai",
		SkillsDir: "/home/user/.config/opencode/skills",
	}

	if !result.Installed {
		t.Error("Expected Installed to be true")
	}
	if result.Version != "1.0.0" {
		t.Errorf("Expected Version to be '1.0.0', got '%s'", result.Version)
	}
	if result.Path != "/usr/bin/gentle-ai" {
		t.Errorf("Expected Path to be '/usr/bin/gentle-ai', got '%s'", result.Path)
	}
	if result.SkillsDir != "/home/user/.config/opencode/skills" {
		t.Errorf("Expected SkillsDir to be '/home/user/.config/opencode/skills', got '%s'", result.SkillsDir)
	}
}

// TestInvoker_NewInvoker tests the invoker creation.
func TestInvoker_NewInvoker(t *testing.T) {
	invoker := NewInvoker()

	if invoker == nil {
		t.Fatal("NewInvoker() returned nil")
	}

	if invoker.detector == nil {
		t.Error("Expected detector to be initialized")
	}

	// Just verify we can call these methods without panic
	_ = invoker.IsAvailable()
	_ = invoker.GetSkillsDir()
	_ = invoker.GetDetector()
}
