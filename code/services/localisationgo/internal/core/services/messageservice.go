package services

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
)

// MessageServiceImpl is the implementation of MessageService
type MessageServiceImpl struct {
	repository        ports.MessageRepository
	cache             ports.MessageCache
	messageLocalesMap map[string]map[string][]string // tenantID -> code -> []locale
}

// NewMessageService creates a new message service
func NewMessageService(repository ports.MessageRepository, cache ports.MessageCache) ports.MessageService {
	return &MessageServiceImpl{
		repository:        repository,
		cache:             cache,
		messageLocalesMap: make(map[string]map[string][]string),
	}
}

// LoadAllMessages loads all messages from the repository and builds the tenant-to-code-to-locales map
func (s *MessageServiceImpl) LoadAllMessages(ctx context.Context) error {
	log.Println("Loading all messages into cache...")
	messages, err := s.repository.FindAllMessages(ctx)
	if err != nil {
		return err
	}

	// Clear the existing map
	s.messageLocalesMap = make(map[string]map[string][]string)

	for _, msg := range messages {
		if _, ok := s.messageLocalesMap[msg.TenantID]; !ok {
			s.messageLocalesMap[msg.TenantID] = make(map[string][]string)
		}
		s.messageLocalesMap[msg.TenantID][msg.Code] = append(s.messageLocalesMap[msg.TenantID][msg.Code], msg.Locale)
	}
	log.Println("Finished loading all messages into cache.")
	return nil
}

// FindMissingMessages finds the missing messages for a given tenant.
// The response can be filtered by providing a list of locales.
func (s *MessageServiceImpl) FindMissingMessages(ctx context.Context, tenantID string, requestedLocales []string) (map[string][]string, error) {
	tenantMessages, ok := s.messageLocalesMap[tenantID]
	if !ok {
		return nil, nil // Or an error indicating tenant not found
	}

	// Determine the set of all unique codes for the tenant.
	allCodesForTenant := make(map[string]struct{})
	for code := range tenantMessages {
		allCodesForTenant[code] = struct{}{}
	}

	// Determine the set of all unique locales for the tenant.
	allLocalesForTenant := make(map[string]struct{})
	for _, msgLocales := range tenantMessages {
		for _, l := range msgLocales {
			allLocalesForTenant[l] = struct{}{}
		}
	}

	// Calculate all missing messages for ALL locales.
	allMissingMessages := make(map[string][]string)
	for locale := range allLocalesForTenant {
		// Create a set of codes that exist for the current locale
		localeCodes := make(map[string]struct{})
		for code, msgLocales := range tenantMessages {
			for _, l := range msgLocales {
				if l == locale {
					localeCodes[code] = struct{}{}
					break
				}
			}
		}

		// Find the codes that are in allCodesForTenant but not in localeCodes
		var missing []string
		for code := range allCodesForTenant {
			if _, exists := localeCodes[code]; !exists {
				missing = append(missing, code)
			}
		}
		if len(missing) > 0 {
			allMissingMessages[locale] = missing
		}
	}

	// If no locales were requested, return all missing messages.
	if len(requestedLocales) == 0 {
		return allMissingMessages, nil
	}

	// Otherwise, filter the results based on the requested locales.
	filteredMissingMessages := make(map[string][]string)
	for _, requestedLocale := range requestedLocales {
		if missingCodes, found := allMissingMessages[requestedLocale]; found {
			filteredMissingMessages[requestedLocale] = missingCodes
		}
	}

	return filteredMissingMessages, nil
}

// UpsertMessages creates or updates localization messages
func (s *MessageServiceImpl) UpsertMessages(ctx context.Context, tenantID string, userID string, messages []domain.Message) ([]domain.Message, error) {
	messagesToCreate := []domain.Message{}
	messagesToUpdate := []domain.Message{}

	for _, msg := range messages {
		if msg.UUID == "" {
			messagesToCreate = append(messagesToCreate, msg)
		} else {
			messagesToUpdate = append(messagesToUpdate, msg)
		}
	}

	now := time.Now()
	var finalMessages []domain.Message

	// Handle creations
	if len(messagesToCreate) > 0 {
		for i := range messagesToCreate {
			messagesToCreate[i].UUID = uuid.New().String()
			messagesToCreate[i].TenantID = tenantID
			messagesToCreate[i].CreatedDate = now
			messagesToCreate[i].LastModifiedDate = now
			messagesToCreate[i].CreatedBy = userID
			messagesToCreate[i].LastModifiedBy = userID
		}
		err := s.repository.SaveMessages(ctx, messagesToCreate)
		if err != nil {
			return nil, err
		}
		finalMessages = append(finalMessages, messagesToCreate...)

		// Update the in-memory map for created messages
		for _, msg := range messagesToCreate {
			if _, ok := s.messageLocalesMap[msg.TenantID]; !ok {
				s.messageLocalesMap[msg.TenantID] = make(map[string][]string)
			}
			// Avoid adding duplicate locales for a code
			found := false
			if _, ok := s.messageLocalesMap[msg.TenantID][msg.Code]; ok {
				for _, locale := range s.messageLocalesMap[msg.TenantID][msg.Code] {
					if locale == msg.Locale {
						found = true
						break
					}
				}
			}
			if !found {
				s.messageLocalesMap[msg.TenantID][msg.Code] = append(s.messageLocalesMap[msg.TenantID][msg.Code], msg.Locale)
			}
		}
	}

	// Handle updates
	if len(messagesToUpdate) > 0 {
		for i := range messagesToUpdate {
			messagesToUpdate[i].TenantID = tenantID
			messagesToUpdate[i].LastModifiedDate = now
			messagesToUpdate[i].LastModifiedBy = userID
		}
		updatedMessages, err := s.repository.UpdateMessages(ctx, tenantID, messagesToUpdate)
		if err != nil {
			return nil, err
		}
		finalMessages = append(finalMessages, updatedMessages...)
	}

	// Invalidate cache
	s.invalidateCacheForMessages(ctx, finalMessages)

	return finalMessages, nil
}

