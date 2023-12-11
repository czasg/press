package third

import "github.com/go-redis/redis"

type RedisConfig struct {
	Address     string
	Password    string
	DB          int
	PoolSize    int
	MaxRetries  int
	MinIdleSize int
}

func NewRedis(cfg RedisConfig) (*redis.Client, error) {
	ins := redis.NewClient(&redis.Options{
		Addr:       cfg.Address,
		Password:   cfg.Password,
		DB:         cfg.DB,
		MaxRetries: 3,
	})
	err := ins.Ping().Err()
	if err != nil {
		return nil, err
	}
	return ins, nil
}
