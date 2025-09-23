package models

import (
	"gorm.io/datatypes"
)

type IDGenTemplate struct {
	TemplateID  string         `gorm:"column:templateid;primaryKey;not null;size:64" json:"templateId"`
	Config      datatypes.JSON `gorm:"column:config;type:jsonb;not null" json:"config"`
	CreatedTime int64          `gorm:"column:createdtime" json:"createdTime"`
	CreatedBy   string         `gorm:"column:createdby;size:64" json:"createdBy"`
}

// TableName overrides the default (id_gen_templates) to match your table
func (IDGenTemplate) TableName() string {
	return "idgen_templates_v2"
}
