// package main

// import (
// 	"log"

// 	"github.com/gin-gonic/gin"
// )

// func main() {
// 	// Initialize router
// 	router := gin.Default()

// 	// Initialize routes
// 	initializeRoutes(router)

// 	// Start server
// 	if err := router.Run(":8081"); err != nil {
// 		log.Fatalf("Failed to start server: %v", err)
// 	}
// }

// func initializeRoutes(router *gin.Engine) {
// 	// Group all boundary related routes
// 	boundaryGroup := router.Group("/boundary")
// 	{
// 		boundaryGroup.POST("", createBoundary)
// 		boundaryGroup.GET("", searchBoundary)
// 		boundaryGroup.PUT("", updateBoundary)
// 	}

// 	// Group all boundary relationship related routes
// 	relationshipGroup := router.Group("/boundary-relationships")
// 	{
// 		relationshipGroup.POST("", createBoundaryRelationship)
// 		relationshipGroup.GET("", searchBoundaryRelationship)
// 		relationshipGroup.PUT("", updateBoundaryRelationship)
// 	}

// 	// Group all hierarchy definition related routes
// 	hierarchyGroup := router.Group("/boundary-hierarchy-definition")
// 	{
// 		hierarchyGroup.POST("", createHierarchyDefinition)
// 		hierarchyGroup.GET("", searchHierarchyDefinition)
// 	}
// } 