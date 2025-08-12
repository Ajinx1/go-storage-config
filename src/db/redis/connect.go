package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisConn struct {
	Client *redis.Client
	Ctx    context.Context
}

func ConnectFromEnv(theConfig RedisConfig) (*RedisConn, error) {
	cfg := LoadRedisConfig(theConfig)

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisConn{
		Client: client,
		Ctx:    ctx,
	}, nil
}
