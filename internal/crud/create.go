package crud

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// Create 用自增 _id 插入一条文档，并对重复键自动重试（防计数器落后于实际 max(_id)）。
// build 回调根据分配到的 id 构造待插入文档。返回新 id。
func Create[T any](ctx context.Context, repo Repo[T], build func(id int64) any) (int64, error) {
	var lastErr error
	for i := 0; i < 5; i++ {
		id, err := repo.NextID(ctx)
		if err != nil {
			return 0, err
		}
		if err := repo.Insert(ctx, build(id)); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				lastErr = err
				continue
			}
			return 0, err
		}
		return id, nil
	}
	return 0, lastErr
}
