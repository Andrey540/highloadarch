package postgres

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	dbDriverName = "postgres"
)

type Config struct {
	MaxConnections     int
	ConnectionLifetime time.Duration
}

type Connector interface {
	Open(dsn DSN, cfg Config) error
	Client() Client
	TransactionalClient() TransactionalClient
	Close() error
}

type connector struct {
	db *sqlx.DB
}

func NewConnector() Connector {
	return &connector{}
}

func (c *connector) Open(dsn DSN, cfg Config) error {
	var err error
	c.db, err = openDBX(dsn, cfg)
	return errors.WithStack(err)
}

func (c *connector) Close() error {
	err := c.db.Close()
	return errors.Wrap(err, "failed to disconnect")
}

func (c *connector) Client() Client {
	return c.db
}

func (c *connector) TransactionalClient() TransactionalClient {
	return &transactionalClient{c.db}
}
