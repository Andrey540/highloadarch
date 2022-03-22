package request

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/gorilla/mux"

	"net/http"
)

const (
	userIDHeader   = "X-UserID"
	userRoleHeader = "X-Role"
	Customer       = "customer"
	Admin          = "admin"
)

func GetIDFromRequest(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["id"]
}

func GetUserIDFromHeader(r *http.Request) string {
	return r.Header.Get(userIDHeader)
}

func GetUserRoleFromHeader(r *http.Request) string {
	return r.Header.Get(userRoleHeader)
}

func GetRequestIDFromRequest(r *http.Request) string {
	return r.Header.Get(server.RequestIDHeader)
}
