package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"yhdm_service/internal/model"
)

// AdminRepository 封装 admin_user / admin_role / authority 三个集合的读取。
type AdminRepository struct {
	db *mongo.Database
}

func NewAdminRepository(db *mongo.Database) *AdminRepository {
	return &AdminRepository{db: db}
}

// FindAdminByUsername 按用户名查管理员。
func (r *AdminRepository) FindAdminByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.Collection("admin_user").FindOne(ctx, bson.M{"username": username}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindAdminByID 按 _id 查管理员。
func (r *AdminRepository) FindAdminByID(ctx context.Context, id int64) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.Collection("admin_user").FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateLoginInfo 更新登录时间与 IP。
func (r *AdminRepository) UpdateLoginInfo(ctx context.Context, id int64, loginAt int64, loginIP string) error {
	_, err := r.db.Collection("admin_user").UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"login_at": loginAt, "login_ip": loginIP}},
	)
	return err
}

// FindRoleByID 按 _id 查角色。
func (r *AdminRepository) FindRoleByID(ctx context.Context, id int64) (*model.AdminRole, error) {
	var role model.AdminRole
	err := r.db.Collection("admin_role").FindOne(ctx, bson.M{"_id": id}).Decode(&role)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// AllAuthorities 取全部权限节点，按 sort 升序（对齐旧系统树构建顺序）。
func (r *AdminRepository) AllAuthorities(ctx context.Context) ([]model.Authority, error) {
	cur, err := r.db.Collection("authority").Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "sort", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []model.Authority
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
