package command

import (
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/pkg/errors"
)

type commandHandlerFactory struct {
}

func NewCommandHandlerFactory() commonapp.CommandHandlerFactory {
	return &commandHandlerFactory{}
}

func (f commandHandlerFactory) CreateHandler(unitOfWork commonapp.UnitOfWork, commandType string) (commonapp.CommandHandler, error) {
	switch t := commandType; t {
	case CreatePostCommand:
		return NewCreatePostCommandHandler(unitOfWork.(app.UnitOfWork)), nil
	default:
		return nil, nil
	}
}

type createPostCommandHandler struct {
	unitOfWork app.UnitOfWork
}

func NewCreatePostCommandHandler(unitOfWork app.UnitOfWork) commonapp.CommandHandler {
	return &createPostCommandHandler{
		unitOfWork: unitOfWork,
	}
}

func (h createPostCommandHandler) Handle(currentCommand commonapp.Command) (interface{}, error) {
	command1 := currentCommand.(CreatePost)
	return executeUnitOfWork(h.unitOfWork, func(service app.PostService) (interface{}, error) {
		authorID, err := uuid.FromString(command1.AuthorID)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return service.CreatePost(authorID, command1.Title, command1.Text)
	}, "")
}

func executeUnitOfWork(unitOfWork app.UnitOfWork, f func(service app.PostService) (interface{}, error), lockName string) (interface{}, error) {
	lockNames := []string{lockName}
	err := unitOfWork.GetLocks(lockNames)
	if err != nil {
		return nil, err
	}
	serviceFactory := app.NewServiceFactory(unitOfWork)
	return f(serviceFactory.CreatePostService())
}
