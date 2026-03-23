// Package loop provides the core Ralph Loop engine for GROVE.
//
// Ralph Loop is an autonomous documentation-to-code execution engine that:
//   - Validates documentation before processing
//   - Loads and manages implementation tasks
//   - Orchestrates execution across multiple phases
//   - Persists state for checkpoint/resume capability
package loop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Gentleman-Programming/grove/internal/sdd"
	"github.com/Gentleman-Programming/grove/internal/types"
)

// Verifier valida implementaciones contra specs delega al skill sdd-verify.
type Verifier struct {
	sddClient      *sdd.Client
	projectPath    string
	verifyAfterAll bool
}

// NewVerifier crea un nuevo Verifier.
func NewVerifier(sddClient *sdd.Client, projectPath string) *Verifier {
	return &Verifier{
		sddClient:      sddClient,
		projectPath:    projectPath,
		verifyAfterAll: true,
	}
}

// VerifyTask valida que una tarea se implementó correctamente.
// Retorna un VerifyReport con el resultado de la verificación.
func (v *Verifier) VerifyTask(ctx context.Context, taskID string, filesChanged []string, spec *types.SpecDocument) (*types.VerifyReport, error) {
	// Preparar input para SDD verify
	input := map[string]interface{}{
		"task_id":       taskID,
		"files_changed": filesChanged,
		"spec":          spec,
		"project_path":  v.projectPath,
	}

	// Ejecutar verificación via SDD client
	result, err := v.sddClient.Execute(ctx, sdd.PhaseVerify, input)
	if err != nil {
		return &types.VerifyReport{
			TaskID:      taskID,
			Timestamp:   time.Now(),
			Status:      types.VerifyStatusFailed,
			Message:     fmt.Sprintf("Verification execution failed: %v", err),
			PassedCount: 0,
			FailedCount: 1,
		}, err
	}

	// Parsear resultado del SDD
	return v.parseVerifyResult(result, taskID)
}

// parseVerifyResult convierte el resultado SDD en un VerifyReport.
func (v *Verifier) parseVerifyResult(result *sdd.Result, taskID string) (*types.VerifyReport, error) {
	report := &types.VerifyReport{
		TaskID:      taskID,
		Timestamp:   time.Now(),
		Status:      types.VerifyStatusPassed,
		Checks:      []types.VerifyCheck{},
		PassedCount: 0,
		FailedCount: 0,
	}

	// Analizar el summary del result para determinar status
	summary := strings.ToLower(result.Summary)
	if strings.Contains(summary, "fail") {
		report.Status = types.VerifyStatusFailed
		report.FailedCount = 1
		report.Message = result.Summary
	} else if strings.Contains(summary, "warning") {
		report.Status = types.VerifyStatusWarning
		report.Message = result.Summary
	} else {
		report.Status = types.VerifyStatusPassed
		report.PassedCount = 1
		report.Message = result.Summary
	}

	// Agregar metadata del result como suggestions
	if len(result.Metadata) > 0 {
		for k, val := range result.Metadata {
			report.Suggestions = append(report.Suggestions, fmt.Sprintf("%s: %v", k, val))
		}
	}

	return report, nil
}

// VerifyAllTasks verifica todas las tareas completadas.
// Útil para verificación final después de ejecutar todas las tareas.
func (v *Verifier) VerifyAllTasks(ctx context.Context, tasks []Task, spec *types.SpecDocument) ([]*types.VerifyReport, error) {
	reports := make([]*types.VerifyReport, 0, len(tasks))

	for i := range tasks {
		task := &tasks[i]
		if !task.Completed {
			continue // Skip incomplete tasks
		}

		report, err := v.VerifyTask(ctx, task.ID, nil, spec)
		if err != nil {
			// Log error pero continuar con otras tareas
			continue
		}
		reports = append(reports, report)
	}

	return reports, nil
}

// GetVerifierConfig retorna la configuración actual del verifier.
func (v *Verifier) GetVerifierConfig() map[string]interface{} {
	return map[string]interface{}{
		"project_path":     v.projectPath,
		"verify_after_all": v.verifyAfterAll,
	}
}
