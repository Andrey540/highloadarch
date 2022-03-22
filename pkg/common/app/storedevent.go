package app

const (
	Created int = iota
	Sent
)

type StoredEvent struct {
	ID     string
	Status int
	Type   string
	Body   string
}

type EventStore interface {
	NewUID() string
	Store(storedEvent StoredEvent) error
	GetCreated() ([]StoredEvent, error)
}

func NewStoredEvent(id, eventType, body string) StoredEvent {
	return StoredEvent{
		ID:   id,
		Type: eventType,
		Body: body,
	}
}

type Transport interface {
	Send(msgBody string, storedEvent StoredEvent) error
}

type ProcessedEvent struct {
	ID string
}

type ProcessedEventStore interface {
	Store(processedEvent ProcessedEvent) error
	GetEvent(ID string) (*ProcessedEvent, error)
}

func NewProcessedEvent(id string) ProcessedEvent {
	return ProcessedEvent{
		ID: id,
	}
}
