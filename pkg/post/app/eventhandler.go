package app

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/app/event"
	"github.com/callicoder/go-docker/pkg/common/uuid"
)

type eventHandlerFactory struct {
	newsLineCache NewsLineCache
}

func NewEventHandlerFactory(newsLineCache NewsLineCache) app.EventHandlerFactory {
	return &eventHandlerFactory{newsLineCache: newsLineCache}
}

func (f eventHandlerFactory) CreateHandler(unitOfWork app.UnitOfWork, eventType string) (app.EventHandler, error) {
	switch t := eventType; t {
	case event.UserFriendAddedEvent:
		return NewUserFriendAddedEventHandler(unitOfWork.(UnitOfWork)), nil
	case event.UserFriendRemovedEvent:
		return NewUserFriendRemovedEventHandler(unitOfWork.(UnitOfWork)), nil
	case event.PostCreatedEvent:
		return NewPostCreatedEventHandler(unitOfWork.(UnitOfWork), f.newsLineCache), nil
	default:
		return nil, nil
	}
}

type userFriendAddedEventHandler struct {
	unitOfWork UnitOfWork
}

func NewUserFriendAddedEventHandler(unitOfWork UnitOfWork) app.EventHandler {
	return &userFriendAddedEventHandler{
		unitOfWork: unitOfWork,
	}
}

func (h userFriendAddedEventHandler) Handle(currentEvent event.Event) error {
	event1 := currentEvent.(event.UserFriendAdded)
	userService := NewUserService(h.unitOfWork.UserFriendRepository())
	userID, err := uuid.FromString(event1.UserID)
	if err != nil {
		return err
	}
	friendID, err := uuid.FromString(event1.FriendID)
	if err != nil {
		return err
	}
	return userService.AddUserFriend(userID, friendID)
}

type userFriendRemovedEventHandler struct {
	unitOfWork UnitOfWork
}

func NewUserFriendRemovedEventHandler(unitOfWork UnitOfWork) app.EventHandler {
	return &userFriendRemovedEventHandler{
		unitOfWork: unitOfWork,
	}
}

func (h userFriendRemovedEventHandler) Handle(currentEvent event.Event) error {
	event1 := currentEvent.(event.UserFriendRemoved)
	userService := NewUserService(h.unitOfWork.UserFriendRepository())
	userID, err := uuid.FromString(event1.UserID)
	if err != nil {
		return err
	}
	friendID, err := uuid.FromString(event1.FriendID)
	if err != nil {
		return err
	}
	return userService.RemoveUserFriend(userID, friendID)
}

type postCreatedEventHandler struct {
	unitOfWork    UnitOfWork
	newsLineCache NewsLineCache
}

func NewPostCreatedEventHandler(unitOfWork UnitOfWork, newsLineCache NewsLineCache) app.EventHandler {
	return &postCreatedEventHandler{
		unitOfWork:    unitOfWork,
		newsLineCache: newsLineCache,
	}
}

func (h postCreatedEventHandler) Handle(currentEvent event.Event) error {
	event1 := currentEvent.(event.PostCreated)
	serviceFactory := NewServiceFactory(h.unitOfWork)
	postService := serviceFactory.CreatePostService()
	postID, err := uuid.FromString(event1.PostID)
	if err != nil {
		return err
	}
	authorID, err := uuid.FromString(event1.AuthorID)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	err = postService.AddNewPost(postID, authorID, event1.Title)
	if err != nil {
		return err
	}
	return nil
}
