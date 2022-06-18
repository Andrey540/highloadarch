package app

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/app/event"
	"github.com/callicoder/go-docker/pkg/common/uuid"

	stdlog "log"
)

type eventHandlerFactory struct {
	errorLogger *stdlog.Logger
}

func NewEventHandlerFactory(errorLogger *stdlog.Logger) app.EventHandlerFactory {
	return &eventHandlerFactory{errorLogger: errorLogger}
}

func (f eventHandlerFactory) CreateHandler(unitOfWork app.UnitOfWork, eventType string) (app.EventHandler, error) {
	switch t := eventType; t {
	case event.MessagesReadEvent:
		return NewMessagesReadEventHandler(unitOfWork.(UnitOfWork), f.errorLogger), nil
	case event.UnreadMessageAddedEvent:
		return NewUnreadMessageAddedEventHandler(unitOfWork.(UnitOfWork), f.errorLogger), nil
	default:
		return nil, nil
	}
}

type messagesReadEventHandler struct {
	unitOfWork  UnitOfWork
	errorLogger *stdlog.Logger
}

func NewMessagesReadEventHandler(unitOfWork UnitOfWork, errorLogger *stdlog.Logger) app.EventHandler {
	return &messagesReadEventHandler{
		unitOfWork:  unitOfWork,
		errorLogger: errorLogger,
	}
}

func (h messagesReadEventHandler) Handle(currentEvent event.Event) error {
	event1 := currentEvent.(event.MessagesRead)
	conversationID, err := uuid.FromString(event1.ConversationID)
	if err != nil {
		return err
	}
	userID, err := uuid.FromString(event1.UserID)
	if err != nil {
		return err
	}
	count := len(event1.MessageIDs)
	unreadMessagesStore := h.unitOfWork.UnreadMessagesStore()
	return unreadMessagesStore.DecreaseUnreadMessages(conversationID, userID, count)
}

type unreadMessageAddedEventHandler struct {
	unitOfWork  UnitOfWork
	errorLogger *stdlog.Logger
}

func NewUnreadMessageAddedEventHandler(unitOfWork UnitOfWork, errorLogger *stdlog.Logger) app.EventHandler {
	return &unreadMessageAddedEventHandler{
		unitOfWork:  unitOfWork,
		errorLogger: errorLogger,
	}
}

func (h unreadMessageAddedEventHandler) Handle(currentEvent event.Event) error {
	event1 := currentEvent.(event.UnreadMessageAdded)
	conversationID, err := uuid.FromString(event1.ConversationID)
	if err != nil {
		return err
	}
	userID, err := uuid.FromString(event1.UserID)
	if err != nil {
		return err
	}
	unreadMessagesStore := h.unitOfWork.UnreadMessagesStore()
	return unreadMessagesStore.IncreaseUnreadMessages(conversationID, userID)
}
