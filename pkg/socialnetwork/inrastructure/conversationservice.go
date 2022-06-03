package inrastructure

import (
	api "github.com/callicoder/go-docker/pkg/common/api"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	"google.golang.org/grpc"

	"context"
	"net/http"
)

type ConversationService struct {
	client api.ConversationClient
}

func (s ConversationService) GetConversationID(r *http.Request) (string, error) {
	userID := request.GetIDFromRequest(r)
	loggedUserID := GetUserIDFromContext(r)
	req := &api.StartConversationRequest{
		User:   loggedUserID,
		Target: userID,
	}
	ctx := getGRPCContext(context.Background(), r)
	res, err := s.client.StartConversation(ctx, req)
	if err != nil {
		return "", err
	}
	return res.ConversationID, nil
}

func (s ConversationService) ListMessages(r *http.Request, conversationID string) ([]*api.Message, error) {
	req := &api.ListMessagesRequest{
		ConversationID: conversationID,
	}
	ctx := getGRPCContext(context.Background(), r)
	res, err := s.client.ListMessages(ctx, req)
	if err != nil {
		return []*api.Message{}, err
	}
	return res.Messages, nil
}

func NewConversationService(conn grpc.ClientConnInterface) ConversationService {
	return ConversationService{
		client: api.NewConversationClient(conn),
	}
}
