package kafka

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"

	stdlog "log"
)

const TopicName = "DomainEvents"

type transport struct {
	connection Connection
}

func (t *transport) Send(msgBody string, storedEvent app.StoredEvent) error {
	return t.connection.Send(storedEvent.ID.String(), msgBody)
}

func (t *transport) SetHandler(handler event.Handler) {
	t.connection.SetMessageHandler(handler)
}

func NewEventTransport(connection Connection) event.Transport {
	return &transport{connection: connection}
}

func CreateTransport(cnf Config, logger, errorLogger *stdlog.Logger) (app.Transport, event.Connection, error) {
	kafkaConnection := NewConnection(&cnf, logger, errorLogger)
	err := kafkaConnection.Start()
	if err != nil {
		return nil, nil, err
	}
	kafkaTransport := NewEventTransport(kafkaConnection)
	return kafkaTransport, kafkaConnection, nil
}
