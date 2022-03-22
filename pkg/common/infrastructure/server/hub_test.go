package server

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestServeNormally(t *testing.T) {
	// Runs server until finished
	hub := NewHub(make(chan struct{}))
	pingServe := make(chan bool)
	hub.Serve(func() error {
		<-pingServe
		pingServe <- true
		return nil
	}, func() error {
		return nil
	})

	go func() {
		pingServe <- true
		<-pingServe
	}()

	err := hub.Wait()
	assert.Nil(t, err)
}

func TestFailedServeFirst(t *testing.T) {
	// Stops other servers when fails to serve the one
	hub := NewHub(make(chan struct{}))

	stopServe := make(chan bool)
	stoppedSecond := false
	hub.Serve(func() error {
		<-stopServe
		stoppedSecond = true
		return nil
	}, func() error {
		stopServe <- true
		return nil
	})

	expectedError := errors.New("simulated error")
	hub.Serve(func() error {
		return expectedError
	}, func() error {
		return nil
	})

	err := hub.Wait()
	assert.True(t, stoppedSecond)
	assert.Equal(t, expectedError, errors.Cause(err))
}

func TestNoStopWithoutStart(t *testing.T) {
	hub := NewHub(make(chan struct{}))

	stopCalled := false
	startCalled := false
	expectedError := errors.New("simulated error")

	// Create and wait until Serve returns error
	hub.Serve(func() error {
		return expectedError
	}, func() error {
		return nil
	})

	// Should nor start neither stop second server
	hub.Serve(func() error {
		startCalled = true
		return nil
	}, func() error {
		stopCalled = true
		return nil
	})

	err := hub.Wait()
	assert.Equal(t, startCalled, stopCalled) // called both or none
	assert.Equal(t, expectedError, errors.Cause(err))
}

func TestStopOnSignal(t *testing.T) {
	stopChan := make(chan struct{})
	hub := NewHub(stopChan)

	stopServe := make(chan bool)
	stopCalled := false
	hub.Serve(func() error {
		<-stopServe
		return nil
	}, func() error {
		stopCalled = true
		stopServe <- true
		return nil
	})

	go func() {
		stopChan <- struct{}{}
	}()
	err := hub.Wait()
	assert.True(t, stopCalled)
	assert.Equal(t, ErrStopped, errors.Cause(err))
}

func TestStoppedOnce(t *testing.T) {
	stopChan := make(chan struct{})
	hub := NewHub(stopChan)

	stopServe := make(chan bool)
	stopCalledCounter := 0
	hub.Serve(func() error {
		<-stopServe
		return nil
	}, func() error {
		stopCalledCounter++
		stopServe <- true
		return nil
	})

	expectedError := errors.New("simulated error")
	hub.Serve(func() error {
		return expectedError
	}, func() error {
		return nil
	})

	go func() {
		stopChan <- struct{}{}
	}()
	err := hub.Wait()
	assert.Equal(t, 1, stopCalledCounter)
	err = errors.Cause(err)
	assert.True(t, err == ErrStopped || err == expectedError)
}
