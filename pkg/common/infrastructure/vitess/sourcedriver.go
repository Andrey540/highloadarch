package vitess

import (
	"embed"
	"net/http"
	stdurl "net/url"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

const (
	EmbeddedHTTPFSSource = "embeddedhttpfs"
)

type SourceDriver interface {
	register(register func(name string, driver source.Driver))
}

func NewEmbeddedHTTPFSDriver(migrations embed.FS) *EmbeddedHTTPFSDriver {
	return &EmbeddedHTTPFSDriver{migrations: migrations}
}

type EmbeddedHTTPFSDriver struct {
	httpfs.PartialDriver

	migrations embed.FS
}

func (driver *EmbeddedHTTPFSDriver) Open(url string) (source.Driver, error) {
	u, err := stdurl.Parse(url)
	if err != nil {
		return nil, err
	}

	path := u.Hostname() + u.Path

	fs := http.FS(driver.migrations)

	var ds EmbeddedHTTPFSDriver

	if err = ds.Init(fs, path); err != nil {
		return nil, err
	}
	return &ds, nil
}

func (driver *EmbeddedHTTPFSDriver) register(register func(name string, driver source.Driver)) {
	register(EmbeddedHTTPFSSource, driver)
}
