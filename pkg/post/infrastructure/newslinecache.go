package infrastructure

import (
	"encoding/json"
	"fmt"
	"time"

	commonredis "github.com/callicoder/go-docker/pkg/common/infrastructure/redis"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

const (
	sessionTimeout = 30 * time.Minute
	userKey        = "post-user-%s"
)

var errNilRedisClient = errors.New("redis client pointer = nil")

type NewsLineCache struct {
	redisClient           *redis.Client
	sessionExpirationTime time.Duration
}

func NewNewsLineCache(cfg *commonredis.Config) (*NewsLineCache, error) {
	res := &NewsLineCache{sessionExpirationTime: sessionTimeout}
	err := res.startRedis(cfg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *NewsLineCache) Stop() error {
	return s.redisClient.Close()
}

func (s *NewsLineCache) InvalidateUsers(userIDs []uuid.UUID) error {
	keys := []string{}
	for _, userID := range userIDs {
		keys = append(keys, s.getUserKey(userID))
	}
	_, err := s.redisClient.Del(keys...).Result()
	return errors.WithStack(err)
}

func (s *NewsLineCache) GetUserNews(userID uuid.UUID) (*[]app.NewsLineItem, error) {
	res, err := s.redisClient.Get(s.getUserKey(userID)).Result()
	switch {
	case errors.Is(err, redis.Nil) || res == "":
		return nil, nil
	case err == nil:
		var result []app.NewsLineItem
		err1 := json.Unmarshal([]byte(res), &result)
		if err1 != nil {
			return nil, errors.WithStack(err1)
		}
		_, err1 = s.redisClient.Set(s.getUserKey(userID), res, s.sessionExpirationTime).Result()
		return &result, errors.WithStack(err1)
	default:
		return nil, errors.WithStack(err)
	}
}

func (s *NewsLineCache) SaveUserNews(userID uuid.UUID, news *[]app.NewsLineItem) error {
	res, err := json.Marshal(news)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = s.redisClient.Set(s.getUserKey(userID), res, s.sessionExpirationTime).Result()
	return errors.WithStack(err)
}

func (s *NewsLineCache) startRedis(cfg *commonredis.Config) error {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host,
		Password: cfg.Password,
	})
	if client == nil {
		return errors.WithStack(errNilRedisClient)
	}
	_, err := client.Ping().Result()
	if err != nil {
		return errors.WithStack(err)
	}
	s.redisClient = client
	return nil
}

func (s *NewsLineCache) getUserKey(userID uuid.UUID) string {
	return fmt.Sprintf(userKey, userID.String())
}
