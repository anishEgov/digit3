package security

import (
	"errors"
	"fmt"
)

// RBACGuard provides a default implementation of the Guard interface
// based on roles, assignees, and attributes.
type RBACGuard struct{}

// NewRBACGuard creates a new RBACGuard.
func NewRBACGuard() Guard {
	return &RBACGuard{}
}

// CanTransition checks if a user is permitted to perform a state transition.
func (g *RBACGuard) CanTransition(ctx GuardContext) (bool, error) {
	// 1. Role Check
	if !hasRequiredRole(ctx.UserRoles, ctx.Action.Roles) {
		return false, errors.New("user does not have the required role for this action")
	}

	// 2. Assignee Check
	if ctx.Action.AttributeValidation != nil && ctx.Action.AttributeValidation.AssigneeCheck {
		if !isAssignee(ctx.UserID, ctx.ProcessInstance.Assignees) {
			return false, errors.New("user is not an assignee for this instance")
		}
	}

	// 3. Attribute Check
	if ctx.Action.AttributeValidation != nil {
		if !hasRequiredAttributes(ctx.ProcessInstance.Attributes, ctx.Action.AttributeValidation.Attributes) {
			return false, errors.New("instance attributes do not match action requirements")
		}
	}

	return true, nil
}

// hasRequiredRole checks if the user has at least one of the roles required by the action.
func hasRequiredRole(userRoles, requiredRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true // No specific roles required
	}
	userRoleSet := make(map[string]struct{}, len(userRoles))
	for _, role := range userRoles {
		userRoleSet[role] = struct{}{}
	}
	for _, requiredRole := range requiredRoles {
		if _, ok := userRoleSet[requiredRole]; ok {
			return true
		}
	}
	return false
}

// isAssignee checks if the user is in the list of assignees for the process instance.
func isAssignee(userID string, assignees []string) bool {
	for _, assignee := range assignees {
		if assignee == userID {
			return true
		}
	}
	return false
}

// hasRequiredAttributes checks if the process instance has the attributes required by the action.
func hasRequiredAttributes(instanceAttributes, requiredAttributes map[string][]string) bool {
	for key, requiredValues := range requiredAttributes {
		instanceValues, ok := instanceAttributes[key]
		if !ok {
			fmt.Printf("Missing required attribute key: %s\n", key)
			return false // Required attribute key is missing from the instance
		}

		// Check if at least one of the instance's values for this key is in the required list
		foundMatch := false
		for _, requiredVal := range requiredValues {
			for _, instanceVal := range instanceValues {
				if instanceVal == requiredVal {
					foundMatch = true
					break
				}
			}
			if foundMatch {
				break
			}
		}

		if !foundMatch {
			fmt.Printf("Attribute mismatch for key '%s'. Required one of %v, but instance has %v\n", key, requiredValues, instanceValues)
			return false // No matching value found for this attribute key
		}
	}
	return true
}
