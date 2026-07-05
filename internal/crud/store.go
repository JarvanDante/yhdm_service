// Package crud 提供任意 Mongo 集合的通用增删改查基座，复刻旧系统的自增 _id 机制，
// 让各业务模块只需写字段映射与校验，不必重复分页/查询/自增等样板代码。
package crud

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Identifiable 约束：模型能返回自身 _id，供通用唯一性校验排除自身。
type Identifiable interface {
	GetID() int64
}

// Repo 是业务模块依赖的通用数据访问接口。生产用 *Store[T]，单测用 *MemStore[T]。
type Repo[T any] interface {
	List(ctx context.Context, query bson.M, sort bson.D, page, size int) ([]T, int64, error)
	FindByID(ctx context.Context, id int64) (*T, error)
	FindOne(ctx context.Context, query bson.M) (*T, error)
	Count(ctx context.Context, query bson.M) (int64, error)
	NextID(ctx context.Context) (int64, error)
	Insert(ctx context.Context, doc any) error
	Update(ctx context.Context, id int64, set bson.M) error
	Delete(ctx context.Context, id int64) (int64, error)
}

// Store 是 Repo 的 Mongo 实现。
type Store[T any] struct {
	db       *mongo.Database
	coll     *mongo.Collection
	collName string
}

// New 构造某集合的通用仓储。
func New[T any](db *mongo.Database, collName string) *Store[T] {
	return &Store[T]{db: db, coll: db.Collection(collName), collName: collName}
}

// List 分页查询。sort 为空时默认按 _id 降序。
func (s *Store[T]) List(ctx context.Context, query bson.M, sort bson.D, page, size int) ([]T, int64, error) {
	if query == nil {
		query = bson.M{}
	}
	total, err := s.coll.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 15
	}
	if len(sort) == 0 {
		sort = bson.D{{Key: "_id", Value: -1}}
	}
	opts := options.Find().SetSort(sort).SetSkip(int64((page - 1) * size)).SetLimit(int64(size))
	cur, err := s.coll.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	var list []T
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindByID 按 _id 查询，不存在返回 (nil, nil)。
func (s *Store[T]) FindByID(ctx context.Context, id int64) (*T, error) {
	return s.FindOne(ctx, bson.M{"_id": id})
}

// FindOne 按条件查一条，不存在返回 (nil, nil)。
func (s *Store[T]) FindOne(ctx context.Context, query bson.M) (*T, error) {
	var doc T
	err := s.coll.FindOne(ctx, query).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// Count 统计条数。
func (s *Store[T]) Count(ctx context.Context, query bson.M) (int64, error) {
	if query == nil {
		query = bson.M{}
	}
	return s.coll.CountDocuments(ctx, query)
}

// NextID 复刻旧 PHP MongoModel::getInsertId：在 collection_ids 集合按集合名自增。
func (s *Store[T]) NextID(ctx context.Context) (int64, error) {
	var res struct {
		ID int64 `bson:"id"`
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	err := s.db.Collection("collection_ids").FindOneAndUpdate(ctx,
		bson.M{"name": s.collName},
		bson.M{"$inc": bson.M{"id": 1}},
		opts,
	).Decode(&res)
	if err != nil {
		return 0, err
	}
	return res.ID, nil
}

// Insert 插入一条文档（_id 需已预分配）。
func (s *Store[T]) Insert(ctx context.Context, doc any) error {
	_, err := s.coll.InsertOne(ctx, doc)
	return err
}

// Update 按 _id 更新指定字段。
func (s *Store[T]) Update(ctx context.Context, id int64, set bson.M) error {
	_, err := s.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set})
	return err
}

// Delete 按 _id 删除，返回删除条数。
func (s *Store[T]) Delete(ctx context.Context, id int64) (int64, error) {
	res, err := s.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}
