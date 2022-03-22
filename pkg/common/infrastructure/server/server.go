package server

import (
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/amqp"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/kafka"

	stdlog "log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/pkg/errors"
)

// ServeFunc - runs server
type ServeFunc func() error

// StopFunc - stops server
type StopFunc func() error

const (
	serverIsCreated int32 = iota
	serverIsRunning
	serverIsStopped
)

type server struct {
	serveFunc ServeFunc
	stopFunc  StopFunc
	state     int32
}

func newServer(serve ServeFunc, stop StopFunc) *server {
	return &server{
		serveFunc: serve,
		stopFunc:  stop,
		state:     serverIsCreated,
	}
}

func (s *server) serve() error {
	if !atomic.CompareAndSwapInt32(&s.state, serverIsCreated, serverIsRunning) {
		if atomic.LoadInt32(&s.state) == serverIsRunning {
			return errAlreadyRun
		}
		return errTryRunStoppedServer
	}
	return s.serveFunc()
}

func (s *server) stop() error {
	stopped := atomic.CompareAndSwapInt32(&s.state, serverIsCreated, hubIsStopped) ||
		atomic.CompareAndSwapInt32(&s.state, serverIsRunning, serverIsStopped)

	if !stopped {
		return errAlreadyStopped
	}
	return s.stopFunc()
}

func ListenOSKillSignals(stopChan chan<- struct{}) {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		<-ch
		stopChan <- struct{}{}
	}()
}

func InitLogger() *stdlog.Logger {
	return stdlog.New(os.Stdout, "http: ", stdlog.LstdFlags)
}

func InitErrorLogger() *stdlog.Logger {
	return stdlog.New(os.Stderr, "http: ", stdlog.LstdFlags)
}

func InitEventTransport(kafkaCnf *kafka.Config, amqpCnf *amqp.Config, logger, errorLogger *stdlog.Logger) ([]commonapp.Transport, []event.Connection, error) {
	var transports []commonapp.Transport
	var connections []event.Connection
	if kafkaCnf != nil {
		transport, connection, err := kafka.CreateTransport(*kafkaCnf, logger, errorLogger)
		if err != nil {
			return nil, nil, err
		}
		transports = append(transports, transport)
		connections = append(connections, connection)
	}
	if amqpCnf != nil {
		transport, connection, err := amqp.CreateTransport(*amqpCnf, logger, errorLogger)
		if err != nil {
			return nil, nil, err
		}
		transports = append(transports, transport)
		connections = append(connections, connection)
	}
	return transports, connections, nil
}

var errAlreadyStopped = errors.New("server is not running, can't change server state")
var errAlreadyRun = errors.New("server is running, can't change server state to running")
var errTryRunStoppedServer = errors.New("server is stopped, can't change server state to running")
