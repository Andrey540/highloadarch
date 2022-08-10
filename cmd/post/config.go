package main

import (
	"encoding/json"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/amqp"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/tarantool"
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
	TarantoolHost     string `envconfig:"tarantool_host" default:"tarantool"`
	TarantoolPort     string `envconfig:"tarantool_port"  default:"3301"`
	TarantoolUser     string `envconfig:"tarantool_user" default:"guest"`
	TarantoolPassword string `envconfig:"tarantool_password" default:""`

	DBHost               string `envconfig:"db_host"`
	DBName               string `envconfig:"db_name"`
	DBUser               string `envconfig:"db_user"`
	DBPassword           string `envconfig:"db_password"`
	DBMaxConn            int    `envconfig:"db_max_conn" default:"0"`
	DBConnectionLifetime int    `envconfig:"db_conn_lifetime" default:"0"`

	AMQPHost                 string `envconfig:"amqp_host"`
	AMQPUser                 string `envconfig:"amqp_user" default:"guest"`
	AMQPPassword             string `envconfig:"amqp_password" default:"guest"`
	AMQPEnabled              int    `envconfig:"amqp_enabled" default:"1"`
	AMQPRoutingKey           string `envconfig:"amqp_routing_key" default:"#"`
	AMQPSuppressEventReading int    `envconfig:"amqp_suppress_event_reading" default:"0"`

	RealtimeHosts string `envconfig:"realtime_hosts" default:"[]"`

	WorkersCount int `envconfig:"workers_count" default:"1"`

	MigrationsDir string `envconfig:"migrations_dir"`

	HTTPServerEnabled int    `envconfig:"http_server_enabled" default:"1"`
	ServeRESTAddress  string `envconfig:"serve_rest_address" default:":80"`
	ServeGRPCAddress  string `envconfig:"serve_grpc_address" default:":81"`

	UseTarantool int `envconfig:"use_tarantool" default:"1"`

	ServiceID string `envconfig:"service_id" default:"1"`
}

func (c *config) dsn() mysql.DSN {
	return mysql.DSN{
		Host:     c.DBHost,
		Database: c.DBName,
		User:     c.DBUser,
		Password: c.DBPassword,
	}
}

func (c *config) amqpConf() *amqp.Config {
	if c.AMQPEnabled != 1 {
		return nil
	}
	return &amqp.Config{
		Host:            c.AMQPHost,
		User:            c.AMQPUser,
		Password:        c.AMQPPassword,
		QueueName:       appID,
		WorkersCount:    c.WorkersCount,
		RoutingKey:      c.AMQPRoutingKey,
		SuppressReading: c.AMQPSuppressEventReading == 1,
	}
}

func (c *config) tarantoolConf() *tarantool.Config {
	return &tarantool.Config{
		Host:     c.TarantoolHost,
		Port:     c.TarantoolPort,
		User:     c.TarantoolUser,
		Password: c.TarantoolPassword,
	}
}

func (c *config) realtimeHosts() ([]string, error) {
	result := []string{}
	err := json.Unmarshal([]byte(c.RealtimeHosts), &result)
	return result, err
}
