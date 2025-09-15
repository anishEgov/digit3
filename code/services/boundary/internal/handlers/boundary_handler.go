package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	commonmodels "boundary/internal/common/models"
	"boundary/internal/models"
	"boundary/internal/service"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// BoundaryHandler handles HTTP requests for boundary operations
type BoundaryHandler struct {
	boundaryService     service.BoundaryService
	hierarchyService    service.BoundaryHierarchyService
	relationshipService service.BoundaryRelationshipService
}

// NewBoundaryHandler creates a new boundary handler
func NewBoundaryHandler(
	boundaryService service.BoundaryService,
	hierarchyService service.BoundaryHierarchyService,
	relationshipService service.BoundaryRelationshipService,
) *BoundaryHandler {
	return &BoundaryHandler{
		boundaryService:     boundaryService,
		hierarchyService:    hierarchyService,
		relationshipService: relationshipService,
	}
}

// RegisterRoutes registers the boundary routes
func (h *BoundaryHandler) RegisterRoutes(router *gin.RouterGroup) {
	boundary := router.Group("/v1/boundary")
	{
		// Boundary endpoints
		boundary.POST("", h.Create)
		boundary.GET("", h.Search)
		boundary.PUT("", h.Update)
	}

	// Hierarchy definition endpoints
	hierarchy := router.Group("/v1/boundary-hierarchy-definition")
	{
		hierarchy.POST("", h.CreateHierarchy)
		hierarchy.GET("", h.GetHierarchy)
	}

	// Relationship endpoints
	relationship := router.Group("/v1/boundary-relationships")
	{
		relationship.POST("", h.CreateRelationship)
		relationship.GET("", h.GetRelationship)
		relationship.PUT("", h.UpdateRelationship)
	}

	// Shapefile boundary create endpoint
	router.POST("/shapefile/v1/boundary/create", ShapefileBoundaryCreateHandler(h.boundaryService))
}

// Create handles the creation of new boundaries (batch)
func (h *BoundaryHandler) Create(ctx *gin.Context) {
	tracer := otel.Tracer("boundary-handler")
	spanCtx, span := tracer.Start(ctx.Request.Context(), "boundary.create")
	defer span.End()

	var request models.BoundaryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid request payload")
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "Invalid request payload", nil)
		return
	}

	tenantID := ctx.GetHeader("X-Tenant-ID")
	clientID := ctx.GetHeader("X-Client-Id")
	if tenantID == "" || clientID == "" {
		span.SetStatus(codes.Error, "Missing headers")
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Missing X-Tenant-ID or X-Client-Id header", "Invalid request payload", nil)
		return
	}

	span.SetAttributes(
		attribute.String("tenant.id", tenantID),
		attribute.String("client.id", clientID),
		attribute.Int("boundary.count", len(request.Boundary)),
	)

	if len(request.Boundary) == 0 {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "boundary array must not be empty", "Invalid request payload", nil)
		return
	}
	for i := range request.Boundary {
		request.Boundary[i].ID = ""
		request.Boundary[i].TenantID = tenantID
		if strings.TrimSpace(request.Boundary[i].Code) == "" {
			errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "code is required for each boundary", "Invalid request payload", nil)
			return
		}
		// Geometry type validation
		var geom map[string]interface{}
		if err := json.Unmarshal(request.Boundary[i].Geometry, &geom); err != nil {
			errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Invalid geometry JSON", "Invalid geometry JSON", nil)
			return
		}
		geomType, ok := geom["type"].(string)
		if !ok || !models.IsValidGeometryType(geomType) {
			errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Invalid geometry type", "Allowed types: Point, Polygon, MultiPolygon", nil)
			return
		}
	}

	if err := h.boundaryService.Create(spanCtx, &request, tenantID, clientID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create boundaries")
		errorResponse(ctx, http.StatusBadRequest, "INTERNAL_SERVER_ERROR", err.Error(), "Failed to create boundaries", nil)
		return
	}

	span.SetStatus(codes.Ok, "Boundaries created successfully")
	span.AddEvent("boundaries.created", trace.WithAttributes(
		attribute.Int("boundary.count", len(request.Boundary)),
	))

	response := models.BoundaryResponse{
		ResponseInfo: &commonmodels.ResponseInfo{
			APIId:    "boundary",
			Ver:      "1.0",
			Ts:       time.Now().UnixMilli(),
			ResMsgId: "",
			MsgId:    "",
			Status:   "successful",
		},
		Boundary: request.Boundary,
	}

	ctx.JSON(http.StatusCreated, response)
}

