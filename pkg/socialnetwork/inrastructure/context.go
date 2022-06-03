package inrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/redis"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"google.golang.org/grpc/metadata"

	"context"
	"net/http"
)

var UserCtxKey = &ContextKey{"user"}

type ContextKey struct {
	name string
}

func GetUserIDFromContext(r *http.Request) string {
	ctx := r.Context()
	userSession, ok := ctx.Value(UserCtxKey).(*redis.UserSession)
	if !ok {
		return ""
	}
	return userSession.UserID
}

func getGRPCContext(ctx context.Context, r *http.Request) context.Context {
	loggedUserID := GetUserIDFromContext(r)
	requestID := server.GetRequestIDFromContext(r)
	md := metadata.New(map[string]string{request.RequestIDHeader: requestID})
	if loggedUserID != "" {
		md = metadata.New(map[string]string{request.UserIDHeader: loggedUserID, request.RequestIDHeader: requestID})
	}
	return metadata.NewOutgoingContext(ctx, md)
}
