package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"template-config/internal/models"
	"template-config/internal/repository"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/oliveagle/jsonpath"
	"gorm.io/gorm"
)

type TemplateConfigService struct {
	repo       *repository.TemplateConfigRepository
	httpClient *resty.Client
}

func NewTemplateConfigService(repo *repository.TemplateConfigRepository) *TemplateConfigService {
	return &TemplateConfigService{
		repo:       repo,
		httpClient: resty.New().SetTimeout(30 * time.Second),
	}
}

func (s *TemplateConfigService) Create(config *models.TemplateConfigDB) error {
	if existing, err := s.repo.GetByTemplateIDAndVersion(config.TemplateID, config.TenantID, config.Version); err != nil && err != gorm.ErrRecordNotFound {
		return err
	} else if existing != nil {
		return fmt.Errorf("template config already exists for templateId: %s, tenantId: %s, version: %s", config.TemplateID, config.TenantID, config.Version)
	}

	now := time.Now().Unix()
	config.ID = uuid.New()
	config.CreatedTime = now
	config.LastModifiedTime = now
	return s.repo.Create(config)
}

func (s *TemplateConfigService) Update(config *models.TemplateConfigDB) error {
	existing, err := s.repo.GetByTemplateIDAndVersion(config.TemplateID, config.TenantID, config.Version)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	config.LastModifiedTime = now
	config.CreatedTime = existing.CreatedTime
	config.CreatedBy = existing.CreatedBy
	return s.repo.Update(config)
}

func (s *TemplateConfigService) Search(search *models.TemplateConfigSearch) ([]models.TemplateConfigDB, error) {
	return s.repo.Search(search)
}

func (s *TemplateConfigService) Delete(templateID, tenantID, version string) error {
	if _, err := s.repo.GetByTemplateIDAndVersion(templateID, tenantID, version); err != nil {
		return err
	}
	return s.repo.Delete(templateID, tenantID, version)
}

func (s *TemplateConfigService) Render(request *models.RenderRequest) (*models.RenderResponse, []models.Error) {
	config, err := s.repo.GetByTemplateIDAndVersion(request.TemplateID, request.TenantID, request.Version)
	if err != nil {
		return nil, []models.Error{{
			Code:        "NOT_FOUND",
			Message:     "Template config not found",
			Description: err.Error(),
		}}
	}

	response := &models.RenderResponse{
		TemplateID: request.TemplateID,
		TenantID:   request.TenantID,
		Version:    request.Version,
		Data:       make(map[string]any),
	}

	for field, jsonPath := range config.FieldMapping {
		if value, err := jsonpath.JsonPathLookup(request.Payload, jsonPath); err == nil {
			response.Data[field] = value
			log.Printf("[FieldMapping] %s: %v", field, value)
		} else {
			log.Printf("[FieldMapping] Failed for %s (%s): %v", field, jsonPath, err)
		}
	}

	if len(config.APIMapping) > 0 {
		if errors := s.executeAPIMappings(config.APIMapping, request.Payload, request.TenantID, response); len(errors) > 0 {
			return nil, errors
		}
	}

	return response, nil
}

func (s *TemplateConfigService) executeAPIMappings(apiMappings []models.APIMapping, payload map[string]interface{}, tenantID string, response *models.RenderResponse) []models.Error {
	var (
		wg        sync.WaitGroup
		errorChan = make(chan models.Error, len(apiMappings))
	)

	for _, mapping := range apiMappings {
		wg.Add(1)
		go func(mapping models.APIMapping) {
			defer wg.Done()
			url := s.buildURL(mapping.Endpoint, payload)
			log.Printf("[APIMapping] Calling: %s", url)

			resp, err := s.httpClient.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-Tenant-ID", tenantID).
				Get(url)

			// Log the raw response body if it exists
			if resp != nil {
				log.Printf("[APIMapping] Response from %s: %s", url, resp.String())
			}

			if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
				var errDesc string
				if err != nil {
					errDesc = err.Error()
				}
				if resp != nil {
					errDesc = fmt.Sprintf("HTTP %d: %s", resp.StatusCode(), resp.String())
				}
				errorChan <- models.Error{
					Code:        "API_CALL_FAILED",
					Message:     "External API call failed",
					Description: errDesc,
					Params:      []string{url, mapping.Method},
				}
				return
			}

			var apiResp interface{}
			if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
				errorChan <- models.Error{
					Code:        "INVALID_JSON",
					Message:     "Failed to parse API response",
					Description: err.Error(),
					Params:      []string{url},
				}
				return
			}

			for field, jsonPath := range mapping.ResponseMapping {
				if value, err := jsonpath.JsonPathLookup(apiResp, jsonPath); err == nil {
					response.Data[field] = value
					log.Printf("[APIResponseMapping] %s: %v", field, value)
				} else {
					log.Printf("[APIResponseMapping] Failed for %s (%s): %v", field, jsonPath, err)
				}
			}
		}(mapping)
	}

	wg.Wait()
	close(errorChan)

	errors := make([]models.Error, 0, len(apiMappings))
	for err := range errorChan {
		errors = append(errors, err)
	}
	return errors
}

func (s *TemplateConfigService) buildURL(endpoint models.EndpointConfig, payload map[string]interface{}) string {
	url := endpoint.Base + endpoint.Path

	for param, path := range endpoint.PathParams {
		if value, err := jsonpath.JsonPathLookup(payload, path); err == nil {
			url = strings.ReplaceAll(url, "{{"+param+"}}", fmt.Sprintf("%v", value))
		}
	}

	if len(endpoint.QueryParams) > 0 {
		var query []string
		for key, path := range endpoint.QueryParams {
			if value, err := jsonpath.JsonPathLookup(payload, path); err == nil {
				query = append(query, fmt.Sprintf("%s=%v", key, value))
			}
		}
		if len(query) > 0 {
			url += "?" + strings.Join(query, "&")
		}
	}
	return url
}
