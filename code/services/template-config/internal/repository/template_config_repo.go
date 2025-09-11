package repository

import (
	"template-config/internal/models"

	"gorm.io/gorm"
)

type TemplateConfigRepository struct {
	db *gorm.DB
}

func NewTemplateConfigRepository(db *gorm.DB) *TemplateConfigRepository {
	return &TemplateConfigRepository{db: db}
}

func (r *TemplateConfigRepository) Create(config *models.TemplateConfigDB) error {
	return r.db.Create(config).Error
}

func (r *TemplateConfigRepository) Update(config *models.TemplateConfigDB) error {
	return r.db.Save(config).Error
}

func (r *TemplateConfigRepository) GetByID(uuid string) (*models.TemplateConfigDB, error) {
	var config models.TemplateConfigDB
	err := r.db.Where("id = ?", uuid).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *TemplateConfigRepository) GetByTemplateIDAndVersion(templateID, tenantID, version string) (*models.TemplateConfigDB, error) {
	var config models.TemplateConfigDB
	err := r.db.Where("templateid = ? AND tenantid = ? AND version = ?", templateID, tenantID, version).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *TemplateConfigRepository) Search(search *models.TemplateConfigSearch) ([]models.TemplateConfigDB, error) {
	var configs []models.TemplateConfigDB
	query := r.db.Where("tenantid = ?", search.TenantID)

	if search.TemplateID != "" {
		query = query.Where("templateid = ?", search.TemplateID)
	}

	if search.Version != "" {
		query = query.Where("version = ?", search.Version)
	}

	if len(search.IDs) > 0 {
		query = query.Where("id IN ?", search.IDs)
	}

	err := query.Find(&configs).Error
	return configs, err
}

func (r *TemplateConfigRepository) Delete(templateID, tenantID, version string) error {
	return r.db.Where("templateid = ? AND tenantid = ? AND version = ?", templateID, tenantID, version).Delete(&models.TemplateConfigDB{}).Error
}
