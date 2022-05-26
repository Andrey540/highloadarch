package app

import (
	"github.com/callicoder/go-docker/pkg/common/uuid"
)

const (
	Created int = iota
	Sent
)

type StoredEvent struct {
	ID        uuid.UUID
	Status    int
	Type      string
	Body      string
	RoutingID string
}

type EventStore interface {
	NewUID() uuid.UUID
	Store(storedEvent StoredEvent) error
	GetCreated() ([]StoredEvent, error)
}

func NewStoredEvent(id uuid.UUID, eventType, routingID, body string) StoredEvent {
	return StoredEvent{
		ID:        id,
		Type:      eventType,
		Body:      body,
		RoutingID: routingID,
	}
}

type Transport interface {
	Send(msgBody string, storedEvent StoredEvent) error
}

type ProcessedEvent struct {
	ID uuid.UUID
}

type ProcessedEventStore interface {
	Store(processedEvent ProcessedEvent) error
	GetEvent(ID uuid.UUID) (*ProcessedEvent, error)
}

func NewProcessedEvent(id uuid.UUID) ProcessedEvent {
	return ProcessedEvent{
		ID: id,
	}
}
