package event

import (
	"database/sql"

	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/postgres"
	"github.com/pkg/errors"
)

type processedEventStore struct {
	client postgres.Client
}

func (store *processedEventStore) Store(processedEvent app.ProcessedEvent) error {
	const query = `INSERT INTO processed_event (id) VALUES ($1);`
	_, err := store.client.Exec(query, processedEvent.ID)
	return err
}

func (store *processedEventStore) GetEvent(id string) (*app.ProcessedEvent, error) {
	query := "SELECT id FROM processed_event WHERE id = $1"
	var event sqlxProcessedEvent
	err := store.client.Get(&event, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	result := app.NewProcessedEvent(event.ID)
	return &result, nil
}

func NewProcessedEventStore(client postgres.Client) app.ProcessedEventStore {
	return &processedEventStore{client: client}
}

type sqlxProcessedEvent struct {
	ID string `db:"id"`
}
