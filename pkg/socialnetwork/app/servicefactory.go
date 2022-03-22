package app

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/app/event"
)

type RepositoryFactory interface {
	UserRepository() UserRepository
	UserFriendRepository() UserFriendRepository
}

type UnitOfWork interface {
	RepositoryFactory
	app.UnitOfWork
}

type ServiceFactory interface {
	CreateUserService() UserService
}

type serviceFactory struct {
	unitOfWork UnitOfWork
}

func NewServiceFactory(unitOfWork UnitOfWork) ServiceFactory {
	return &serviceFactory{
		unitOfWork: unitOfWork,
	}
}

func (f serviceFactory) CreateUserService() UserService {
	eventDispatcher := f.createEventDispatcher()
	userRepository := f.unitOfWork.UserRepository()
	userFriendRepository := f.unitOfWork.UserFriendRepository()
	return NewUserService(userRepository, userFriendRepository, eventDispatcher)
}

func (f serviceFactory) createEventDispatcher() event.Dispatcher {
	eventStore := f.unitOfWork.EventStore()
	eventDispatcher := event.NewEventDispatcher()
	storingHandler := app.NewStoringHandler(eventStore, app.NewSerializer())
	eventDispatcher.Subscribe(storingHandler)
	return eventDispatcher
}
