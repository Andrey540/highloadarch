package sql

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
	var result sql.NullInt32
	rows, err := l.client.Query(sqlQuery, l.lockName, l.timeoutSeconds)
	if err != nil {
		return errors.WithStack(ErrLockTimeout)
	}
	if rows.Err() != nil {
		return errors.WithStack(rows.Err())
	}
	err = rows.Scan(result)
	if result.Int32 == 0 && err == nil {
		return errors.WithStack(ErrLockTimeout)
	}
	defer rows.Close()
	return errors.WithStack(err)
}

func (l *Lock) Release() error {
	const sqlQuery = "SELECT RELEASE_LOCK(SUBSTRING(CONCAT(?, '.', DATABASE()), 1, 64))"
	var result sql.NullInt32
	rows, err := l.client.Query(sqlQuery, l.lockName)
	if err != nil {
		return errors.WithStack(ErrLockTimeout)
	}
	if rows.Err() != nil {
		return errors.WithStack(rows.Err())
	}
	err = rows.Scan(result)
	if err == nil {
		if !result.Valid {
			return ErrLockNotFound
		}
		if result.Int32 == 0 {
			return errors.WithStack(ErrLockNotAcquired)
		}
	}
	defer rows.Close()
	return errors.WithStack(err)
}

func (l *Lock) SetTimeout(timeoutSeconds int) {
	l.timeoutSeconds = timeoutSeconds
}