// CreateMessages creates new localization messages
func (s *MessageServiceImpl) CreateMessages(ctx context.Context, tenantID string, userID string, messages []domain.Message) ([]domain.Message, error) {
	// Enrich messages with tenant ID and timestamps
	now := time.Now()
	for i := range messages {
		messages[i].UUID = uuid.New().String()
		messages[i].TenantID = tenantID
		messages[i].CreatedDate = now
		messages[i].LastModifiedDate = now
		messages[i].CreatedBy = userID
		messages[i].LastModifiedBy = userID
	}

	// Save to database
	err := s.repository.SaveMessages(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Invalidate cache for affected tenant+module+locale combinations
	s.invalidateCacheForMessages(ctx, messages)

	// Update the in-memory map
	for _, msg := range messages {
		if _, ok := s.messageLocalesMap[msg.TenantID]; !ok {
			s.messageLocalesMap[msg.TenantID] = make(map[string][]string)
		}
		// Avoid adding duplicate locales for a code
		found := false
		if _, ok := s.messageLocalesMap[msg.TenantID][msg.Code]; ok {
			for _, locale := range s.messageLocalesMap[msg.TenantID][msg.Code] {
				if locale == msg.Locale {
					found = true
					break
				}
			}
		}
		if !found {
			s.messageLocalesMap[msg.TenantID][msg.Code] = append(s.messageLocalesMap[msg.TenantID][msg.Code], msg.Locale)
		}
	}

	return messages, nil
}

// UpdateMessages updates existing messages by their UUIDs
func (s *MessageServiceImpl) UpdateMessages(ctx context.Context, tenantID string, userID string, messages []domain.Message) ([]domain.Message, error) {
	// Enrich messages with timestamp
	now := time.Now()
	for i := range messages {
		messages[i].TenantID = tenantID
		messages[i].LastModifiedDate = now
		messages[i].LastModifiedBy = userID
	}

	// Update in database
	updatedMessages, err := s.repository.UpdateMessages(ctx, tenantID, messages)
	if err != nil {
		return nil, err
	}

	// Invalidate cache for affected tenant+module+locale combinations
	s.invalidateCacheForMessages(ctx, updatedMessages)

	return updatedMessages, nil
}

// DeleteMessages deletes messages by their UUIDs
func (s *MessageServiceImpl) DeleteMessages(ctx context.Context, tenantID string, uuids []string) error {
	// Note: To invalidate the cache correctly, we would first need to fetch the messages
	// to get their module and locale. For simplicity in this example, we will
	// bust the entire cache for the tenant. A more optimized approach would fetch
	// the details before deleting.
	if err := s.cache.BustCache(ctx, tenantID, "", ""); err != nil {
		log.Printf("Failed to bust cache for tenant %s during delete, continuing...", tenantID)
	}

	return s.repository.DeleteMessages(ctx, tenantID, uuids)
}

// BustCache clears the cache based on tenant, and optionally module and locale
func (s *MessageServiceImpl) BustCache(ctx context.Context, tenantID, module, locale string) error {
	return s.cache.BustCache(ctx, tenantID, module, locale)
}

// SearchMessages searches for messages, checking the cache first
func (s *MessageServiceImpl) SearchMessages(ctx context.Context, tenantID, module, locale string) ([]domain.Message, error) {
	// Try to get from cache first
	if cachedMessages, err := s.cache.GetMessages(ctx, tenantID, module, locale); err == nil && len(cachedMessages) > 0 {
		return cachedMessages, nil
	}

	// If not in cache, get from repository
	messages, err := s.repository.FindMessages(ctx, tenantID, module, locale)
	if err != nil {
		return nil, err
	}

	// Store in cache
	_ = s.cache.SetMessages(ctx, tenantID, module, locale, messages)

	return messages, nil
}

// SearchMessagesByCodes searches for messages by specific codes, checking the cache first
func (s *MessageServiceImpl) SearchMessagesByCodes(ctx context.Context, tenantID, locale string, codes []string) ([]domain.Message, error) {
	// For searching by codes, it is complex to implement a partial cache read.
	// For now, we will bypass the cache and go straight to the repository.
	// A more advanced caching strategy could be implemented here in the future.
	return s.repository.FindMessagesByCode(ctx, tenantID, locale, codes)
}

// invalidateCacheForMessages invalidates the cache for the given messages
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
