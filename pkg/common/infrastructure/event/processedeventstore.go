package event

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	commonsql "github.com/callicoder/go-docker/pkg/common/infrastructure/sql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/pkg/errors"
)

type processedEventStore struct {
	client commonsql.Client
}

func (store *processedEventStore) Store(processedEvent app.ProcessedEvent) error {
	const query = `INSERT INTO processed_event (id) VALUES (?);`
	_, err := store.client.Exec(query, processedEvent.ID)
	return err
}

func (store *processedEventStore) GetEvent(id uuid.UUID) (*app.ProcessedEvent, error) {
	query := "SELECT id FROM processed_event WHERE id = ?"
	var event sqlxProcessedEvent
	rows, err := store.client.Query(query, commonsql.BinaryUUID(id))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	err = rows.Scan(&event.ID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result := app.NewProcessedEvent(event.ID)
	return &result, nil
}

func NewProcessedEventStore(client commonsql.Client) app.ProcessedEventStore {
	return &processedEventStore{client: client}
}

type sqlxProcessedEvent struct {
	ID uuid.UUID
}
