package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"yhdm_service/internal/model"
)

// 系统管理-权限资源(authority) 的数据访问。

// AuthorityFilter 是权限资源列表的筛选条件（nil 指针表示不筛选）。
type AuthorityFilter struct {
	Name     string
	Key      string
	ParentID *int64
	IsMenu   *int
}

// ListAuthorities 分页查询权限资源。
func (r *AdminRepository) ListAuthorities(ctx context.Context, f AuthorityFilter, page, size int) ([]model.Authority, int64, error) {
	query := bson.M{}
	if f.Name != "" {
		query["name"] = bson.M{"$regex": f.Name, "$options": "i"}
	}
	if f.Key != "" {
		query["key"] = bson.M{"$regex": f.Key, "$options": "i"}
	}
	if f.ParentID != nil {
		query["parent_id"] = *f.ParentID
	}
	if f.IsMenu != nil {
		query["is_menu"] = *f.IsMenu
	}

	coll := r.db.Collection("authority")
	total, err := coll.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 15
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetSkip(int64((page - 1) * size)).
		SetLimit(int64(size))
	cur, err := coll.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	var list []model.Authority
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindAuthorityByID 按 _id 查询。
func (r *AdminRepository) FindAuthorityByID(ctx context.Context, id int64) (*model.Authority, error) {
	var a model.Authority
	err := r.db.Collection("authority").FindOne(ctx, bson.M{"_id": id}).Decode(&a)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// FindAuthorityByKey 按 key 查询（用于唯一性校验）。
func (r *AdminRepository) FindAuthorityByKey(ctx context.Context, key string) (*model.Authority, error) {
	var a model.Authority
	err := r.db.Collection("authority").FindOne(ctx, bson.M{"key": key}).Decode(&a)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// InsertAuthority 插入权限资源（_id 需已由 NextID 预分配）。
func (r *AdminRepository) InsertAuthority(ctx context.Context, a *model.Authority) error {
	_, err := r.db.Collection("authority").InsertOne(ctx, a)
	return err
}

// UpdateAuthorityFields 按 _id 更新字段。
func (r *AdminRepository) UpdateAuthorityFields(ctx context.Context, id int64, set bson.M) error {
	_, err := r.db.Collection("authority").UpdateOne(ctx,
		bson.M{"_id": id}, bson.M{"$set": set})
	return err
}

// DeleteAuthority 按 _id 删除。
func (r *AdminRepository) DeleteAuthority(ctx context.Context, id int64) (int64, error) {
	res, err := r.db.Collection("authority").DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}
