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
	DBHost               string `envconfig:"db_host"`
	DBName               string `envconfig:"db_name"`
	DBUser               string `envconfig:"db_user"`
	DBPassword           string `envconfig:"db_password"`
	DBMaxConn            int    `envconfig:"db_max_conn" default:"0"`
	DBConnectionLifetime int    `envconfig:"db_conn_lifetime" default:"0"`

	RedisHost     string `envconfig:"redis_host" default:"localhost"`
	RedisPort     string `envconfig:"redis_port"  default:"6379"`
	RedisPassword string `envconfig:"redis_password" default:""`

	MigrationsDir string `envconfig:"migrations_dir"`

	ServiceHost      string `envconfig:"service_host" default:"http://socialnetwork:80"`
	ServeRESTAddress string `envconfig:"serve_rest_address" default:":80"`
}

func (c *config) dsn() mysql.DSN {
	return mysql.DSN{
		Host:     c.DBHost,
		Database: c.DBName,
		User:     c.DBUser,
		Password: c.DBPassword,
	}
}
