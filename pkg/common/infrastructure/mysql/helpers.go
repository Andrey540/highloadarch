package mysql

import (
	"github.com/pkg/errors"
)

func BeginTransaction(client TransactionalClient, lockNames []string) (Transaction, []Lock, error) {
	transaction, err := client.BeginTransaction()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	var locks = make([]Lock, 0)
	for _, lockName := range lockNames {
		var lock Lock
		if lockName != "" {
			lock = NewLock(transaction, lockName)
			err = lock.Acquire()
			if err != nil {
				err2 := transaction.Rollback()
				if err2 != nil {
					return nil, nil, errors.Wrap(err, err2.Error())
				}
				return nil, nil, err
			}
		}
		locks = append(locks, lock)
	}
	return transaction, locks, err
}

func CompleteTransaction(transaction Transaction, locks []Lock, err error) (returnErr error) {
	for _, lock := range locks {
		if lock.lockName != "" {
			lockErr := lock.Release()
			if err != nil {
				if lockErr != nil {
					err = errors.Wrap(err, lockErr.Error())
				}
			} else {
				err = lockErr
			}
		}
	}

	if err != nil {
		err2 := transaction.Rollback()
		if err2 != nil {
			return errors.Wrap(err, err2.Error())
		}
		return err
	}
	return errors.WithStack(transaction.Commit())
}
