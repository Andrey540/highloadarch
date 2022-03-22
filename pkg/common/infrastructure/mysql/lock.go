package mysql

import (
	"database/sql"
	stderrors "errors"

	"github.com/pkg/errors"
)

const lockTimeoutSeconds = 5

var ErrLockTimeout = stderrors.New("timeout is reached when waiting for lock")
var ErrLockNotAcquired = stderrors.New("cannot release a lock that is not acquired")
var ErrLockNotFound = errors.New("cannot release a lock that is not found")

func NewLock(client Client, lockName string) Lock {
	return Lock{client: client, lockName: lockName, timeoutSeconds: lockTimeoutSeconds}
}

type Lock struct {
	client         Client
	lockName       string
	timeoutSeconds int
}

func (l *Lock) Acquire() error {
	const sqlQuery = "SELECT GET_LOCK(SUBSTRING(CONCAT(?, '.', DATABASE()), 1, 64), ?)"
	var result int
	err := l.client.Get(&result, sqlQuery, l.lockName, l.timeoutSeconds)
	if result == 0 && err == nil {
		return errors.WithStack(ErrLockTimeout)
	}
	return errors.WithStack(err)
}

func (l *Lock) Release() error {
	const sqlQuery = "SELECT RELEASE_LOCK(SUBSTRING(CONCAT(?, '.', DATABASE()), 1, 64))"
	var result sql.NullInt32
	err := l.client.Get(&result, sqlQuery, l.lockName)
	if err == nil {
		if !result.Valid {
			return ErrLockNotFound
		}
		if result.Int32 == 0 {
			return errors.WithStack(ErrLockNotAcquired)
		}
	}
	return errors.WithStack(err)
}

func (l *Lock) SetTimeout(timeoutSeconds int) {
	l.timeoutSeconds = timeoutSeconds
}
