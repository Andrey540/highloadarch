package app

import (
	"github.com/callicoder/go-docker/pkg/common/app"
)

type eventHandlerFactory struct {
}

func NewEventHandlerFactory() app.EventHandlerFactory {
	return &eventHandlerFactory{}
}

func (f eventHandlerFactory) CreateHandler(unitOfWork app.UnitOfWork, eventType string) (app.EventHandler, error) {
	return nil, nil
}
