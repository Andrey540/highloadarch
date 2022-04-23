package command

const (
	StartConversationCommand = "conversation.start"
	AddMessageCommand        = "conversation.add_message"
)

type StartUserConversation struct {
	ID     string `json:"id"`
	User   string `json:"user"`
	Target string `json:"target"`
}

func (command StartUserConversation) CommandType() string {
	return StartConversationCommand
}

func (command StartUserConversation) CommandID() string {
	return command.ID
}

type AddMessage struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	Text           string `json:"text"`
}

func (command AddMessage) CommandType() string {
	return AddMessageCommand
}

func (command AddMessage) CommandID() string {
	return command.ID
}
