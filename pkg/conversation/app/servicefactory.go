package app

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/app/event"
)

type RepositoryFactory interface {
	ConversationRepository() ConversationRepository
	MessageRepository() MessageRepository
}

type UnitOfWork interface {
	RepositoryFactory
	app.UnitOfWork
}

type ServiceFactory interface {
	CreateConversationService() ConversationService
}

type serviceFactory struct {
	unitOfWork UnitOfWork
}

func NewServiceFactory(unitOfWork UnitOfWork) ServiceFactory {
	return &serviceFactory{
		unitOfWork: unitOfWork,
	}
}

func (f serviceFactory) CreateConversationService() ConversationService {
	eventDispatcher := f.createEventDispatcher()
	conversationRepository := f.unitOfWork.ConversationRepository()
	messageRepository := f.unitOfWork.MessageRepository()
	return NewConversationService(conversationRepository, messageRepository, eventDispatcher)
}

func (f serviceFactory) createEventDispatcher() event.Dispatcher {
	eventStore := f.unitOfWork.EventStore()
	eventDispatcher := event.NewEventDispatcher()
	storingHandler := app.NewStoringHandler(eventStore, app.NewSerializer())
	eventDispatcher.Subscribe(storingHandler)
	return eventDispatcher
}
