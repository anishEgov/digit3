package services

import (
	"context"
	"log"
	"strings"
	"time"

	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
	"localisationgo/pkg/dtos"
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
	s.invalidateCacheForMessages(ctx, messages)

	// Update the in-memory map
	for _, msg := range messages {
		if _, ok := s.messageLocalesMap[msg.TenantID]; !ok {
			s.messageLocalesMap[msg.TenantID] = make(map[string][]string)
		}
		// Avoid adding duplicate locales for a code
		found := false
		for _, locale := range s.messageLocalesMap[msg.TenantID][msg.Code] {
			if locale == msg.Locale {
				found = true
				break
			}
		}
		if !found {
			s.messageLocalesMap[msg.TenantID][msg.Code] = append(s.messageLocalesMap[msg.TenantID][msg.Code], msg.Locale)
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

	// Update the in-memory map
	for _, msg := range messages {
		if _, ok := s.messageLocalesMap[msg.TenantID]; !ok {
			s.messageLocalesMap[msg.TenantID] = make(map[string][]string)
		}
		// Avoid adding duplicate locales for a code
		found := false
		for _, locale := range s.messageLocalesMap[msg.TenantID][msg.Code] {
			if locale == msg.Locale {
				found = true
				break
			}
		}
		if !found {
			s.messageLocalesMap[msg.TenantID][msg.Code] = append(s.messageLocalesMap[msg.TenantID][msg.Code], msg.Locale)
		}
	}

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

	// We don't need to update the in-memory map here, because the locales don't change on update.
	// The message content changes, but the missing messages logic only cares about the presence of a (code, locale) pair.

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

				// Update the in-memory map by reloading all messages.
				// This is a simple approach. A more optimized approach would be to remove the specific entries.
				if err := s.LoadAllMessages(ctx); err != nil {
					log.Printf("Error reloading messages into cache after deletion: %v", err)
					// Decide if we should return this error or just log it
				}
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
