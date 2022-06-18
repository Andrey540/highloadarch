package vitess

import (
	"database/sql"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"vitess.io/vitess/go/vt/vitessdriver"
)

func openDB(dsn DSN, cfg Config, target string) (*sql.DB, error) {
	vitessDB, err := vitessdriver.Open(dsn.String(), target)
	if err != nil {
		return nil, errors.WithStack(errors.Wrapf(err, "failed to open database"))
	}
	err = setupDB(vitessDB, cfg)
	return vitessDB, errors.WithStack(err)
}

func setupDB(db *sql.DB, cfg Config) error {
	// Limit max connections count,
	//  next goroutine will wait once reached limit.
	db.SetMaxOpenConns(cfg.MaxConnections)
	// Limits the maximum amount of time the connection may be reused
	// This value must be lower than wait_timeout value on MySQL
	db.SetConnMaxLifetime(cfg.ConnectionLifetime)

	err := backoff.Retry(func() error {
		tryError := db.Ping()
		return tryError
	}, newExponentialBackOff(30))
	if err != nil {
		dbCloseErr := db.Close()
		if dbCloseErr != nil {
			err = errors.Wrap(err, dbCloseErr.Error())
		}
		return errors.Wrapf(err, "failed to ping database")
	}
	return errors.Wrapf(err, "failed to change database")
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
