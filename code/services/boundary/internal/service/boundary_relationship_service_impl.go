package service

import (
	commonmodels "boundary/internal/common/models"
	"boundary/internal/models"
	"boundary/internal/repository"
	"boundary/internal/validator"
	"boundary/pkg/cache"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BoundaryRelationshipServiceImpl implements BoundaryRelationshipService
type BoundaryRelationshipServiceImpl struct {
	repo  repository.BoundaryRelationshipRepository
	cache cache.Cache
}

// NewBoundaryRelationshipService creates a new boundary relationship service
func NewBoundaryRelationshipService(repo repository.BoundaryRelationshipRepository, cache cache.Cache) BoundaryRelationshipService {
	return &BoundaryRelationshipServiceImpl{
		repo:  repo,
		cache: cache,
	}
}

// Create implements BoundaryRelationshipService.Create
func (s *BoundaryRelationshipServiceImpl) Create(ctx context.Context, request *models.BoundaryRelationshipRequest, tenantID, clientID string) error {
	validator := &validator.BoundaryRelationshipValidator{Repo: s.repo}
	request.BoundaryRelationship.TenantID = tenantID
	if err := validator.ValidateRelationship(ctx, &request.BoundaryRelationship); err != nil {
		return err
	}
	epoch := time.Now().UnixMilli()
	request.BoundaryRelationship.ID = uuid.New().String()
	request.BoundaryRelationship.AuditDetails = &commonmodels.AuditDetails{
		CreatedBy:        clientID,
		LastModifiedBy:   clientID,
		CreatedTime:      epoch,
		LastModifiedTime: epoch,
	}
	return s.repo.Create(ctx, request)
}

// Search implements BoundaryRelationshipService.Search
func (s *BoundaryRelationshipServiceImpl) Search(ctx context.Context, criteria *models.BoundaryRelationshipSearchCriteria) (*models.BoundarySearchResponse, error) {
	cacheKey := criteria.TenantID + ":relationship:search:" + strings.Join(criteria.Codes, ",")
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		var result models.BoundarySearchResponse
		if err := json.Unmarshal([]byte(cached.(string)), &result); err == nil {
			return &result, nil
		}
	}

	// Use materialized path for efficient search with Java-like logic
	relationships, err := s.searchWithInternalParameters(ctx, criteria)
	if err != nil {
		return nil, err
	}

	// Convert to hierarchical structure using materialized path
	hierarchy := s.buildHierarchyFromMaterializedPath(relationships, criteria)

	response := &models.BoundarySearchResponse{
		ResponseInfo: &commonmodels.ResponseInfo{
			APIId:    "boundary",
			Ver:      "1.0",
			Ts:       time.Now().UnixMilli(),
			ResMsgId: "",
			MsgId:    "",
			Status:   "successful",
		},
		TenantBoundary: []models.HierarchyRelation{hierarchy},
	}
	if b, err := json.Marshal(response); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(b))
	}
	return response, nil
}

// Update implements BoundaryRelationshipService.Update
func (s *BoundaryRelationshipServiceImpl) Update(ctx context.Context, request *models.BoundaryRelationshipRequest, tenantID, clientID string) error {
	epoch := time.Now().UnixMilli()
	request.BoundaryRelationship.TenantID = tenantID
	// Fetch existing record
	existing, err := s.repo.GetByID(ctx, request.BoundaryRelationship.ID, tenantID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("boundary relationship with id %s does not exist", request.BoundaryRelationship.ID)
	}
	if request.BoundaryRelationship.AuditDetails == nil {
		request.BoundaryRelationship.AuditDetails = &commonmodels.AuditDetails{}
	}
	// Preserve createdBy and createdTime from DB
	request.BoundaryRelationship.AuditDetails.CreatedBy = existing.AuditDetails.CreatedBy
	request.BoundaryRelationship.AuditDetails.CreatedTime = existing.AuditDetails.CreatedTime
	// Set last modified fields
	request.BoundaryRelationship.AuditDetails.LastModifiedBy = clientID
	request.BoundaryRelationship.AuditDetails.LastModifiedTime = epoch
	err = s.repo.Update(ctx, request)
	if err != nil {
		return err
	}
	// Invalidate cache entries for this relationship
	cacheKey := tenantID + ":relationship:search:" + request.BoundaryRelationship.Code
	_ = s.cache.Delete(ctx, cacheKey)
	// Also invalidate any cached searches that might include this relationship
	// This is a simplified approach - in production you might want pattern-based cache invalidation
	for _, code := range []string{request.BoundaryRelationship.Code, request.BoundaryRelationship.Parent} {
		if code != "" {
			cacheKeyPattern := tenantID + ":relationship:search:" + code
			_ = s.cache.Delete(ctx, cacheKeyPattern)
		}
	}

	// Invalidate cache for any searches that might include children of this boundary
	childrenCacheKey := tenantID + ":relationship:children:" + request.BoundaryRelationship.Code
	_ = s.cache.Delete(ctx, childrenCacheKey)
	return nil
}

