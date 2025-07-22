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

	// Transition route
	v3.POST("/transition", transitionHandler.Transition)
}
