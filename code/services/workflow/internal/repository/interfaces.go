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
	UpdateState(ctx context.Context, state *models.State) error
	DeleteState(ctx context.Context, tenantID, id string) error
}

// ActionRepository defines the interface for action data operations.
type ActionRepository interface {
	CreateAction(ctx context.Context, action *models.Action) error
	GetActionsByStateID(ctx context.Context, tenantID, stateID string) ([]*models.Action, error)
	GetActionByID(ctx context.Context, tenantID, id string) (*models.Action, error)
	UpdateAction(ctx context.Context, action *models.Action) error
	DeleteAction(ctx context.Context, tenantID, id string) error
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
	UpdateProcessInstance(ctx context.Context, instance *models.ProcessInstance) error
}

// More repository interfaces (ActionRepository, etc.) will be added here later.
