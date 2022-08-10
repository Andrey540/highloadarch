package app

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/app/event"
	"github.com/callicoder/go-docker/pkg/common/uuid"

	stdlog "log"
)

type eventHandlerFactory struct {
	userNotifier UserNotifier
	errorLogger  *stdlog.Logger
}

func NewEventHandlerFactory(userNotifier UserNotifier, errorLogger *stdlog.Logger) app.EventHandlerFactory {
	return &eventHandlerFactory{userNotifier: userNotifier, errorLogger: errorLogger}
}

func (f eventHandlerFactory) CreateHandler(unitOfWork app.UnitOfWork, eventType string) (app.EventHandler, error) {
	switch t := eventType; t {
	case event.MessageAddedEvent:
		return NewMessageAddedEventHandler(unitOfWork.(UnitOfWork), f.userNotifier, f.errorLogger), nil
	default:
		return nil, nil
	}
}

type messageAddedEventHandler struct {
	unitOfWork   UnitOfWork
	userNotifier UserNotifier
	errorLogger  *stdlog.Logger
}

func NewMessageAddedEventHandler(unitOfWork UnitOfWork, userNotifier UserNotifier, errorLogger *stdlog.Logger) app.EventHandler {
	return &messageAddedEventHandler{
		unitOfWork:   unitOfWork,
		userNotifier: userNotifier,
		errorLogger:  errorLogger,
	}
}

func (h messageAddedEventHandler) Handle(currentEvent event.Event) error {
	event1 := currentEvent.(event.MessageAdded)
	userIDs, err := uuid.FromStrings(event1.UserIDs)
	if err != nil {
		return err
	}
	conversationID, err := uuid.FromString(event1.ConversationID)
	if err != nil {
		return err
	}
	messageID, err := uuid.FromString(event1.MessageID)
	if err != nil {
		return err
	}
	authorID, err := uuid.FromString(event1.AuthorID)
	if err != nil {
		return err
	}
	err = h.userNotifier.Notify(userIDs, conversationID, messageID, authorID, event1.Text)
	if err != nil {
		h.errorLogger.Println(err)
	}
	return nil
}
