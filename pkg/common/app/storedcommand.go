package app

import "github.com/callicoder/go-docker/pkg/common/uuid"

type ProcessedCommand struct {
	ID uuid.UUID
}

type ProcessedCommandStore interface {
	Store(processedEvent ProcessedCommand) error
	GetCommand(ID uuid.UUID) (*ProcessedCommand, error)
}

func NewProcessedCommand(id uuid.UUID) ProcessedCommand {
	return ProcessedCommand{
		ID: id,
	}
}
