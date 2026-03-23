package sdd

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("/test/project")
	if client == nil {
		t.Fatal("Client should not be nil")
	}
	if client.projectDir != "/test/project" {
		t.Errorf("Expected projectDir '/test/project', got '%s'", client.projectDir)
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	client := NewClientWithTimeout("/test/project", 10*time.Minute)
	if client.timeout != 10*time.Minute {
		t.Errorf("Expected timeout 10m, got %v", client.timeout)
	}
}

func TestPhaseConstants(t *testing.T) {
	phases := []Phase{
		PhaseExplore,
		PhasePropose,
		PhaseSpec,
		PhaseDesign,
		PhaseTasks,
		PhaseApply,
		PhaseVerify,
		PhaseArchive,
	}

	expected := []string{
		"explore",
		"propose",
		"spec",
		"design",
		"tasks",
		"apply",
		"verify",
		"archive",
	}

	for i, phase := range phases {
		if string(phase) != expected[i] {
			t.Errorf("Phase %d: expected '%s', got '%s'", i, expected[i], string(phase))
		}
	}
}

func TestExecute_UnknownPhase(t *testing.T) {
	client := NewClient("/test/project")
	ctx := context.Background()

	_, err := client.Execute(ctx, Phase("unknown"), nil)
	if err == nil {
		t.Error("Should return error for unknown phase")
	}
}

func TestResult_toJSON(t *testing.T) {
	result := &Result{
		Phase:    PhaseVerify,
		Status:   "success",
		Summary:  "Task verified",
		Duration: 5 * time.Second,
	}

	json := result.toJSON()
	if json == "" {
		t.Error("JSON should not be empty")
	}
}

func TestGetSkillsDir(t *testing.T) {
	dir := getSkillsDir()
	if dir == "" {
		t.Error("Skills dir should not be empty")
	}
}
