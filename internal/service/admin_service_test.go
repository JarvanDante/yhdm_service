package service

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"yhdm_service/internal/model"
	"yhdm_service/internal/pkg/password"
)

// fakeMgmtRepo 是 AdminMgmtRepo 的内存实现。
type fakeMgmtRepo struct {
	byID     map[int64]*model.AdminUser
	byName   map[string]*model.AdminUser
	roles    []model.AdminRole
	nextID   int64
	inserted *model.AdminUser
	updated  bson.M
	updateID int64
	deleted  int64
}

func newFakeMgmtRepo() *fakeMgmtRepo {
	admin := &model.AdminUser{ID: 1, Username: "admin", RoleID: 0, RealName: "超管",
		Slat: "12345", Password: password.Make("x", "12345")}
	editor := &model.AdminUser{ID: 2, Username: "editor", RoleID: 10, RealName: "编辑",
		IsDisabled: 1, GoogleCode: "SECRET", Slat: "12345"}
	return &fakeMgmtRepo{
		byID:   map[int64]*model.AdminUser{1: admin, 2: editor},
		byName: map[string]*model.AdminUser{"admin": admin, "editor": editor},
		roles:  []model.AdminRole{{ID: 10, Name: "内容编辑"}},
		nextID: 100,
	}
}

func (f *fakeMgmtRepo) ListAdmins(_ context.Context, username string, status *int, _, _ int) ([]model.AdminUser, int64, error) {
	var out []model.AdminUser
	for _, u := range f.byID {
		if username != "" && u.Username != username {
			continue
		}
		if status != nil {
			disabled := 0
			if *status == 0 {
				disabled = 1
			}
			if u.IsDisabled != disabled {
				continue
			}
		}
		out = append(out, *u)
	}
	return out, int64(len(out)), nil
}
func (f *fakeMgmtRepo) AllRoles(_ context.Context) ([]model.AdminRole, error) { return f.roles, nil }
func (f *fakeMgmtRepo) FindAdminByUsername(_ context.Context, username string) (*model.AdminUser, error) {
	return f.byName[username], nil
}
func (f *fakeMgmtRepo) FindAdminByID(_ context.Context, id int64) (*model.AdminUser, error) {
	return f.byID[id], nil
}
func (f *fakeMgmtRepo) NextID(_ context.Context, _ string) (int64, error) {
	f.nextID++
	return f.nextID, nil
}
func (f *fakeMgmtRepo) InsertAdmin(_ context.Context, u *model.AdminUser) error {
	f.inserted = u
	f.byID[u.ID] = u
	f.byName[u.Username] = u
	return nil
}
func (f *fakeMgmtRepo) UpdateAdminFields(_ context.Context, id int64, set bson.M) error {
	f.updateID = id
	f.updated = set
	return nil
}
func (f *fakeMgmtRepo) DeleteAdmin(_ context.Context, id int64) (int64, error) {
	if _, ok := f.byID[id]; ok {
		delete(f.byID, id)
		f.deleted = id
		return 1, nil
	}
	return 0, nil
}

func TestAdminList_MapsFields(t *testing.T) {
	svc := NewAdminService(newFakeMgmtRepo())
	res, err := svc.List(context.Background(), "editor", nil, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.List) != 1 {
		t.Fatalf("应返回 1 条, got %d", len(res.List))
	}
	it := res.List[0]
	// editor: is_disabled=1 → status=0；有 google_code → switch=1；role_id=10 → 内容编辑
	if it.Status != 0 {
		t.Errorf("禁用用户 status 应为 0, got %d", it.Status)
	}
	if it.SwitchGoogle2FA != 1 {
		t.Errorf("绑定谷歌应为 1, got %d", it.SwitchGoogle2FA)
	}
	if it.Role != "内容编辑" {
		t.Errorf("角色名映射错误, got %q", it.Role)
	}
	if it.Nickname != "编辑" {
		t.Errorf("nickname 应映射 real_name, got %q", it.Nickname)
	}
}

