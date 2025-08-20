package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
)

// MessageRepositoryImpl is an implementation of the MessageRepository using PostgreSQL
type MessageRepositoryImpl struct {
	db *sql.DB
}

// NewMessageRepository creates a new instance of MessageRepositoryImpl
func NewMessageRepository(db *sql.DB) ports.MessageRepository {
	return &MessageRepositoryImpl{
		db: db,
	}
}

// isValidUUID checks if a string is a valid UUID format
func isValidUUID(u string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(u)
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
		result, err := stmt.ExecContext(ctx,
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

		// Check if the row was actually inserted (not ignored due to conflict)
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return fmt.Errorf("message with code '%s' already exists for module '%s' and locale '%s'. Duplicate codes are not allowed",
				messages[i].Code, messages[i].Module, messages[i].Locale)
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
	if len(messages) == 0 {
		return nil, nil
	}

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

	// First, validate UUID formats
	uuids := make([]string, len(messages))
	for i, msg := range messages {
		uuids[i] = msg.UUID
		if !isValidUUID(msg.UUID) {
			return nil, fmt.Errorf("invalid UUID format: '%s'. UUID must be in format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", msg.UUID)
		}
	}

	// Create placeholders for the IN clause for validation
	placeholders := make([]string, len(uuids))
	args := make([]interface{}, len(uuids)+1)
	args[0] = tenantID

	for i, id := range uuids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	// Check how many records exist with the given UUIDs and tenant
	checkQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM localisation
		WHERE tenant_id = $1 AND uuid IN (%s)
	`, strings.Join(placeholders, ","))

	var count int64
	err = tx.QueryRowContext(ctx, checkQuery, args...).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("database error while validating UUIDs: %w", err)
	}

	// If the count doesn't match the number of UUIDs provided, some UUIDs are missing
	if count != int64(len(uuids)) {
		missingCount := int64(len(uuids)) - count
		return nil, fmt.Errorf("update failed: %d UUID(s) not found in database. All UUIDs must exist to perform atomic update", missingCount)
	}

	// All UUIDs exist, proceed with updates
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
			// This should not happen since we validated all UUIDs exist above
			return nil, fmt.Errorf("unexpected error updating UUID %s: %w", msg.UUID, err)
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

	// First, validate UUID formats
	for _, uuid := range uuids {
		if !isValidUUID(uuid) {
			return fmt.Errorf("invalid UUID format: '%s'. UUID must be in format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", uuid)
		}
	}

	// Create a placeholder string for the IN clause for validation
	placeholders := make([]string, len(uuids))
	args := make([]interface{}, len(uuids)+1)
	args[0] = tenantID

	for i, id := range uuids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	// Check how many records exist with the given UUIDs and tenant
	checkQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM localisation
		WHERE tenant_id = $1 AND uuid IN (%s)
	`, strings.Join(placeholders, ","))

	var count int64
	err = tx.QueryRowContext(ctx, checkQuery, args...).Scan(&count)
	if err != nil {
		return fmt.Errorf("database error while validating UUIDs: %w", err)
	}

	// If the count doesn't match the number of UUIDs provided, some UUIDs are missing
	if count != int64(len(uuids)) {
		missingCount := int64(len(uuids)) - count
		return fmt.Errorf("delete failed: %d UUID(s) not found in database. All UUIDs must exist to perform atomic deletion", missingCount)
	}

	// All UUIDs exist, proceed with deletion
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

// FindMessages finds messages based on the search criteria, with pagination and sorting
func (r *MessageRepositoryImpl) FindMessages(ctx context.Context, tenantID, module, locale string, limit, offset int) ([]domain.Message, error) {
	// Build dynamic query based on provided parameters
	query := `
        SELECT uuid, tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date
        FROM localisation
        WHERE tenant_id = $1`

	args := []interface{}{tenantID}
	paramCount := 1

	// Add module filter if provided
	if module != "" {
		paramCount++
		query += fmt.Sprintf(" AND module = $%d", paramCount)
		args = append(args, module)
	}

	// Add locale filter if provided
	if locale != "" {
		paramCount++
		query += fmt.Sprintf(" AND locale = $%d", paramCount)
		args = append(args, locale)
	}

	query += " ORDER BY code ASC"

	// Add pagination if limit is specified
	if limit > 0 {
		paramCount++
		query += fmt.Sprintf(" LIMIT $%d", paramCount)
		args = append(args, limit)

		paramCount++
		query += fmt.Sprintf(" OFFSET $%d", paramCount)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []domain.Message{}
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(
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

	// Build the query dynamically based on whether locale is provided
	var query string
	var args []interface{}

	if locale == "" {
		// Search across all locales when locale is not specified
		placeholders := make([]string, len(codes))
		args = make([]interface{}, len(codes)+1)
		args[0] = tenantID

		for i, code := range codes {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			args[i+1] = code
		}

		query = fmt.Sprintf(`
			SELECT id, uuid, tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date
			FROM localisation
			WHERE tenant_id = $1 AND code IN (%s)
		`, strings.Join(placeholders, ","))
	} else {
		// Search within specific locale when locale is provided
		placeholders := make([]string, len(codes))
		args = make([]interface{}, len(codes)+2)
		args[0] = tenantID
		args[1] = locale

		for i, code := range codes {
			placeholders[i] = fmt.Sprintf("$%d", i+3)
			args[i+2] = code
		}

		query = fmt.Sprintf(`
			SELECT id, uuid, tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date
			FROM localisation
			WHERE tenant_id = $1 AND locale = $2 AND code IN (%s)
		`, strings.Join(placeholders, ","))
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
