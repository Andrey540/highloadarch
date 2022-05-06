package event

import (
	"encoding/json"
	"fmt"
	stdlog "log"

	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/pkg/errors"
)

const (
	eventLockName = "event-%s"
)

func NewEventsHandler(unitOfWorkFactory app.UnitOfWorkFactory, eventHandlerFactory app.EventHandlerFactory, logger, errorLogger *stdlog.Logger) Handler {
	return &handler{unitOfWorkFactory: unitOfWorkFactory, eventHandlerFactory: eventHandlerFactory, serializer: app.NewSerializer(), logger: logger, errorLogger: errorLogger}
}

type handler struct {
	unitOfWorkFactory   app.UnitOfWorkFactory
	eventHandlerFactory app.EventHandlerFactory
	serializer          app.Serializer
	logger              *stdlog.Logger
	errorLogger         *stdlog.Logger
}

func (handler *handler) Handle(msg string) error {
	handler.logger.Println("Event received: " + msg)
	storedEvent := app.NewStoredEvent(uuid.Nil, "", "")
	err := json.Unmarshal([]byte(msg), &storedEvent)
	if err != nil {
		handler.errorLogger.Println(err)
		return errors.WithStack(err)
	}
	lockID := ""
	if storedEvent.ID != uuid.Nil {
		lockID = storedEvent.ID.String()
	}
	lockName := handler.getEventLockName(lockID)
	return handler.executeUnitOfWork(func(unitOfWork app.UnitOfWork) error {
		processedEventStore := unitOfWork.ProcessedEventStore()
		processedEvent, err := processedEventStore.GetEvent(storedEvent.ID)
		if err != nil {
			handler.errorLogger.Println(err)
			return errors.WithStack(err)
		}
		if processedEvent != nil {
			handler.logger.Println("Event already processed")
			return nil
		}
		domainEvent, err := handler.serializer.Deserialize(storedEvent.Type, storedEvent.Body)
		if err != nil {
			handler.errorLogger.Println(err)
			return errors.WithStack(err)
		}
		eventHandler, err := handler.eventHandlerFactory.CreateHandler(unitOfWork, storedEvent.Type)
		if err != nil {
			handler.errorLogger.Println(err)
			return errors.WithStack(err)
		}
		if eventHandler != nil {
			err = eventHandler.Handle(domainEvent)
			if err != nil {
				handler.errorLogger.Println(err)
				return errors.WithStack(err)
			}
			handler.logger.Println("Event processed")
		} else {
			handler.logger.Println("Event skipped")
		}

		err = processedEventStore.Store(app.NewProcessedEvent(storedEvent.ID))
		return errors.WithStack(err)
	}, lockName)
}

func (handler *handler) executeUnitOfWork(f func(app.UnitOfWork) error, lockName string) (err error) {
	var unitOfWork app.UnitOfWork
	lockNames := []string{lockName}
	unitOfWork, err = handler.unitOfWorkFactory.NewUnitOfWork(lockNames)
	if err != nil {
		return err
	}
	defer func() {
		err = unitOfWork.Complete(err)
	}()
	err = f(unitOfWork)
	return err
}

func (handler *handler) getEventLockName(eventID string) string {
	if eventID == "" {
		return ""
	}

	return fmt.Sprintf(eventLockName, eventID)
}
