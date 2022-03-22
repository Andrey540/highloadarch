package app

import (
	"github.com/callicoder/go-docker/pkg/common/app/event"
)

type storingEventHandler struct {
	eventStore EventStore
	serializer Serializer
}

func (h *storingEventHandler) Handle(_ event.Event) error {
	return nil
}

func NewStoringHandler(eventStore EventStore, serializer Serializer) event.Handler {
	return &storingEventHandler{eventStore: eventStore, serializer: serializer}
}
