package inrastructure

import (
	"net/http"
	"strings"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/httpclient"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	conversationrequest "github.com/callicoder/go-docker/pkg/common/infrastructure/request/conversation"
	conversationresponse "github.com/callicoder/go-docker/pkg/common/infrastructure/response/conversation"
)

type ConversationService struct {
	baseURL string
	client  httpclient.HTTPClient
}

func (s ConversationService) GetConversationID(r *http.Request) (string, error) {
	var response conversationresponse.Conversation
	headers := getHeaders(r)
	userID := request.GetIDFromRequest(r)
	loggedUserID := getUserIDFromContext(r)
	startConversationRequest := conversationrequest.StartUserConversation{User: loggedUserID, Target: userID}
	err := s.client.MakeJSONRequest(startConversationRequest, &response, http.MethodPost, s.baseURL+conversationrequest.StartConversationURL, headers)
	return response.ConversationID, err
}

func (s ConversationService) ListMessages(r *http.Request, conversationID string) ([]conversationresponse.MessageData, error) {
	var response []conversationresponse.MessageData
	headers := getHeaders(r)
	url := strings.ReplaceAll(s.baseURL+conversationrequest.GetConversationURL, "{id}", conversationID)
	err := s.client.MakeJSONRequest(nil, &response, http.MethodGet, url, headers)
	return response, err
}

func NewConversationService(baseURL string, client httpclient.HTTPClient) ConversationService {
	return ConversationService{
		client:  client,
		baseURL: baseURL,
	}
}
