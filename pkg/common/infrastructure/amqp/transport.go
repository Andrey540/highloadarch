package amqp

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"
	"github.com/streadway/amqp"

	stdlog "log"
	"time"
)

const (
	domainEventsExchangeName = "domain_event"
	domainEventsExchangeType = "topic"
	routingKey               = "#"
	routingPrefix            = ""
	contentType              = "application/json; charset=utf-8"
)

type Transport interface {
	event.Transport
	Channel
}

type transport struct {
	conn                  *amqp.Connection
	writeChannel          *amqp.Channel
	handler               event.Handler
	queueName             string
	suppressEventsReading bool
}

func (t *transport) Send(msgBody string, storedEvent app.StoredEvent) error {
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  contentType,
		Body:         []byte(msgBody),
	}
	routingKey := routingPrefix + storedEvent.Type
	return t.writeChannel.Publish(domainEventsExchangeName, routingKey, false, false, msg)
}

func (t *transport) Connect(conn *amqp.Connection) error {
	t.writeChannel = nil

	t.conn = conn

	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	t.writeChannel = channel

	err = channel.ExchangeDeclare(domainEventsExchangeName, domainEventsExchangeType, true, false, false, false, nil)
	if err != nil {
		return err
	}

	if !t.suppressEventsReading {
		return t.connectReadChannel(err, channel)
	}
	return nil
}

func (t *transport) connectReadChannel(err error, channel *amqp.Channel) error {
	readQueue, err := channel.QueueDeclare(t.queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = channel.QueueBind(readQueue.Name, routingKey, domainEventsExchangeName, false, nil)
	if err != nil {
		return err
	}

	readChan, err := channel.Consume(readQueue.Name, "", false, false, false, false, nil)

	go func() {
		for msg := range readChan {
			if t.handler == nil {
				err = msg.Nack(false, true)
				time.Sleep(1 * time.Second)
				continue
			}
			err = t.handler.Handle(string(msg.Body))
			if err == nil {
				err = msg.Ack(false)
			} else {
				err = msg.Nack(false, true)
			}
			_ = err
		}
	}()

	return err
}

func (t *transport) SetHandler(handler event.Handler) {
	t.handler = handler
}

func NewEventTransport(queueName string, suppressEventsReading bool) Transport {
	return &transport{queueName: queueName, suppressEventsReading: suppressEventsReading}
}

func CreateTransport(cnf Config, logger, errorLogger *stdlog.Logger) (app.Transport, event.Connection, error) {
	amqpConnection := NewAMQPConnection(&cnf, logger, errorLogger)

	integrationEventTransport := NewEventTransport(cnf.QueueName, false)
	amqpConnection.AddChannel(integrationEventTransport)

	err := amqpConnection.Start()
	if err != nil {
		return nil, nil, err
	}
	return integrationEventTransport, amqpConnection, nil
}
