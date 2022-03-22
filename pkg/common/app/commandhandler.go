package app

type CommandHandler interface {
	Handle(command Command) (interface{}, error)
}

type CommandHandlerFactory interface {
	CreateHandler(unitOfWork UnitOfWork, commandType string) (CommandHandler, error)
}
