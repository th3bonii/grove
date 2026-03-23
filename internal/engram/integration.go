// Package engram provides integration with the Engram persistent memory system.
//
// This file contains GROVE-specific integration functions that use the EngramClient
// to persist and retrieve GROVE-specific data like spec decisions, loop checkpoints,
// and optimization patterns.
package engram

import (
	"encoding/json"
	"fmt"
	"time"

	gerrors "github.com/Gentleman-Programming/grove/internal/errors"
)

// =============================================================================
// Spec Decision Integration
// =============================================================================

// SpecDecision represents a specification decision made during GROVE Spec execution.
type SpecDecision struct {
	ID            string                 `json:"id"`
	ChangeName    string                 `json:"change_name"`
	Decision      string                 `json:"decision"`
	Justification string                 `json:"justification"`
	Alternatives  []string               `json:"alternatives,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// SaveSpecDecision saves a specification decision to Engram.
// It stores the decision indexed by change name for easy retrieval.
func (c *EngramClient) SaveSpecDecision(changeName string, decision *SpecDecision) error {
	if changeName == "" {
		return gerrors.NewValidationError("changeName", "required", fmt.Errorf("change name cannot be empty"))
	}
	if decision == nil {
		return gerrors.NewValidationError("decision", "required", fmt.Errorf("decision cannot be nil"))
	}

	// Set timestamp if not set
	if decision.Timestamp.IsZero() {
		decision.Timestamp = time.Now()
	}

	// Store with key pattern: spec-decision/{changeName}/{decisionID}
	key := fmt.Sprintf("spec-decision/%s/%s", changeName, decision.ID)
	return c.Save(key, decision)
}

// LoadSpecDecisions loads all spec decisions for a given change name.
func (c *EngramClient) LoadSpecDecisions(changeName string) ([]SpecDecision, error) {
	if changeName == "" {
		return nil, gerrors.NewValidationError("changeName", "required", fmt.Errorf("change name cannot be empty"))
	}

	// Search for decisions matching the change name
	query := fmt.Sprintf("spec-decision/%s", changeName)
	keys, err := c.Search(query)
	if err != nil {
		// If Engram unavailable, return empty slice (graceful degradation)
		if IsEngramUnavailable(err) {
			return []SpecDecision{}, nil
		}
		return nil, fmt.Errorf("search spec decisions: %w", err)
	}

	var decisions []SpecDecision
	for _, key := range keys {
		value, err := c.Load(key)
		if err != nil {
			// Skip keys that can't be loaded (might be stale)
			continue
		}

		// Try to parse as SpecDecision
		data, err := json.Marshal(value)
		if err != nil {
			continue
		}

		var decision SpecDecision
		if err := json.Unmarshal(data, &decision); err != nil {
			continue
		}

		decisions = append(decisions, decision)
	}

	return decisions, nil
}

// =============================================================================
// Loop Checkpoint Integration
// =============================================================================

// LoopCheckpoint represents a Ralph Loop checkpoint for resuming execution.
type LoopCheckpoint struct {
	ChangeName string                 `json:"change_name"`
	LoopNumber int                    `json:"loop_number"`
	Phase      string                 `json:"phase"`
	State      map[string]interface{} `json:"state"`
	Artifacts  []string               `json:"artifacts"`
	LastAction string                 `json:"last_action"`
	CreatedAt  time.Time              `json:"created_at"`
	ExpiresAt  time.Time              `json:"expires_at"`
}

// NewLoopCheckpoint creates a new LoopCheckpoint with default expiration (24 hours).
func NewLoopCheckpoint(changeName string, loopNumber int, phase string) *LoopCheckpoint {
	now := time.Now()
	return &LoopCheckpoint{
		ChangeName: changeName,
		LoopNumber: loopNumber,
		Phase:      phase,
		State:      make(map[string]interface{}),
		Artifacts:  []string{},
		CreatedAt:  now,
		ExpiresAt:  now.Add(24 * time.Hour),
	}
}

// IsExpired checks if the checkpoint has expired.
func (cp *LoopCheckpoint) IsExpired() bool {
	return time.Now().After(cp.ExpiresAt)
}

// SaveLoopCheckpoint saves a Ralph Loop checkpoint to Engram.
func (c *EngramClient) SaveLoopCheckpoint(checkpoint *LoopCheckpoint) error {
	if checkpoint == nil {
		return gerrors.NewValidationError("checkpoint", "required", fmt.Errorf("checkpoint cannot be nil"))
	}
	if checkpoint.ChangeName == "" {
		return gerrors.NewValidationError("checkpoint.ChangeName", "required", fmt.Errorf("change name cannot be empty"))
	}

	// Store with key pattern: loop-checkpoint/{changeName}
	key := fmt.Sprintf("loop-checkpoint/%s", checkpoint.ChangeName)
	return c.Save(key, checkpoint)
}

// LoadLoopCheckpoint loads the most recent checkpoint for a change.
// Returns nil if no checkpoint exists or if it's expired.
func (c *EngramClient) LoadLoopCheckpoint(changeName string) (*LoopCheckpoint, error) {
	if changeName == "" {
		return nil, gerrors.NewValidationError("changeName", "required", fmt.Errorf("change name cannot be empty"))
	}

	key := fmt.Sprintf("loop-checkpoint/%s", changeName)

	value, err := c.Load(key)
	if err != nil {
		// If not found, return nil (no checkpoint)
		if err != nil && !IsEngramUnavailable(err) {
			// Check if it's a "not found" type error
			errStr := err.Error()
			if contains(errStr, "not found") || contains(errStr, "key not found") {
				return nil, nil
			}
		}
		// If Engram unavailable or other error, return nil for graceful degradation
		return nil, nil
	}

	// Parse the checkpoint
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal checkpoint: %w", err)
	}

	var checkpoint LoopCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("unmarshal checkpoint: %w", err)
	}

	// Check if expired
	if checkpoint.IsExpired() {
		// Delete expired checkpoint
		_ = c.Delete(key)
		return nil, nil
	}

	return &checkpoint, nil
}

// DeleteLoopCheckpoint removes a checkpoint for a change.
func (c *EngramClient) DeleteLoopCheckpoint(changeName string) error {
	if changeName == "" {
		return gerrors.NewValidationError("changeName", "required", fmt.Errorf("change name cannot be empty"))
	}

	key := fmt.Sprintf("loop-checkpoint/%s", changeName)
	return c.Delete(key)
}

// =============================================================================
// Optimization Patterns Integration
// =============================================================================

// OptiPattern represents a prompt optimization pattern learned over time.
type OptiPattern struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Pattern     string                 `json:"pattern"`
	Description string                 `json:"description"`
	Examples    []string               `json:"examples,omitempty"`
	SuccessRate float64                `json:"success_rate"`
	UsageCount  int                    `json:"usage_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// SaveOptiPattern saves a prompt optimization pattern to Engram.
func (c *EngramClient) SaveOptiPattern(pattern *OptiPattern) error {
	if pattern == nil {
		return gerrors.NewValidationError("pattern", "required", fmt.Errorf("pattern cannot be nil"))
	}
	if pattern.Name == "" {
		return gerrors.NewValidationError("pattern.Name", "required", fmt.Errorf("pattern name cannot be empty"))
	}
	if pattern.Category == "" {
		return gerrors.NewValidationError("pattern.Category", "required", fmt.Errorf("pattern category cannot be empty"))
	}

	// Set timestamps
	now := time.Now()
	if pattern.CreatedAt.IsZero() {
		pattern.CreatedAt = now
	}
	pattern.UpdatedAt = now

	// Store with key pattern: opti-pattern/{category}/{patternID}
	key := fmt.Sprintf("opti-pattern/%s/%s", pattern.Category, pattern.ID)
	return c.Save(key, pattern)
}

// LoadOptiPatterns loads all optimization patterns, optionally filtered by category.
func (c *EngramClient) LoadOptiPatterns(category string) ([]OptiPattern, error) {
	// If category specified, search for it specifically
	var keys []string
	var err error

	if category != "" {
		query := fmt.Sprintf("opti-pattern/%s", category)
		keys, err = c.Search(query)
	} else {
		// Load all patterns - search for prefix
		keys, err = c.Search("opti-pattern")
	}

	if err != nil {
		if IsEngramUnavailable(err) {
			return []OptiPattern{}, nil
		}
		return nil, fmt.Errorf("search opti patterns: %w", err)
	}

	var patterns []OptiPattern
	seen := make(map[string]bool)

	for _, key := range keys {
		// Avoid duplicates
		if seen[key] {
			continue
		}
		seen[key] = true

		value, err := c.Load(key)
		if err != nil {
			continue
		}

		data, err := json.Marshal(value)
		if err != nil {
			continue
		}

		var pattern OptiPattern
		if err := json.Unmarshal(data, &pattern); err != nil {
			continue
		}

		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// SaveOptiPatterns saves multiple optimization patterns at once.
func (c *EngramClient) SaveOptiPatterns(patterns []OptiPattern) error {
	for i := range patterns {
		if err := c.SaveOptiPattern(&patterns[i]); err != nil {
			return fmt.Errorf("save pattern %d: %w", i, err)
		}
	}
	return nil
}

// GetBestPatterns returns the top N patterns by success rate.
func (c *EngramClient) GetBestPatterns(category string, count int) ([]OptiPattern, error) {
	patterns, err := c.LoadOptiPatterns(category)
	if err != nil {
		return nil, err
	}

	if len(patterns) <= count {
		return patterns, nil
	}

	// Sort by success rate (descending)
	for i := 0; i < len(patterns)-1; i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[j].SuccessRate > patterns[i].SuccessRate {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	return patterns[:count], nil
}

// =============================================================================
// Change Metadata Integration
// =============================================================================

// ChangeMetadata stores metadata about a GROVE change for context tracking.
type ChangeMetadata struct {
	ChangeName string                 `json:"change_name"`
	Project    string                 `json:"project"`
	Status     string                 `json:"status"` // "exploring", "proposing", "specifying", "implementing", "verifying", "archived"
	Artifacts  []string               `json:"artifacts"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	LastPhase  string                 `json:"last_phase"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SaveChangeMetadata saves metadata about a change.
func (c *EngramClient) SaveChangeMetadata(meta *ChangeMetadata) error {
	if meta == nil {
		return gerrors.NewValidationError("meta", "required", fmt.Errorf("metadata cannot be nil"))
	}
	if meta.ChangeName == "" {
		return gerrors.NewValidationError("meta.ChangeName", "required", fmt.Errorf("change name cannot be empty"))
	}

	now := time.Now()
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = now
	}
	meta.UpdatedAt = now

	key := fmt.Sprintf("change-metadata/%s", meta.ChangeName)
	return c.Save(key, meta)
}

// LoadChangeMetadata loads metadata for a specific change.
func (c *EngramClient) LoadChangeMetadata(changeName string) (*ChangeMetadata, error) {
	if changeName == "" {
		return nil, gerrors.NewValidationError("changeName", "required", fmt.Errorf("change name cannot be empty"))
	}

	key := fmt.Sprintf("change-metadata/%s", changeName)

	value, err := c.Load(key)
	if err != nil {
		if IsEngramUnavailable(err) {
			return nil, nil
		}
		errStr := err.Error()
		if contains(errStr, "not found") || contains(errStr, "key not found") {
			return nil, nil
		}
		return nil, err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	var meta ChangeMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}

	return &meta, nil
}

// ListChanges returns all changes tracked by GROVE, optionally filtered by project.
func (c *EngramClient) ListChanges(project string) ([]ChangeMetadata, error) {
	keys, err := c.Search("change-metadata")
	if err != nil {
		if IsEngramUnavailable(err) {
			return []ChangeMetadata{}, nil
		}
		return nil, fmt.Errorf("list changes: %w", err)
	}

	var changes []ChangeMetadata
	seen := make(map[string]bool)

	for _, key := range keys {
		if seen[key] {
			continue
		}
		seen[key] = true

		value, err := c.Load(key)
		if err != nil {
			continue
		}

		data, err := json.Marshal(value)
		if err != nil {
			continue
		}

		var meta ChangeMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}

		// Filter by project if specified
		if project != "" && meta.Project != project {
			continue
		}

		changes = append(changes, meta)
	}

	return changes, nil
}

// =============================================================================
// Session Integration
// =============================================================================

// SessionSummary represents a GROVE session summary for persistence.
type SessionSummary struct {
	SessionID     string    `json:"session_id"`
	Project       string    `json:"project"`
	Goal          string    `json:"goal"`
	Instructions  string    `json:"instructions,omitempty"`
	Discoveries   []string  `json:"discoveries"`
	Accomplished  []string  `json:"accomplished"`
	NextSteps     []string  `json:"next_steps"`
	RelevantFiles []string  `json:"relevant_files"`
	Timestamp     time.Time `json:"timestamp"`
}

// SaveSessionSummary saves a session summary to Engram.
func (c *EngramClient) SaveSessionSummary(summary *SessionSummary) error {
	if summary == nil {
		return gerrors.NewValidationError("summary", "required", fmt.Errorf("summary cannot be nil"))
	}
	if summary.SessionID == "" {
		return gerrors.NewValidationError("summary.SessionID", "required", fmt.Errorf("session ID cannot be empty"))
	}

	if summary.Timestamp.IsZero() {
		summary.Timestamp = time.Now()
	}

	key := fmt.Sprintf("session/%s", summary.SessionID)
	return c.Save(key, summary)
}

// LoadSessionSummary loads a session summary by ID.
func (c *EngramClient) LoadSessionSummary(sessionID string) (*SessionSummary, error) {
	if sessionID == "" {
		return nil, gerrors.NewValidationError("sessionID", "required", fmt.Errorf("session ID cannot be empty"))
	}

	key := fmt.Sprintf("session/%s", sessionID)

	value, err := c.Load(key)
	if err != nil {
		if IsEngramUnavailable(err) {
			return nil, nil
		}
		errStr := err.Error()
		if contains(errStr, "not found") || contains(errStr, "key not found") {
			return nil, nil
		}
		return nil, err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal summary: %w", err)
	}

	var summary SessionSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("unmarshal summary: %w", err)
	}

	return &summary, nil
}

// GetRecentSessions returns the most recent session summaries.
func (c *EngramClient) GetRecentSessions(project string, count int) ([]SessionSummary, error) {
	// Search for session keys - filter by project in code
	query := "session"

	keys, err := c.Search(query)
	if err != nil {
		if IsEngramUnavailable(err) {
			return []SessionSummary{}, nil
		}
		return nil, fmt.Errorf("search sessions: %w", err)
	}

	var summaries []SessionSummary
	seen := make(map[string]bool)

	for _, key := range keys {
		if seen[key] {
			continue
		}
		seen[key] = true

		value, err := c.Load(key)
		if err != nil {
			continue
		}

		data, err := json.Marshal(value)
		if err != nil {
			continue
		}

		var summary SessionSummary
		if err := json.Unmarshal(data, &summary); err != nil {
			continue
		}

		// Filter by project if specified
		if project != "" && summary.Project != project {
			continue
		}

		summaries = append(summaries, summary)
	}

	// Sort by timestamp (most recent first) and limit
	if len(summaries) > 1 {
		for i := 0; i < len(summaries)-1; i++ {
			for j := i + 1; j < len(summaries); j++ {
				if summaries[j].Timestamp.After(summaries[i].Timestamp) {
					summaries[i], summaries[j] = summaries[j], summaries[i]
				}
			}
		}
	}

	if len(summaries) > count {
		summaries = summaries[:count]
	}

	return summaries, nil
}
