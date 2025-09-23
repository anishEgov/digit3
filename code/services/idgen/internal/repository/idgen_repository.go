package repository

import (
	"fmt"
	"idgen/internal/models"

	"gorm.io/gorm"
)

type IDGenRepository struct {
	db *gorm.DB
}

func NewIDGenRepository(db *gorm.DB) *IDGenRepository {
	return &IDGenRepository{db: db}
}

func (r *IDGenRepository) SaveTemplate(template *models.IDGenTemplate) error {
	return r.db.Create(template).Error
}

func (r *IDGenRepository) GetTemplate(id string) (*models.IDGenTemplate, error) {
	var tmpl models.IDGenTemplate
	if err := r.db.First(&tmpl, "templateid = ?", id).Error; err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (r *IDGenRepository) CreateSequence(templateID string, start int) error {
	seqName := fmt.Sprintf("seq_%s", templateID)
	query := fmt.Sprintf(`
		CREATE SEQUENCE IF NOT EXISTS %s
		START WITH %d
		INCREMENT BY 1
		MINVALUE 1
		CACHE 1;
	`, seqName, start)
	return r.db.Exec(query).Error
}

func (r *IDGenRepository) NextSequenceValue(templateID string) (int64, error) {
	seqName := fmt.Sprintf("seq_%s", templateID)
	var value int64
	query := fmt.Sprintf("SELECT nextval('%s')", seqName)
	if err := r.db.Raw(query).Scan(&value).Error; err != nil {
		return 0, err
	}
	return value, nil
}

func (r *IDGenRepository) EnsureScopeReset(templateID, scopeKey string, start int) error {
	var count int64
	err := r.db.Raw(
		"SELECT COUNT(*) FROM idgen_sequence_resets_v2 WHERE templateid = ? AND scopekey = ?",
		templateID, scopeKey,
	).Scan(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		// Reset the underlying Postgres sequence
		seqName := fmt.Sprintf("seq_%s", templateID)
		resetSQL := fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH %d", seqName, start)
		if err := r.db.Exec(resetSQL).Error; err != nil {
			return err
		}

		// Track this reset so we don't reset again within the scope
		return r.db.Exec(
			"INSERT INTO idgen_sequence_resets_v2 (templateid, scopekey, lastvalue) VALUES (?, ?, ?)",
			templateID, scopeKey, 0,
		).Error
	}
	return nil
}
