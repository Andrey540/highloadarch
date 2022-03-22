package postgres

import (
	"github.com/pkg/errors"
)

type Lock interface {
	Acquire() error
	Release() error
}

func NewLock(client Client, lockName string) Lock {
	return &lock{client: client, lockName: lockName}
}

type lock struct {
	client   Client
	lockName string
}

func (l *lock) Acquire() error {
	var result string
	err := l.client.Get(&result, "SELECT pg_advisory_xact_lock(hashtext(CONCAT($1::text, '.', current_database())))", l.lockName)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(err)
}

func (l *lock) Release() error {
	return nil
}
