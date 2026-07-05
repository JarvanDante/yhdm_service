package service

import (
	"context"
	"errors"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"yhdm_service/internal/model"
)

var (
	ErrRoleNameRequired = errors.New("角色名不能为空")
	ErrRoleNotFound     = errors.New("角色不存在")
)

// RoleMgmtRepo 是角色管理所需的数据访问接口。
type RoleMgmtRepo interface {
	AllRoles(ctx context.Context) ([]model.AdminRole, error)
	FindRoleByID(ctx context.Context, id int64) (*model.AdminRole, error)
	AllAuthorities(ctx context.Context) ([]model.Authority, error)
	NextID(ctx context.Context, collectionName string) (int64, error)
	InsertRole(ctx context.Context, role *model.AdminRole) error
	UpdateRoleFields(ctx context.Context, id int64, set bson.M) error
	DeleteRole(ctx context.Context, id int64) (int64, error)
}

// RoleListItem 对齐前端 role.ts 的 RoleItem。
type RoleListItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Permissions string `json:"permissions"` // ← rights 以逗号连接
	CreatedAt   string `json:"created_at"`
}

// PermissionNode 对齐前端 role.ts 的 PermissionItem（权限树节点）。
type PermissionNode struct {
	ID       int64            `json:"id"`
	Name     string           `json:"name"`
	ParentID int64            `json:"parent_id"`
	Children []PermissionNode `json:"children"`
}

// RoleService 承载系统管理里"角色"的增删改查与权限分配。
type RoleService struct {
	repo RoleMgmtRepo
}

func NewRoleService(repo RoleMgmtRepo) *RoleService {
	return &RoleService{repo: repo}
}

// List 返回全部角色。
func (s *RoleService) List(ctx context.Context) ([]RoleListItem, error) {
	roles, err := s.repo.AllRoles(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]RoleListItem, 0, len(roles))
	for _, r := range roles {
		items = append(items, RoleListItem{
			ID:          r.ID,
			Name:        r.Name,
			Permissions: r.Rights.String(),
			CreatedAt:   fmtUnix(r.CreatedAt),
		})
	}
	return items, nil
}

// Permissions 返回权限树（authority 全部节点，构造两级树）。
func (s *RoleService) Permissions(ctx context.Context) ([]PermissionNode, error) {
	auths, err := s.repo.AllAuthorities(ctx)
	if err != nil {
		return nil, err
	}
	childrenByParent := make(map[int64][]model.Authority)
	var roots []model.Authority
	for _, a := range auths {
		if a.ParentID == 0 {
			roots = append(roots, a)
		} else {
			childrenByParent[a.ParentID] = append(childrenByParent[a.ParentID], a)
		}
	}
	sortBySort := func(list []model.Authority) {
		sort.SliceStable(list, func(i, j int) bool { return list[i].Sort < list[j].Sort })
	}
	sortBySort(roots)

	nodes := make([]PermissionNode, 0, len(roots))
	for _, root := range roots {
		kids := childrenByParent[root.ID]
		sortBySort(kids)
		children := make([]PermissionNode, 0, len(kids))
		for _, kid := range kids {
			children = append(children, PermissionNode{ID: kid.ID, Name: kid.Name, ParentID: kid.ParentID, Children: []PermissionNode{}})
		}
		nodes = append(nodes, PermissionNode{ID: root.ID, Name: root.Name, ParentID: 0, Children: children})
	}
	return nodes, nil
}

// Create 新增角色（仅名称，权限后续用 SavePermission 分配）。
func (s *RoleService) Create(ctx context.Context, name string) (int64, error) {
	if name == "" {
		return 0, ErrRoleNameRequired
	}
	id, err := s.repo.NextID(ctx, "admin_role")
	if err != nil {
		return 0, err
	}
	now := time.Now().Unix()
	role := &model.AdminRole{
		ID:          id,
		Name:        name,
		Rights:      model.Rights{},
		IsDisabled:  0,
		Description: "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.InsertRole(ctx, role); err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateName 更新角色名称。
func (s *RoleService) UpdateName(ctx context.Context, id int64, name string) error {
	if name == "" {
		return ErrRoleNameRequired
	}
	role, err := s.repo.FindRoleByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}
	return s.repo.UpdateRoleFields(ctx, id, bson.M{"name": name, "updated_at": time.Now().Unix()})
}

// SavePermission 给角色分配权限（permissionList 为逗号分隔的 authority.key）。
func (s *RoleService) SavePermission(ctx context.Context, id int64, permissionList string) error {
	role, err := s.repo.FindRoleByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}
	rights := model.SplitCSV(permissionList) // 存为数组，对齐旧 PHP array_values
	return s.repo.UpdateRoleFields(ctx, id, bson.M{"rights": rights, "updated_at": time.Now().Unix()})
}

// Delete 删除角色。
func (s *RoleService) Delete(ctx context.Context, id int64) error {
	n, err := s.repo.DeleteRole(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrRoleNotFound
	}
	return nil
}