// Search handles the search for boundaries
func (h *BoundaryHandler) Search(ctx *gin.Context) {
	tracer := otel.Tracer("boundary-handler")
	spanCtx, span := tracer.Start(ctx.Request.Context(), "boundary.search")
	defer span.End()

	criteria := &models.BoundarySearchCriteria{}
	tenantID := ctx.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		span.SetStatus(codes.Error, "Missing tenant ID")
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Missing X-Tenant-ID header", "Invalid request payload", nil)
		return
	}
	criteria.TenantID = tenantID

	span.SetAttributes(
		attribute.String("tenant.id", tenantID),
	)

	if codes := ctx.QueryArray("codes"); len(codes) > 0 {
		criteria.Codes = codes
	} else {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Missing required query parameter: codes", "Invalid request payload", nil)
		return
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Invalid limit parameter", "Invalid request payload", nil)
			return
		}
		criteria.Limit = limit
	}

	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Invalid offset parameter", "Invalid request payload", nil)
			return
		}
		criteria.Offset = offset
	}

	span.SetAttributes(
		attribute.StringSlice("boundary.codes", criteria.Codes),
		attribute.Int("boundary.limit", criteria.Limit),
		attribute.Int("boundary.offset", criteria.Offset),
	)

	boundaries, err := h.boundaryService.Search(spanCtx, criteria)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to search boundaries")
		errorResponse(ctx, http.StatusBadRequest, "INTERNAL_SERVER_ERROR", err.Error(), "Failed to search boundaries", nil)
		return
	}

	span.SetStatus(codes.Ok, "Boundaries searched successfully")
	span.AddEvent("boundaries.found", trace.WithAttributes(
		attribute.Int("boundary.count", len(boundaries)),
	))

	response := models.BoundaryResponse{
		ResponseInfo: &commonmodels.ResponseInfo{
			APIId:    "boundary",
			Ver:      "1.0",
			Ts:       time.Now().UnixMilli(),
			ResMsgId: "",
			MsgId:    "",
			Status:   "successful",
		},
		Boundary: boundaries,
	}

	ctx.JSON(http.StatusOK, response)
}

// Update handles the update of existing boundaries (batch)
func (h *BoundaryHandler) Update(ctx *gin.Context) {
	var request models.BoundaryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "Invalid request payload", nil)
		return
	}

	tenantID := ctx.GetHeader("X-Tenant-ID")
	clientID := ctx.GetHeader("X-Client-Id")
	if tenantID == "" || clientID == "" {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Missing X-Tenant-ID or X-Client-Id header", "Invalid request payload", nil)
		return
	}

	if len(request.Boundary) == 0 {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "boundary array must not be empty", "Invalid request payload", nil)
		return
	}
	for i := range request.Boundary {
		if request.Boundary[i].ID == "" {
			errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "ID is required for update", "Invalid request payload", nil)
			return
		}
		request.Boundary[i].TenantID = tenantID
		// Geometry type validation
		var geom map[string]interface{}
		if err := json.Unmarshal(request.Boundary[i].Geometry, &geom); err != nil {
			errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Invalid geometry JSON", "Invalid geometry JSON", nil)
			return
		}
		geomType, ok := geom["type"].(string)
		if !ok || !models.IsValidGeometryType(geomType) {
			errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Invalid geometry type", "Allowed types: Point, Polygon, MultiPolygon", nil)
			return
		}
	}

	if err := h.boundaryService.Update(ctx.Request.Context(), &request, tenantID, clientID); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			errorResponse(ctx, http.StatusNotFound, "NOT_FOUND", err.Error(), "Boundary not found", nil)
			return
		}
		errorResponse(ctx, http.StatusBadRequest, "INTERNAL_SERVER_ERROR", err.Error(), "Failed to update boundaries", nil)
		return
	}

	response := models.BoundaryResponse{
		ResponseInfo: &commonmodels.ResponseInfo{
			APIId:    "boundary",
			Ver:      "1.0",
			Ts:       time.Now().UnixMilli(),
			ResMsgId: "",
			MsgId:    "",
			Status:   "successful",
		},
		Boundary: request.Boundary,
	}

	ctx.JSON(http.StatusOK, response)
}

