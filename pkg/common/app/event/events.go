package event

type Event interface {
	EventType() string
}

const (
	UserCreatedEvent = "user.user_created"
	UserUpdatedEvent = "user.user_updated"
	UserRemovedEvent = "user.user_removed"

	ConversationCreatedEvent = "conversation.created"
	MessageAddedEvent        = "conversation.message_added"
)

type UserCreated struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func (event UserCreated) EventType() string {
	return UserCreatedEvent
}

type UserUpdated struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func (event UserUpdated) EventType() string {
	return UserUpdatedEvent
}

type UserRemoved struct {
	UserID string `json:"user_id"`
}

func (event UserRemoved) EventType() string {
	return UserRemovedEvent
}

type ConversationCreated struct {
	ConversationID string   `json:"conversation_id"`
	UserIDs        []string `json:"user_ids"`
}

func (event ConversationCreated) EventType() string {
	return ConversationCreatedEvent
}

type MessageAdded struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	Text           string `json:"text"`
}

func (event MessageAdded) EventType() string {
	return MessageAddedEvent
}
