package app

type ProcessedCommand struct {
	ID string
}

type ProcessedCommandStore interface {
	Store(processedEvent ProcessedCommand) error
	GetCommand(ID string) (*ProcessedCommand, error)
}

func NewProcessedCommand(id string) ProcessedCommand {
	return ProcessedCommand{
		ID: id,
	}
}
