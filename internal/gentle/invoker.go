// Package gentle provides detection and invocation capabilities for gentle-ai.
// It enables GROVE to detect gentle-ai installation and gracefully degrade when unavailable.
package gentle

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// Invoker invokes gentle-ai skills.
type Invoker struct {
	detector  *Detector
	skillsDir string
}

// NewInvoker creates a new Invoker instance.
// It detects gentle-ai availability at initialization time.
func NewInvoker() *Invoker {
	d := NewDetector()
	result := d.Detect()
	return &Invoker{
		detector:  d,
		skillsDir: result.SkillsDir,
	}
}

// InvokeSkill invokes a gentle-ai skill by name with the provided arguments.
// Returns an error if gentle-ai is not available.
func (i *Invoker) InvokeSkill(ctx context.Context, skillName string, args map[string]interface{}) error {
	if !i.detector.IsAvailable() {
		return fmt.Errorf("gentle-ai not installed: to enable full functionality, install gentle-ai from https://github.com/Gentleman-Programming/gentle-ai")
	}

	// Invoke via CLI or API based on availability
	// TODO: Implement actual skill invocation when gentle-ai CLI is available
	// For now, this serves as a placeholder for future integration
	return nil
}

// ListSkills returns the list of available skills from the skills directory.
// Returns an error if no skills directory is found.
func (i *Invoker) ListSkills() ([]string, error) {
	if i.skillsDir == "" {
		return nil, fmt.Errorf("no skills directory found")
	}

	entries, err := os.ReadDir(i.skillsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills directory: %w", err)
	}

	var skills []string
	for _, entry := range entries {
		if entry.IsDir() && !isHidden(entry.Name()) {
			skills = append(skills, entry.Name())
		}
	}

	return skills, nil
}

// isHidden checks if a directory name starts with a dot (hidden file/dir).
func isHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

// GetDetector returns the underlying detector.
func (i *Invoker) GetDetector() *Detector {
	return i.detector
}

// GetSkillsDir returns the configured skills directory.
func (i *Invoker) GetSkillsDir() string {
	return i.skillsDir
}

// IsAvailable checks if gentle-ai is available for invocation.
func (i *Invoker) IsAvailable() bool {
	return i.detector.IsAvailable()
}
