package app

import (
	"github.com/callicoder/go-docker/pkg/common/app/event"
)

type storingEventHandler struct {
	eventStore EventStore
	serializer Serializer
}

func (h *storingEventHandler) Handle(e event.Event) error {
	msg, err := h.serializer.Serialize(e)
	if err != nil {
		return err
	}

	storedEvent := NewStoredEvent(h.eventStore.NewUID(), e.EventType(), e.RoutingID(), msg)
	return h.eventStore.Store(storedEvent)
}

func NewStoringHandler(eventStore EventStore, serializer Serializer) event.Handler {
	return &storingEventHandler{eventStore: eventStore, serializer: serializer}
}
