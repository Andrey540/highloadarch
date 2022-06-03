package request

import (
	"github.com/gorilla/mux"

	"net/http"
)

const (
	UserIDHeader    = "X-AuthorID"
	userRoleHeader  = "X-Role"
	RequestIDHeader = "X-Request-Id"
	Customer        = "customer"
	Admin           = "admin"
)

func GetIDFromRequest(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["id"]
}

func GetUserIDFromRequest(r *http.Request) string {
	return r.Header.Get(UserIDHeader)
}

func GetUserRoleFromHeader(r *http.Request) string {
	return r.Header.Get(userRoleHeader)
}

func GetRequestIDFromRequest(r *http.Request) string {
	return r.Header.Get(RequestIDHeader)
}
