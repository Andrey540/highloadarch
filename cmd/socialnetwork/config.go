package main

import (
	"encoding/json"

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
	RedisHost     string `envconfig:"redis_host" default:"localhost"`
	RedisPort     string `envconfig:"redis_port"  default:"6379"`
	RedisPassword string `envconfig:"redis_password" default:""`

	RealtimeHosts string `envconfig:"realtime_hosts" default:"[]"`

	MigrationsDir string `envconfig:"migrations_dir"`

	ServeRESTAddress   string `envconfig:"serve_rest_address" default:":80"`
	ServiceRESTAddress string `envconfig:"service_rest_address" default:"http://service:80"`
	ServiceGRPCAddress string `envconfig:"service_grpc_address" default:"service:81"`
}

func (c *config) realtimeHosts() ([]string, error) {
	result := []string{}
	err := json.Unmarshal([]byte(c.RealtimeHosts), &result)
	return result, err
}
