package service

import (
	"context"
	"testing"

	"yhdm_service/internal/model"
	"yhdm_service/internal/pkg/jwtauth"
	"yhdm_service/internal/pkg/password"
)

// fakeRepo 是 AdminRepo 的内存实现，用于无 Mongo 的单元测试。
type fakeRepo struct {
	users       map[string]*model.AdminUser // key=username
	usersByID   map[int64]*model.AdminUser
	roles       map[int64]*model.AdminRole
	authorities []model.Authority
	updated     bool
}

func (f *fakeRepo) FindAdminByUsername(_ context.Context, username string) (*model.AdminUser, error) {
	return f.users[username], nil
}
func (f *fakeRepo) FindAdminByID(_ context.Context, id int64) (*model.AdminUser, error) {
	return f.usersByID[id], nil
}
func (f *fakeRepo) UpdateLoginInfo(_ context.Context, _ int64, _ int64, _ string) error {
	f.updated = true
	return nil
}
func (f *fakeRepo) FindRoleByID(_ context.Context, id int64) (*model.AdminRole, error) {
	return f.roles[id], nil
}
func (f *fakeRepo) AllAuthorities(_ context.Context) ([]model.Authority, error) {
	return f.authorities, nil
}

// 构造一套测试夹具：一个超管、一个受限角色用户，以及两级 authority 树。
func newFixture() *fakeRepo {
	// slat=12345, password=admin123 -> 复用 password 包生成真实哈希
	super := &model.AdminUser{ID: 1, RoleID: 0, Username: "super", Slat: "12345",
		Password: password.Make("admin123", "12345")}
	editor := &model.AdminUser{ID: 2, RoleID: 10, Username: "editor", Slat: "12345",
		Password: password.Make("editor123", "12345")}
	disabled := &model.AdminUser{ID: 3, RoleID: 0, Username: "banned", IsDisabled: 1,
		Slat: "12345", Password: password.Make("x", "12345")}

	return &fakeRepo{
		users:     map[string]*model.AdminUser{"super": super, "editor": editor, "banned": disabled},
		usersByID: map[int64]*model.AdminUser{1: super, 2: editor, 3: disabled},
		roles: map[int64]*model.AdminRole{
			10: {ID: 10, Name: "内容编辑", Rights: model.Rights{"content", "content_movie"}},
		},
		authorities: []model.Authority{
			{ID: 100, ParentID: 0, Name: "内容管理", Key: "content", IsMenu: 1, Sort: 1, Link: "/content", ClassName: "video"},
			{ID: 101, ParentID: 100, Name: "动漫", Key: "content_movie", IsMenu: 1, Sort: 1, Link: "/content/movie"},
			{ID: 102, ParentID: 100, Name: "漫画", Key: "content_comics", IsMenu: 1, Sort: 2, Link: "/content/comics"},
			{ID: 200, ParentID: 0, Name: "系统管理", Key: "system", IsMenu: 1, Sort: 2, Link: "/system"},
			{ID: 201, ParentID: 200, Name: "管理员", Key: "system_admin", IsMenu: 1, Sort: 1, Link: "/system/admin"},
		},
	}
}

func newSvc(repo AdminRepo, dev bool) *AuthService {
	return NewAuthService(repo, jwtauth.New("test-secret", 3600), dev)
}

func TestLogin_Success(t *testing.T) {
	svc := newSvc(newFixture(), true) // dev 模式跳过 2FA
	token, u, err := svc.Login(context.Background(), "super", "admin123", "", "127.0.0.1")
	if err != nil {
		t.Fatalf("登录应成功, got err=%v", err)
	}
	if token == "" || u == nil || u.ID != 1 {
		t.Fatalf("返回 token/user 异常: token=%q user=%+v", token, u)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newSvc(newFixture(), true)
	_, _, err := svc.Login(context.Background(), "super", "wrong", "", "127.0.0.1")
	if err != ErrPasswordWrong {
		t.Fatalf("应返回密码错误, got %v", err)
	}
}

func TestLogin_DisabledUser(t *testing.T) {
	svc := newSvc(newFixture(), true)
	_, _, err := svc.Login(context.Background(), "banned", "x", "", "127.0.0.1")
	if err != ErrUserNotFoundOrDisabled {
		t.Fatalf("禁用用户应拒绝, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := newSvc(newFixture(), true)
	_, _, err := svc.Login(context.Background(), "ghost", "x", "", "127.0.0.1")
	if err != ErrUserNotFoundOrDisabled {
		t.Fatalf("不存在用户应拒绝, got %v", err)
	}
}

func TestLogin_GoogleCodeRequiredInProd(t *testing.T) {
	svc := newSvc(newFixture(), false) // 非 dev：需要谷歌验证码
	_, _, err := svc.Login(context.Background(), "super", "admin123", "", "127.0.0.1")
	if err != ErrGoogleCodeRequired {
		t.Fatalf("生产模式无 google_code 应报错, got %v", err)
	}
}

func TestAccessCodes_SuperAdmin(t *testing.T) {
	svc := newSvc(newFixture(), true)
	codes, err := svc.AccessCodes(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != 1 || codes[0] != "*" {
		t.Fatalf("超管权限码应为 [*], got %v", codes)
	}
}

func TestAccessCodes_RoleLimited(t *testing.T) {
	svc := newSvc(newFixture(), true)
	codes, err := svc.AccessCodes(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	// editor 角色 rights=content,content_movie
	got := map[string]bool{}
	for _, c := range codes {
		got[c] = true
	}
	if !got["content"] || !got["content_movie"] {
		t.Fatalf("应包含 content/content_movie, got %v", codes)
	}
	if got["content_comics"] || got["system_admin"] {
		t.Fatalf("不应包含未授权项, got %v", codes)
	}
}

func TestMenus_SuperAdminSeesAll(t *testing.T) {
	svc := newSvc(newFixture(), true)
	menus, err := svc.Menus(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(menus) != 2 {
		t.Fatalf("超管应看到 2 个顶级菜单, got %d", len(menus))
	}
	// 校验排序与 type、子节点数
	if menus[0].Name != "内容管理" || menus[0].Type != 1 || len(menus[0].Children) != 2 {
		t.Fatalf("内容管理菜单异常: %+v", menus[0])
	}
}

func TestMenus_RoleFiltered(t *testing.T) {
	svc := newSvc(newFixture(), true)
	menus, err := svc.Menus(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	// editor 只有 content/content_movie：系统管理整块应被过滤掉
	if len(menus) != 1 || menus[0].Name != "内容管理" {
		t.Fatalf("受限角色应只见内容管理, got %+v", menus)
	}
	if len(menus[0].Children) != 1 || menus[0].Children[0].Name != "动漫" {
		t.Fatalf("受限角色子菜单应只有动漫, got %+v", menus[0].Children)
	}
}
