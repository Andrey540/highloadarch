package infrastructure

import (
	"context"

	api "github.com/callicoder/go-docker/pkg/common/api"
	commonserver "github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/counter/app"
)

type server struct {
	queryService app.UnreadMessagesQueryService
}

// NewGRPCServer creates GRPC server which accesses to Service
func NewGRPCServer(queryService app.UnreadMessagesQueryService) api.CounterServer {
	return &server{
		queryService: queryService,
	}
}

func (s *server) ListUserUnreadMessages(ctx context.Context, request *api.ListUserUnreadMessagesRequest) (*api.ListUserUnreadMessagesResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.FromString(commonserver.GetUserIDFromGRPCMetadata(ctx))
	if err != nil {
		return nil, err
	}
	messages, err := s.queryService.ListConversations(userID)
	if err != nil {
		return nil, err
	}
	result := make([]*api.UnreadMessage, 0, len(messages))
	for _, message := range messages {
		result = append(result, &api.UnreadMessage{ConversationID: message.ConversationID.String(), Count: int64(message.Count)})
	}
	return &api.ListUserUnreadMessagesResponse{UnreadMessages: result}, err
}
