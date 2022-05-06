package event

type Event interface {
	EventType() string
}

const (
	UserCreatedEvent = "user.user_created"
	UserUpdatedEvent = "user.user_updated"
	UserRemovedEvent = "user.user_removed"

	UserFriendAddedEvent   = "user.friend_added"
	UserFriendRemovedEvent = "user.friend_removed"

	ConversationCreatedEvent = "conversation.created"
	MessageAddedEvent        = "conversation.message_added"

	PostCreatedEvent = "post.created"
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

type UserFriendAdded struct {
	UserID   string `json:"user_id"`
	FriendID string `json:"friend_id"`
}

func (event UserFriendAdded) EventType() string {
	return UserFriendAddedEvent
}

type UserFriendRemoved struct {
	UserID   string `json:"user_id"`
	FriendID string `json:"friend_id"`
}

func (event UserFriendRemoved) EventType() string {
	return UserFriendRemovedEvent
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

type PostCreated struct {
	PostID   string `json:"post_id"`
	AuthorID string `json:"author_id"`
	Title    string `json:"title"`
	Text     string `json:"text"`
}

func (event PostCreated) EventType() string {
	return PostCreatedEvent
}