// CreateHierarchy handles the creation of a new boundary hierarchy
func (h *BoundaryHandler) CreateHierarchy(ctx *gin.Context) {
	var request models.BoundaryHierarchyRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "Invalid request payload", nil)
		return
	}

	tenantID := ctx.GetHeader("X-Tenant-ID")
	clientID := ctx.GetHeader("X-Client-Id")
	if tenantID == "" || clientID == "" {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Missing X-Tenant-ID or X-Client-Id header", "Invalid request payload", nil)
		return
	}

	request.BoundaryHierarchy.TenantID = tenantID

	if err := h.hierarchyService.Create(ctx.Request.Context(), &request, tenantID, clientID); err != nil {
		errorResponse(ctx, http.StatusBadRequest, "INTERNAL_SERVER_ERROR", err.Error(), "Failed to create hierarchy", nil)
		return
	}

	response := models.BoundaryTypeHierarchyResponse{
		ResponseInfo: &commonmodels.ResponseInfo{
			APIId:    "boundary",
			Ver:      "1.0",
			Ts:       time.Now().UnixMilli(),
			ResMsgId: "",
			MsgId:    "",
			Status:   "successful",
		},
		Hierarchy: []models.BoundaryHierarchy{request.BoundaryHierarchy},
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetHierarchy handles the retrieval of a boundary hierarchy
func (h *BoundaryHandler) GetHierarchy(ctx *gin.Context) {
	tenantID := ctx.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Missing X-Tenant-ID header", "Invalid request payload", nil)
		return
	}
	hierarchyType := ctx.Query("hierarchyType")
	if hierarchyType == "" {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Missing hierarchyType query parameter", "Invalid request payload", nil)
		return
	}
	criteria := &models.BoundaryHierarchySearchCriteria{
		TenantID:      tenantID,
		HierarchyType: hierarchyType,
	}

	hierarchies, err := h.hierarchyService.Search(ctx.Request.Context(), criteria)
	if err != nil {
		errorResponse(ctx, http.StatusBadRequest, "INTERNAL_SERVER_ERROR", err.Error(), "Failed to get hierarchy", nil)
		return
	}

	response := models.BoundaryTypeHierarchyResponse{
		ResponseInfo: &commonmodels.ResponseInfo{
			APIId:    "boundary",
			Ver:      "1.0",
			Ts:       time.Now().UnixMilli(),
			ResMsgId: "",
			MsgId:    "",
			Status:   "successful",
		},
		Hierarchy: hierarchies,
	}

	ctx.JSON(http.StatusOK, response)
}

