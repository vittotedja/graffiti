package redis

import (
	"context"
	"crypto/tls"
	"github.com/redis/go-redis/v9"
	"github.com/vittotedja/graffiti/graffiti-backend/util"
	"log"
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
	err := client.Ping(context.Background())
	if err != nil {
		log.Fatalf("failed to connect to redis: %v at this address: %s", err, cfg.RedisHost)
	}

	return client
}
