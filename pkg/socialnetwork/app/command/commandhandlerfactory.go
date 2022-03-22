package command

import (
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/socialnetwork/app"
)

type commandHandlerFactory struct {
}

func NewCommandHandlerFactory() commonapp.CommandHandlerFactory {
	return &commandHandlerFactory{}
}

func (f commandHandlerFactory) CreateHandler(unitOfWork commonapp.UnitOfWork, commandType string) (commonapp.CommandHandler, error) {
	switch t := commandType; t {
	case RegisterUserCommand:
		return NewRegisterUserCommandHandler(unitOfWork.(app.UnitOfWork)), nil
	case UpdateUserCommand:
		return NewUpdateUserCommandHandler(unitOfWork.(app.UnitOfWork)), nil
	case RemoveUserCommand:
		return NewRemoveUserCommandHandler(unitOfWork.(app.UnitOfWork)), nil
	case AddUserFriendCommand:
		return NewAddUserFriendCommandHandler(unitOfWork.(app.UnitOfWork)), nil
	case RemoveUserFriendCommand:
		return NewRemoveUserFriendCommandHandler(unitOfWork.(app.UnitOfWork)), nil
	default:
		return nil, nil
	}
}

type registerUserCommandHandler struct {
	unitOfWork app.UnitOfWork
}

func NewRegisterUserCommandHandler(unitOfWork app.UnitOfWork) commonapp.CommandHandler {
	return &registerUserCommandHandler{
		unitOfWork: unitOfWork,
	}
}

func (h registerUserCommandHandler) Handle(currentCommand commonapp.Command) (interface{}, error) {
	command1 := currentCommand.(RegisterUser)
	return executeUnitOfWork(h.unitOfWork, func(service app.UserService) (interface{}, error) {
		return service.RegisterUser(command1.Username, command1.FirstName, command1.LastName, command1.Interests,
			command1.City, command1.Password, command1.Age, command1.Sex)
	}, "")
}

type updateUserCommandHandler struct {
	unitOfWork app.UnitOfWork
}

func NewUpdateUserCommandHandler(unitOfWork app.UnitOfWork) commonapp.CommandHandler {
	return &updateUserCommandHandler{
		unitOfWork: unitOfWork,
	}
}

func (h updateUserCommandHandler) Handle(currentCommand commonapp.Command) (interface{}, error) {
	command1 := currentCommand.(UpdateUser)
	return executeUnitOfWork(h.unitOfWork, func(service app.UserService) (interface{}, error) {
		userID, err := uuid.FromString(command1.UserID)
		if err != nil {
			return nil, err
		}
		return service.UpdateUser(userID, command1.Username, command1.FirstName, command1.LastName, command1.Interests,
			command1.City, command1.Password, command1.Age, command1.Sex)
	}, getUserLockName(command1.UserID))
}

type removeUserCommandHandler struct {
	unitOfWork app.UnitOfWork
}

func NewRemoveUserCommandHandler(unitOfWork app.UnitOfWork) commonapp.CommandHandler {
	return &removeUserCommandHandler{
		unitOfWork: unitOfWork,
	}
}

func (h removeUserCommandHandler) Handle(currentCommand commonapp.Command) (interface{}, error) {
	command1 := currentCommand.(RemoveUser)
	return executeUnitOfWork(h.unitOfWork, func(service app.UserService) (interface{}, error) {
		userID, err := uuid.FromString(command1.UserID)
		if err != nil {
			return nil, err
		}
		err = service.DeleteUser(userID)
		return nil, err
	}, getUserLockName(command1.UserID))
}

type addUserFriendCommandHandler struct {
	unitOfWork app.UnitOfWork
}

func NewAddUserFriendCommandHandler(unitOfWork app.UnitOfWork) commonapp.CommandHandler {
	return &addUserFriendCommandHandler{
		unitOfWork: unitOfWork,
	}
}

func (h addUserFriendCommandHandler) Handle(currentCommand commonapp.Command) (interface{}, error) {
	command1 := currentCommand.(AddUserFriend)
	return executeUnitOfWork(h.unitOfWork, func(service app.UserService) (interface{}, error) {
		userID, err := uuid.FromString(command1.UserID)
		if err != nil {
			return nil, err
		}
		friendID, err := uuid.FromString(command1.FriendID)
		if err != nil {
			return nil, err
		}
		err = service.AddUserFriend(userID, friendID)
		return nil, err
	}, getUserLockName(command1.UserID))
}

type removeUserFriendCommandHandler struct {
	unitOfWork app.UnitOfWork
}

func NewRemoveUserFriendCommandHandler(unitOfWork app.UnitOfWork) commonapp.CommandHandler {
	return &removeUserFriendCommandHandler{
		unitOfWork: unitOfWork,
	}
}

func (h removeUserFriendCommandHandler) Handle(currentCommand commonapp.Command) (interface{}, error) {
	command1 := currentCommand.(RemoveUserFriend)
	return executeUnitOfWork(h.unitOfWork, func(service app.UserService) (interface{}, error) {
		userID, err := uuid.FromString(command1.UserID)
		if err != nil {
			return nil, err
		}
		friendID, err := uuid.FromString(command1.FriendID)
		if err != nil {
			return nil, err
		}
		err = service.RemoveUserFriend(userID, friendID)
		return nil, err
	}, getUserLockName(command1.UserID))
}

func executeUnitOfWork(unitOfWork app.UnitOfWork, f func(service app.UserService) (interface{}, error), lockName string) (interface{}, error) {
	lockNames := []string{lockName}
	err := unitOfWork.GetLocks(lockNames)
	if err != nil {
		return nil, err
	}
	serviceFactory := app.NewServiceFactory(unitOfWork)
	return f(serviceFactory.CreateUserService())
}

func getUserLockName(userID string) string {
	return ""
}
