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
	case event.UserFriendAddedEvent:
		return NewUserFriendAddedEventHandler(unitOfWork.(UnitOfWork)), nil
	case event.UserFriendRemovedEvent:
		return NewUserFriendRemovedEventHandler(unitOfWork.(UnitOfWork)), nil
	case event.UserCreatedEvent:
		return NewUserCreatedEventHandler(unitOfWork.(UnitOfWork)), nil
	case event.UserRemovedEvent:
		return NewUserRemovedEventHandler(unitOfWork.(UnitOfWork)), nil
	case event.PostCreatedEvent:
		return NewPostCreatedEventHandler(unitOfWork.(UnitOfWork), f.userNotifier, f.errorLogger), nil
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
	serviceFactory := NewServiceFactory(h.unitOfWork)
	userService := serviceFactory.CreateUserService()
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
	serviceFactory := NewServiceFactory(h.unitOfWork)
	userService := serviceFactory.CreateUserService()
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

type userCreatedEventHandler struct {
	unitOfWork UnitOfWork
}

func NewUserCreatedEventHandler(unitOfWork UnitOfWork) app.EventHandler {
	return &userCreatedEventHandler{
		unitOfWork: unitOfWork,
	}
}

func (h userCreatedEventHandler) Handle(currentEvent event.Event) error {
	event1 := currentEvent.(event.UserCreated)
	serviceFactory := NewServiceFactory(h.unitOfWork)
	userService := serviceFactory.CreateUserService()
	userID, err := uuid.FromString(event1.UserID)
	if err != nil {
		return err
	}
	return userService.AddUser(userID, event1.Username)
}

type userRemovedEventHandler struct {
	unitOfWork UnitOfWork
}

func NewUserRemovedEventHandler(unitOfWork UnitOfWork) app.EventHandler {
	return &userRemovedEventHandler{
		unitOfWork: unitOfWork,
	}
}

func (h userRemovedEventHandler) Handle(currentEvent event.Event) error {
	event1 := currentEvent.(event.UserRemoved)
	serviceFactory := NewServiceFactory(h.unitOfWork)
	userService := serviceFactory.CreateUserService()
	userID, err := uuid.FromString(event1.UserID)
	if err != nil {
		return err
	}
	return userService.RemoveUser(userID)
}

type postCreatedEventHandler struct {
	unitOfWork   UnitOfWork
	userNotifier UserNotifier
	errorLogger  *stdlog.Logger
}

func NewPostCreatedEventHandler(unitOfWork UnitOfWork, userNotifier UserNotifier, errorLogger *stdlog.Logger) app.EventHandler {
	return &postCreatedEventHandler{
		unitOfWork:   unitOfWork,
		userNotifier: userNotifier,
		errorLogger:  errorLogger,
	}
}

func (h postCreatedEventHandler) Handle(currentEvent event.Event) error {
	event1 := currentEvent.(event.PostCreated)
	serviceFactory := NewServiceFactory(h.unitOfWork)
	postService := serviceFactory.CreatePostService()
	userService := serviceFactory.CreateUserService()
	postID, err := uuid.FromString(event1.PostID)
	if err != nil {
		return err
	}
	authorID, err := uuid.FromString(event1.AuthorID)
	if err != nil {
		return err
	}
	err = postService.AddNewPost(postID, authorID, event1.Title)
	if err != nil {
		return err
	}
	subscribers, err := h.unitOfWork.UserProvider().ListUserSubscribers(authorID)
	if err != nil {
		h.errorLogger.Println(err)
		return nil
	}
	if len(subscribers) == 0 {
		return nil
	}
	author, err := userService.GetUserName(authorID)
	if err != nil {
		h.errorLogger.Println(err)
		return nil
	}
	if author == "" {
		return nil
	}
	err = h.userNotifier.Notify(subscribers, postID, author, event1.Title)
	if err != nil {
		h.errorLogger.Println(err)
	}
	return nil
}
