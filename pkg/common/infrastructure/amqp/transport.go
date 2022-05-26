package amqp

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"
	"github.com/streadway/amqp"

	stdlog "log"
	"strconv"
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
	errorLogger           *stdlog.Logger
	queueName             string
	suppressEventsReading bool
	workersCount          int
	routingKey            string
}

func (t *transport) Send(msgBody string, storedEvent app.StoredEvent) error {
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  contentType,
		Body:         []byte(msgBody),
	}

	workersCount := t.workersCount
	if workersCount <= 0 {
		workersCount = 1
	}
	routing := routingPrefix + strconv.Itoa(t.getRoutingKey(storedEvent.RoutingID, workersCount))
	return t.writeChannel.Publish(domainEventsExchangeName, routing, false, false, msg)
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

	routing := routingKey
	if t.routingKey != "" {
		routing = t.routingKey
	}
	err = channel.QueueBind(readQueue.Name, routing, domainEventsExchangeName, false, nil)
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
				t.errorLogger.Println(err)
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

func (t *transport) getRoutingKey(routingID string, publishingMode int) int {
	bytes := []byte(routingID)
	sum := 0
	for _, item := range bytes {
		sum += int(item)
	}
	return sum % publishingMode
}

func NewEventTransport(queueName string, errorLogger *stdlog.Logger, suppressEventsReading bool, workersCount int, routingKey string) Transport {
	return &transport{
		queueName:             queueName,
		errorLogger:           errorLogger,
		suppressEventsReading: suppressEventsReading,
		workersCount:          workersCount,
		routingKey:            routingKey,
	}
}

func CreateTransport(cnf Config, logger, errorLogger *stdlog.Logger) (app.Transport, event.Connection, error) {
	amqpConnection := NewAMQPConnection(&cnf, logger, errorLogger)

	integrationEventTransport := NewEventTransport(cnf.QueueName, errorLogger, cnf.SuppressReading, cnf.WorkersCount, cnf.RoutingKey)
	amqpConnection.AddChannel(integrationEventTransport)

	err := amqpConnection.Start()
	if err != nil {
		return nil, nil, err
	}
	return integrationEventTransport, amqpConnection, nil
}
