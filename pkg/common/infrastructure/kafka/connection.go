package kafka

import (
	"context"
	"fmt"
	stdlog "log"
	"time"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	Host           string
	Port           int
	Partition      int
	Topic          string
	GroupID        string
	ConnectTimeout time.Duration // 0 means default timeout (60 seconds)
}

type Connection interface {
	event.Connection
	Send(key, message string) error
	SetMessageHandler(handler event.Handler)
}

func NewConnection(cfg *Config, logger, errorLogger *stdlog.Logger) Connection {
	return &connection{cfg: cfg, logger: logger, errorLogger: errorLogger}
}

type connection struct {
	cfg         *Config
	kafkaConn   *kafka.Conn
	logger      *stdlog.Logger
	errorLogger *stdlog.Logger
	writer      *kafka.Writer
	reader      *kafka.Reader
	handler     event.Handler
}

func (c *connection) Start() error {
	address := fmt.Sprintf("%s:%v", c.cfg.Host, c.cfg.Port)

	err := backoff.Retry(func() error {
		connection, cErr := kafka.DialLeader(context.Background(), "tcp", address, c.cfg.Topic, c.cfg.Partition)
		c.kafkaConn = connection
		return errors.Wrap(cErr, "failed to connect to kafka")
	}, newBackOff(c.cfg.ConnectTimeout))
	if err != nil {
		return err
	}
	c.writer = &kafka.Writer{
		Addr:     kafka.TCP(address),
		Topic:    c.cfg.Topic,
		Balancer: &kafka.LeastBytes{},
	}
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{address},
		GroupID:  c.cfg.GroupID,
		Topic:    c.cfg.Topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	go func() {
		for {
			if c.handler == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			message, err1 := c.reader.FetchMessage(context.Background())
			if err != nil {
				c.errorLogger.Println("failed to reader:", err1)
				break
			}
			err1 = c.handler.Handle(string(message.Value))
			if err1 != nil {
				c.errorLogger.Println("failed to reader:", err1)
				break
			}

			if err := c.reader.CommitMessages(context.Background(), message); err != nil {
				c.errorLogger.Println("failed to reader:", err1)
				break
			}
		}
	}()

	return nil
}

func (c *connection) Stop() error {
	_ = c.writer.Close()
	_ = c.reader.Close()
	return c.kafkaConn.Close()
}

func (c *connection) Send(key, message string) error {
	return c.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(message),
		},
	)
}

func (c *connection) SetMessageHandler(handler event.Handler) {
	c.handler = handler
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
