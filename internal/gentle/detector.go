// Package gentle provides detection and invocation capabilities for gentle-ai.
// It enables GROVE to detect gentle-ai installation and gracefully degrade when unavailable.
package gentle

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Detector detects if gentle-ai is installed and available.
type Detector struct{}

// DetectionResult contains the results of the gentle-ai detection.
type DetectionResult struct {
	Installed bool   // Whether gentle-ai binary is in PATH
	Version   string // Version string if available
	Path      string // Full path to gentle-ai binary
	SkillsDir string // Path to skills directory (~/.config/opencode/skills or similar)
}

// NewDetector creates a new Detector instance.
func NewDetector() *Detector {
	return &Detector{}
}

// Detect checks if gentle-ai is installed by searching PATH and common directories.
func (d *Detector) Detect() *DetectionResult {
	result := &DetectionResult{}

	// 1. Search in PATH
	if path, err := exec.LookPath("gentle-ai"); err == nil {
		result.Installed = true
		result.Path = path
		result.Version = d.getVersion(path)
	}

	// 2. Search in common directories for skills
	if !result.Installed || result.SkillsDir == "" {
		result.SkillsDir = d.findSkillsDir()
	}

	return result
}

// getVersion attempts to get the version of gentle-ai.
func (d *Detector) getVersion(path string) string {
	cmd := exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// findSkillsDir searches for the skills directory in common locations.
func (d *Detector) findSkillsDir() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		// Windows fallback
		homeDir = os.Getenv("USERPROFILE")
	}

	dirs := []string{
		filepath.Join(homeDir, ".config", "opencode", "skills"),
		filepath.Join(homeDir, ".gentle-ai", "skills"),
		filepath.Join(homeDir, ".opencode", "skills"),
	}

	for _, dir := range dirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return ""
}

// IsAvailable returns true if gentle-ai binary is installed and in PATH.
func (d *Detector) IsAvailable() bool {
	return d.Detect().Installed
}

// HasSkills returns true if a skills directory was found.
func (d *Detector) HasSkills() bool {
	return d.Detect().SkillsDir != ""
}
