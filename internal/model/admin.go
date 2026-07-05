package model

// 说明：旧系统 _id 为自增整数（非 ObjectId），故用 int64 映射。
// 集合名与字段名严格对齐旧 PHP Model，保证 Go 与 PHP 读写同一份数据。

// AdminUser 对应集合 admin_user。
type AdminUser struct {
	ID         int64  `bson:"_id" json:"id"`
	RoleID     int64  `bson:"role_id" json:"role_id"`
	RealName   string `bson:"real_name" json:"real_name"`
	Email      string `bson:"email" json:"email"`
	Username   string `bson:"username" json:"username"`
	Password   string `bson:"password" json:"-"`
	Slat       string `bson:"slat" json:"-"`        // 密码盐（旧字段名拼写即 slat）
	GoogleCode string `bson:"google_code" json:"-"` // 谷歌 2FA 密钥
	LoginAt    int64  `bson:"login_at" json:"login_at"`
	LoginIP    string `bson:"login_ip" json:"login_ip"`
	IsDisabled int    `bson:"is_disabled" json:"is_disabled"`
	CreatedAt  int64  `bson:"created_at" json:"created_at"`
	UpdatedAt  int64  `bson:"updated_at" json:"updated_at"`
}

// IsSuperAdmin：role_id 为空(0) 即超级管理员，拥有全部权限。
func (a *AdminUser) IsSuperAdmin() bool { return a.RoleID == 0 }

// AdminRole 对应集合 admin_role。rights 为逗号分隔的 authority.key 列表。
type AdminRole struct {
	ID          int64  `bson:"_id" json:"id"`
	Name        string `bson:"name" json:"name"`
	Rights      string `bson:"rights" json:"rights"`
	IsDisabled  int    `bson:"is_disabled" json:"is_disabled"`
	Description string `bson:"description" json:"description"`
	CreatedAt   int64  `bson:"created_at" json:"created_at"`
	UpdatedAt   int64  `bson:"updated_at" json:"updated_at"`
}

// Authority 对应集合 authority（系统资源/菜单权限节点）。
type Authority struct {
	ID        int64  `bson:"_id" json:"id"`
	Name      string `bson:"name" json:"name"`
	ParentID  int64  `bson:"parent_id" json:"parent_id"`
	Sort      int    `bson:"sort" json:"sort"`
	Key       string `bson:"key" json:"key"`
	ClassName string `bson:"class_name" json:"class_name"` // 图标样式名
	IsMenu    int    `bson:"is_menu" json:"is_menu"`
	Link      string `bson:"link" json:"link"`
	CreatedAt int64  `bson:"created_at" json:"created_at"`
	UpdatedAt int64  `bson:"updated_at" json:"updated_at"`
}
