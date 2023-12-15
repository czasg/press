package third

import (
	"github.com/czasg/press/internal/config"
	"github.com/go-redis/redis"
	"net/url"
	"strconv"
)

func NewRedis(cfg *config.Config) (*redis.Client, error) {
	parsedURL, err := url.Parse(cfg.Metadata.Annotations.PressClusterBrokerRedisUrl)
	if err != nil {
		return nil, err
	}
	password, _ := parsedURL.User.Password()
	db, err := strconv.Atoi(parsedURL.Path[1:])
	if err != nil {
		return nil, err
	}
	ins := redis.NewClient(&redis.Options{
		Addr:       parsedURL.Host,
		Password:   password,
		DB:         db,
		MaxRetries: 3,
	})
	err = ins.Ping().Err()
	if err != nil {
		return nil, err
	}
	return ins, nil
}
