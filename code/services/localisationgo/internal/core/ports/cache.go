package ports

import (
	"context"

	"localisationgo/internal/core/domain"
)

// MessageCache defines the caching operations for messages
type MessageCache interface {
	// SetMessages adds messages to the cache
	SetMessages(ctx context.Context, tenantID, module, locale string, messages []domain.Message) error

	// GetMessages retrieves messages from the cache
	GetMessages(ctx context.Context, tenantID, module, locale string) ([]domain.Message, error)

	// Invalidate removes cached messages for a specific tenant+module+locale
	Invalidate(ctx context.Context, tenantID, module, locale string) error

	// BustCache clears the entire cache
	BustCache(ctx context.Context) error
}
