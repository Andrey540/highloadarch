package inrastructure

import (
	api "github.com/callicoder/go-docker/pkg/common/api"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"context"
	"net/http"
)

type CounterService struct {
	client api.CounterClient
}

func (s CounterService) ListUnreadMessages(r *http.Request) ([]*api.UnreadMessage, error) {
	loggedUserID := GetUserIDFromContext(r)
	req := &api.ListUserUnreadMessagesRequest{
		User: loggedUserID,
	}
	ctx := getGRPCContext(context.Background(), r)
	res, err := s.client.ListUserUnreadMessages(ctx, req)
	if err != nil {
		return []*api.UnreadMessage{}, errors.WithStack(err)
	}
	return res.UnreadMessages, nil
}

func NewCounterService(conn grpc.ClientConnInterface) CounterService {
	return CounterService{
		client: api.NewCounterClient(conn),
	}
}
