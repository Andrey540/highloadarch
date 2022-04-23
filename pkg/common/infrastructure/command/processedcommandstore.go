package command

import (
	"github.com/callicoder/go-docker/pkg/common/uuid"

	"github.com/callicoder/go-docker/pkg/common/app"
	commonsql "github.com/callicoder/go-docker/pkg/common/infrastructure/sql"
	"github.com/pkg/errors"
)

type processedCommandStore struct {
	client commonsql.Client
}

func (store *processedCommandStore) Store(processedCommand app.ProcessedCommand) error {
	const query = `INSERT INTO processed_command (id) VALUES (?);`
	_, err := store.client.Exec(query, processedCommand.ID)
	return err
}

func (store *processedCommandStore) GetCommand(id uuid.UUID) (*app.ProcessedCommand, error) {
	query := "SELECT id FROM processed_command WHERE id = ?"
	var command sqlxProcessedCommand
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
	err = rows.Scan(&command.ID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result := app.NewProcessedCommand(command.ID)
	return &result, nil
}

func NewProcessedCommandStore(client commonsql.Client) app.ProcessedCommandStore {
	return &processedCommandStore{client: client}
}

type sqlxProcessedCommand struct {
	ID uuid.UUID
}
