package app

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

const (
	dispatchDelay       = 1 * time.Second
	eventDispatcherLock = "event_dispatcher"
)

var ErrEmptyTransport = errors.New("Not enough money")

type StoredEventDispatcher interface {
	Activate()
	Start()
	Stop()
}

type storedEventDispatcher struct {
	unitOfWorkFactory UnitOfWorkFactory
	transports        []Transport
	errorsChan        chan error
	controlChan       chan struct{}
	waitGroup         sync.WaitGroup
	dispatchEvents    int32
}

func (dispatcher *storedEventDispatcher) Start() {
	ticker := time.NewTicker(dispatchDelay)

	go func() {
		for {
			select {
			case <-ticker.C:
				dispatchEvents := atomic.LoadInt32(&dispatcher.dispatchEvents)
				if dispatchEvents > 0 {
					err := dispatcher.dispatchNewEvents(dispatchEvents)
					if err != nil {
						dispatcher.errorsChan <- err
					}
				}
			case <-dispatcher.controlChan:
				dispatcher.waitGroup.Done()
				return
			}
		}
	}()
}

func (dispatcher *storedEventDispatcher) Activate() {
	done := false
	for !done {
		dispatchEvents := dispatcher.dispatchEvents
		done = atomic.CompareAndSwapInt32(&dispatcher.dispatchEvents, dispatchEvents, dispatchEvents+1)
	}
}

func (dispatcher *storedEventDispatcher) Stop() {
	dispatcher.controlChan <- struct{}{}
	dispatcher.waitGroup.Wait()
}

func (dispatcher *storedEventDispatcher) dispatchNewEvents(dispatchEvents int32) (err error) {
	err = dispatcher.executeUnitOfWork(func(eventStore EventStore) error {
		allDispatched := true

		events, err2 := eventStore.GetCreated()
		if err2 != nil {
			return err2
		}

		if len(events) == 0 {
			return nil
		}

		if len(dispatcher.transports) == 0 {
			return ErrEmptyTransport
		}

		for _, storedEvent := range events {
			payload, err2 := dispatcher.serializeEvent(storedEvent)
			if err2 != nil {
				allDispatched = false
				dispatcher.errorsChan <- err2
				break
			}
			for _, transport := range dispatcher.transports {
				err = transport.Send(payload, storedEvent)
				if err != nil {
					break
				}
			}

			if err != nil {
				allDispatched = false
				dispatcher.errorsChan <- err
				break
			}

			storedEvent.Status = Sent
			err = eventStore.Store(storedEvent)
			if err != nil {
				allDispatched = false
				dispatcher.errorsChan <- err
			}
		}

		if allDispatched {
			atomic.CompareAndSwapInt32(&dispatcher.dispatchEvents, dispatchEvents, 0)
		}

		return err
	}, eventDispatcherLock)
	return err
}

func (dispatcher *storedEventDispatcher) serializeEvent(storedEvent StoredEvent) (string, error) {
	payload, err := json.Marshal(storedEvent)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func (dispatcher *storedEventDispatcher) executeUnitOfWork(f func(eventStore EventStore) error, lockName string) (err error) {
	var unitOfWork UnitOfWork
	lockNames := []string{lockName}
	unitOfWork, err = dispatcher.unitOfWorkFactory.NewUnitOfWork(lockNames)
	if err != nil {
		return err
	}
	defer func() {
		err = unitOfWork.Complete(err)
	}()
	err = f(unitOfWork.EventStore())
	return err
}

func NewStoredEventDispatcher(unitOfWorkFactory UnitOfWorkFactory, transports []Transport, errorsChan chan error) StoredEventDispatcher {
	controlChan := make(chan struct{})

	p := &storedEventDispatcher{
		unitOfWorkFactory: unitOfWorkFactory,
		transports:        transports,
		errorsChan:        errorsChan,
		controlChan:       controlChan,
		dispatchEvents:    1,
	}
	p.waitGroup.Add(1)
	return p
}
