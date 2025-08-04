package repository

import (
	"context"

	"digit.org/workflow/internal/models"
)

// ProcessRepository defines the interface for process data operations.
type ProcessRepository interface {
	CreateProcess(ctx context.Context, process *models.Process) error
	GetProcessByID(ctx context.Context, tenantID, id string) (*models.Process, error)
	GetProcessByCode(ctx context.Context, tenantID, code string) (*models.Process, error)
	GetProcesses(ctx context.Context, tenantID string, ids []string, names []string) ([]*models.Process, error)
	UpdateProcess(ctx context.Context, process *models.Process) error
	DeleteProcess(ctx context.Context, tenantID, id string) error
}

// StateRepository defines the interface for state data operations.
type StateRepository interface {
	CreateState(ctx context.Context, state *models.State) error
	GetStatesByProcessID(ctx context.Context, tenantID, processID string) ([]*models.State, error)
	GetStateByID(ctx context.Context, tenantID, id string) (*models.State, error)
	GetStateByCodeAndProcess(ctx context.Context, tenantID, processID, code string) (*models.State, error)
	UpdateState(ctx context.Context, state *models.State) error
	DeleteState(ctx context.Context, tenantID, id string) error
}

// ActionRepository defines the interface for action-related database operations.
type ActionRepository interface {
	CreateAction(ctx context.Context, action *models.Action) (*models.Action, error)
	GetActionsByStateID(ctx context.Context, tenantID, stateID string) ([]*models.Action, error)
	GetActionByID(ctx context.Context, tenantID, id string) (*models.Action, error)
	UpdateAction(ctx context.Context, action *models.Action) error
	DeleteAction(ctx context.Context, tenantID, id string) error
}

// ParallelExecutionRepository defines the interface for parallel workflow coordination operations.
type ParallelExecutionRepository interface {
	CreateParallelExecution(ctx context.Context, execution *models.ParallelExecution) error
	GetParallelExecution(ctx context.Context, tenantID, entityID, processID, parallelStateID string) (*models.ParallelExecution, error)
	UpdateParallelExecution(ctx context.Context, execution *models.ParallelExecution) error
	MarkBranchCompleted(ctx context.Context, tenantID, entityID, processID, branchID string) error
	GetActiveParallelExecutions(ctx context.Context, tenantID, entityID, processID string) ([]*models.ParallelExecution, error)
}

// AttributeValidationRepository defines the interface for attribute validation data operations.
type AttributeValidationRepository interface {
	CreateAttributeValidation(ctx context.Context, validation *models.AttributeValidation) error
	GetAttributeValidationByID(ctx context.Context, tenantID, id string) (*models.AttributeValidation, error)
}

// ProcessInstanceRepository defines the interface for process instance data operations.
type ProcessInstanceRepository interface {
	CreateProcessInstance(ctx context.Context, instance *models.ProcessInstance) error
	GetProcessInstanceByID(ctx context.Context, tenantID, id string) (*models.ProcessInstance, error)
	GetProcessInstanceByEntityID(ctx context.Context, tenantID, entityID, processID string) (*models.ProcessInstance, error)
	GetLatestProcessInstanceByEntityID(ctx context.Context, tenantID, entityID, processID string) (*models.ProcessInstance, error)
	GetProcessInstancesByEntityID(ctx context.Context, tenantID, entityID, processID string, history bool) ([]*models.ProcessInstance, error)
	// Parallel workflow methods
	GetActiveParallelInstances(ctx context.Context, tenantID, entityID, processID string) ([]*models.ProcessInstance, error)
	GetInstancesByBranch(ctx context.Context, tenantID, entityID, processID, branchID string) ([]*models.ProcessInstance, error)
}

// EscalationConfigRepository handles persistence operations for escalation configurations.
type EscalationConfigRepository interface {
	CreateEscalationConfig(ctx context.Context, config *models.EscalationConfig) (*models.EscalationConfig, error)
	GetEscalationConfigByID(ctx context.Context, tenantID, id string) (*models.EscalationConfig, error)
	GetEscalationConfigsByProcessID(ctx context.Context, tenantID, processID string, stateCode string, isActive *bool) ([]*models.EscalationConfig, error)
	UpdateEscalationConfig(ctx context.Context, config *models.EscalationConfig) (*models.EscalationConfig, error)
	DeleteEscalationConfig(ctx context.Context, tenantID, id string) error
	GetActiveEscalationConfigs(ctx context.Context, tenantID string) ([]*models.EscalationConfig, error)
}

// More repository interfaces (ActionRepository, etc.) will be added here later.
