package mysql

import (
	"database/sql"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type dbClient interface {
	SetMaxOpenConns(maxConnections int)
	SetConnMaxLifetime(d time.Duration)
	Ping() error
	Close() error
}

func openDB(dsn DSN, cfg Config) (*sql.DB, error) {
	db, err := sql.Open(dbDriverName, dsn.String())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open database")
	}
	err = setupDB(db, cfg)
	return db, errors.WithStack(err)
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
	}, newExponentialBackOff(cfg.ConnectTimeout))
	if err != nil {
		dbCloseErr := db.Close()
		if dbCloseErr != nil {
			err = errors.Wrap(err, dbCloseErr.Error())
		}
		return errors.Wrapf(err, "failed to ping database")
	}
	return nil
}

func newExponentialBackOff(timeout time.Duration) *backoff.ExponentialBackOff {
	exponentialBackOff := backoff.NewExponentialBackOff()
	const maxReconnectWaitingTime = 15 * time.Second
	if timeout != 0 {
		exponentialBackOff.MaxElapsedTime = timeout
	} else {
		exponentialBackOff.MaxElapsedTime = maxReconnectWaitingTime
	}
	return exponentialBackOff
}
