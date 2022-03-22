package mysql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Client interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)

	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row

	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
}

type ClientContext interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row

	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type Transaction interface {
	ClientContext
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
	*sqlx.DB
}

func (t *transactionalClient) BeginTransaction() (Transaction, error) {
	return t.Beginx()
}

func (t *transactionalClient) Connection(ctx context.Context) (TransactionalConnection, error) {
	connx, err := t.Connx(ctx)
	if err != nil {
		return nil, err
	}

	return &transactionalConnection{Conn: connx}, nil
}

type transactionalConnection struct {
	*sqlx.Conn
}

func (t *transactionalConnection) BeginTransaction(ctx context.Context, opts *sql.TxOptions) (Transaction, error) {
	return t.BeginTxx(ctx, opts)
}
