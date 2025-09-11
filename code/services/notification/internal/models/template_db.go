package models

import (
	"github.com/google/uuid"
)

// TemplateDB is the database model that matches the table schema
type TemplateDB struct {
	ID               uuid.UUID `gorm:"column:id;type:uuid;primary_key"`
	TemplateID       string    `gorm:"column:templateid;not null"`
	Version          string    `gorm:"column:version;not null"`
	TenantID         string    `gorm:"column:tenantid;not null"`
	Type             string    `gorm:"column:type;not null"`
	Subject          string    `gorm:"column:subject;type:text"`
	Content          string    `gorm:"column:content;type:text;not null"`
	IsHTML           bool      `gorm:"column:ishtml"`
	CreatedBy        string    `gorm:"column:createdby"`
	LastModifiedBy   string    `gorm:"column:lastmodifiedby"`
	CreatedTime      int64     `gorm:"column:createdtime"`
	LastModifiedTime int64     `gorm:"column:lastmodifiedtime"`
}

func (TemplateDB) TableName() string {
	return "notification_template"
}

// ToDTO converts TemplateDB to Template (DB to API)
func (t *TemplateDB) ToDTO() Template {
	return Template{
		ID:         t.ID,
		TemplateID: t.TemplateID,
		TenantID:   t.TenantID,
		Version:    t.Version,
		Type:       TemplateType(t.Type),
		Subject:    t.Subject,
		Content:    t.Content,
		IsHTML:     t.IsHTML,
		AuditDetails: AuditDetails{
			CreatedBy:        t.CreatedBy,
			CreatedTime:      t.CreatedTime,
			LastModifiedBy:   t.LastModifiedBy,
			LastModifiedTime: t.LastModifiedTime,
		},
	}
}

// FromDTO converts Template to TemplateDB (API to DB)
func FromDTO(dto *Template) TemplateDB {
	return TemplateDB{
		ID:               dto.ID,
		TemplateID:       dto.TemplateID,
		TenantID:         dto.TenantID,
		Version:          dto.Version,
		Type:             string(dto.Type),
		Subject:          dto.Subject,
		Content:          dto.Content,
		IsHTML:           dto.IsHTML,
		CreatedBy:        dto.AuditDetails.CreatedBy,
		CreatedTime:      dto.AuditDetails.CreatedTime,
		LastModifiedBy:   dto.AuditDetails.LastModifiedBy,
		LastModifiedTime: dto.AuditDetails.LastModifiedTime,
	}
}
