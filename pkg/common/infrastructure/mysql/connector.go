package mysql

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff"
	_ "github.com/go-sql-driver/mysql" // provides MySQL driver
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source"
	_ "github.com/golang-migrate/migrate/v4/source/file" // provides filesystem source
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	dbDriverName      = "mysql"
	shardedFolderName = "sharded"
)

type Config struct {
	MaxConnections     int
	ConnectionLifetime time.Duration
	ConnectTimeout     time.Duration // 0 means default timeout (15 seconds)
}

type Connector interface {
	Open(dsn DSN, cfg Config) error
	MigrateUp(dsn DSN, migrationsDir string) error
	AddSourceDriver(sourceDriver SourceDriver) error
	Client() Client
	TransactionalClient() TransactionalClient
	Close() error
	Ping() error
}

type connector struct {
	db                 *sqlx.DB
	shardMigrationsDir string
}

func NewConnector() Connector {
	return &connector{}
}

func (c *connector) MigrateUp(dsn DSN, migrationsDir string) (err error) {
	// Db connections will be closed when migration object is closed, so new connection must be opened
	db, err := openDB(dsn, Config{MaxConnections: 1, ConnectionLifetime: time.Minute})
	if err != nil {
		return errors.WithStack(err)
	}
	c.shardMigrationsDir = filepath.Join(migrationsDir, shardedFolderName)
	m, err := createMigrator(db, migrationsDir)
	if err != nil {
		return errors.WithStack(err)
	}
	// noinspection GoUnhandledErrorResult
	defer m.Close()

	err = m.Up()
	if err == migrate.ErrNoChange {
		return nil
	}

	return errors.Wrap(err, "failed to migrate")
}

func (c *connector) Open(dsn DSN, cfg Config) error {
	var err error
	c.db, err = openDBX(dsn, cfg)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(err)
}

func (c *connector) AddSourceDriver(sourceDriver SourceDriver) error {
	sourceDriver.register(source.Register)
	return nil
}

func (c *connector) Close() error {
	err := c.db.Close()
	return errors.Wrap(err, "failed to disconnect")
}

func (c *connector) Ping() error {
	return c.db.Ping()
}

func (c *connector) Client() Client {
	return c.db
}

func (c *connector) TransactionalClient() TransactionalClient {
	return &transactionalClient{c.db}
}

func createMigrator(db *sql.DB, migrationsDir string) (*migrate.Migrate, error) {
	migrationsURL, err := makeMigrationsURL(migrationsDir)
	if err != nil {
		return nil, err
	}
	driver, err := createMigrationDriver(db)
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsURL, dbDriverName, driver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create migrator")
	}
	return m, nil
}

func createMigrationDriver(db *sql.DB) (driver database.Driver, err error) {
	err = backoff.Retry(func() error {
		var tryError error
		driver, tryError = mysql.WithInstance(db, &mysql.Config{})
		return tryError
	}, newExponentialBackOff(0))
	return driver, errors.Wrapf(err, "cannot create migrations driver")
}

func makeMigrationsURL(migrationsDir string) (string, error) {
	// if already url with scheme just return
	if u, err := url.Parse(migrationsDir); err == nil && u.Scheme != "" {
		return migrationsDir, nil
	}

	_, err := os.Stat(migrationsDir)
	if err != nil {
		return "", errors.Wrapf(err, "cannot use migrations from %s", migrationsDir)
	}
	migrationsDir, err = filepath.Abs(migrationsDir)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return fmt.Sprintf("file://%s", migrationsDir), nil
}
