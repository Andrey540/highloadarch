package app

import "github.com/callicoder/go-docker/pkg/common/app/event"

type EventHandler interface {
	Handle(event event.Event) error
}

type EventHandlerFactory interface {
	CreateHandler(unitOfWork UnitOfWork, eventType string) (EventHandler, error)
}
