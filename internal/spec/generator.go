package spec

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/Gentleman-Programming/grove/internal/types"
)

// Generator creates specification documents from templates and content.
type Generator struct {
	config *types.Config
}

// NewGenerator creates a new Generator instance.
func NewGenerator(config *types.Config) *Generator {
	return &Generator{
		config: config,
	}
}

// GenerateSpecMD generates a SPEC.md document from a SpecDocument struct.
func (g *Generator) GenerateSpecMD(spec *types.SpecDocument) (string, error) {
	if spec == nil {
		return "", fmt.Errorf("spec is nil")
	}

	var buf bytes.Buffer

	// Header
	buf.WriteString("# SPEC.md - Especificación de Requisitos\n\n")
	buf.WriteString(fmt.Sprintf("**Cambio:** %s\n", spec.Title))
	buf.WriteString(fmt.Sprintf("**Versión:** %s\n", spec.Version))
	buf.WriteString(fmt.Sprintf("**Fecha:** %s\n\n", spec.CreatedAt.Format("2006-01-02")))

	// Overview section
	if spec.Overview != "" {
		buf.WriteString("## Resumen\n\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", spec.Overview))
	}

	// Requirements section
	buf.WriteString("## Requisitos\n\n")
	if len(spec.Requirements) == 0 {
		buf.WriteString("*No se han definido requisitos todavía.*\n\n")
	} else {
		for _, req := range spec.Requirements {
			buf.WriteString(fmt.Sprintf("### %s: %s\n\n", req.ID, req.Description))
			buf.WriteString(fmt.Sprintf("**Tipo:** %s\n", req.Type))
			buf.WriteString(fmt.Sprintf("**Prioridad:** %s\n\n", req.Priority))
		}
	}

	// Components section
	buf.WriteString("## Componentes\n\n")
	if len(spec.Components) == 0 {
		buf.WriteString("*No se han definido componentes todavía.*\n\n")
	} else {
		for _, comp := range spec.Components {
			buf.WriteString(fmt.Sprintf("### %s (%s)\n\n", comp.Name, comp.Type))
			buf.WriteString(fmt.Sprintf("%s\n\n", comp.Description))
		}
	}

	// User Flows section
	buf.WriteString("## Flujos de Usuario\n\n")
	if len(spec.UserFlows) == 0 {
		buf.WriteString("*No se han definido flujos todavía.*\n\n")
	} else {
		for _, flow := range spec.UserFlows {
			buf.WriteString(fmt.Sprintf("### %s: %s\n\n", flow.ID, flow.Name))
			buf.WriteString(fmt.Sprintf("%s\n\n", flow.Description))
			buf.WriteString("**Pasos:**\n")
			for _, step := range flow.Steps {
				buf.WriteString(fmt.Sprintf("- %d. %s (%s)\n", step.StepNumber, step.Action, step.ComponentID))
			}
			buf.WriteString("\n")
		}
	}

	// Assumptions section
	if len(spec.Assumptions) > 0 {
		buf.WriteString("## Supuestos\n\n")
		for _, assumption := range spec.Assumptions {
			buf.WriteString(fmt.Sprintf("- **%s:** %s\n", assumption.ID, assumption.Statement))
			if assumption.Rationale != "" {
				buf.WriteString(fmt.Sprintf("  - Justificación: %s\n", assumption.Rationale))
			}
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// GenerateDesignMD generates a DESIGN.md document from a DesignDocument struct.
func (g *Generator) GenerateDesignMD(design *types.DesignDocument) (string, error) {
	if design == nil {
		return "", fmt.Errorf("design is nil")
	}

	var buf bytes.Buffer

	// Header
	buf.WriteString("# DESIGN.md - Diseño Técnico\n\n")
	buf.WriteString(fmt.Sprintf("**Cambio:** %s\n", design.Title))
	buf.WriteString(fmt.Sprintf("**Versión:** %s\n", design.Version))
	buf.WriteString(fmt.Sprintf("**Fecha:** %s\n\n", design.CreatedAt.Format("2006-01-02")))

	// Architecture section
	buf.WriteString("## Arquitectura\n\n")
	buf.WriteString(fmt.Sprintf("%s\n\n", design.Architecture))

	// Tech Stack section
	if len(design.TechStack) > 0 {
		buf.WriteString("## Stack Tecnológico\n\n")
		buf.WriteString("| Tecnología | Versión | Propósito |\n")
		buf.WriteString("|------------|---------|-----------|\n")
		for _, tech := range design.TechStack {
			buf.WriteString(fmt.Sprintf("| %s | %s | %s |\n", tech.Name, tech.Version, tech.Purpose))
		}
		buf.WriteString("\n")
	}

	// Decisions section
	buf.WriteString("## Decisiones de Arquitectura\n\n")
	if len(design.Decisions) == 0 {
		buf.WriteString("*No se han registrado decisiones todavía.*\n\n")
	} else {
		for _, dec := range design.Decisions {
			buf.WriteString(fmt.Sprintf("### %s: %s\n\n", dec.ID, dec.Title))
			buf.WriteString(fmt.Sprintf("**Decisión:** %s\n\n", dec.Decision))
			if len(dec.Alternatives) > 0 {
				buf.WriteString(fmt.Sprintf("**Alternativas:** %s\n\n", strings.Join(dec.Alternatives, ", ")))
			}
			buf.WriteString(fmt.Sprintf("**Justificación:** %s\n\n", dec.Justification))
			if dec.Consequences != "" {
				buf.WriteString(fmt.Sprintf("**Consecuencias:** %s\n\n", dec.Consequences))
			}
		}
	}

	// Directory structure section
	if design.DirectoryStructure != "" {
		buf.WriteString("## Estructura de Directorios\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", design.DirectoryStructure))
	}

	return buf.String(), nil
}

// GenerateTasksMD generates a TASKS.md document from a TasksDocument struct.
func (g *Generator) GenerateTasksMD(tasks *types.TasksDocument) (string, error) {
	if tasks == nil {
		return "", fmt.Errorf("tasks is nil")
	}

	var buf bytes.Buffer

	// Header
	buf.WriteString("# TASKS.md - Lista de Tareas\n\n")
	buf.WriteString(fmt.Sprintf("**Cambio:** %s\n", tasks.Title))
	buf.WriteString(fmt.Sprintf("**Versión:** %s\n", tasks.Version))
	buf.WriteString(fmt.Sprintf("**Fecha:** %s\n\n", tasks.CreatedAt.Format("2006-01-02")))
	buf.WriteString(fmt.Sprintf("**Total de tareas:** %d\n\n", len(tasks.Tasks)))

	// Group tasks by status
	byStatus := make(map[types.TaskStatus][]types.Task)
	for _, task := range tasks.Tasks {
		byStatus[task.Status] = append(byStatus[task.Status], task)
	}

	// Status summary
	buf.WriteString("## Resumen por Estado\n\n")
	buf.WriteString("| Estado | Cantidad |\n")
	buf.WriteString("|--------|----------|\n")
	statuses := []types.TaskStatus{types.TaskStatusCompleted, types.TaskStatusInProgress, types.TaskStatusBlocked, types.TaskStatusPending}
	for _, status := range statuses {
		count := len(byStatus[status])
		if count > 0 {
			buf.WriteString(fmt.Sprintf("| %s | %d |\n", status, count))
		}
	}
	buf.WriteString("\n")

	// Tasks by status
	for _, status := range statuses {
		taskList, ok := byStatus[status]
		if !ok || len(taskList) == 0 {
			continue
		}

		statusLabel := mapStatusToLabel(status)
		buf.WriteString(fmt.Sprintf("## %s\n\n", statusLabel))

		for _, task := range taskList {
			checkbox := "[ ]"
			if task.Status == types.TaskStatusCompleted {
				checkbox = "[x]"
			}

			priorityBadge := mapPriorityToBadgeString(task.Priority)
			buf.WriteString(fmt.Sprintf("- %s %s **%s** (%s)\n", checkbox, priorityBadge, task.Title, task.ID))
			buf.WriteString(fmt.Sprintf("  - %s\n", task.Description))

			if task.ComponentID != "" {
				buf.WriteString(fmt.Sprintf("  - Componente: %s\n", task.ComponentID))
			}
			if task.EstimatedEffort != "" {
				buf.WriteString(fmt.Sprintf("  - Estimación: %s\n", task.EstimatedEffort))
			}
			if len(task.Dependencies) > 0 {
				buf.WriteString(fmt.Sprintf("  - Depende de: %s\n", strings.Join(task.Dependencies, ", ")))
			}
			if len(task.Skills) > 0 {
				buf.WriteString(fmt.Sprintf("  - Habilidades: %s\n", strings.Join(task.Skills, ", ")))
			}
			buf.WriteString("\n")
		}
	}

	// Milestones section
	if len(tasks.Milestones) > 0 {
		buf.WriteString("## Hitos\n\n")
		for _, milestone := range tasks.Milestones {
			buf.WriteString(fmt.Sprintf("### %s: %s\n\n", milestone.ID, milestone.Name))
			buf.WriteString(fmt.Sprintf("%s\n\n", milestone.Description))
			buf.WriteString(fmt.Sprintf("**Tareas:** %s\n\n", strings.Join(milestone.TaskIDs, ", ")))
		}
	}

	return buf.String(), nil
}

// GenerateAgentsMD generates an AGENTS.md document.
func (g *Generator) GenerateAgentsMD(changeName string, tasks []types.Task) (string, error) {
	var buf bytes.Buffer

	buf.WriteString("# AGENTS.md - Configuración de Agentes\n\n")
	buf.WriteString(fmt.Sprintf("**Cambio:** %s\n", changeName))
	buf.WriteString(fmt.Sprintf("**Fecha:** %s\n\n", time.Now().Format("2006-01-02")))

	buf.WriteString("## Habilidades Disponibles\n\n")
	buf.WriteString("| Habilidad | Descripción |\n")
	buf.WriteString("|-----------|-------------|\n")
	buf.WriteString("| sdd-apply | Implementar tareas siguiendo specs y diseño |\n")
	buf.WriteString("| sdd-verify | Validar implementación contra specs |\n")
	buf.WriteString("| sdd-archive | Sincronizar specs y archivar cambios |\n")
	buf.WriteString("| go-testing | Patrones de testing en Go |\n")
	buf.WriteString("\n")

	buf.WriteString("## Tareas Asignables\n\n")
	for _, task := range tasks {
		if task.Status == types.TaskStatusPending {
			buf.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", task.ID, task.Title, task.Priority))
		}
	}
	buf.WriteString("\n")

	return buf.String(), nil
}

// GenerateVerifyReport generates a verification report.
func (g *Generator) GenerateVerifyReport(ctx *types.Context) (*types.Report, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	report := &types.Report{
		ChangeName: ctx.Change.Name,
		Type:       types.ReportVerify,
		CreatedAt:  time.Now(),
		Checks:     make([]types.ReportCheck, 0),
		Issues:     make([]types.Issue, 0),
	}

	// Check spec existence
	if ctx.Spec != nil {
		report.Checks = append(report.Checks, types.ReportCheck{
			Name:   "Spec Document",
			Status: types.CheckPass,
			Message: fmt.Sprintf("Spec found with %d requirements and %d components",
				len(ctx.Spec.Requirements), len(ctx.Spec.Components)),
		})
	} else {
		report.Checks = append(report.Checks, types.ReportCheck{
			Name:    "Spec Document",
			Status:  types.CheckFail,
			Message: "Spec document is missing",
		})
	}

	// Check design existence
	if ctx.Design != nil {
		report.Checks = append(report.Checks, types.ReportCheck{
			Name:   "Design Document",
			Status: types.CheckPass,
			Message: fmt.Sprintf("Design found with %d decisions",
				len(ctx.Design.Decisions)),
		})
	} else {
		report.Checks = append(report.Checks, types.ReportCheck{
			Name:    "Design Document",
			Status:  types.CheckFail,
			Message: "Design document is missing",
		})
	}

	// Check tasks existence
	if ctx.Tasks != nil && len(ctx.Tasks.Tasks) > 0 {
		pendingTasks := 0
		doneTasks := 0
		for _, task := range ctx.Tasks.Tasks {
			if task.Status == types.TaskStatusCompleted {
				doneTasks++
			} else {
				pendingTasks++
			}
		}
		report.Checks = append(report.Checks, types.ReportCheck{
			Name:    "Task List",
			Status:  types.CheckPass,
			Message: fmt.Sprintf("Tasks found: %d done, %d pending", doneTasks, pendingTasks),
		})
	} else {
		report.Checks = append(report.Checks, types.ReportCheck{
			Name:    "Task List",
			Status:  types.CheckFail,
			Message: "Task list is missing or empty",
		})
	}

	// Check quality score
	if ctx.Score != nil {
		if ctx.Score.Overall >= 7.0 {
			report.Checks = append(report.Checks, types.ReportCheck{
				Name:    "Quality Score",
				Status:  types.CheckPass,
				Message: fmt.Sprintf("Quality score: %.1f/10", ctx.Score.Overall),
			})
		} else if ctx.Score.Overall >= 5.0 {
			report.Checks = append(report.Checks, types.ReportCheck{
				Name:    "Quality Score",
				Status:  types.CheckWarning,
				Message: fmt.Sprintf("Quality score: %.1f/10 (needs improvement)", ctx.Score.Overall),
			})
		} else {
			report.Checks = append(report.Checks, types.ReportCheck{
				Name:    "Quality Score",
				Status:  types.CheckFail,
				Message: fmt.Sprintf("Quality score too low: %.1f/10", ctx.Score.Overall),
			})
		}
	}

	// Generate summary
	passed := 0
	failed := 0
	warnings := 0
	for _, check := range report.Checks {
		switch check.Status {
		case types.CheckPass:
			passed++
		case types.CheckFail:
			failed++
		case types.CheckWarning:
			warnings++
		}
	}

	report.Summary = fmt.Sprintf("Verification complete: %d passed, %d warnings, %d failed",
		passed, warnings, failed)

	return report, nil
}

// GenerateFromTemplate generates content from input template.
func (g *Generator) GenerateFromTemplate(input string, phase types.Phase) (string, error) {
	var buf bytes.Buffer

	switch phase {
	case types.PhaseSpec:
		buf.WriteString("# SPEC.md - Especificación de Requisitos\n\n")
		buf.WriteString("**Cambio:** [NOMBRE DEL CAMBIO]\n")
		buf.WriteString(fmt.Sprintf("**Fecha:** %s\n\n", time.Now().Format("2006-01-02")))
		buf.WriteString("## Resumen\n\n")
		buf.WriteString(input)
		buf.WriteString("\n\n## Requisitos\n\n")
		buf.WriteString("### REQ-001: [Título del Requisito]\n\n")
		buf.WriteString("**Prioridad:** [critical/high/medium/low]\n\n")
		buf.WriteString("Descripción del requisito...\n\n")
		buf.WriteString("## Escenarios de Prueba\n\n")
		buf.WriteString("### SCN-001: [Título del Escenario]\n\n")
		buf.WriteString("**Dado** [contexto inicial]\n")
		buf.WriteString("**Cuando** [acción realizada]\n")
		buf.WriteString("**Entonces** [resultado esperado]\n\n")
		buf.WriteString("## Restricciones\n\n")
		buf.WriteString("- Restricción 1\n")
		buf.WriteString("- Restricción 2\n\n")

	case types.PhaseDesign:
		buf.WriteString("# DESIGN.md - Diseño Técnico\n\n")
		buf.WriteString("**Cambio:** [NOMBRE DEL CAMBIO]\n")
		buf.WriteString(fmt.Sprintf("**Fecha:** %s\n\n", time.Now().Format("2006-01-02")))
		buf.WriteString("## Arquitectura\n\n")
		buf.WriteString(input)
		buf.WriteString("\n\n## Componentes\n\n")
		buf.WriteString("### [Nombre del Componente]\n\n")
		buf.WriteString("**Tipo:** [service/component/module]\n\n")
		buf.WriteString("**Responsabilidad:** Descripción de responsabilidad\n\n")
		buf.WriteString("**Interfaces:** Lista de interfaces expuestas\n\n")
		buf.WriteString("## Decisiones de Arquitectura\n\n")
		buf.WriteString("### ADR-001: [Título]\n\n")
		buf.WriteString("**Problema:** Descripción del problema\n\n")
		buf.WriteString("**Decisión:** Solución adoptada\n\n")
		buf.WriteString("**Consecuencias:** Impacto de la decisión\n\n")

	case types.PhaseTasks:
		buf.WriteString("# TASKS.md - Lista de Tareas\n\n")
		buf.WriteString("**Cambio:** [NOMBRE DEL CAMBIO]\n")
		buf.WriteString(fmt.Sprintf("**Fecha:** %s\n\n", time.Now().Format("2006-01-02")))
		buf.WriteString("## Resumen\n\n")
		buf.WriteString(input)
		buf.WriteString("\n\n## Tareas\n\n")
		buf.WriteString("- [ ] 🔴 TASK-001: [Título de la tarea]\n")
		buf.WriteString("  - Descripción detallada\n")
		buf.WriteString("  - Estimación: [esfuerzo]\n")
		buf.WriteString("  - Depende de: [otras tareas]\n\n")
		buf.WriteString("- [ ] 🟡 TASK-002: [Título de la tarea]\n")
		buf.WriteString("  - Descripción detallada\n")
		buf.WriteString("  - Estimación: [esfuerzo]\n\n")

	default:
		buf.WriteString(input)
	}

	return buf.String(), nil
}

// GenerateArchiveReport generates an archive report.
func (g *Generator) GenerateArchiveReport(ctx *types.Context) (*types.Report, error) {
	if ctx == nil || ctx.Change == nil {
		return nil, fmt.Errorf("context or change is nil")
	}

	report := &types.Report{
		ChangeName: ctx.Change.Name,
		Type:       types.ReportArchive,
		CreatedAt:  time.Now(),
		Checks:     make([]types.ReportCheck, 0),
		Issues:     make([]types.Issue, 0),
		Summary:    "Archive completed successfully",
	}

	// Verify all phases completed
	phases := []types.Phase{
		types.PhaseSpec,
		types.PhaseDesign,
		types.PhaseTasks,
		types.PhaseApply,
		types.PhaseVerify,
	}

	for _, phase := range phases {
		report.Checks = append(report.Checks, types.ReportCheck{
			Name:    fmt.Sprintf("Phase: %s", phase),
			Status:  types.CheckPass,
			Message: fmt.Sprintf("Phase %s completed", phase),
		})
	}

	return report, nil
}

// Helper functions

func mapStatusToLabel(status types.TaskStatus) string {
	switch status {
	case types.TaskStatusCompleted:
		return "✅ Completadas"
	case types.TaskStatusInProgress:
		return "🔄 En Progreso"
	case types.TaskStatusBlocked:
		return "🚫 Bloqueadas"
	case types.TaskStatusPending:
		return "📋 Pendientes"
	default:
		return string(status)
	}
}

func mapPriorityToBadge(priority types.Priority) string {
	switch priority {
	case types.PriorityCritical:
		return "🔴"
	case types.PriorityHigh:
		return "🟠"
	case types.PriorityMedium:
		return "🟡"
	case types.PriorityLow:
		return "🟢"
	default:
		return "⚪"
	}
}

func mapPriorityToBadgeString(priority string) string {
	switch priority {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "⚪"
	}
}
