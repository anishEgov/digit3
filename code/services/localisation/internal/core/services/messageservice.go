package services

import (
	"context"
	"strings"
	"time"

	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
)

// MessageServiceImpl is the implementation of MessageService
type MessageServiceImpl struct {
	repository ports.MessageRepository
	cache      ports.MessageCache
}

// NewMessageService creates a new message service
func NewMessageService(repository ports.MessageRepository, cache ports.MessageCache) ports.MessageService {
	return &MessageServiceImpl{
		repository: repository,
		cache:      cache,
	}
}

// UpsertMessages creates or updates localization messages
func (s *MessageServiceImpl) UpsertMessages(ctx context.Context, tenantID string, messages []domain.Message) ([]domain.Message, error) {
	// Enrich messages with tenant ID and timestamps
	now := time.Now()
	for i := range messages {
		messages[i].TenantID = tenantID
		// Only set created date for new messages (ID == 0)
		if messages[i].ID == 0 {
			messages[i].CreatedDate = now
		}
		messages[i].LastModifiedDate = now
	}

	// Save to database
	err := s.repository.SaveMessages(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Invalidate cache for affected tenant+module+locale combinations
	cacheKeys := make(map[string]struct{})
	for _, msg := range messages {
		key := msg.TenantID + ":" + msg.Module + ":" + msg.Locale
		cacheKeys[key] = struct{}{}
	}

	for key := range cacheKeys {
		parts := strings.Split(key, ":")
		if len(parts) == 3 {
			_ = s.cache.Invalidate(ctx, parts[0], parts[1], parts[2])
		}
	}

	return messages, nil
}

// SearchMessages retrieves messages based on search criteria
func (s *MessageServiceImpl) SearchMessages(ctx context.Context, tenantID, module, locale string) ([]domain.Message, error) {
	// When module is empty, we're searching across all modules - bypass cache
	// to ensure we get all results from the database
	if module == "" {
		return s.repository.FindMessages(ctx, tenantID, module, locale)
	}

	// For module-specific queries, try to get from cache first
	cachedMessages, err := s.cache.GetMessages(ctx, tenantID, module, locale)
	if err == nil && len(cachedMessages) > 0 {
		return cachedMessages, nil
	}

	// If not in cache or there was an error, get from database
	messages, err := s.repository.FindMessages(ctx, tenantID, module, locale)
	if err != nil {
		return nil, err
	}

	// Update cache with new results
	if len(messages) > 0 {
		_ = s.cache.SetMessages(ctx, tenantID, module, locale, messages)
	}

	return messages, nil
}

// SearchMessagesByCodes retrieves messages for specific codes
func (s *MessageServiceImpl) SearchMessagesByCodes(ctx context.Context, tenantID, locale string, codes []string) ([]domain.Message, error) {
	// For code-specific searches, we don't use cache to ensure accuracy
	return s.repository.FindMessagesByCode(ctx, tenantID, locale, codes)
}
