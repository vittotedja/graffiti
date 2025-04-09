package redis

import (
	"crypto/tls"
	"github.com/redis/go-redis/v9"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

func NewRedisClient(cfg util.Config) redis.UniversalClient {
	opts := &redis.ClusterOptions{
		Addrs: []string{cfg.RedisHost},
	}

	if cfg.RedisTLS {
		opts.TLSConfig = &tls.Config{}
	}

	if cfg.RedisAuth != "" {
		opts.Password = cfg.RedisAuth
	}

	client := redis.NewClusterClient(opts)

	return client
}
