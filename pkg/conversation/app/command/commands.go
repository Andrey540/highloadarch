package command

const (
	StartConversationCommand = "conversation.start"
	AddMessageCommand        = "conversation.add_message"
)

type StartConversation struct {
	ID    string   `json:"id"`
	Users []string `json:"users"`
}

func (command StartConversation) CommandType() string {
	return StartConversationCommand
}

func (command StartConversation) CommandID() string {
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
