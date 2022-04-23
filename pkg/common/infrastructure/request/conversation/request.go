package conversation

const (
	AppID = "conversation"

	urlPrefix = "/" + AppID

	StartConversationURL = urlPrefix + "/api/v1/start"
	AddMessageURL        = urlPrefix + "/api/v1/message/add"
	GetConversationURL   = urlPrefix + "/api/v1/{id}"
)

type StartUserConversation struct {
	User   string `json:"user"`
	Target string `json:"target"`
}

type AddMessage struct {
	ConversationID string `json:"conversationId"`
	Text           string `json:"text"`
}
