package app

import (
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/pkg/errors"
)

func NewCommandsHandler(unitOfWorkFactory UnitOfWorkFactory, commandHandlerFactory CommandHandlerFactory) CommandHandler {
	return &handler{unitOfWorkFactory: unitOfWorkFactory, commandHandlerFactory: commandHandlerFactory}
}

type handler struct {
	unitOfWorkFactory     UnitOfWorkFactory
	commandHandlerFactory CommandHandlerFactory
}

func (handler *handler) Handle(command Command) (interface{}, error) {
	lockName := handler.getCommandLockName(command.CommandID())
	return handler.executeUnitOfWork(func(unitOfWork UnitOfWork) (interface{}, error) {
		processedCommandStore := unitOfWork.ProcessedCommandStore()
		if command.CommandID() != "" {
			commandID, err := uuid.FromString(command.CommandID())
			if err != nil {
				return nil, errors.WithStack(err)
			}
			processedCommand, err := processedCommandStore.GetCommand(commandID)
			if err != nil {
				return nil, err
			}
			if processedCommand != nil {
				return nil, ErrCommandAlreadyProcessed
			}
		}
		commandHandler, err := handler.commandHandlerFactory.CreateHandler(unitOfWork, command.CommandType())
		if err != nil {
			return nil, err
		}
		if commandHandler == nil {
			return nil, ErrCommandHandlerNotFound
		}
		result, err := commandHandler.Handle(command)
		if err != nil {
			return nil, err
		}
		if command.CommandID() != "" {
			commandID, err1 := uuid.FromString(command.CommandID())
			if err1 != nil {
				return nil, errors.WithStack(err1)
			}
			err = processedCommandStore.Store(NewProcessedCommand(commandID))
		}
		return result, err
	}, lockName)
}

func (handler *handler) executeUnitOfWork(f func(UnitOfWork) (interface{}, error), lockName string) (result interface{}, err error) {
	var unitOfWork UnitOfWork
	lockNames := []string{lockName}
	unitOfWork, err = handler.unitOfWorkFactory.NewUnitOfWork(lockNames)
	result = nil
	if err != nil {
		return result, errors.WithStack(err)
	}
	defer func() {
		err = unitOfWork.Complete(err)
	}()
	result, err = f(unitOfWork)
	return result, errors.WithStack(err)
}

func (handler *handler) getCommandLockName(commandID string) string {
	return ""
}
