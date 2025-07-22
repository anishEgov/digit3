package service

import (
	"context"
	"errors"
	"fmt"

	"digit.org/workflow/internal/models"
	"digit.org/workflow/internal/repository"
	"digit.org/workflow/internal/security"
	"github.com/looplab/fsm"
)

type transitionService struct {
	instanceRepo repository.ProcessInstanceRepository
	stateRepo    repository.StateRepository
	actionRepo   repository.ActionRepository
	guard        security.Guard
}

// NewTransitionService creates a new instance of TransitionService.
func NewTransitionService(
	instanceRepo repository.ProcessInstanceRepository,
	stateRepo repository.StateRepository,
	actionRepo repository.ActionRepository,
	guard security.Guard,
) TransitionService {
	return &transitionService{
		instanceRepo: instanceRepo,
		stateRepo:    stateRepo,
		actionRepo:   actionRepo,
		guard:        guard,
	}
}

func (s *transitionService) Transition(ctx context.Context, processInstanceID *string, entityID, processCode, action string, comment *string, documents []models.Document, assignees *[]string, tenantID string) (*models.ProcessInstance, error) {
	// Build ProcessInstance from parameters
	instance := &models.ProcessInstance{
		ProcessID: processCode, // processCode parameter maps to ProcessID field
		EntityID:  entityID,
		Action:    action,
		Comment:   comment,
		Documents: documents,
		TenantID:  tenantID,
	}

	if processInstanceID != nil {
		instance.ID = *processInstanceID
	}

	if assignees != nil {
		instance.Assignees = *assignees
	}

	// 1. Get or Create Process Instance
	existingInstance, err := s.getOrCreateInstance(ctx, tenantID, instance)
	if err != nil {
		return nil, err
	}

	// 2. Find the action being performed
	actions, err := s.actionRepo.GetActionsByStateID(ctx, tenantID, existingInstance.CurrentState)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve actions for current state: %w", err)
	}
	var targetAction *models.Action
	for _, a := range actions {
		if a.Name == instance.Action {
			targetAction = a
			break
		}
	}
	if targetAction == nil {
		return nil, fmt.Errorf("action '%s' is not valid for the current state", instance.Action)
	}

	// 3. Authorization Guard Check
	userID := "anonymous"
	var userRoles []string

	// Extract user information from context (for testing)
	if uid := ctx.Value("userID"); uid != nil {
		if uidStr, ok := uid.(string); ok {
			userID = uidStr
		}
	}
	if roles := ctx.Value("userRoles"); roles != nil {
		if rolesSlice, ok := roles.([]string); ok {
			userRoles = rolesSlice
		}
	}

	guardCtx := security.GuardContext{
		UserRoles:       userRoles,
		UserID:          userID,
		ProcessInstance: existingInstance,
		Action:          targetAction,
	}
	can, err := s.guard.CanTransition(guardCtx)
	if err != nil {
		return nil, fmt.Errorf("guard check failed: %w", err)
	}
	if !can {
		return nil, errors.New("transition not permitted by guard")
	}

	// 4. Construct FSM and transition
	fsm, err := s.buildFSM(ctx, tenantID, existingInstance, eventsForState(actions))
	if err != nil {
		return nil, err
	}

	err = fsm.Event(ctx, instance.Action)
	if err != nil {
		return nil, fmt.Errorf("invalid state transition: %w", err)
	}

	// 5. Persist updated state
	existingInstance.CurrentState = fsm.Current()
	if err := s.instanceRepo.UpdateProcessInstance(ctx, existingInstance); err != nil {
		return nil, err
	}

	// 6. Populate NextActions with available actions from the current state
	nextActions, err := s.actionRepo.GetActionsByStateID(ctx, tenantID, existingInstance.CurrentState)
	if err != nil {
		// Log warning but don't fail the transition
		existingInstance.NextActions = []string{}
	} else {
		existingInstance.NextActions = make([]string, len(nextActions))
		for i, action := range nextActions {
			existingInstance.NextActions[i] = action.Name
		}
	}

	return existingInstance, nil
}

func (s *transitionService) getOrCreateInstance(ctx context.Context, tenantID string, instance *models.ProcessInstance) (*models.ProcessInstance, error) {
	// Try to find existing instance
	existingInstance, err := s.instanceRepo.GetProcessInstanceByEntityID(ctx, tenantID, instance.EntityID, instance.ProcessID)
	if err == nil {
		// Found it, update with new data from request
		existingInstance.Comment = instance.Comment
		existingInstance.Documents = instance.Documents
		existingInstance.Assignees = instance.Assignees
		existingInstance.Attributes = instance.Attributes
		return existingInstance, nil
	}

	// Not found, create a new one
	states, err := s.stateRepo.GetStatesByProcessID(ctx, tenantID, instance.ProcessID)
	if err != nil || len(states) == 0 {
		return nil, errors.New("cannot find any states for this process definition")
	}
	var initialState *models.State
	for _, state := range states {
		if state.IsInitial {
			initialState = state
			break
		}
	}
	if initialState == nil {
		return nil, errors.New("no initial state configured for this process")
	}

	instance.CurrentState = initialState.ID
	instance.Status = "ACTIVE" // Set default status for new instances
	if err := s.instanceRepo.CreateProcessInstance(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to create new process instance: %w", err)
	}
	return instance, nil
}

func (s *transitionService) buildFSM(ctx context.Context, tenantID string, instance *models.ProcessInstance, events fsm.Events) (*fsm.FSM, error) {
	return fsm.NewFSM(instance.CurrentState, events, fsm.Callbacks{}), nil
}

func eventsForState(actions []*models.Action) fsm.Events {
	var events fsm.Events
	for _, action := range actions {
		events = append(events, fsm.EventDesc{
			Name: action.Name,
			Src:  []string{action.CurrentState},
			Dst:  action.NextState,
		})
	}
	return events
}
