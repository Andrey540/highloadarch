package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerStartAfterStop(t *testing.T) {
	stopCalled := false
	startCalled := false

	server := newServer(func() error {
		startCalled = true
		return nil
	}, func() error {
		stopCalled = true
		return nil
	})

	err := server.serve()
	assert.Nil(t, err)
	err = server.stop()
	assert.Nil(t, err)
	err = server.serve()
	assert.Equal(t, errTryRunStoppedServer, err)
	assert.Equal(t, startCalled, stopCalled)
}

func TestServerStartTwiceFail(t *testing.T) {
	stopCalled := false
	startCalled := false

	server := newServer(func() error {
		startCalled = true
		return nil
	}, func() error {
		stopCalled = true
		return nil
	})

	err := server.serve()
	assert.Nil(t, err)
	err = server.serve()
	assert.Equal(t, errAlreadyRun, err)
	assert.Equal(t, startCalled, !stopCalled)
}
