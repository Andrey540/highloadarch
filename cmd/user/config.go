package main

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

func parseEnv() (*config, error) {
	c := new(config)
	if err := envconfig.Process(appID, c); err != nil {
		return nil, errors.Wrap(err, "failed to parse env")
	}
	return c, nil
}

type config struct {
	DBMasterHost         string `envconfig:"db_master_host"`
	DBSlaveHost          string `envconfig:"db_slave_host"`
	DBName               string `envconfig:"db_name"`
	DBUser               string `envconfig:"db_user"`
	DBPassword           string `envconfig:"db_password"`
	DBMaxConn            int    `envconfig:"db_max_conn" default:"0"`
	DBConnectionLifetime int    `envconfig:"db_conn_lifetime" default:"0"`

	MigrationsDir string `envconfig:"migrations_dir"`

	ServiceHost      string `envconfig:"service_host" default:"http://socialnetwork:80"`
	ServeRESTAddress string `envconfig:"serve_rest_address" default:":80"`
}

func (c *config) masterDSN() mysql.DSN {
	return mysql.DSN{
		Host:     c.DBMasterHost,
		Database: c.DBName,
		User:     c.DBUser,
		Password: c.DBPassword,
	}
}

func (c *config) slaveDSN() mysql.DSN {
	return mysql.DSN{
		Host:     c.DBSlaveHost,
		Database: c.DBName,
		User:     c.DBUser,
		Password: c.DBPassword,
	}
}
