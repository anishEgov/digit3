package dtos

// MessageResponse is a simplified message structure for responses
type MessageResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Module  string `json:"module"`
	Locale  string `json:"locale"`
}

// UpsertMessagesResponse represents the response for the upsert localization messages API
type UpsertMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}

// SearchMessagesResponse represents the response for the search localization messages API
type SearchMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}
