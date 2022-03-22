package command

import (
	"database/sql"

	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/postgres"
	"github.com/pkg/errors"
)

type processedCommandStore struct {
	client postgres.Client
}

func (store *processedCommandStore) Store(processedCommand app.ProcessedCommand) error {
	const query = `INSERT INTO processed_command (id) VALUES ($1);`
	_, err := store.client.Exec(query, processedCommand.ID)
	return err
}

func (store *processedCommandStore) GetCommand(id string) (*app.ProcessedCommand, error) {
	query := "SELECT id FROM processed_command WHERE id = $1"
	var command sqlxProcessedCommand
	err := store.client.Get(&command, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	result := app.NewProcessedCommand(command.ID)
	return &result, nil
}

func NewProcessedCommandStore(client postgres.Client) app.ProcessedCommandStore {
	return &processedCommandStore{client: client}
}

type sqlxProcessedCommand struct {
	ID string `db:"id"`
}
