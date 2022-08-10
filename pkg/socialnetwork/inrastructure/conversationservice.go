package inrastructure

import (
	api "github.com/callicoder/go-docker/pkg/common/api"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"context"
	"net/http"
)

type ConversationService struct {
	client api.ConversationClient
}

func (s ConversationService) ListConversations(r *http.Request) ([]*api.UserConversation, error) {
	loggedUserID := GetUserIDFromContext(r)
	req := &api.ListConversationsRequest{
		User: loggedUserID,
	}
	ctx := getGRPCContext(context.Background(), r)
	res, err := s.client.ListConversations(ctx, req)
	if err != nil {
		return []*api.UserConversation{}, errors.WithStack(err)
	}
	return res.Conversations, nil
}

func (s ConversationService) GetCompanion(r *http.Request, conversationID string) (string, error) {
	req := &api.GetConversationRequest{
		ConversationID: conversationID,
	}
	ctx := getGRPCContext(context.Background(), r)
	res, err := s.client.GetConversation(ctx, req)
	return res.CompanionID, err
}

func (s ConversationService) ListMessages(r *http.Request, conversationID string) ([]*api.Message, error) {
	req := &api.ListMessagesRequest{
		ConversationID: conversationID,
	}
	ctx := getGRPCContext(context.Background(), r)
	res, err := s.client.ListMessages(ctx, req)
	if err != nil {
		return []*api.Message{}, errors.WithStack(err)
	}
	return res.Messages, nil
}

func (s ConversationService) ReadMessages(r *http.Request, conversationID string, messageIDs []string) error {
	req := &api.ReadMessagesRequest{
		ConversationID: conversationID,
		Messages:       messageIDs,
	}
	ctx := getGRPCContext(context.Background(), r)
	_, err := s.client.ReadMessages(ctx, req)
	return errors.WithStack(err)
}

func NewConversationService(conn grpc.ClientConnInterface) ConversationService {
	return ConversationService{
		client: api.NewConversationClient(conn),
	}
}
