package models

// Process represents a workflow process definition.
// It corresponds to the 'Process' schema in the OpenAPI specification.
type Process struct {
	ID          string      `json:"id,omitempty" db:"id"`
	TenantID    string      `json:"-" db:"tenant_id"` // Never serialize tenantId in JSON - comes from header only
	Name        string      `json:"name" db:"name"`
	Code        string      `json:"code" db:"code"`
	Description *string     `json:"description,omitempty" db:"description"`
	Version     *string     `json:"version,omitempty" db:"version"`
	SLA         *int64      `json:"sla,omitempty" db:"sla"`
	AuditDetail AuditDetail `json:"auditDetail,omitempty" db:",inline"`
}

// State represents a state within a workflow process.
// It corresponds to the 'State' schema in the OpenAPI specification.
type State struct {
	ID           string      `json:"id,omitempty" db:"id"`
	TenantID     string      `json:"-" db:"tenant_id"` // Never serialize tenantId in JSON - comes from header only
	ProcessID    string      `json:"processId" db:"process_id"`
	Code         string      `json:"code" db:"code"`
	Name         string      `json:"name" db:"name"`
	Description  *string     `json:"description,omitempty" db:"description"`
	SLA          *int64      `json:"sla,omitempty" db:"sla"`
	IsInitial    bool        `json:"isInitial,omitempty" db:"is_initial"`
	IsParallel   bool        `json:"isParallel,omitempty" db:"is_parallel"`
	IsJoin       bool        `json:"isJoin,omitempty" db:"is_join"`
	BranchStates []string    `json:"branchStates,omitempty" db:"branch_states"`
	AuditDetail  AuditDetail `json:"auditDetail,omitempty" db:",inline"`
}

// Action represents a transition between two states.
// It corresponds to the 'Action' schema in the OpenAPI specification.
type Action struct {
	ID                    string               `json:"id,omitempty" db:"id"`
	TenantID              string               `json:"-" db:"tenant_id"` // Never serialize tenantId in JSON - comes from header only
	Name                  string               `json:"name" db:"name"`
	Label                 *string              `json:"label,omitempty" db:"label"`
	CurrentState          string               `json:"currentState" db:"current_state_id"`
	NextState             string               `json:"nextState" db:"next_state_id"`
	Roles                 []string             `json:"roles,omitempty" db:"roles"`
	AttributeValidationID *string              `json:"-" db:"attribute_validation_id"` // Internal DB relation
	AttributeValidation   *AttributeValidation `json:"attributeValidation,omitempty" db:"-"`
	AuditDetail           AuditDetail          `json:"auditDetail,omitempty" db:",inline"`
}

// AttributeValidation defines guard conditions for an action.
// It corresponds to the 'AttributeValidation' schema in the OpenAPI specification.
type AttributeValidation struct {
	ID            string              `json:"id,omitempty" db:"id"`
	TenantID      string              `json:"-" db:"tenant_id"` // Never serialize tenantId in JSON - comes from header only
	Attributes    map[string][]string `json:"attributes,omitempty" db:"attributes"`
	AssigneeCheck bool                `json:"assigneeCheck,omitempty" db:"assignee_check"`
	AuditDetail   AuditDetail         `json:"auditDetail,omitempty" db:",inline"`
}

// ProcessInstance represents a running instance of a process for a specific entity.
// It corresponds to the 'ProcessInstance' schema in the OpenAPI specification.
type ProcessInstance struct {
	ID           string              `json:"id,omitempty" db:"id"`
	TenantID     string              `json:"-" db:"tenant_id"` // Never serialize tenantId in JSON - comes from header only
	ProcessID    string              `json:"processId" db:"process_id"`
	EntityID     string              `json:"entityId" db:"entity_id"`
	Action       string              `json:"action,omitempty" db:"action"`
	Status       string              `json:"status,omitempty" db:"status"`
	Comment      *string             `json:"comment,omitempty" db:"comment"`
	Documents    []Document          `json:"documents,omitempty" db:"documents"`
	Assigner     *string             `json:"assigner,omitempty" db:"assigner"`
	Assignees    []string            `json:"assignees,omitempty" db:"assignees"`
	CurrentState string              `json:"currentState" db:"current_state_id"`
	StateSLA     *int64              `json:"stateSla,omitempty" db:"state_sla"`
	ProcessSLA   *int64              `json:"processSla,omitempty" db:"process_sla"`
	Attributes   map[string][]string `json:"attributes,omitempty" db:"attributes"`
	NextActions  []string            `json:"nextActions,omitempty" db:"-"` // Not stored in DB, computed from current state's actions
	AuditDetails AuditDetail         `json:"auditDetails,omitempty" db:",inline"`
}

// StateDetail is a response model that includes a state's information along with its possible actions.
type StateDetail struct {
	State
	Actions []Action `json:"actions,omitempty"`
}

// ProcessDefinitionDetail is a response model for the full definition of a process, including all states and actions.
type ProcessDefinitionDetail struct {
	Process
	States []StateDetail `json:"states,omitempty"`
}