// buildHierarchyFromMaterializedPath builds hierarchical structure from flat list using materialized paths
func (s *BoundaryRelationshipServiceImpl) buildHierarchyFromMaterializedPath(relationships []models.BoundaryRelationship, criteria *models.BoundaryRelationshipSearchCriteria) models.HierarchyRelation {
	// Create boundary map for quick lookup
	boundaryMap := make(map[string]*models.EnrichedBoundary)
	for _, r := range relationships {
		boundary := &models.EnrichedBoundary{
			ID:           r.ID,
			Code:         r.Code,
			BoundaryType: r.BoundaryType,
			AuditDetails: r.AuditDetails,
			Parent:       r.Parent,
			Children:     []models.EnrichedBoundary{},
		}
		boundaryMap[r.Code] = boundary
	}

	// Build parent-child relationships
	for _, r := range relationships {
		if r.Parent != "" {
			if parent, exists := boundaryMap[r.Parent]; exists {
				if child, childExists := boundaryMap[r.Code]; childExists {
					parent.Children = append(parent.Children, *child)
				}
			}
		}
	}

	// Find root nodes (nodes without parents in the result set)
	var rootBoundaries []models.EnrichedBoundary
	for _, r := range relationships {
		if r.Parent == "" || boundaryMap[r.Parent] == nil {
			if boundary, exists := boundaryMap[r.Code]; exists {
				rootBoundaries = append(rootBoundaries, *boundary)
			}
		}
	}

	// If no specific filtering and we have includeParents/includeChildren,
	// we might need to organize differently
	if len(criteria.Codes) > 0 && (criteria.IncludeParents || criteria.IncludeChildren) {
		// For parent/children inclusion, we want to show the full hierarchy
		// starting from the actual root nodes
		rootBoundaries = s.findActualRoots(relationships, boundaryMap)
	}

	return models.HierarchyRelation{
		TenantID:      criteria.TenantID,
		HierarchyType: criteria.HierarchyType,
		Boundary:      rootBoundaries,
	}
}

// findActualRoots finds the actual root nodes from the materialized paths
func (s *BoundaryRelationshipServiceImpl) findActualRoots(relationships []models.BoundaryRelationship, boundaryMap map[string]*models.EnrichedBoundary) []models.EnrichedBoundary {
	rootCodes := make(map[string]bool)

	// Extract all root codes from materialized paths
	for _, r := range relationships {
		if r.AncestralMaterializedPath != "" {
			parts := strings.Split(r.AncestralMaterializedPath, "|")
			if len(parts) > 0 && parts[0] != "" {
				rootCodes[parts[0]] = true
			}
		}
	}

	var roots []models.EnrichedBoundary
	for rootCode := range rootCodes {
		if boundary, exists := boundaryMap[rootCode]; exists {
			roots = append(roots, *boundary)
		}
	}

	// If no roots found, return boundaries without parents
	if len(roots) == 0 {
		for _, r := range relationships {
			if r.Parent == "" {
				if boundary, exists := boundaryMap[r.Code]; exists {
					roots = append(roots, *boundary)
				}
			}
		}
	}

	return roots
}

// searchWithInternalParameters implements Java-like search logic using internal parameters
func (s *BoundaryRelationshipServiceImpl) searchWithInternalParameters(ctx context.Context, criteria *models.BoundaryRelationshipSearchCriteria) ([]models.BoundaryRelationship, error) {
	// Step 1: Get main boundaries based on search criteria
	mainBoundaries, err := s.repo.SearchWithMaterializedPath(ctx, criteria)
	if err != nil {
		return nil, err
	}

	// Step 2: Get parent boundaries if includeParents flag is set (Java-like logic)
	parentBoundaries, err := s.getParentBoundaries(ctx, mainBoundaries, criteria)
	if err != nil {
		return nil, err
	}

	// Step 3: Get children boundaries if includeChildren flag is set (Java-like logic)
	childrenBoundaries, err := s.getChildrenBoundaries(ctx, mainBoundaries, criteria)
	if err != nil {
		return nil, err
	}

	// Step 4: Combine all boundaries and remove duplicates
	allBoundaries := s.combineBoundaries(mainBoundaries, parentBoundaries, childrenBoundaries)

	return allBoundaries, nil
}

