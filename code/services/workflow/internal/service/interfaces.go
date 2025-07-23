package service

import (
	"context"

	"digit.org/workflow/internal/models"
)

// ProcessService defines the interface for process business logic.
type ProcessService interface {
	CreateProcess(ctx context.Context, process *models.Process) (*models.Process, error)
	GetProcessByID(ctx context.Context, tenantID, id string) (*models.Process, error)
	GetProcesses(ctx context.Context, tenantID string, ids []string, names []string) ([]*models.Process, error)
	GetProcessDefinitions(ctx context.Context, tenantID string, ids []string, names []string) ([]*models.ProcessDefinitionDetail, error)
	UpdateProcess(ctx context.Context, process *models.Process) (*models.Process, error)
	DeleteProcess(ctx context.Context, tenantID, id string) error
}

// StateService defines the interface for state business logic.
type StateService interface {
	CreateState(ctx context.Context, state *models.State) (*models.State, error)
	GetStatesByProcessID(ctx context.Context, tenantID, processID string) ([]*models.State, error)
	GetStateByID(ctx context.Context, tenantID, id string) (*models.State, error)
	UpdateState(ctx context.Context, state *models.State) (*models.State, error)
	DeleteState(ctx context.Context, tenantID, id string) error
}

// ActionService defines the interface for action business logic.
type ActionService interface {
	CreateAction(ctx context.Context, action *models.Action) (*models.Action, error)
	GetActionsByStateID(ctx context.Context, tenantID, stateID string) ([]*models.Action, error)
	GetActionByID(ctx context.Context, tenantID, id string) (*models.Action, error)
	UpdateAction(ctx context.Context, action *models.Action) (*models.Action, error)
	DeleteAction(ctx context.Context, tenantID, id string) error
}

// TransitionService defines the interface for handling state transitions.
type TransitionService interface {
	Transition(ctx context.Context, processInstanceID *string, entityID, processCode, action string, comment *string, documents []models.Document, assignees *[]string, attributes map[string][]string, tenantID string) (*models.ProcessInstance, error)
}

// More service interfaces will be added here later.
