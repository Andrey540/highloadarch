package server

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	hubIsCreated int32 = iota
	hubIsRunning
	hubIsStopped
)

// Hub of servers that can run and stop together,
//  runs all servers until receive SIGINT/SIGTERM,
//  stops all if cannot run any one.
type Hub struct {
	state           int32
	wg              sync.WaitGroup
	reportErrorOnce sync.Once
	errs            chan error
	stopChan        chan struct{}
	stoppers        []StopFunc
	startServerChan chan *server
}

// NewHub - creates new servers group and starts to listen OS signals SIGINT/SIGTERM
func NewHub(stopChan chan struct{}) *Hub {
	startServerChan := make(chan *server)
	h := &Hub{
		state:           hubIsCreated,
		errs:            make(chan error, 1), // intentionally buffered
		stopChan:        stopChan,
		startServerChan: startServerChan,
	}

	go func() {
		for server := range startServerChan {
			h.startServer(server)
		}
	}()

	return h
}

// Serve - registers new server and it's stop function
func (h *Hub) Serve(serve ServeFunc, stop StopFunc) {
	h.startServerChan <- newServer(serve, stop)
}

// Wait - waits until all servers completed
// If one of the servers generates error, stops all servers and returns first error
func (h *Hub) Wait() error {
	var err error

	// Wait for error or stopChan message and stop all servers
	h.wg.Add(1)
	go func() {
		select {
		case err = <-h.errs:
			_ = h.stop()
		case <-h.stopChan:
			err = h.stop()
			if err == nil {
				err = ErrStopped
			}
		}
		h.wg.Done()
	}()

	// Wait until all goroutines finished
	h.wg.Wait()
	return errors.WithStack(err)
}

// Serve - registers new server and it's stop function
func (h *Hub) startServer(s *server) {
	started := atomic.CompareAndSwapInt32(&h.state, hubIsCreated, hubIsRunning)

	h.wg.Add(1)
	if !started && atomic.LoadInt32(&h.state) == hubIsStopped {
		// There is no reason to serve if hub is stopped.
		h.wg.Done()
		return
	}

	h.stoppers = append(h.stoppers, s.stop)
	go func() {
		h.run(s.serve)
		h.wg.Done()
	}()
}

func (h *Hub) run(serve ServeFunc) {
	defer func() {
		h.recoverReportError()
	}()
	err := serve()
	h.reportError(err)
}

// Stop all servers, store first error
func (h *Hub) stop() error {
	stopped := atomic.CompareAndSwapInt32(&h.state, hubIsCreated, hubIsStopped) ||
		atomic.CompareAndSwapInt32(&h.state, hubIsRunning, hubIsStopped)
	if !stopped {
		return nil
	}

	var err error
	for _, stop := range h.stoppers {
		stopErr := stop()
		if err == nil && stopErr != errAlreadyStopped {
			err = stopErr
		}
	}
	return errors.WithStack(err)
}

// Recovers and reports error if there was a panic
func (h *Hub) recoverReportError() {
	if value := recover(); value != nil {
		switch x := value.(type) {
		case error:
			h.reportError(x)
		case string:
			h.reportError(fmt.Errorf("%s", x))
		default:
			h.reportError(fmt.Errorf("%v", value))
		}
	}
}

// Saves error if it's first
func (h *Hub) reportError(err error) {
	// errs channel is not listened for stopped hub. For all other cases error is unexpected even nil error.
	if atomic.LoadInt32(&h.state) == hubIsStopped {
		return
	}
	h.reportErrorOnce.Do(func() {
		h.errs <- err
	})
}

var ErrStopped = errors.New("hub is stopped by signal without errors")
