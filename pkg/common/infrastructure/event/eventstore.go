package event

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/postgres"
	"github.com/pkg/errors"

	"strconv"

	uuid "github.com/satori/go.uuid"
)

type eventStore struct {
	client postgres.Client
}

func (store *eventStore) NewUID() string {
	return uuid.UUID(infrastructure.NewUUID()).String()
}

func (store *eventStore) Store(storedEvent app.StoredEvent) error {
	const query = `INSERT INTO stored_event (id, status, type, body) VALUES ($1, $2, $3, $4)
	               ON CONFLICT (id) DO UPDATE
	               SET status = excluded.status;`
	_, err := store.client.Exec(query, storedEvent.ID, storedEvent.Status, storedEvent.Type, storedEvent.Body)
	return err
}

func (store *eventStore) GetCreated() ([]app.StoredEvent, error) {
	query := "SELECT id, status, type, body FROM stored_event WHERE status = $1 ORDER BY id"
	var storedEvents []sqlxStoredEvent
	err := store.client.Select(&storedEvents, query, app.Created)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var result = make([]app.StoredEvent, 0, len(storedEvents))
	for _, event := range storedEvents {
		status, err := strconv.Atoi(event.Status)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		result = append(result, app.StoredEvent{ID: event.ID, Status: status, Type: event.EventType, Body: event.Body})
	}

	return result, nil
}

func NewEventStore(client postgres.Client) app.EventStore {
	return &eventStore{client: client}
}

type sqlxStoredEvent struct {
	ID        string `db:"id"`
	Status    string `db:"status"`
	EventType string `db:"type"`
	Body      string `db:"body"`
}
