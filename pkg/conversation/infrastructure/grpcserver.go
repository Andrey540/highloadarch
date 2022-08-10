package infrastructure

import (
	api "github.com/callicoder/go-docker/pkg/common/api"
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	commonserver "github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/conversation/app"
	"github.com/callicoder/go-docker/pkg/conversation/app/command"
	"github.com/golang/protobuf/ptypes/empty"

	"context"
)

type server struct {
	queryService    app.ConversationQueryService
	commandsHandler commonapp.CommandHandler
}

// NewGRPCServer creates GRPC server which accesses to Service
func NewGRPCServer(queryService app.ConversationQueryService, commandsHandler commonapp.CommandHandler) api.ConversationServer {
	return &server{
		queryService:    queryService,
		commandsHandler: commandsHandler,
	}
}

func (s *server) StartConversation(ctx context.Context, request *api.StartConversationRequest) (*api.StartConversationResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	startConversationCommand := command.StartUserConversation{
		User:   request.User,
		Target: request.Target,
	}
	id, err := s.commandsHandler.Handle(startConversationCommand)
	if err != nil {
		return nil, err
	}
	return &api.StartConversationResponse{ConversationID: id.(uuid.UUID).String()}, err
}

func (s *server) GetConversation(ctx context.Context, request *api.GetConversationRequest) (*api.GetConversationResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	conversationID, err := uuid.FromString(request.ConversationID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.FromString(commonserver.GetUserIDFromGRPCMetadata(ctx))
	if err != nil {
		return nil, err
	}
	conversation, err := s.queryService.GetConversation(conversationID, userID)
	if err != nil {
		return nil, err
	}
	return &api.GetConversationResponse{CompanionID: conversation.UserID.String()}, err
}

func (s *server) ListConversations(ctx context.Context, request *api.ListConversationsRequest) (*api.ListConversationsResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.FromString(request.User)
	if err != nil {
		return nil, err
	}
	conversations, err := s.queryService.ListConversations(userID)
	if err != nil {
		return nil, err
	}
	result := make([]*api.UserConversation, 0, len(conversations))
	for _, conversation := range conversations {
		result = append(result, &api.UserConversation{Id: conversation.ID.String(), UserID: conversation.UserID.String()})
	}
	return &api.ListConversationsResponse{Conversations: result}, nil
}

func (s *server) AddMessage(ctx context.Context, request *api.AddMessageRequest) (*api.AddMessageResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	userID := commonserver.GetUserIDFromGRPCMetadata(ctx)
	addMessageCommand := command.AddMessage{
		ID:             commonserver.GetRequestIDFromGRPCMetadata(ctx),
		ConversationID: request.ConversationID,
		UserID:         userID,
		Text:           request.Text,
	}
	id, err := s.commandsHandler.Handle(addMessageCommand)
	if err != nil {
		return nil, err
	}
	return &api.AddMessageResponse{MessageID: id.(uuid.UUID).String()}, err
}

func (s *server) ReadMessages(ctx context.Context, request *api.ReadMessagesRequest) (*empty.Empty, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	userID := commonserver.GetUserIDFromGRPCMetadata(ctx)
	readMessagesCommand := command.ReadMessages{
		ConversationID: request.ConversationID,
		UserID:         userID,
		Messages:       request.Messages,
	}
	_, err = s.commandsHandler.Handle(readMessagesCommand)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, err
}

func (s *server) ListMessages(ctx context.Context, request *api.ListMessagesRequest) (*api.ListMessagesResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	conversationID, err := uuid.FromString(request.ConversationID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.FromString(commonserver.GetUserIDFromGRPCMetadata(ctx))
	if err != nil {
		return nil, err
	}
	messages, err := s.queryService.ListMessages(userID, conversationID)
	if err != nil {
		return nil, err
	}
	result := make([]*api.Message, 0, len(messages))
	for _, message := range messages {
		result = append(result, &api.Message{Id: message.ID.String(), ConversationID: message.ConversationID.String(), UserID: message.UserID.String(), Text: message.Text, Unread: message.Unread})
	}
	return &api.ListMessagesResponse{Messages: result}, nil
}
