package app

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	commandLockName = "command-%s"
)

var ErrCommandAlreadyProcessed = errors.New("Command already processed")
var ErrCommandHandlerNotFound = errors.New("Command handler not found")

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
			processedCommand, err := processedCommandStore.GetCommand(command.CommandID())
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
			err = processedCommandStore.Store(NewProcessedCommand(command.CommandID()))
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
		return result, err
	}
	defer func() {
		err = unitOfWork.Complete(err)
	}()
	result, err = f(unitOfWork)
	return result, err
}

func (handler *handler) getCommandLockName(commandID string) string {
	if commandID == "" {
		return ""
	}

	return fmt.Sprintf(commandLockName, commandID)
}
