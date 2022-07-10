package redis

import (
	"encoding/json"

	"github.com/go-redis/redis"
	"github.com/lithammer/shortuuid/v3"
	"github.com/pkg/errors"

	"time"
)

const sessionTimeout = 30 * time.Minute

var errNilRedisClient = errors.New("redis client pointer = nil")

type UserSession struct {
	UserID string `json:"userID"`
}

// Config contains parameters for Redis connection
type Config struct {
	Password string
	Host     string
}

// Service interface provides access to
type SessionService interface {
	GetUserSession(sessionKey string) (*UserSession, error)
	SaveSession(userID string) (string, error)
	RemoveSession(sessionKey string) error
	Stop() error
	Ping() error
}

type service struct {
	redisClient           *redis.Client
	sessionExpirationTime time.Duration
}

func NewSessionService(cfg *Config) (SessionService, error) {
	res := &service{sessionExpirationTime: sessionTimeout}
	err := res.startRedis(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *service) startRedis(cfg *Config) error {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host,
		Password: cfg.Password,
	})
	if client == nil {
		return errors.WithStack(errNilRedisClient)
	}
	_, err := client.Ping().Result()
	if err != nil {
		return err
	}
	s.redisClient = client
	return nil
}

func (s *service) Stop() error {
	return s.redisClient.Close()
}

func (s *service) Ping() error {
	return s.redisClient.Ping().Err()
}

func (s *service) GetUserSession(sessionKey string) (*UserSession, error) {
	res, err := s.redisClient.Get(sessionKey).Result()
	switch {
	case errors.Is(err, redis.Nil) || res == "":
		return nil, nil
	case err == nil:
		var session UserSession
		err1 := json.Unmarshal([]byte(res), &session)
		if err1 != nil {
			return nil, errors.WithStack(err1)
		}
		_, err1 = s.redisClient.Set(sessionKey, res, s.sessionExpirationTime).Result()
		return &session, errors.WithStack(err1)
	default:
		return nil, errors.WithStack(err)
	}
}

func (s *service) SaveSession(userID string) (string, error) {
	sessionKey := shortuuid.New()
	res, err := json.Marshal(UserSession{
		UserID: userID,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}
	_, err = s.redisClient.Set(sessionKey, res, s.sessionExpirationTime).Result()
	return sessionKey, errors.WithStack(err)
}

func (s *service) RemoveSession(sessionKey string) error {
	_, err := s.redisClient.Del(sessionKey).Result()
	return errors.WithStack(err)
}
