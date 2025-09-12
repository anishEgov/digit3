package repository

import (
	"notification/internal/models"

	"gorm.io/gorm"
)

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Create(template *models.TemplateDB) error {
	return r.db.Create(template).Error
}

func (r *TemplateRepository) Update(template *models.TemplateDB) error {
	return r.db.Save(template).Error
}

func (r *TemplateRepository) GetByTemplateIDAndVersion(templateID, tenantID, version string) (*models.TemplateDB, error) {
	var template models.TemplateDB
	err := r.db.Where("tenantid = ? AND templateid = ? AND version = ?", tenantID, templateID, version).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) Search(searchReq *models.TemplateSearch) ([]models.TemplateDB, error) {
	var templates []models.TemplateDB
	query := r.db.Model(&models.TemplateDB{})

	if searchReq.TenantID != "" {
		query = query.Where("tenantid = ?", searchReq.TenantID)
	}

	if searchReq.TemplateID != "" {
		query = query.Where("templateid = ?", searchReq.TemplateID)
	}

	if searchReq.Version != "" {
		query = query.Where("version = ?", searchReq.Version)
	}

	if searchReq.Type != "" {
		query = query.Where("type = ?", searchReq.Type)
	}

	if len(searchReq.IDs) > 0 {
		query = query.Where("id IN ?", searchReq.IDs)
	}

	err := query.Find(&templates).Error
	return templates, err
}

func (r *TemplateRepository) Delete(templateID, tenantID, version string) error {
	return r.db.Where("tenantid = ? AND templateid = ? AND version = ?", tenantID, templateID, version).Delete(&models.TemplateDB{}).Error
}
