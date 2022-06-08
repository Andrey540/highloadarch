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

	ServeRESTAddress               string `envconfig:"serve_rest_address" default:":80"`
	UserServiceRESTAddress         string `envconfig:"user_service_rest_address" default:"http://user:80"`
	UserServiceGRPCAddress         string `envconfig:"user_service_grpc_address" default:"user:81"`
	ConversationServiceRESTAddress string `envconfig:"conversation_service_rest_address" default:"http://conversation:80"`
	ConversationServiceGRPCAddress string `envconfig:"conversation_service_grpc_address" default:"conversation:81"`
	PostServiceRESTAddress         string `envconfig:"post_service_rest_address" default:"http://post:80"`
	PostServiceGRPCAddress         string `envconfig:"post_service_grpc_address" default:"post:81"`
}

func (c *config) realtimeHosts() ([]string, error) {
	result := []string{}
	err := json.Unmarshal([]byte(c.RealtimeHosts), &result)
	return result, err
}
