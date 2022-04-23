package command

import (
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/conversation/app"
	"github.com/pkg/errors"
)

type commandHandlerFactory struct {
}

func NewCommandHandlerFactory() commonapp.CommandHandlerFactory {
	return &commandHandlerFactory{}
}

func (f commandHandlerFactory) CreateHandler(unitOfWork commonapp.UnitOfWork, commandType string) (commonapp.CommandHandler, error) {
	switch t := commandType; t {
	case StartConversationCommand:
		return NewStartConversationCommandHandler(unitOfWork.(app.UnitOfWork)), nil
	case AddMessageCommand:
		return NewAddMessageCommandHandler(unitOfWork.(app.UnitOfWork)), nil
	default:
		return nil, nil
	}
}

type startConversationCommandHandler struct {
	unitOfWork app.UnitOfWork
}

func NewStartConversationCommandHandler(unitOfWork app.UnitOfWork) commonapp.CommandHandler {
	return &startConversationCommandHandler{
		unitOfWork: unitOfWork,
	}
}

func (h startConversationCommandHandler) Handle(currentCommand commonapp.Command) (interface{}, error) {
	command1 := currentCommand.(StartUserConversation)
	return executeUnitOfWork(h.unitOfWork, func(service app.ConversationService) (interface{}, error) {
		user, err := uuid.FromString(command1.User)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		target, err := uuid.FromString(command1.Target)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return service.StartUserConversation(user, target)
	}, "")
}

type addMessageCommandHandler struct {
	unitOfWork app.UnitOfWork
}

func NewAddMessageCommandHandler(unitOfWork app.UnitOfWork) commonapp.CommandHandler {
	return &addMessageCommandHandler{
		unitOfWork: unitOfWork,
	}
}

func (h addMessageCommandHandler) Handle(currentCommand commonapp.Command) (interface{}, error) {
	command1 := currentCommand.(AddMessage)
	return executeUnitOfWork(h.unitOfWork, func(service app.ConversationService) (interface{}, error) {
		conversationID, err := uuid.FromString(command1.ConversationID)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		userID, err := uuid.FromString(command1.UserID)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return service.AddMessage(conversationID, userID, command1.Text)
	}, "")
}

func executeUnitOfWork(unitOfWork app.UnitOfWork, f func(service app.ConversationService) (interface{}, error), lockName string) (interface{}, error) {
	lockNames := []string{lockName}
	err := unitOfWork.GetLocks(lockNames)
	if err != nil {
		return nil, err
	}
	serviceFactory := app.NewServiceFactory(unitOfWork)
	return f(serviceFactory.CreateConversationService())
}
