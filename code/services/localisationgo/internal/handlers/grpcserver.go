package handlers

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"localisationgo/api/proto/localization/v1"
	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
	"localisationgo/pkg/dtos"
)

// GRPCServer implements the gRPC server for the localization service
type GRPCServer struct {
	localizationv1.UnimplementedLocalizationServiceServer
	service ports.MessageService
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(service ports.MessageService) *GRPCServer {
	return &GRPCServer{
		service: service,
	}
}

// SearchMessages implements the SearchMessages gRPC endpoint
func (s *GRPCServer) SearchMessages(ctx context.Context, req *localizationv1.SearchMessagesRequest) (*localizationv1.SearchMessagesResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	var messages []domain.Message
	var err error

	if len(req.Codes) > 0 {
		messages, err = s.service.SearchMessagesByCodes(ctx, req.TenantId, req.Locale, req.Codes)
	} else {
		messages, err = s.service.SearchMessages(ctx, req.TenantId, req.Module, req.Locale)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to search messages: %v", err))
	}

	response := &localizationv1.SearchMessagesResponse{
		Messages: make([]*localizationv1.Message, len(messages)),
	}

	for i, msg := range messages {
		response.Messages[i] = &localizationv1.Message{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  msg.Module,
			Locale:  msg.Locale,
		}
	}

	return response, nil
}

// CreateMessages implements the CreateMessages gRPC endpoint
func (s *GRPCServer) CreateMessages(ctx context.Context, req *localizationv1.CreateMessagesRequest) (*localizationv1.CreateMessagesResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if len(req.Messages) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one message is required")
	}

	// Convert proto messages to domain messages
	domainMessages := make([]domain.Message, len(req.Messages))
	for i, msg := range req.Messages {
		domainMessages[i] = domain.Message{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  msg.Module,
			Locale:  msg.Locale,
		}
	}

	messages, err := s.service.CreateMessages(ctx, req.TenantId, domainMessages)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create messages: %v", err))
	}

	response := &localizationv1.CreateMessagesResponse{
		Messages: make([]*localizationv1.Message, len(messages)),
	}

	for i, msg := range messages {
		response.Messages[i] = &localizationv1.Message{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  msg.Module,
			Locale:  msg.Locale,
		}
	}

	return response, nil
}

// UpdateMessages implements the UpdateMessages gRPC endpoint
func (s *GRPCServer) UpdateMessages(ctx context.Context, req *localizationv1.UpdateMessagesRequest) (*localizationv1.UpdateMessagesResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if req.Locale == "" {
		return nil, status.Error(codes.InvalidArgument, "locale is required")
	}

	if req.Module == "" {
		return nil, status.Error(codes.InvalidArgument, "module is required")
	}

	if len(req.Messages) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one message is required")
	}

	// Convert proto messages to domain messages
	domainMessages := make([]domain.Message, len(req.Messages))
	for i, msg := range req.Messages {
		domainMessages[i] = domain.Message{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  req.Module,
			Locale:  req.Locale,
		}
	}

	messages, err := s.service.UpdateMessagesForModule(ctx, req.TenantId, req.Locale, req.Module, domainMessages)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update messages: %v", err))
	}

	response := &localizationv1.UpdateMessagesResponse{
		Messages: make([]*localizationv1.Message, len(messages)),
	}

	for i, msg := range messages {
		response.Messages[i] = &localizationv1.Message{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  msg.Module,
			Locale:  msg.Locale,
		}
	}

	return response, nil
}

// UpsertMessages implements the UpsertMessages gRPC endpoint
func (s *GRPCServer) UpsertMessages(ctx context.Context, req *localizationv1.UpsertMessagesRequest) (*localizationv1.UpsertMessagesResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if len(req.Messages) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one message is required")
	}

	// Convert proto messages to domain messages
	domainMessages := make([]domain.Message, len(req.Messages))
	for i, msg := range req.Messages {
		domainMessages[i] = domain.Message{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  msg.Module,
			Locale:  msg.Locale,
		}
	}

	messages, err := s.service.UpsertMessages(ctx, req.TenantId, domainMessages)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to upsert messages: %v", err))
	}

	response := &localizationv1.UpsertMessagesResponse{
		Messages: make([]*localizationv1.Message, len(messages)),
	}

	for i, msg := range messages {
		response.Messages[i] = &localizationv1.Message{
			Code:    msg.Code,
			Message: msg.Message,
			Module:  msg.Module,
			Locale:  msg.Locale,
		}
	}

	return response, nil
}

// DeleteMessages implements the DeleteMessages gRPC endpoint
func (s *GRPCServer) DeleteMessages(ctx context.Context, req *localizationv1.DeleteMessagesRequest) (*localizationv1.DeleteMessagesResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	if len(req.Messages) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one message is required")
	}

	// Convert proto message identities to domain message identities
	messageIdentities := make([]dtos.MessageIdentity, len(req.Messages))
	for i, msg := range req.Messages {
		messageIdentities[i] = dtos.MessageIdentity{
			TenantId: req.TenantId,
			Module:   msg.Module,
			Locale:   msg.Locale,
			Code:     msg.Code,
		}
	}

	err := s.service.DeleteMessages(ctx, messageIdentities)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete messages: %v", err))
	}

	return &localizationv1.DeleteMessagesResponse{
		Success: true,
	}, nil
}

// BustCache implements the BustCache gRPC endpoint
func (s *GRPCServer) BustCache(ctx context.Context, req *localizationv1.BustCacheRequest) (*localizationv1.BustCacheResponse, error) {
	err := s.service.BustCache(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to bust cache: %v", err))
	}

	return &localizationv1.BustCacheResponse{
		Message: "Cache cleared successfully",
		Success: true,
	}, nil
} 