// CreateRelationship handles the creation of a new boundary relationship
func (h *BoundaryHandler) CreateRelationship(ctx *gin.Context) {
	var request models.BoundaryRelationshipRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "Invalid request payload", nil)
		return
	}

	tenantID := ctx.GetHeader("X-Tenant-ID")
	clientID := ctx.GetHeader("X-Client-Id")
	if tenantID == "" || clientID == "" {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Missing X-Tenant-ID or X-Client-Id header", "Invalid request payload", nil)
		return
	}

	request.BoundaryRelationship.TenantID = tenantID

	if err := h.relationshipService.Create(ctx.Request.Context(), &request, tenantID, clientID); err != nil {
		errorResponse(ctx, http.StatusBadRequest, "INTERNAL_SERVER_ERROR", err.Error(), "Failed to create relationship", nil)
		return
	}

	response := models.BoundaryRelationshipResponse{
		ResponseInfo: &commonmodels.ResponseInfo{
			APIId:    "boundary",
			Ver:      "1.0",
			Ts:       time.Now().UnixMilli(),
			ResMsgId: "",
			MsgId:    "",
			Status:   "successful",
		},
		Relationship: []models.BoundaryRelationship{request.BoundaryRelationship},
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetRelationship handles the retrieval of boundary relationships
func (h *BoundaryHandler) GetRelationship(ctx *gin.Context) {
	criteria := &models.BoundaryRelationshipSearchCriteria{}
	tenantID := ctx.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Missing X-Tenant-ID header", "Invalid request payload", nil)
		return
	}
	criteria.TenantID = tenantID

	criteria.HierarchyType = ctx.Query("hierarchyType")
	criteria.BoundaryType = ctx.Query("boundaryType")
	if codes := ctx.QueryArray("codes"); len(codes) > 0 {
		criteria.Codes = codes
	}
	if parent := ctx.Query("parent"); parent != "" {
		criteria.Parent = parent
	}
	if ctx.Query("includeChildren") == "true" {
		criteria.IncludeChildren = true
	}
	if ctx.Query("includeParents") == "true" {
		criteria.IncludeParents = true
	}
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			criteria.Limit = limit
		}
	}
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			criteria.Offset = offset
		}
	}

	response, err := h.relationshipService.Search(ctx.Request.Context(), criteria)
	if err != nil {
		errorResponse(ctx, http.StatusBadRequest, "INTERNAL_SERVER_ERROR", err.Error(), "Failed to get relationships", nil)
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// UpdateRelationship handles the update of an existing boundary relationship
func (h *BoundaryHandler) UpdateRelationship(ctx *gin.Context) {
	var request models.BoundaryRelationshipRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", err.Error(), "Invalid request payload", nil)
		return
	}

	tenantID := ctx.GetHeader("X-Tenant-ID")
	clientID := ctx.GetHeader("X-Client-Id")
	if tenantID == "" || clientID == "" {
		errorResponse(ctx, http.StatusBadRequest, "BAD_REQUEST", "Missing X-Tenant-ID or X-Client-Id header", "Invalid request payload", nil)
		return
	}

	request.BoundaryRelationship.TenantID = tenantID

	if err := h.relationshipService.Update(ctx.Request.Context(), &request, tenantID, clientID); err != nil {
		errorResponse(ctx, http.StatusBadRequest, "INTERNAL_SERVER_ERROR", err.Error(), "Failed to update relationship", nil)
		return
	}

	response := models.BoundaryRelationshipResponse{
		ResponseInfo: &commonmodels.ResponseInfo{
			APIId:    "boundary",
			Ver:      "1.0",
			Ts:       time.Now().UnixMilli(),
			ResMsgId: "",
			MsgId:    "",
			Status:   "successful",
		},
		Relationship: []models.BoundaryRelationship{request.BoundaryRelationship},
	}

	ctx.JSON(http.StatusOK, response)
}

func errorResponse(ctx *gin.Context, status int, code, message, description string, params []string) {
	ctx.JSON(status, commonmodels.ErrorResponse{
		ResponseInfo: commonmodels.ResponseInfo{
			APIId:    "boundary",
			Ver:      "1.0",
			Ts:       time.Now().UnixMilli(),
			ResMsgId: "",
			MsgId:    "",
			Status:   "FAILED",
		},
		Errors: []commonmodels.Error{{
			Code:        code,
			Message:     message,
			Description: description,
			Params:      params,
		}},
	})
}
