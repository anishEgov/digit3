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

// BoundaryServiceImpl implements the BoundaryService interface
type BoundaryServiceImpl struct {
	repo  repository.BoundaryRepository
	cache cache.Cache
}

// NewBoundaryService creates a new boundary service
func NewBoundaryService(repo repository.BoundaryRepository, cache cache.Cache) BoundaryService {
	return &BoundaryServiceImpl{
		repo:  repo,
		cache: cache,
	}
}

// Create implements BoundaryService.Create
func (s *BoundaryServiceImpl) Create(ctx context.Context, request *models.BoundaryRequest, tenantID, clientID string) error {
	epoch := time.Now().UnixMilli()
	validator := &validator.BoundaryValidator{Repo: s.repo}
	for i := range request.Boundary {
		request.Boundary[i].TenantID = tenantID
		if err := validator.ValidateBoundary(ctx, &request.Boundary[i]); err != nil {
			return err
		}
	}
	for i := range request.Boundary {
		request.Boundary[i].ID = uuid.New().String()
		request.Boundary[i].AuditDetails = &commonmodels.AuditDetails{
			CreatedBy:        clientID,
			LastModifiedBy:   clientID,
			CreatedTime:      epoch,
			LastModifiedTime: epoch,
		}
	}
	return s.repo.Create(ctx, request)
}

// Search implements BoundaryService.Search
func (s *BoundaryServiceImpl) Search(ctx context.Context, criteria *models.BoundarySearchCriteria) ([]models.Boundary, error) {
	cacheKey := criteria.TenantID + ":boundary:search:" + strings.Join(criteria.Codes, ",")
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		var result []models.Boundary
		if err := json.Unmarshal([]byte(cached.(string)), &result); err == nil {
			return result, nil
		}
	}
	result, err := s.repo.Search(ctx, criteria)
	if err != nil {
		return nil, err
	}
	if b, err := json.Marshal(result); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(b))
	}
	return result, nil
}

// Update implements BoundaryService.Update
func (s *BoundaryServiceImpl) Update(ctx context.Context, request *models.BoundaryRequest, tenantID, clientID string) error {
	epoch := time.Now().UnixMilli()
	var updatedCodes []string
	for i := range request.Boundary {
		request.Boundary[i].TenantID = tenantID
		// Fetch existing record
		existing, err := s.repo.GetByID(ctx, request.Boundary[i].ID, tenantID)
		if err != nil {
			return err
		}
		if existing == nil {
			return fmt.Errorf("boundary with id %s does not exist", request.Boundary[i].ID)
		}
		if request.Boundary[i].AuditDetails == nil {
			request.Boundary[i].AuditDetails = &commonmodels.AuditDetails{}
		}
		// Preserve createdBy and createdTime from DB
		request.Boundary[i].AuditDetails.CreatedBy = existing.AuditDetails.CreatedBy
		request.Boundary[i].AuditDetails.CreatedTime = existing.AuditDetails.CreatedTime
		// Set last modified fields
		request.Boundary[i].AuditDetails.LastModifiedBy = clientID
		request.Boundary[i].AuditDetails.LastModifiedTime = epoch
		// Track codes for cache invalidation
		updatedCodes = append(updatedCodes, request.Boundary[i].Code)
	}
	err := s.repo.Update(ctx, request)
	if err != nil {
		return err
	}
	// Invalidate cache entries for updated boundaries
	for _, code := range updatedCodes {
		cacheKey := tenantID + ":boundary:search:" + code
		_ = s.cache.Delete(ctx, cacheKey)
	}
	// Also invalidate general search cache (codes might be in combined searches)
	cacheKey := tenantID + ":boundary:search:" + strings.Join(updatedCodes, ",")
	_ = s.cache.Delete(ctx, cacheKey)
	return nil
}
