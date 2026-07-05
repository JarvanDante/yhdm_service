package service

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"yhdm_service/internal/model"
	"yhdm_service/internal/pkg/password"
)

// 管理员管理相关错误。
var (
	ErrUsernameRequired = errors.New("用户名不能为空")
	ErrPasswordRequired = errors.New("新增管理员密码不能为空")
	ErrUsernameExists   = errors.New("用户名已存在")
	ErrAdminNotFound    = errors.New("管理员不存在")
)

// AdminMgmtRepo 是管理员管理所需的数据访问接口（便于单元测试注入 fake）。
type AdminMgmtRepo interface {
	ListAdmins(ctx context.Context, usernameFilter string, status *int, page, size int) ([]model.AdminUser, int64, error)
	AllRoles(ctx context.Context) ([]model.AdminRole, error)
	FindAdminByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	FindAdminByID(ctx context.Context, id int64) (*model.AdminUser, error)
	NextID(ctx context.Context, collectionName string) (int64, error)
	InsertAdmin(ctx context.Context, u *model.AdminUser) error
	UpdateAdminFields(ctx context.Context, id int64, set bson.M) error
	DeleteAdmin(ctx context.Context, id int64) (int64, error)
}

// AdminListItem 是管理员列表项（字段对齐前端 admin.ts 的 AdminItem）。
type AdminListItem struct {
	ID              int64  `json:"id"`
	Username        string `json:"username"`
	Nickname        string `json:"nickname"`         // ← real_name
	Role            string `json:"role"`             // ← 角色名（role_id=0 为超级管理员）
	LastLoginIP     string `json:"last_login_ip"`    // ← login_ip
	LastLoginTime   string `json:"last_login_time"`  // ← login_at 格式化
	CreatedAt       string `json:"created_at"`       // ← created_at 格式化
	Status          int    `json:"status"`           // ← is_disabled 取反：1 启用 / 0 禁用
	SwitchGoogle2FA int    `json:"switch_google2fa"` // ← 有 google_code 为 1
}

