package event

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/sql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/pkg/errors"
)

type eventStore struct {
	client sql.Client
	dbName string
}

func (store *eventStore) NewUID() uuid.UUID {
	return uuid.UUID(infrastructure.NewUUID())
}

func (store *eventStore) Store(storedEvent app.StoredEvent) error {
	query := `INSERT INTO ` + store.dbName + `stored_event (id, status, type, body) VALUES (?, ?, ?, ?)
                   ON DUPLICATE KEY UPDATE status=VALUES(status);`
	_, err := store.client.Exec(query, sql.BinaryUUID(storedEvent.ID), storedEvent.Status, storedEvent.Type, storedEvent.Body)
	return errors.WithStack(err)
}

func (store *eventStore) GetCreated() ([]app.StoredEvent, error) {
	query := `SELECT id, status, type, body FROM ` + store.dbName + `stored_event WHERE status = ? ORDER BY id`
	rows, err := store.client.Query(query, app.Created)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}

	var result []app.StoredEvent
	for rows.Next() {
		var storedEvent app.StoredEvent
		err1 := rows.Scan(&storedEvent.ID, &storedEvent.Status, &storedEvent.Type, &storedEvent.Body)
		if err1 != nil {
			return []app.StoredEvent{}, errors.WithStack(err)
		}
		result = append(result, storedEvent)
	}
	defer rows.Close()
	return result, nil
}

func NewEventStore(client sql.Client, dbName string) app.EventStore {
	if dbName != "" {
		dbName += "."
	}
	return &eventStore{client: client, dbName: dbName}
}
