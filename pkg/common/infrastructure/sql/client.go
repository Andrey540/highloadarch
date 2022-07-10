package sql

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

type Client interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type HealthCheckClient interface {
	Ping() error
}

type ClientContext interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Transaction interface {
	Client
	Commit() error
	Rollback() error
}

type TransactionalConnection interface {
	ClientContext
	BeginTransaction(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
	Close() error
}

type TransactionalClient interface {
	Client
	ClientContext
	BeginTransaction() (Transaction, error)
	Connection(ctx context.Context) (TransactionalConnection, error)
}

type transactionalClient struct {
	*sql.DB
}

func (t *transactionalClient) BeginTransaction() (Transaction, error) {
	return t.Begin()
}

func (t *transactionalClient) Connection(ctx context.Context) (TransactionalConnection, error) {
	connx, err := t.Conn(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &transactionalConnection{Conn: connx}, nil
}

func NewTransactionalClient(db *sql.DB) TransactionalClient {
	return &transactionalClient{db}
}

type transactionalConnection struct {
	*sql.Conn
}

func (t *transactionalConnection) BeginTransaction(ctx context.Context, opts *sql.TxOptions) (Transaction, error) {
	return t.BeginTx(ctx, opts)
}

func NewTransactionalConnection(conn *sql.Conn) TransactionalConnection {
	return &transactionalConnection{conn}
}
