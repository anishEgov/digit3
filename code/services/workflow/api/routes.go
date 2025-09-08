package api

import (
	"digit.org/workflow/api/handlers"
	"github.com/gin-gonic/gin"
)

// RegisterAllRoutes registers all API routes for the service.
func RegisterAllRoutes(
	router *gin.Engine,
	processHandler *handlers.ProcessHandler,
	stateHandler *handlers.StateHandler,
	actionHandler *handlers.ActionHandler,
	transitionHandler *handlers.TransitionHandler,
	escalationConfigHandler *handlers.EscalationConfigHandler,
	autoEscalationHandler *handlers.AutoEscalationHandler,
) {
	v3 := router.Group("/workflow/v3")

	// Process routes
	processGroup := v3.Group("/process")
	{
		processGroup.POST("", processHandler.CreateProcess)
		processGroup.GET("", processHandler.GetProcesses)
		processGroup.GET("/definition", processHandler.GetProcessDefinitions) // New route
		processGroup.GET("/:id", processHandler.GetProcess)
		processGroup.PUT("/:id", processHandler.UpdateProcess)
		processGroup.DELETE("/:id", processHandler.DeleteProcess)

		// Nested State routes
		processGroup.POST("/:id/state", stateHandler.CreateState)
		processGroup.GET("/:id/state", stateHandler.GetStates)

		// Nested Escalation Config routes
		processGroup.POST("/:id/escalation", escalationConfigHandler.CreateEscalationConfig)
		processGroup.GET("/:id/escalation", escalationConfigHandler.GetEscalationConfigs)
	}

	// State routes (for operations on a state by its own ID)
	stateGroup := v3.Group("/state")
	{
		stateGroup.GET("/:id", stateHandler.GetState)
		stateGroup.PUT("/:id", stateHandler.UpdateState)
		stateGroup.DELETE("/:id", stateHandler.DeleteState)

		// Nested Action routes
		stateGroup.POST("/:id/action", actionHandler.CreateAction)
		stateGroup.GET("/:id/action", actionHandler.GetActions)
	}

	// Action routes (for operations on an action by its own ID)
	actionGroup := v3.Group("/action")
	{
		actionGroup.GET("/:id", actionHandler.GetAction)
		actionGroup.PUT("/:id", actionHandler.UpdateAction)
		actionGroup.DELETE("/:id", actionHandler.DeleteAction)
	}

	// Escalation Config routes (for operations on an escalation config by its own ID)
	escalationGroup := v3.Group("/escalation")
	{
		escalationGroup.GET("/:id", escalationConfigHandler.GetEscalationConfig)
		escalationGroup.PUT("/:id", escalationConfigHandler.UpdateEscalationConfig)
		escalationGroup.DELETE("/:id", escalationConfigHandler.DeleteEscalationConfig)
	}

	// Transition routes
	v3.POST("/transition", transitionHandler.Transition)
	v3.GET("/transition", transitionHandler.GetTransitions)

	// Auto-escalation routes
	v3.POST("/auto/:processCode/_escalate", autoEscalationHandler.EscalateApplications)
	v3.GET("/auto/_search", autoEscalationHandler.SearchEscalatedApplications)
}
