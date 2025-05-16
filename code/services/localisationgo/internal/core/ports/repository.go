package ports

import (
	"context"

	"localisationgo/internal/core/domain"
)

// MessageRepository defines the operations available for message storage
type MessageRepository interface {
	// SaveMessages persists messages to the database
	SaveMessages(ctx context.Context, messages []domain.Message) error

	// FindMessages finds messages based on the search criteria
	FindMessages(ctx context.Context, tenantID, module, locale string) ([]domain.Message, error)

	// FindMessagesByCode finds messages with specific codes
	FindMessagesByCode(ctx context.Context, tenantID, locale string, codes []string) ([]domain.Message, error)

	// UpdateMessages updates existing messages
	UpdateMessages(ctx context.Context, tenantID, locale, module string, messages []domain.Message) error

	// DeleteMessages deletes messages by tenantID, locale, module and codes
	DeleteMessages(ctx context.Context, tenantID, locale, module string, codes []string) error
}
