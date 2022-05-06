package main

import (
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

	MigrationsDir string `envconfig:"migrations_dir"`

	ServiceHost            string `envconfig:"service_host" default:"http://socialnetwork:80"`
	ServeRESTAddress       string `envconfig:"serve_rest_address" default:":80"`
	UserServiceURL         string `envconfig:"user_service_url" default:"http://user:80"`
	ConversationServiceURL string `envconfig:"conversation_service_url" default:"http://conversation:80"`
	PostServiceURL         string `envconfig:"post_service_url" default:"http://post:80"`
}