// getParentBoundaries fetches parent boundaries using materialized path (Java-like implementation)
func (s *BoundaryRelationshipServiceImpl) getParentBoundaries(ctx context.Context, boundaries []models.BoundaryRelationship, criteria *models.BoundaryRelationshipSearchCriteria) ([]models.BoundaryRelationship, error) {
	if !criteria.IncludeParents || len(boundaries) == 0 {
		return []models.BoundaryRelationship{}, nil
	}

	// Extract all ancestor codes from materialized paths
	ancestorCodes := make(map[string]bool)
	for _, boundary := range boundaries {
		if boundary.AncestralMaterializedPath != "" {
			parts := strings.Split(boundary.AncestralMaterializedPath, "|")
			for _, part := range parts {
				if part != "" && part != boundary.Code {
					ancestorCodes[part] = true
				}
			}
		}
	}

	if len(ancestorCodes) == 0 {
		return []models.BoundaryRelationship{}, nil
	}

	// Convert map keys to slice
	codes := make([]string, 0, len(ancestorCodes))
	for code := range ancestorCodes {
		codes = append(codes, code)
	}

	// Search for parent boundaries
	parentCriteria := &models.BoundaryRelationshipSearchCriteria{
		TenantID:      criteria.TenantID,
		HierarchyType: criteria.HierarchyType,
		Codes:         codes,
	}

	return s.repo.SearchWithMaterializedPath(ctx, parentCriteria)
}

// getChildrenBoundaries fetches children boundaries using currentBoundaryCodes (Java-like implementation)
func (s *BoundaryRelationshipServiceImpl) getChildrenBoundaries(ctx context.Context, boundaries []models.BoundaryRelationship, criteria *models.BoundaryRelationshipSearchCriteria) ([]models.BoundaryRelationship, error) {
	if !criteria.IncludeChildren || len(boundaries) == 0 {
		return []models.BoundaryRelationship{}, nil
	}

	// Extract current boundary codes
	currentBoundaryCodes := make([]string, len(boundaries))
	for i, boundary := range boundaries {
		currentBoundaryCodes[i] = boundary.Code
	}

	// Search for children using currentBoundaryCodes (Java-like logic)
	childrenCriteria := &models.BoundaryRelationshipSearchCriteria{
		TenantID:             criteria.TenantID,
		HierarchyType:        criteria.HierarchyType,
		CurrentBoundaryCodes: currentBoundaryCodes,
	}

	return s.repo.SearchWithMaterializedPath(ctx, childrenCriteria)
}

// combineBoundaries combines main, parent, and children boundaries, removing duplicates
func (s *BoundaryRelationshipServiceImpl) combineBoundaries(main, parents, children []models.BoundaryRelationship) []models.BoundaryRelationship {
	boundaryMap := make(map[string]models.BoundaryRelationship)

	// Add main boundaries
	for _, boundary := range main {
		boundaryMap[boundary.Code] = boundary
	}

	// Add parent boundaries
	for _, boundary := range parents {
		boundaryMap[boundary.Code] = boundary
	}

	// Add children boundaries
	for _, boundary := range children {
		boundaryMap[boundary.Code] = boundary
	}

	// Convert map to slice
	result := make([]models.BoundaryRelationship, 0, len(boundaryMap))
	for _, boundary := range boundaryMap {
		result = append(result, boundary)
	}

	return result
}

// GetRootBoundaries searches for root boundaries using isSearchForRootNode flag
func (s *BoundaryRelationshipServiceImpl) GetRootBoundaries(ctx context.Context, tenantID, hierarchyType string) ([]models.BoundaryRelationship, error) {
	criteria := &models.BoundaryRelationshipSearchCriteria{
		TenantID:            tenantID,
		HierarchyType:       hierarchyType,
		IsSearchForRootNode: true,
	}

	return s.repo.SearchWithMaterializedPath(ctx, criteria)
}