func TestAdminList_SuperAdminRoleName(t *testing.T) {
	svc := NewAdminService(newFakeMgmtRepo())
	res, _ := svc.List(context.Background(), "admin", nil, 1, 10)
	if res.List[0].Role != "超级管理员" {
		t.Fatalf("role_id=0 应显示超级管理员, got %q", res.List[0].Role)
	}
	if res.List[0].Status != 1 {
		t.Fatalf("启用用户 status 应为 1, got %d", res.List[0].Status)
	}
}

func TestAdminCreate_Success(t *testing.T) {
	repo := newFakeMgmtRepo()
	svc := NewAdminService(repo)
	id, err := svc.Create(context.Background(), CreateAdminInput{
		Username: "newone", Password: "pass123", Nickname: "新人", Role: 10, Status: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != 101 {
		t.Fatalf("应使用自增 id 101, got %d", id)
	}
	if repo.inserted == nil || repo.inserted.Username != "newone" {
		t.Fatal("未正确插入")
	}
	// status=1 → is_disabled=0；密码应被哈希且能校验
	if repo.inserted.IsDisabled != 0 {
		t.Errorf("status=1 应 is_disabled=0")
	}
	if !password.Verify("pass123", repo.inserted.Slat, repo.inserted.Password) {
		t.Errorf("密码哈希不可校验")
	}
}

func TestAdminCreate_DuplicateUsername(t *testing.T) {
	svc := NewAdminService(newFakeMgmtRepo())
	_, err := svc.Create(context.Background(), CreateAdminInput{
		Username: "admin", Password: "x", Nickname: "重名", Role: 0, Status: 1,
	})
	if err != ErrUsernameExists {
		t.Fatalf("应报用户名已存在, got %v", err)
	}
}

func TestAdminCreate_EmptyPassword(t *testing.T) {
	svc := NewAdminService(newFakeMgmtRepo())
	_, err := svc.Create(context.Background(), CreateAdminInput{
		Username: "nopass", Password: "", Nickname: "x", Role: 0, Status: 1,
	})
	if err != ErrPasswordRequired {
		t.Fatalf("空密码应报错, got %v", err)
	}
}

func TestAdminUpdate_ChangesFields(t *testing.T) {
	repo := newFakeMgmtRepo()
	svc := NewAdminService(repo)
	err := svc.Update(context.Background(), UpdateAdminInput{
		ID: 2, Username: "editor", Nickname: "改名", RoleID: 10, Status: 0, Password: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.updateID != 2 {
		t.Fatalf("更新目标 id 错误")
	}
	if repo.updated["real_name"] != "改名" {
		t.Errorf("real_name 未更新")
	}
	if repo.updated["is_disabled"] != 1 { // status=0 → is_disabled=1
		t.Errorf("status=0 应 is_disabled=1, got %v", repo.updated["is_disabled"])
	}
	if _, hasPwd := repo.updated["password"]; hasPwd {
		t.Errorf("空密码不应更新 password 字段")
	}
}

func TestAdminUpdate_NotFound(t *testing.T) {
	svc := NewAdminService(newFakeMgmtRepo())
	err := svc.Update(context.Background(), UpdateAdminInput{ID: 999, Status: 1})
	if err != ErrAdminNotFound {
		t.Fatalf("不存在应报错, got %v", err)
	}
}

func TestAdminDelete(t *testing.T) {
	repo := newFakeMgmtRepo()
	svc := NewAdminService(repo)
	if err := svc.Delete(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	if repo.deleted != 2 {
		t.Fatalf("应删除 id=2")
	}
	if err := svc.Delete(context.Background(), 999); err != ErrAdminNotFound {
		t.Fatalf("删除不存在应报错, got %v", err)
	}
}

func TestRoleOptions(t *testing.T) {
	svc := NewAdminService(newFakeMgmtRepo())
	opts, err := svc.RoleOptions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) != 1 || opts[0].Name != "内容编辑" {
		t.Fatalf("角色选项异常, got %+v", opts)
	}
}
