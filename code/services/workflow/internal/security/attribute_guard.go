package security

import (
	"fmt"
	"strings"
)

// AttributeGuard provides simple attribute-based validation
// It compares request attributes with action requirements using direct key-value matching
type AttributeGuard struct{}

// NewAttributeGuard creates a new AttributeGuard
func NewAttributeGuard() Guard {
	return &AttributeGuard{}
}

// CanTransition validates if a user can perform a transition based on attribute matching
func (g *AttributeGuard) CanTransition(ctx GuardContext) (bool, error) {
	var validationErrors []string

	// 1. Check assignee validation only if required by the action
	shouldCheckAssignee := false
	if ctx.Action.AttributeValidation != nil && ctx.Action.AttributeValidation.AssigneeCheck {
		shouldCheckAssignee = true
	}

	if shouldCheckAssignee && !isAssignee(ctx.UserID, ctx.ProcessInstance.Assignees) {
		validationErrors = append(validationErrors, "user is not assigned to this process instance")
	}

	// 2. Check attribute validation - compare request attributes with action requirements
	if ctx.Action.AttributeValidation != nil && len(ctx.Action.AttributeValidation.Attributes) > 0 {
		if valid, err := g.validateAttributes(ctx.ProcessInstance.Attributes, ctx.Action.AttributeValidation.Attributes); !valid {
			validationErrors = append(validationErrors, fmt.Sprintf("attribute validation failed: %s", err.Error()))
		}
	}

	// 3. Return combined error if any validation failed
	if len(validationErrors) > 0 {
		return false, fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return true, nil
}

// validateAttributes compares request attributes with action requirements
// requestAttrs: attributes from the transition request
// requiredAttrs: attributes required by the action
func (g *AttributeGuard) validateAttributes(requestAttrs, requiredAttrs map[string][]string) (bool, error) {
	var missingKeys []string
	var mismatchedValues []string

	// Iterate through all required attributes from the action
	for requiredKey, requiredValues := range requiredAttrs {
		// Check if the request contains this required key
		requestValues, keyExists := requestAttrs[requiredKey]
		if !keyExists || len(requestValues) == 0 {
			missingKeys = append(missingKeys, requiredKey)
			continue
		}

		// Check if at least one request value matches one of the required values
		hasMatch := false
		for _, reqVal := range requestValues {
			for _, requiredVal := range requiredValues {
				if strings.TrimSpace(reqVal) == strings.TrimSpace(requiredVal) {
					hasMatch = true
					break
				}
			}
			if hasMatch {
				break
			}
		}

		if !hasMatch {
			mismatchedValues = append(mismatchedValues, fmt.Sprintf("key '%s': request has %v, but requires one of %v", requiredKey, requestValues, requiredValues))
		}
	}

	// Build error message
	var errorParts []string
	if len(missingKeys) > 0 {
		errorParts = append(errorParts, fmt.Sprintf("missing required attributes: %v", missingKeys))
	}
	if len(mismatchedValues) > 0 {
		errorParts = append(errorParts, fmt.Sprintf("attribute mismatches: [%s]", strings.Join(mismatchedValues, ", ")))
	}

	if len(errorParts) > 0 {
		return false, fmt.Errorf(strings.Join(errorParts, "; "))
	}

	return true, nil
}
