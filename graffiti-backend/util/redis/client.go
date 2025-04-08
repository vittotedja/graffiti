package redis

import (
	"crypto/tls"
	"github.com/redis/go-redis/v9"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
)

func NewRedisClient(cfg util.Config) *redis.Client {
	opt := &redis.Options{
		Addr: cfg.RedisHost,
	}
	if cfg.RedisAuth != "" {
		opt.Password = cfg.RedisAuth
	}
	if cfg.RedisTLS {
		opt.TLSConfig = &tls.Config{}
	}

	return redis.NewClient(opt)
}
