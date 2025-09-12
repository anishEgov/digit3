package service

import (
	"context"
	"fmt"
	"log"
	"notification/internal/config"
	"notification/internal/models"
	"notification/internal/repository"
	"notification/internal/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TemplateService struct {
	repo             *repository.TemplateRepository
	enrichmentClient *utils.EnrichmentClient
	templateRenderer *TemplateRenderer
	config           *config.Config
}

func NewTemplateService(repo *repository.TemplateRepository, cfg *config.Config) *TemplateService {
	// Create enrichment client with base path and endpoint from config
	enrichmentClient := utils.NewEnrichmentClient(cfg.TemplateConfigHost, cfg.TemplateConfigPath)

	return &TemplateService{
		repo:             repo,
		enrichmentClient: enrichmentClient,
		templateRenderer: NewTemplateRenderer(),
		config:           cfg,
	}
}

func (s *TemplateService) Create(templateDB *models.TemplateDB) error {
	if existing, err := s.repo.GetByTemplateIDAndVersion(templateDB.TemplateID, templateDB.TenantID, templateDB.Version); err != nil && err != gorm.ErrRecordNotFound {
		return err
	} else if existing != nil {
		return fmt.Errorf("template already exists for templateId: %s, tenantId: %s, version: %s", templateDB.TemplateID, templateDB.TenantID, templateDB.Version)
	}

	now := time.Now().Unix()
	templateDB.ID = uuid.New()
	templateDB.CreatedTime = now
	templateDB.LastModifiedTime = now
	return s.repo.Create(templateDB)
}

func (s *TemplateService) Update(templateDB *models.TemplateDB) error {
	existing, err := s.repo.GetByTemplateIDAndVersion(templateDB.TemplateID, templateDB.TenantID, templateDB.Version)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	templateDB.LastModifiedTime = now
	templateDB.CreatedTime = existing.CreatedTime
	templateDB.CreatedBy = existing.CreatedBy
	return s.repo.Update(templateDB)
}

func (s *TemplateService) Search(searchReq *models.TemplateSearch) ([]models.TemplateDB, error) {
	return s.repo.Search(searchReq)
}

func (s *TemplateService) Delete(deleteReq *models.TemplateDelete) error {
	if _, err := s.repo.GetByTemplateIDAndVersion(deleteReq.TemplateID, deleteReq.TenantID, deleteReq.Version); err != nil {
		return err
	}
	return s.repo.Delete(deleteReq.TemplateID, deleteReq.TenantID, deleteReq.Version)
}

func (s *TemplateService) Preview(request *models.TemplatePreviewRequest) (*models.TemplatePreviewResponse, []models.Error) {
	// Step 1: Template Fetching
	templateDB, err := s.repo.GetByTemplateIDAndVersion(request.TemplateID, request.TenantID, request.Version)
	if err != nil {
		return nil, []models.Error{{
			Code:        "NOT_FOUND",
			Message:     "Template not found",
			Description: err.Error(),
		}}
	}

	// Step 2: Payload Enrichment (Conditional)
	var templateData map[string]interface{}
	if request.Enrich {
		// Create context for enrichment
		ctx := context.Background()

		// Enrich payload using the template-config service
		enrichedData, err := s.enrichmentClient.EnrichPayload(ctx, request.TemplateID, request.TenantID, request.Version, request.Payload)
		if err != nil {
			log.Printf("Failed to enrich payload: %v", err)
			return nil, []models.Error{{
				Code:        "ENRICHMENT_FAILED",
				Message:     "Failed to enrich payload",
				Description: err.Error(),
			}}
		}
		templateData = enrichedData
	} else {
		// Use raw payload for template rendering
		templateData = request.Payload
	}

	// Step 3: Template Rendering
	var renderedSubject, renderedContent string
	var errors []models.Error

	// Render subject if it exists
	if templateDB.Type == string(models.TemplateTypeEmail) && templateDB.Subject != "" {
		renderedSubject, err = s.templateRenderer.RenderSubject(templateDB.Subject, templateData)
		if err != nil {
			errors = append(errors, models.Error{
				Code:        "TEMPLATE_RENDER_ERROR",
				Message:     "Failed to render subject template",
				Description: err.Error(),
			})
		}
	}

	// Render content
	renderedContent, err = s.templateRenderer.RenderContent(templateDB.Content, templateData, templateDB.IsHTML)
	if err != nil {
		errors = append(errors, models.Error{
			Code:        "TEMPLATE_RENDER_ERROR",
			Message:     "Failed to render content template",
			Description: err.Error(),
		})
	}

	// If there were rendering errors, return them
	if len(errors) > 0 {
		return nil, errors
	}

	response := &models.TemplatePreviewResponse{
		Type:            models.TemplateType(templateDB.Type),
		IsHTML:          templateDB.IsHTML,
		RenderedSubject: renderedSubject,
		RenderedContent: renderedContent,
	}

	return response, nil
}
