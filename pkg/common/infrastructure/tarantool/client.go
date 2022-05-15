package tarantool

import (
	"fmt"

	"github.com/tarantool/go-tarantool"
)

type Config struct {
	User     string
	Password string
	Host     string
	Port     string
}

type Client interface {
	Open() error
	Close() error
	Select(space, index string, offset, limit uint32, value string, result interface{}) error
}

type client struct {
	cfg        *Config
	connection *tarantool.Connection
}

func (c *client) Select(space, index string, offset, limit uint32, value string, result interface{}) error {
	return c.connection.SelectTyped(space, index, offset, limit, tarantool.IterReq, []interface{}{value}, &result)
}

func (c *client) Open() error {
	opts := tarantool.Opts{User: c.cfg.User, Pass: c.cfg.Password}
	address := fmt.Sprintf("%s:%s", c.cfg.Host, c.cfg.Port)
	connection, err := tarantool.Connect(address, opts)
	c.connection = connection
	return err
}

func (c *client) Close() error {
	return c.connection.Close()
}

func NewClient(cfg *Config) Client {
	return &client{cfg: cfg}
}
