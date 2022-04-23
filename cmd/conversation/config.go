package main

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/vitess"
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
	SchemaHost string `envconfig:"schema_host"`
	SchemaPort string `envconfig:"schema_port"`

	VSchemaPath string `envconfig:"vschema_path"`

	DBHost               string `envconfig:"db_host"`
	DBName               string `envconfig:"db_name"`
	DBPort               string `envconfig:"db_port"`
	DBMaxConn            int    `envconfig:"db_max_conn" default:"0"`
	DBConnectionLifetime int    `envconfig:"db_conn_lifetime" default:"0"`

	MigrationsDir string `envconfig:"migrations_dir"`

	ServiceHost      string `envconfig:"service_host" default:"http://conversation:80"`
	ServeRESTAddress string `envconfig:"serve_rest_address" default:":80"`
}

func (c *config) dbDsn() vitess.DSN {
	return vitess.DSN{
		Host: c.DBHost,
		Port: c.DBPort,
	}
}

func (c *config) schemaDsn() vitess.DSN {
	return vitess.DSN{
		Host: c.SchemaHost,
		Port: c.SchemaPort,
	}
}
