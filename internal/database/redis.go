package database

import (
	"context"

	"github.com/redis/go-redis/v9"

	"yhdm_service/internal/config"
)

// NewRedis 连接 Redis 并返回客户端。
func NewRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}
