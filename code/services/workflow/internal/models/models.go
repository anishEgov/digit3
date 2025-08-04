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

// Action represents a transition between two states in a process.
// It corresponds to the 'Action' schema in the OpenAPI specification.
type Action struct {
	ID                    string               `json:"id,omitempty" db:"id"`
	TenantID              string               `json:"-" db:"tenant_id"` // Never serialize tenantId in JSON - comes from header only
	Name                  string               `json:"name" db:"name"`
	Label                 *string              `json:"label,omitempty" db:"label"`
	CurrentState          string               `json:"currentState" db:"current_state_id"`
	NextState             string               `json:"nextState" db:"next_state_id"`
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
	// Parallel workflow fields
	ParentInstanceID *string     `json:"parentInstanceId,omitempty" db:"parent_instance_id"` // For tracking parallel branches
	BranchID         *string     `json:"branchId,omitempty" db:"branch_id"`                  // Which parallel branch this represents
	IsParallelBranch bool        `json:"isParallelBranch,omitempty" db:"is_parallel_branch"` // Flag for parallel instances
	AuditDetails     AuditDetail `json:"auditDetails,omitempty" db:",inline"`
}

// ParallelExecution tracks the coordination of parallel workflow branches
type ParallelExecution struct {
	ID                string      `json:"id,omitempty" db:"id"`
	TenantID          string      `json:"-" db:"tenant_id"`
	EntityID          string      `json:"entityId" db:"entity_id"`
	ProcessID         string      `json:"processId" db:"process_id"`
	ParallelStateID   string      `json:"parallelStateId" db:"parallel_state_id"`    // State that created branches
	JoinStateID       string      `json:"joinStateId" db:"join_state_id"`            // State where branches merge
	ActiveBranches    []string    `json:"activeBranches" db:"active_branches"`       // Which branches are still running
	CompletedBranches []string    `json:"completedBranches" db:"completed_branches"` // Which branches reached join
	Status            string      `json:"status" db:"status"`                        // ACTIVE, WAITING_FOR_JOIN, COMPLETED
	AuditDetail       AuditDetail `json:"auditDetail,omitempty" db:",inline"`
}

// EscalationConfig represents an auto-escalation rule for a process state.
// It defines when and how to automatically escalate workflow instances based on SLA breaches.
type EscalationConfig struct {
	ID                string      `json:"id,omitempty" db:"id"`
	TenantID          string      `json:"-" db:"tenant_id"` // Never serialize tenantId in JSON - comes from header only
	ProcessID         string      `json:"processId" db:"process_id"`
	StateCode         string      `json:"stateCode" db:"state_code"`
	EscalationAction  string      `json:"escalationAction" db:"escalation_action"`
	StateSlaMinutes   *int        `json:"stateSlaMinutes,omitempty" db:"state_sla_minutes"`
	ProcessSlaMinutes *int        `json:"processSlaMinutes,omitempty" db:"process_sla_minutes"`
	AuditDetail       AuditDetail `json:"auditDetail,omitempty" db:",inline"`
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
