package infrastructure

import (
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/command"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/sql"
	"github.com/callicoder/go-docker/pkg/conversation/app"
	"github.com/pkg/errors"
)

func NewUnitOfWorkFactory(client sql.TransactionalClient) commonapp.UnitOfWorkFactory {
	return &unitOfWorkFactory{client: client}
}

type unitOfWorkFactory struct {
	client sql.TransactionalClient
}

func (s *unitOfWorkFactory) NewUnitOfWork(lockNames []string) (commonapp.UnitOfWork, error) {
	transaction, locks, err := sql.BeginTransaction(s.client, lockNames)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &unitOfWork{transaction: transaction, locks: locks}, nil
}

type unitOfWork struct {
	transaction sql.Transaction
	locks       []sql.Lock
}

func (u *unitOfWork) ConversationRepository() app.ConversationRepository {
	return NewConversationRepository(u.transaction)
}

func (u *unitOfWork) MessageRepository() app.MessageRepository {
	return NewMessageRepository(u.transaction)
}

func (u *unitOfWork) EventStore() commonapp.EventStore {
	return event.NewEventStore(u.transaction)
}

func (u *unitOfWork) ProcessedEventStore() commonapp.ProcessedEventStore {
	return event.NewProcessedEventStore(u.transaction)
}

func (u *unitOfWork) ProcessedCommandStore() commonapp.ProcessedCommandStore {
	return command.NewProcessedCommandStore(u.transaction)
}

func (u *unitOfWork) GetLocks(lockNames []string) error {
	for _, lockName := range lockNames {
		var lock sql.Lock
		if lockName != "" {
			lock = sql.NewLock(u.transaction, lockName)
			err := lock.Acquire()
			if err != nil {
				return err
			}
		}
		u.locks = append(u.locks, lock)
	}
	return nil
}

func (u *unitOfWork) Complete(err error) error {
	return sql.CompleteTransaction(u.transaction, u.locks, err)
}
