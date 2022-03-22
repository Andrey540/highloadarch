package app

type UnitOfWork interface {
	EventStore() EventStore
	ProcessedEventStore() ProcessedEventStore
	ProcessedCommandStore() ProcessedCommandStore
	GetLocks(lockNames []string) error
	Complete(err error) error
}

type UnitOfWorkFactory interface {
	NewUnitOfWork(lockNames []string) (UnitOfWork, error)
}
