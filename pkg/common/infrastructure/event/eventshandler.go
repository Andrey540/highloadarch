package event

import (
	"encoding/json"
	"fmt"

	"github.com/callicoder/go-docker/pkg/common/app"

	stdlog "log"
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
	storedEvent := app.NewStoredEvent("", "", "")
	err := json.Unmarshal([]byte(msg), &storedEvent)
	if err != nil {
		handler.errorLogger.Println(err)
		return err
	}
	lockName := handler.getEventLockName(storedEvent.ID)
	return handler.executeUnitOfWork(func(unitOfWork app.UnitOfWork) error {
		processedEventStore := unitOfWork.ProcessedEventStore()
		processedEvent, err := processedEventStore.GetEvent(storedEvent.ID)
		if err != nil {
			handler.errorLogger.Println(err)
			return err
		}
		if processedEvent != nil {
			handler.logger.Println("Event already processed")
			return nil
		}
		domainEvent, err := handler.serializer.Deserialize(storedEvent.Type, storedEvent.Body)
		if err != nil {
			handler.errorLogger.Println(err)
			return err
		}
		eventHandler, err := handler.eventHandlerFactory.CreateHandler(unitOfWork, storedEvent.Type)
		if err != nil {
			handler.errorLogger.Println(err)
			return err
		}
		if eventHandler != nil {
			err = eventHandler.Handle(domainEvent)
			if err != nil {
				handler.errorLogger.Println(err)
				return err
			}
			handler.logger.Println("Event processed")
		} else {
			handler.logger.Println("Event skipped")
		}

		return processedEventStore.Store(app.NewProcessedEvent(storedEvent.ID))
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