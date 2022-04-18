package inrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/redis"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"

	"net/http"
)

var UserCtxKey = &ContextKey{"user"}

type ContextKey struct {
	name string
}

func getHeaders(r *http.Request) map[string]string {
	loggedUserID := getUserIDFromContext(r)
	headers := make(map[string]string)
	headers[request.UserIDHeader] = loggedUserID
	return headers
}

func getUserIDFromContext(r *http.Request) string {
	ctx := r.Context()
	userSession := ctx.Value(UserCtxKey).(*redis.UserSession)
	return userSession.UserID
}
