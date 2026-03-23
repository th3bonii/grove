package spec

import "fmt"

// Task represents a prioritizable task.
type Task struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Priority     string   `json:"priority"`
	Dependencies []string `json:"dependencies"`
	Component    string   `json:"component"`
}

// PrioritizedTask represents a task with execution order.
type PrioritizedTask struct {
	Task
	Order          int      `json:"order"`
	ResolvedDeps   []string `json:"resolved_deps"`
	PriorityReason string   `json:"priority_reason"`
	CanStart       bool     `json:"can_start"`
	BlockedBy      []string `json:"blocked_by"`
}

// Prioritizer handles task prioritization.
type Prioritizer struct{}

// NewPrioritizer creates a new prioritizer.
func NewPrioritizer() *Prioritizer {
	return &Prioritizer{}
}

// PrioritizeTasks orders tasks by dependencies and priority.
func (p *Prioritizer) PrioritizeTasks(tasks []Task) []PrioritizedTask {
	// Build dependency graph
	depGraph := make(map[string][]string)
	taskMap := make(map[string]Task)
	for _, t := range tasks {
		depGraph[t.ID] = t.Dependencies
		taskMap[t.ID] = t
	}

	// Detect cycles
	if err := ValidateNoCycles(depGraph); err != nil {
		// If cycles detected, return tasks as-is with warning
		return p.fallbackOrder(tasks)
	}

	// Topological sort
	sorted := topologicalSort(depGraph)

	// Create prioritized tasks
	result := make([]PrioritizedTask, 0, len(tasks))
	for i, id := range sorted {
		task := taskMap[id]
		deps := depGraph[id]

		canStart := true
		blockedBy := []string{}
		for _, dep := range deps {
			if !isCompleted(dep, result[:i]) {
				canStart = false
				blockedBy = append(blockedBy, dep)
			}
		}

		result = append(result, PrioritizedTask{
			Task:           task,
			Order:          i + 1,
			ResolvedDeps:   deps,
			PriorityReason: getPriorityReason(task, deps),
			CanStart:       canStart,
			BlockedBy:      blockedBy,
		})
	}

	return result
}

// ValidateNoCycles checks for circular dependencies.
func ValidateNoCycles(deps map[string][]string) error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var check func(node string) bool
	check = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range deps[node] {
			if !visited[dep] {
				if check(dep) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for node := range deps {
		if !visited[node] {
			if check(node) {
				return fmt.Errorf("circular dependency detected involving %s", node)
			}
		}
	}

	return nil
}

// DetectDependencies analyzes tasks for implicit dependencies.
func DetectDependencies(tasks []Task) map[string][]string {
	deps := make(map[string][]string)

	for _, task := range tasks {
		taskDeps := []string{}

		// Check for explicit dependencies
		if len(task.Dependencies) > 0 {
			taskDeps = append(taskDeps, task.Dependencies...)
		}

		// Check for implicit dependencies based on component
		if task.Component != "" {
			for _, other := range tasks {
				if other.ID != task.ID && other.Component == task.Component {
					// Same component = implicit dependency
					taskDeps = append(taskDeps, other.ID)
				}
			}
		}

		deps[task.ID] = taskDeps
	}

	return deps
}

func topologicalSort(graph map[string][]string) []string {
	visited := make(map[string]bool)
	result := []string{}

	var visit func(node string)
	visit = func(node string) {
		if visited[node] {
			return
		}
		visited[node] = true
		for _, dep := range graph[node] {
			visit(dep)
		}
		result = append([]string{node}, result...)
	}

	for node := range graph {
		visit(node)
	}

	return result
}

func isCompleted(id string, completed []PrioritizedTask) bool {
	for _, t := range completed {
		if t.ID == id {
			return true
		}
	}
	return false
}

func getPriorityReason(task Task, deps []string) string {
	if len(deps) == 0 {
		return "No dependencies — can start immediately"
	}
	if task.Priority == "critical" {
		return "Critical priority — blocks other tasks"
	}
	if task.Priority == "high" {
		return "High priority — important for core functionality"
	}
	return fmt.Sprintf("Depends on %d tasks", len(deps))
}

func (p *Prioritizer) fallbackOrder(tasks []Task) []PrioritizedTask {
	result := make([]PrioritizedTask, len(tasks))
	for i, task := range tasks {
		result[i] = PrioritizedTask{
			Task:           task,
			Order:          i + 1,
			ResolvedDeps:   task.Dependencies,
			PriorityReason: "Fallback order (cycles detected)",
			CanStart:       true,
			BlockedBy:      []string{},
		}
	}
	return result
}
