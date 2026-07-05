package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"yhdm_service/internal/config"
)

// NewMongo 连接 MongoDB 并返回目标库的 *mongo.Database。
func NewMongo(ctx context.Context, cfg config.MongoConfig) (*mongo.Client, *mongo.Database, error) {
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client, err := mongo.Connect(cctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, nil, err
	}
	if err := client.Ping(cctx, nil); err != nil {
		return nil, nil, err
	}
	return client, client.Database(cfg.Database), nil
}
