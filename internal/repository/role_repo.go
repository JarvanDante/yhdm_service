package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"yhdm_service/internal/model"
)

// 系统管理-角色(admin_role) 的数据访问。

// InsertRole 插入角色（_id 需已由 NextID 预分配）。
func (r *AdminRepository) InsertRole(ctx context.Context, role *model.AdminRole) error {
	_, err := r.db.Collection("admin_role").InsertOne(ctx, role)
	return err
}

// UpdateRoleFields 按 _id 更新角色字段。
func (r *AdminRepository) UpdateRoleFields(ctx context.Context, id int64, set bson.M) error {
	_, err := r.db.Collection("admin_role").UpdateOne(ctx,
		bson.M{"_id": id}, bson.M{"$set": set})
	return err
}

// DeleteRole 按 _id 删除角色。
func (r *AdminRepository) DeleteRole(ctx context.Context, id int64) (int64, error) {
	res, err := r.db.Collection("admin_role").DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}
