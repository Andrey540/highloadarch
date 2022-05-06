package app

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/app/event"
)

type RepositoryFactory interface {
	PostRepository() PostRepository
	NewsLineStore() NewsLineStore
	UserFriendRepository() UserFriendRepository
	UserProvider() UserProvider
}

type UnitOfWork interface {
	RepositoryFactory
	app.UnitOfWork
}

type ServiceFactory interface {
	CreateUserService() UserService
	CreatePostService() PostService
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
	userFriendRepository := f.unitOfWork.UserFriendRepository()
	return NewUserService(userFriendRepository)
}

func (f serviceFactory) CreatePostService() PostService {
	eventDispatcher := f.createEventDispatcher()
	postRepository := f.unitOfWork.PostRepository()
	newsLineStore := f.unitOfWork.NewsLineStore()
	return NewPostService(postRepository, newsLineStore, eventDispatcher)
}

func (f serviceFactory) createEventDispatcher() event.Dispatcher {
	eventStore := f.unitOfWork.EventStore()
	eventDispatcher := event.NewEventDispatcher()
	storingHandler := app.NewStoringHandler(eventStore, app.NewSerializer())
	eventDispatcher.Subscribe(storingHandler)
	return eventDispatcher
}
