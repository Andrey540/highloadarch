package conversation

type Conversation struct {
	ConversationID string `json:"conversation_id"`
}

type Message struct {
	MessageID string `json:"message_id"`
}

type MessageData struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	Text           string `json:"text"`
}
