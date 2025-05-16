package ports

import (
	"context"

	"localisationgo/internal/core/domain"
)

// MessageService defines the business operations for localization messages
type MessageService interface {
	// UpsertMessages creates or updates localization messages
	UpsertMessages(ctx context.Context, tenantID string, messages []domain.Message) ([]domain.Message, error)

	// SearchMessages retrieves messages based on search criteria
	SearchMessages(ctx context.Context, tenantID, module, locale string) ([]domain.Message, error)

	// SearchMessagesByCodes retrieves messages for specific codes
	SearchMessagesByCodes(ctx context.Context, tenantID, locale string, codes []string) ([]domain.Message, error)
}
