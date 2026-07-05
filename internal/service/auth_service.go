package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"yhdm_service/internal/model"
	"yhdm_service/internal/pkg/jwtauth"
	"yhdm_service/internal/pkg/password"
	"yhdm_service/internal/pkg/totp"
)

// AdminRepo 是 AuthService 依赖的数据访问接口。抽成接口便于用 fake 做单元测试，
// 生产环境由 repository.AdminRepository（Mongo 实现）满足。
type AdminRepo interface {
	FindAdminByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	FindAdminByID(ctx context.Context, id int64) (*model.AdminUser, error)
	UpdateLoginInfo(ctx context.Context, id int64, loginAt int64, loginIP string) error
	FindRoleByID(ctx context.Context, id int64) (*model.AdminRole, error)
	AllAuthorities(ctx context.Context) ([]model.Authority, error)
}

// 业务错误（handler 层据此返回统一失败响应）。
var (
	ErrUserNotFoundOrDisabled = errors.New("用户不存在或已被禁用")
	ErrPasswordWrong          = errors.New("密码错误")
	ErrGoogleCodeRequired     = errors.New("请联系管理员绑定谷歌验证码")
	ErrGoogleCodeWrong        = errors.New("谷歌验证码错误")
)

// UserInfo 是 /app/get-info 返回给 Vben 的用户信息。
type UserInfo struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	RealName    string   `json:"realName"`
	Avatar      string   `json:"avatar"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// MenuItem 是 /app/menus 返回给 Vben 的菜单节点（对齐前端 menu.ts 约定）。
type MenuItem struct {
	ID       int64      `json:"id"`
	Type     int        `json:"type"` // 前端只渲染 type==1 的节点
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Sort     int        `json:"sort"`
	Icon     string     `json:"icon"`
	Children []MenuItem `json:"children"`
}

// AuthService 承载登录、用户信息、权限码、菜单等后台认证逻辑。
type AuthService struct {
	repo AdminRepo
	jwt  *jwtauth.Manager
	dev  bool
}

func NewAuthService(repo AdminRepo, jwt *jwtauth.Manager, dev bool) *AuthService {
	return &AuthService{repo: repo, jwt: jwt, dev: dev}
}

// Login 复刻旧 PHP AdminUserRepository::login。
// 返回签发的 JWT。dev 模式下跳过谷歌 2FA（对应旧 isDevEnv()）。
// TODO: 旧系统还有 IP 白名单校验（读 configs 集合），后续补上。
func (s *AuthService) Login(ctx context.Context, username, plainPwd, googleCode, clientIP string) (string, *model.AdminUser, error) {
	u, err := s.repo.FindAdminByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}
	if u == nil || u.IsDisabled != 0 {
		return "", nil, ErrUserNotFoundOrDisabled
	}
	if !s.dev {
		if u.GoogleCode == "" {
			return "", nil, ErrGoogleCodeRequired
		}
		if !totp.Verify(u.GoogleCode, googleCode) {
			return "", nil, ErrGoogleCodeWrong
		}
	}
	if !password.Verify(plainPwd, u.Slat, u.Password) {
		return "", nil, ErrPasswordWrong
	}
	// 更新登录信息（尽力而为，不阻断登录）
	_ = s.repo.UpdateLoginInfo(ctx, u.ID, time.Now().Unix(), clientIP)

	token, err := s.jwt.Generate(u.ID, u.Username, u.RoleID)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

// GetUserInfo 返回当前管理员的信息（含角色、权限码）。
func (s *AuthService) GetUserInfo(ctx context.Context, adminID int64) (*UserInfo, error) {
	u, err := s.repo.FindAdminByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if u == nil || u.IsDisabled != 0 {
		return nil, ErrUserNotFoundOrDisabled
	}
	codes, roleName, err := s.accessCodes(ctx, u)
	if err != nil {
		return nil, err
	}
	return &UserInfo{
		ID:          u.ID,
		Username:    u.Username,
		RealName:    u.RealName,
		Avatar:      "",
		Roles:       []string{roleName},
		Permissions: codes,
	}, nil
}

// AccessCodes 返回权限码列表（Vben 按钮级权限用）。超管返回 ["*"]。
func (s *AuthService) AccessCodes(ctx context.Context, adminID int64) ([]string, error) {
	u, err := s.repo.FindAdminByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return []string{}, nil
	}
	codes, _, err := s.accessCodes(ctx, u)
	return codes, err
}

// accessCodes 计算该管理员拥有的所有 authority.key，并返回角色名。
func (s *AuthService) accessCodes(ctx context.Context, u *model.AdminUser) ([]string, string, error) {
	auths, err := s.repo.AllAuthorities(ctx)
	if err != nil {
		return nil, "", err
	}
	if u.IsSuperAdmin() {
		return []string{"*"}, "超级管理员", nil
	}
	role, err := s.repo.FindRoleByID(ctx, u.RoleID)
	if err != nil {
		return nil, "", err
	}
	if role == nil || role.IsDisabled != 0 {
		return []string{}, "", nil
	}
	allowed := splitRights(role.Rights)
	codes := make([]string, 0, len(auths))
	for _, a := range auths {
		if _, ok := allowed[a.Key]; ok {
			codes = append(codes, a.Key)
		}
	}
	return codes, role.Name, nil
}

// Menus 返回当前管理员可见的菜单树（对齐旧 getMenus：仅 is_menu 节点）。
func (s *AuthService) Menus(ctx context.Context, adminID int64) ([]MenuItem, error) {
	u, err := s.repo.FindAdminByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if u == nil || u.IsDisabled != 0 {
		return []MenuItem{}, nil
	}
	auths, err := s.repo.AllAuthorities(ctx)
	if err != nil {
		return nil, err
	}

	// 超管：全部；否则按 role.rights 过滤子节点。
	var allowed map[string]struct{}
	if !u.IsSuperAdmin() {
		role, err := s.repo.FindRoleByID(ctx, u.RoleID)
		if err != nil {
			return nil, err
		}
		if role == nil || role.IsDisabled != 0 {
			return []MenuItem{}, nil
		}
		allowed = splitRights(role.Rights)
	}

	return buildMenuTree(auths, allowed), nil
}

// buildMenuTree 构建两级菜单树。allowed 为 nil 表示超管（放行全部）。
func buildMenuTree(auths []model.Authority, allowed map[string]struct{}) []MenuItem {
	// 顶层：parent_id == 0 且 is_menu；子层挂在其下。
	childrenByParent := make(map[int64][]model.Authority)
	var roots []model.Authority
	for _, a := range auths {
		if a.ParentID == 0 {
			roots = append(roots, a)
		} else {
			childrenByParent[a.ParentID] = append(childrenByParent[a.ParentID], a)
		}
	}

	sortAuths := func(list []model.Authority) {
		sort.SliceStable(list, func(i, j int) bool { return list[i].Sort < list[j].Sort })
	}
	sortAuths(roots)

	var menus []MenuItem
	for _, root := range roots {
		if root.IsMenu == 0 {
			continue
		}
		kids := childrenByParent[root.ID]
		sortAuths(kids)
		var children []MenuItem
		for _, kid := range kids {
			if kid.IsMenu == 0 {
				continue
			}
			if allowed != nil {
				if _, ok := allowed[kid.Key]; !ok {
					continue
				}
			}
			children = append(children, toMenuItem(kid))
		}
		// 旧逻辑：子节点被过滤空的父节点不展示（超管例外，保留有菜单子项的父级）。
		if allowed != nil && len(children) == 0 {
			continue
		}
		item := toMenuItem(root)
		item.Children = children
		menus = append(menus, item)
	}
	return menus
}

func toMenuItem(a model.Authority) MenuItem {
	path := a.Link
	if path == "" {
		path = a.Key
	}
	return MenuItem{
		ID:       a.ID,
		Type:     1, // 前端只渲染 type==1
		Name:     a.Name,
		Path:     path,
		Sort:     a.Sort,
		Icon:     a.ClassName,
		Children: []MenuItem{},
	}
}

func splitRights(rights string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, k := range strings.Split(rights, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			m[k] = struct{}{}
		}
	}
	return m
}
