package amqp

import (
	stderrors "errors"
	"fmt"
	stdlog "log"
	"time"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"
	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Config struct {
	User            string
	Password        string
	Host            string
	QueueName       string
	WorkersCount    int
	RoutingKey      string
	SuppressReading bool
	ConnectTimeout  time.Duration // 0 means default timeout (60 seconds)
}

type Connection interface {
	event.Connection
	AddChannel(channel Channel)
}

type Channel interface {
	Connect(conn *amqp.Connection) error
}

func NewAMQPConnection(cfg *Config, logger, errorLogger *stdlog.Logger) Connection {
	return &connection{cfg: cfg, logger: logger, errorLogger: errorLogger}
}

var (
	errNilAMQPConnection    = stderrors.New("amqp connection is empty")
	errClosedAMQPConnection = stderrors.New("amqp connection is closed")
)

type connection struct {
	cfg         *Config
	amqpConn    *amqp.Connection
	logger      *stdlog.Logger
	errorLogger *stdlog.Logger
	channels    []Channel
}

func (c *connection) Start() error {
	url := fmt.Sprintf("amqp://%s:%s@%s/", c.cfg.User, c.cfg.Password, c.cfg.Host)

	err := backoff.Retry(func() error {
		connection, cErr := amqp.Dial(url)
		c.amqpConn = connection
		return errors.Wrap(cErr, "failed to connect to amqp")
	}, newBackOff(c.cfg.ConnectTimeout))

	if err == nil {
		if err = c.validateConnection(c.amqpConn); err != nil {
			return err
		}

		for _, channel := range c.channels {
			if err = channel.Connect(c.amqpConn); err != nil {
				return err
			}
		}

		connErrorChan := c.amqpConn.NotifyClose(make(chan *amqp.Error))
		go c.processConnectErrors(connErrorChan)
	}
	return err
}

func (c *connection) Stop() error {
	return c.amqpConn.Close()
}

func (c *connection) AddChannel(channel Channel) {
	c.channels = append(c.channels, channel)
}

func (c *connection) validateConnection(conn *amqp.Connection) error {
	if conn == nil {
		return errors.WithStack(errNilAMQPConnection)
	}
	if conn.IsClosed() {
		return errors.WithStack(errClosedAMQPConnection)
	}
	return nil
}

func (c *connection) processConnectErrors(ch chan *amqp.Error) {
	err := <-ch
	if err == nil {
		return
	}

	c.errorLogger.Println(err, "AMQP connection error, trying to reconnect")
	for {
		err := c.Start()
		if err == nil {
			c.logger.Println("AMQP connection restored")
			break
		} else {
			c.errorLogger.Println(err, "failed to reconnect to AMQP")
		}
	}
}

func newBackOff(timeout time.Duration) backoff.BackOff {
	exponentialBackOff := backoff.NewExponentialBackOff()
	const defaultTimeout = 60 * time.Second
	if timeout != 0 {
		exponentialBackOff.MaxElapsedTime = timeout
	} else {
		exponentialBackOff.MaxElapsedTime = defaultTimeout
	}
	exponentialBackOff.MaxInterval = 5 * time.Second
	return exponentialBackOff
}
