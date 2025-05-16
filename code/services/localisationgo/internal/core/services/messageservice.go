package services

import (
	"context"
	"strings"
	"time"

	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
	"localisationgo/pkg/dtos"
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

// CreateMessages creates new localization messages
func (s *MessageServiceImpl) CreateMessages(ctx context.Context, tenantID string, messages []domain.Message) ([]domain.Message, error) {
	// Enrich messages with tenant ID and timestamps
	now := time.Now()
	for i := range messages {
		messages[i].TenantID = tenantID
		messages[i].CreatedDate = now
		messages[i].LastModifiedDate = now
	}

	// Save to database
	err := s.repository.SaveMessages(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Invalidate cache for affected tenant+module+locale combinations
	s.invalidateCacheForMessages(ctx, messages)

	return messages, nil
}

// UpdateMessagesForModule updates existing messages for a specific module
func (s *MessageServiceImpl) UpdateMessagesForModule(ctx context.Context, tenantID, locale, module string, messages []domain.Message) ([]domain.Message, error) {
	// Enrich messages with tenant ID, locale, module and timestamp
	now := time.Now()
	for i := range messages {
		messages[i].TenantID = tenantID
		messages[i].Locale = locale
		messages[i].Module = module
		messages[i].LastModifiedDate = now
	}

	// Update in database
	err := s.repository.UpdateMessages(ctx, tenantID, locale, module, messages)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	_ = s.cache.Invalidate(ctx, tenantID, module, locale)

	return messages, nil
}

// DeleteMessages deletes messages matching the given identities
func (s *MessageServiceImpl) DeleteMessages(ctx context.Context, messageIdentities []dtos.MessageIdentity) error {
	// Group message identities by tenant+locale+module for efficient deletion
	tenantLocaleModuleMap := make(map[string]map[string]map[string][]string)

	for _, identity := range messageIdentities {
		// Initialize maps if they don't exist
		if _, ok := tenantLocaleModuleMap[identity.TenantId]; !ok {
			tenantLocaleModuleMap[identity.TenantId] = make(map[string]map[string][]string)
		}
		if _, ok := tenantLocaleModuleMap[identity.TenantId][identity.Locale]; !ok {
			tenantLocaleModuleMap[identity.TenantId][identity.Locale] = make(map[string][]string)
		}
		if _, ok := tenantLocaleModuleMap[identity.TenantId][identity.Locale][identity.Module]; !ok {
			tenantLocaleModuleMap[identity.TenantId][identity.Locale][identity.Module] = []string{}
		}

		// Add code to the list
		tenantLocaleModuleMap[identity.TenantId][identity.Locale][identity.Module] =
			append(tenantLocaleModuleMap[identity.TenantId][identity.Locale][identity.Module], identity.Code)
	}

	// Delete messages for each tenant+locale+module combination
	for tenantID, localeMap := range tenantLocaleModuleMap {
		for locale, moduleMap := range localeMap {
			for module, codes := range moduleMap {
				err := s.repository.DeleteMessages(ctx, tenantID, locale, module, codes)
				if err != nil {
					return err
				}

				// Invalidate cache
				_ = s.cache.Invalidate(ctx, tenantID, module, locale)
			}
		}
	}

	return nil
}

// BustCache clears the entire cache
func (s *MessageServiceImpl) BustCache(ctx context.Context) error {
	return s.cache.BustCache(ctx)
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

// invalidateCacheForMessages invalidates cache entries for the given messages
func (s *MessageServiceImpl) invalidateCacheForMessages(ctx context.Context, messages []domain.Message) {
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
}