// AdminListResult 是列表响应（对齐前端 AdminListResponse）。
type AdminListResult struct {
	List  []AdminListItem `json:"list"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

// RoleOption 是角色下拉选项。
type RoleOption struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// AdminService 承载系统管理里"管理员"的增删改查。
type AdminService struct {
	repo AdminMgmtRepo
	rng  *rand.Rand
}

func NewAdminService(repo AdminMgmtRepo) *AdminService {
	// 使用固定源即可（盐随机性对安全非关键，密码哈希已带业务固定串）。
	return &AdminService{repo: repo, rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

const timeLayout = "2006-01-02 15:04:05"

// fmtUnix 把 unix 秒格式化为 Y-m-d H:i:s，0 值返回空串。
func fmtUnix(sec int64) string {
	if sec <= 0 {
		return ""
	}
	return time.Unix(sec, 0).Format(timeLayout)
}

// statusToDisabled：前端 status(1 启用/0 禁用) → is_disabled(0/1)。
func statusToDisabled(status int) int {
	if status == 1 {
		return 0
	}
	return 1
}

// List 返回分页的管理员列表。
func (s *AdminService) List(ctx context.Context, username string, status *int, page, size int) (*AdminListResult, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	admins, total, err := s.repo.ListAdmins(ctx, username, status, page, size)
	if err != nil {
		return nil, err
	}
	roleNames, err := s.roleNameMap(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]AdminListItem, 0, len(admins))
	for _, a := range admins {
		items = append(items, AdminListItem{
			ID:              a.ID,
			Username:        a.Username,
			Nickname:        a.RealName,
			Role:            roleName(roleNames, a.RoleID),
			LastLoginIP:     a.LoginIP,
			LastLoginTime:   fmtUnix(a.LoginAt),
			CreatedAt:       fmtUnix(a.CreatedAt),
			Status:          boolToInt(a.IsDisabled == 0),
			SwitchGoogle2FA: boolToInt(a.GoogleCode != ""),
		})
	}
	return &AdminListResult{List: items, Total: total, Page: page, Size: size}, nil
}

// RoleOptions 返回角色下拉选项。
func (s *AdminService) RoleOptions(ctx context.Context) ([]RoleOption, error) {
	roles, err := s.repo.AllRoles(ctx)
	if err != nil {
		return nil, err
	}
	opts := make([]RoleOption, 0, len(roles))
	for _, r := range roles {
		opts = append(opts, RoleOption{ID: r.ID, Name: r.Name})
	}
	return opts, nil
}

// CreateAdminInput 是新增管理员的入参。
type CreateAdminInput struct {
	Username string
	Password string
	Nickname string
	Role     int64
	Status   int
}

// Create 新增管理员，返回新 _id。
func (s *AdminService) Create(ctx context.Context, in CreateAdminInput) (int64, error) {
	if in.Username == "" {
		return 0, ErrUsernameRequired
	}
	if in.Password == "" {
		return 0, ErrPasswordRequired
	}
	exists, err := s.repo.FindAdminByUsername(ctx, in.Username)
	if err != nil {
		return 0, err
	}
	if exists != nil {
		return 0, ErrUsernameExists
	}

	salt := s.newSalt()
	now := time.Now().Unix()
	u := &model.AdminUser{
		RoleID:     in.Role,
		RealName:   in.Nickname,
		Email:      "",
		Username:   in.Username,
		Password:   password.Make(in.Password, salt),
		Slat:       salt,
		GoogleCode: "",
		LoginAt:    0,
		LoginIP:    "",
		IsDisabled: statusToDisabled(in.Status),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 自增 id 复刻旧 PHP(collection_ids)。兜底：若计数器落后于实际 max(_id)
	// 导致 _id 冲突（如本地库计数器未初始化），自动重取几次，避免抛裸的重复键错误。
	// 生产库计数器由 PHP 持续维护，正常一次成功。
	var lastErr error
	for i := 0; i < 5; i++ {
		id, err := s.repo.NextID(ctx, "admin_user")
		if err != nil {
			return 0, err
		}
		u.ID = id
		if err := s.repo.InsertAdmin(ctx, u); err != nil {
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

// UpdateAdminInput 是更新管理员的入参。Password 为空表示不改密码。
type UpdateAdminInput struct {
	ID       int64
	Username string
	Nickname string
	RoleID   int64
	Status   int
	Password string
}

// Update 更新管理员信息。
func (s *AdminService) Update(ctx context.Context, in UpdateAdminInput) error {
	cur, err := s.repo.FindAdminByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if cur == nil {
		return ErrAdminNotFound
	}
	// 若改了用户名，校验唯一性
	if in.Username != "" && in.Username != cur.Username {
		other, err := s.repo.FindAdminByUsername(ctx, in.Username)
		if err != nil {
			return err
		}
		if other != nil && other.ID != in.ID {
			return ErrUsernameExists
		}
	}

	set := bson.M{
		"real_name":   in.Nickname,
		"role_id":     in.RoleID,
		"is_disabled": statusToDisabled(in.Status),
		"updated_at":  time.Now().Unix(),
	}
	if in.Username != "" {
		set["username"] = in.Username
	}
	if in.Password != "" {
		salt := s.newSalt()
		set["slat"] = salt
		set["password"] = password.Make(in.Password, salt)
	}
	return s.repo.UpdateAdminFields(ctx, in.ID, set)
}

// Delete 删除管理员。
func (s *AdminService) Delete(ctx context.Context, id int64) error {
	n, err := s.repo.DeleteAdmin(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrAdminNotFound
	}
	return nil
}

// roleNameMap 构建 role_id → name 映射。
func (s *AdminService) roleNameMap(ctx context.Context) (map[int64]string, error) {
	roles, err := s.repo.AllRoles(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]string, len(roles))
	for _, r := range roles {
		m[r.ID] = r.Name
	}
	return m, nil
}

// newSalt 复刻旧 PHP：strval(mt_rand(10000, 50000))。
func (s *AdminService) newSalt() string {
	return strconv.Itoa(10000 + s.rng.Intn(40001))
}

func roleName(m map[int64]string, roleID int64) string {
	if roleID == 0 {
		return "超级管理员"
	}
	if n, ok := m[roleID]; ok {
		return n
	}
	return ""
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
