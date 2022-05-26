package infrastructure

import (
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/command"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/pkg/errors"
)

func NewUnitOfWorkFactory(client mysql.TransactionalClient) commonapp.UnitOfWorkFactory {
	return &unitOfWorkFactory{client: client}
}

type unitOfWorkFactory struct {
	client mysql.TransactionalClient
}

func (s *unitOfWorkFactory) NewUnitOfWork(lockNames []string) (commonapp.UnitOfWork, error) {
	transaction, locks, err := mysql.BeginTransaction(s.client, lockNames)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &unitOfWork{transaction: transaction, locks: locks}, nil
}

type unitOfWork struct {
	transaction mysql.Transaction
	locks       []mysql.Lock
}

func (u *unitOfWork) PostRepository() app.PostRepository {
	return NewPostRepository(u.transaction)
}

func (u *unitOfWork) NewsLineStore() app.NewsLineStore {
	return NewNewsLineStore(u.transaction)
}

func (u *unitOfWork) UserFriendRepository() app.UserFriendRepository {
	return NewUserFriendRepository(u.transaction)
}

func (u *unitOfWork) UserRepository() app.UserRepository {
	return NewUserRepository(u.transaction)
}

func (u *unitOfWork) UserProvider() app.UserProvider {
	return NewUserProvider(u.transaction)
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
		var lock mysql.Lock
		if lockName != "" {
			lock = mysql.NewLock(u.transaction, lockName)
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
	return mysql.CompleteTransaction(u.transaction, u.locks, err)
}
