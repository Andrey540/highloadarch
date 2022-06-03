package server

import (
	"context"
	stdlog "log"
	"net/http"
	"strings"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/redis"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/response"
)

const (
	RequestIDKey key = 0
)

type key int

func LoggingMiddleware(log *stdlog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := []byte{}
			defer func() {
				requestID := GetRequestIDFromContext(r)
				log.Println("RequestID: "+requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
				log.Println("Response: " + string(body))
			}()
			rw := &responseWriter{w, body}
			next.ServeHTTP(rw, r)
			body = rw.body
		})
	}
}

func TracingMiddleware(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(request.RequestIDHeader)
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			w.Header().Set(request.RequestIDHeader, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RecoverMiddleware(log *stdlog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Println("panic:", err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func AuthAppMiddleware(sessionService redis.SessionService, userCtxKey interface{}, sessionName, redirectURL string, routeMasks, excludedMasks []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !containsUlr(r.URL.Path, routeMasks) || containsUlr(r.URL.Path, excludedMasks) {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie(sessionName)
			if err != nil || cookie == nil {
				http.Redirect(w, r, redirectURL, http.StatusSeeOther)
				return
			}
			userSession, err := sessionService.GetUserSession(cookie.Value)
			if err != nil {
				response.WriteErrorResponse(err, w)
				return
			}
			if userSession == nil {
				http.Redirect(w, r, redirectURL, http.StatusSeeOther)
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, userSession)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func AuthAPIMiddleware(sessionService redis.SessionService, userCtxKey interface{}, sessionName string, routeMasks, excludedMasks []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !containsUlr(r.URL.Path, routeMasks) || containsUlr(r.URL.Path, excludedMasks) {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie(sessionName)
			if err != nil {
				response.WriteErrorResponse(err, w)
				return
			}
			if cookie == nil {
				response.WriteUnauthorizedResponse("Empty session", w)
				return
			}
			userSession, err := sessionService.GetUserSession(cookie.Value)
			if err != nil {
				response.WriteErrorResponse(err, w)
				return
			}
			if userSession == nil {
				response.WriteUnauthorizedResponse("Session not found", w)
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, userSession)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func containsUlr(url string, masks []string) bool {
	for _, mask := range masks {
		if strings.Contains(url, mask) {
			return true
		}
	}
	return false
}

type responseWriter struct {
	http.ResponseWriter
	body []byte
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	rw.body = body
	return rw.ResponseWriter.Write(body)
}
