package event

type Event interface {
	EventType() string
}

const (
	UserCreatedEvent = "user.user_created"
	UserUpdatedEvent = "user.user_updated"
	UserRemovedEvent = "user.user_removed"
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
