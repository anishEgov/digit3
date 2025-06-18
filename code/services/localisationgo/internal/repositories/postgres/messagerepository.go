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

	// Use upsert (INSERT ... ON CONFLICT UPDATE) for each message
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO localisation (tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (tenant_id, locale, module, code) 
		DO UPDATE SET 
			message = EXCLUDED.message,
			last_modified_by = EXCLUDED.last_modified_by,
			last_modified_date = EXCLUDED.last_modified_date
		RETURNING id
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range messages {
		var id int64
		err = stmt.QueryRowContext(ctx,
			messages[i].TenantID,
			messages[i].Module,
			messages[i].Locale,
			messages[i].Code,
			messages[i].Message,
			messages[i].CreatedBy,
			messages[i].CreatedDate,
			messages[i].LastModifiedBy,
			messages[i].LastModifiedDate,
		).Scan(&id)
		if err != nil {
			return err
		}
		messages[i].ID = id
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// UpdateMessages updates existing messages
func (r *MessageRepositoryImpl) UpdateMessages(ctx context.Context, tenantID, locale, module string, messages []domain.Message) error {
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

	// Prepare statement for updating messages
	stmt, err := tx.PrepareContext(ctx, `
		UPDATE localisation
		SET message = $1, last_modified_by = $2, last_modified_date = $3
		WHERE tenant_id = $4 AND locale = $5 AND module = $6 AND code = $7
		RETURNING id
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range messages {
		var id int64
		err = stmt.QueryRowContext(ctx,
			messages[i].Message,
			messages[i].LastModifiedBy,
			messages[i].LastModifiedDate,
			tenantID,
			locale,
			module,
			messages[i].Code,
		).Scan(&id)
		if err != nil {
			// If no rows are affected, the message doesn't exist - create it
			if err == sql.ErrNoRows {
				insertStmt, err := tx.PrepareContext(ctx, `
					INSERT INTO localisation (tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
					RETURNING id
				`)
				if err != nil {
					return err
				}
				defer insertStmt.Close()

				err = insertStmt.QueryRowContext(ctx,
					tenantID,
					module,
					locale,
					messages[i].Code,
					messages[i].Message,
					messages[i].CreatedBy,
					messages[i].CreatedDate,
					messages[i].LastModifiedBy,
					messages[i].LastModifiedDate,
				).Scan(&id)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		messages[i].ID = id
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// DeleteMessages deletes messages by tenantID, locale, module and codes
func (r *MessageRepositoryImpl) DeleteMessages(ctx context.Context, tenantID, locale, module string, codes []string) error {
	// If no codes provided, return immediately
	if len(codes) == 0 {
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
	placeholders := make([]string, len(codes))
	args := make([]interface{}, len(codes)+3)
	args[0] = tenantID
	args[1] = locale
	args[2] = module

	// Build the query dynamically
	for i, code := range codes {
		placeholders[i] = fmt.Sprintf("$%d", i+4)
		args[i+3] = code
	}

	// Create the delete query
	query := fmt.Sprintf(`
		DELETE FROM localisation
		WHERE tenant_id = $1 AND locale = $2 AND module = $3 AND code IN (%s)
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

	log.Printf("Deleted %d messages with tenantID=%s, locale=%s, module=%s, codes=%v",
		rowsAffected, tenantID, locale, module, codes)

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// FindMessages finds messages based on the search criteria
func (r *MessageRepositoryImpl) FindMessages(ctx context.Context, tenantID, module, locale string) ([]domain.Message, error) {
	// Base query
	query := `
		SELECT id, tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date
		FROM localisation 
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argCount := 1

	// Add optional filters
	if module != "" {
		argCount++
		query += fmt.Sprintf(" AND module = $%d", argCount)
		args = append(args, module)
	}

	if locale != "" {
		argCount++
		query += fmt.Sprintf(" AND locale = $%d", argCount)
		args = append(args, locale)
	}

	// Add order by
	query += " ORDER BY module, code"

	// Debug information
	log.Printf("Executing query: %s with args: %v", query, args)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(
			&msg.ID,
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
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning rows: %v", err)
		return nil, err
	}

	log.Printf("Found %d messages for tenantID=%s, module=%s, locale=%s", len(messages), tenantID, module, locale)
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

	// Create the complete query
	query := fmt.Sprintf(`
		SELECT id, tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date
		FROM localisation 
		WHERE tenant_id = $1 AND locale = $2 AND code IN (%s)
		ORDER BY id
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
