package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
)

// MessageRepositoryImpl is an implementation of the MessageRepository using PostgreSQL
type MessageRepositoryImpl struct {
	db *sql.DB
}

// NewMessageRepository creates a new PostgreSQL message repository
func NewMessageRepository(db *sql.DB) ports.MessageRepository {
	return &MessageRepositoryImpl{
		db: db,
	}
}

// SaveMessages persists messages to the database
func (r *MessageRepositoryImpl) SaveMessages(ctx context.Context, messages []domain.Message) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
	}()

	// Use simple INSERT ON CONFLICT DO NOTHING for each message.
	// The service layer is responsible for generating UUIDs and handling conflicts.
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO localisation (uuid, tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id, locale, module, code) DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range messages {
		_, err = stmt.ExecContext(ctx,
			messages[i].UUID,
			messages[i].TenantID,
			messages[i].Module,
			messages[i].Locale,
			messages[i].Code,
			messages[i].Message,
			messages[i].CreatedBy,
			messages[i].CreatedDate,
			messages[i].LastModifiedBy,
			messages[i].LastModifiedDate,
		)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// UpdateMessages updates existing messages by their UUIDs
func (r *MessageRepositoryImpl) UpdateMessages(ctx context.Context, tenantID string, messages []domain.Message) ([]domain.Message, error) {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
	}()

	// Prepare statement for updating messages
	stmt, err := tx.PrepareContext(ctx, `
		UPDATE localisation
		SET message = $1, last_modified_by = $2, last_modified_date = $3
		WHERE tenant_id = $4 AND uuid = $5
		RETURNING id, uuid, module, locale, code, created_by, created_date
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	updatedMessages := make([]domain.Message, 0, len(messages))
	for _, msg := range messages {
		var updatedMsg domain.Message
		err = stmt.QueryRowContext(ctx,
			msg.Message,
			msg.LastModifiedBy,
			msg.LastModifiedDate,
			tenantID,
			msg.UUID,
		).Scan(
			&updatedMsg.ID,
			&updatedMsg.UUID,
			&updatedMsg.Module,
			&updatedMsg.Locale,
			&updatedMsg.Code,
			&updatedMsg.CreatedBy,
			&updatedMsg.CreatedDate,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("No message found with UUID %s for tenant %s. Skipping update.", msg.UUID, tenantID)
				continue
			}
			return nil, err
		}
		// Preserve the fields that are not returned from the DB
		updatedMsg.Message = msg.Message
		updatedMsg.LastModifiedBy = msg.LastModifiedBy
		updatedMsg.LastModifiedDate = msg.LastModifiedDate
		updatedMsg.TenantID = tenantID
		updatedMessages = append(updatedMessages, updatedMsg)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return updatedMessages, nil
}

// DeleteMessages deletes messages by their UUIDs for a given tenant
func (r *MessageRepositoryImpl) DeleteMessages(ctx context.Context, tenantID string, uuids []string) error {
	// If no uuids provided, return immediately
	if len(uuids) == 0 {
		return nil
	}

	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
	}()

	// Create a placeholder string for the IN clause
	placeholders := make([]string, len(uuids))
	args := make([]interface{}, len(uuids)+1)
	args[0] = tenantID

	// Build the query dynamically
	for i, id := range uuids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	// Create the delete query
	query := fmt.Sprintf(`
		DELETE FROM localisation
		WHERE tenant_id = $1 AND uuid IN (%s)
	`, strings.Join(placeholders, ","))

	// Execute the delete query
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	log.Printf("Deleted %d messages with tenantID=%s, uuids=%v",
		rowsAffected, tenantID, uuids)

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// FindMessages finds messages based on the search criteria
func (r *MessageRepositoryImpl) FindMessages(ctx context.Context, tenantID, module, locale string) ([]domain.Message, error) {
	query := `
		SELECT id, uuid, tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date
		FROM localisation
		WHERE tenant_id = $1 AND locale = $2
	`
	args := []interface{}{tenantID, locale}

	if module != "" {
		query += " AND module = $3"
		args = append(args, module)
	}

	// Execute the query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.UUID,
			&msg.TenantID,
			&msg.Module,
			&msg.Locale,
			&msg.Code,
			&msg.Message,
			&msg.CreatedBy,
			&msg.CreatedDate,
			&msg.LastModifiedBy,
			&msg.LastModifiedDate,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// FindMessagesByCode finds messages with specific codes
func (r *MessageRepositoryImpl) FindMessagesByCode(ctx context.Context, tenantID, locale string, codes []string) ([]domain.Message, error) {
	// If no codes provided, return empty slice
	if len(codes) == 0 {
		return []domain.Message{}, nil
	}

	// Create a placeholder string for the IN clause
	placeholders := make([]string, len(codes))
	args := make([]interface{}, len(codes)+2)
	args[0] = tenantID
	args[1] = locale

	// Build the query dynamically
	for i, code := range codes {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args[i+2] = code
	}

	// Create the select query
	query := fmt.Sprintf(`
		SELECT id, uuid, tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date
		FROM localisation
		WHERE tenant_id = $1 AND locale = $2 AND code IN (%s)
	`, strings.Join(placeholders, ","))

	// Execute the query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.UUID,
			&msg.TenantID,
			&msg.Module,
			&msg.Locale,
			&msg.Code,
			&msg.Message,
			&msg.CreatedBy,
			&msg.CreatedDate,
			&msg.LastModifiedBy,
			&msg.LastModifiedDate,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// FindAllMessages fetches all messages from the database
func (r *MessageRepositoryImpl) FindAllMessages(ctx context.Context) ([]domain.Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, uuid, tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date
		FROM localisation
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.UUID,
			&msg.TenantID,
			&msg.Module,
			&msg.Locale,
			&msg.Code,
			&msg.Message,
			&msg.CreatedBy,
			&msg.CreatedDate,
			&msg.LastModifiedBy,
			&msg.LastModifiedDate,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
