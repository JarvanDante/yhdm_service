package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"yhdm_service/internal/model"
)

// 系统管理（管理员）相关的数据访问，均为 *AdminRepository 的方法。

// NextID 复刻旧 PHP MongoModel::getInsertId：在 collection_ids 集合里按集合名自增。
// findAndModify {name: collectionName} $inc id, upsert, returnNew。
func (r *AdminRepository) NextID(ctx context.Context, collectionName string) (int64, error) {
	var res struct {
		ID int64 `bson:"id"`
	}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)
	err := r.db.Collection("collection_ids").FindOneAndUpdate(ctx,
		bson.M{"name": collectionName},
		bson.M{"$inc": bson.M{"id": 1}},
		opts,
	).Decode(&res)
	if err != nil {
		return 0, err
	}
	return res.ID, nil
}

// ListAdmins 分页查询管理员。usernameFilter 非空则按用户名模糊匹配；
// status 非 nil 则按状态过滤（前端 status 1=启用/0=禁用，映射到 is_disabled 取反）。
func (r *AdminRepository) ListAdmins(ctx context.Context, usernameFilter string, status *int, page, size int) ([]model.AdminUser, int64, error) {
	query := bson.M{}
	if usernameFilter != "" {
		query["username"] = bson.M{"$regex": usernameFilter, "$options": "i"}
	}
	if status != nil {
		disabled := 0
		if *status == 0 {
			disabled = 1
		}
		query["is_disabled"] = disabled
	}

	coll := r.db.Collection("admin_user")
	total, err := coll.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	findOpts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetSkip(int64((page - 1) * size)).
		SetLimit(int64(size))
	cur, err := coll.Find(ctx, query, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	var list []model.AdminUser
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AllRoles 取全部角色（用于列表里的角色名映射和下拉选项）。
func (r *AdminRepository) AllRoles(ctx context.Context) ([]model.AdminRole, error) {
	cur, err := r.db.Collection("admin_role").Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []model.AdminRole
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// InsertAdmin 插入管理员（_id 需已由 NextID 预分配）。
func (r *AdminRepository) InsertAdmin(ctx context.Context, u *model.AdminUser) error {
	_, err := r.db.Collection("admin_user").InsertOne(ctx, u)
	return err
}

// UpdateAdminFields 按 _id 更新指定字段。
func (r *AdminRepository) UpdateAdminFields(ctx context.Context, id int64, set bson.M) error {
	_, err := r.db.Collection("admin_user").UpdateOne(ctx,
		bson.M{"_id": id}, bson.M{"$set": set})
	return err
}

// DeleteAdmin 按 _id 删除管理员。
func (r *AdminRepository) DeleteAdmin(ctx context.Context, id int64) (int64, error) {
	res, err := r.db.Collection("admin_user").DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}
