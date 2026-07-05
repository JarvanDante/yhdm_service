package service

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"yhdm_service/internal/model"
)

// fakeRoleRepo 是 RoleMgmtRepo 的内存实现。
type fakeRoleRepo struct {
	roles       map[int64]*model.AdminRole
	authorities []model.Authority
	nextID      int64
	inserted    *model.AdminRole
	updateID    int64
	updated     bson.M
	deleted     int64
}

func newFakeRoleRepo() *fakeRoleRepo {
	return &fakeRoleRepo{
		roles: map[int64]*model.AdminRole{
			10: {ID: 10, Name: "内容编辑", Rights: model.Rights{"content", "content_movie"}, CreatedAt: 1700000000},
		},
		authorities: []model.Authority{
			{ID: 100, ParentID: 0, Name: "内容管理", Key: "content", Sort: 1},
			{ID: 101, ParentID: 100, Name: "动漫", Key: "content_movie", Sort: 1},
			{ID: 102, ParentID: 100, Name: "漫画", Key: "content_comics", Sort: 2},
			{ID: 200, ParentID: 0, Name: "系统管理", Key: "system", Sort: 2},
		},
		nextID: 20,
	}
}

func (f *fakeRoleRepo) AllRoles(_ context.Context) ([]model.AdminRole, error) {
	out := make([]model.AdminRole, 0, len(f.roles))
	for _, r := range f.roles {
		out = append(out, *r)
	}
	return out, nil
}
func (f *fakeRoleRepo) FindRoleByID(_ context.Context, id int64) (*model.AdminRole, error) {
	return f.roles[id], nil
}
func (f *fakeRoleRepo) AllAuthorities(_ context.Context) ([]model.Authority, error) {
	return f.authorities, nil
}
func (f *fakeRoleRepo) NextID(_ context.Context, _ string) (int64, error) {
	f.nextID++
	return f.nextID, nil
}
func (f *fakeRoleRepo) InsertRole(_ context.Context, role *model.AdminRole) error {
	f.inserted = role
	f.roles[role.ID] = role
	return nil
}
func (f *fakeRoleRepo) UpdateRoleFields(_ context.Context, id int64, set bson.M) error {
	f.updateID = id
	f.updated = set
	return nil
}
func (f *fakeRoleRepo) DeleteRole(_ context.Context, id int64) (int64, error) {
	if _, ok := f.roles[id]; ok {
		delete(f.roles, id)
		f.deleted = id
		return 1, nil
	}
	return 0, nil
}

func TestRoleList_JoinsRights(t *testing.T) {
	svc := NewRoleService(newFakeRoleRepo())
	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("应有 1 个角色, got %d", len(list))
	}
	if list[0].Permissions != "content,content_movie" {
		t.Fatalf("rights 应逗号连接, got %q", list[0].Permissions)
	}
}

func TestRolePermissions_Tree(t *testing.T) {
	svc := NewRoleService(newFakeRoleRepo())
	tree, err := svc.Permissions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != 2 {
		t.Fatalf("应有 2 个顶级权限, got %d", len(tree))
	}
	// 内容管理(sort=1) 在前，有 2 个子节点
	if tree[0].Name != "内容管理" || len(tree[0].Children) != 2 {
		t.Fatalf("内容管理树异常: %+v", tree[0])
	}
}

func TestRoleCreate(t *testing.T) {
	repo := newFakeRoleRepo()
	svc := NewRoleService(repo)
	id, err := svc.Create(context.Background(), "运营")
	if err != nil {
		t.Fatal(err)
	}
	if id != 21 {
		t.Fatalf("应用自增 id 21, got %d", id)
	}
	if repo.inserted == nil || repo.inserted.Name != "运营" {
		t.Fatal("未正确插入角色")
	}
	if repo.inserted.Rights == nil {
		t.Fatal("新角色 rights 应为空数组而非 nil")
	}
}

func TestRoleCreate_EmptyName(t *testing.T) {
	svc := NewRoleService(newFakeRoleRepo())
	_, err := svc.Create(context.Background(), "")
	if err != ErrRoleNameRequired {
		t.Fatalf("空名应报错, got %v", err)
	}
}

func TestRoleUpdateName(t *testing.T) {
	repo := newFakeRoleRepo()
	svc := NewRoleService(repo)
	if err := svc.UpdateName(context.Background(), 10, "新编辑"); err != nil {
		t.Fatal(err)
	}
	if repo.updated["name"] != "新编辑" {
		t.Fatalf("名称未更新, got %v", repo.updated["name"])
	}
}

func TestRoleUpdateName_NotFound(t *testing.T) {
	svc := NewRoleService(newFakeRoleRepo())
	if err := svc.UpdateName(context.Background(), 999, "x"); err != ErrRoleNotFound {
		t.Fatalf("不存在应报错, got %v", err)
	}
}

func TestRoleSavePermission(t *testing.T) {
	repo := newFakeRoleRepo()
	svc := NewRoleService(repo)
	if err := svc.SavePermission(context.Background(), 10, "content, content_movie ,content_comics"); err != nil {
		t.Fatal(err)
	}
	rights, ok := repo.updated["rights"].(model.Rights)
	if !ok {
		t.Fatalf("rights 应存为数组, got %T", repo.updated["rights"])
	}
	if len(rights) != 3 {
		t.Fatalf("应解析出 3 个权限(去空格), got %v", rights)
	}
}

func TestRoleDelete(t *testing.T) {
	repo := newFakeRoleRepo()
	svc := NewRoleService(repo)
	if err := svc.Delete(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	if repo.deleted != 10 {
		t.Fatal("应删除 id=10")
	}
	if err := svc.Delete(context.Background(), 999); err != ErrRoleNotFound {
		t.Fatalf("删不存在应报错, got %v", err)
	}
}
