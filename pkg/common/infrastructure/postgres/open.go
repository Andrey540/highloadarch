package postgres

import (
	"time"

	"github.com/cenkalti/backoff"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	maxReconnectWaitingTime = 15 * time.Second
)

type dbClient interface {
	SetMaxOpenConns(maxConnections int)
	SetConnMaxLifetime(d time.Duration)
	Ping() error
	Close() error
}

func openDBX(dsn DSN, cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open(dbDriverName, dsn.String())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open database")
	}
	err = setupDB(db, cfg)
	return db, errors.WithStack(err)
}

func setupDB(db dbClient, cfg Config) error {
	// Limit max connections count,
	//  next goroutine will wait once reached limit.
	db.SetMaxOpenConns(cfg.MaxConnections)
	// Limits the maximum amount of time the connection may be reused
	// This value must be lower than wait_timeout value on MySQL
	db.SetConnMaxLifetime(cfg.ConnectionLifetime)

	err := backoff.Retry(func() error {
		tryError := db.Ping()
		return tryError
	}, newExponentialBackOff())
	if err != nil {
		dbCloseErr := db.Close()
		if dbCloseErr != nil {
			err = errors.Wrap(err, dbCloseErr.Error())
		}
		return errors.Wrapf(err, "failed to ping database")
	}
	return nil
}

func newExponentialBackOff() *backoff.ExponentialBackOff {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = maxReconnectWaitingTime
	return exponentialBackOff
}
