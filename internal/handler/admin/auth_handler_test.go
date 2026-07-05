package admin_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	adminHandler "yhdm_service/internal/handler/admin"
	"yhdm_service/internal/middleware"
	"yhdm_service/internal/model"
	"yhdm_service/internal/pkg/jwtauth"
	"yhdm_service/internal/pkg/password"
	"yhdm_service/internal/service"
)

// fakeRepo 内存实现 service.AdminRepo，供 handler 端到端(无 Mongo)测试。
type fakeRepo struct {
	user *model.AdminUser
}

func (f *fakeRepo) FindAdminByUsername(_ context.Context, username string) (*model.AdminUser, error) {
	if f.user != nil && f.user.Username == username {
		return f.user, nil
	}
	return nil, nil
}
func (f *fakeRepo) FindAdminByID(_ context.Context, id int64) (*model.AdminUser, error) {
	if f.user != nil && f.user.ID == id {
		return f.user, nil
	}
	return nil, nil
}
func (f *fakeRepo) UpdateLoginInfo(_ context.Context, _ int64, _ int64, _ string) error { return nil }
func (f *fakeRepo) FindRoleByID(_ context.Context, _ int64) (*model.AdminRole, error) {
	return nil, nil
}
func (f *fakeRepo) AllAuthorities(_ context.Context) ([]model.Authority, error) {
	return []model.Authority{}, nil
}

func setupRouter() (*gin.Engine, *jwtauth.Manager) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepo{user: &model.AdminUser{
		ID: 1, RoleID: 0, Username: "admin", Slat: "12345",
		Password: password.Make("admin123", "12345"),
	}}
	jwtMgr := jwtauth.New("test-secret", 3600)
	svc := service.NewAuthService(repo, jwtMgr, true) // dev 模式
	h := adminHandler.NewAuthHandler(svc)

	r := gin.New()
	r.POST("/app/login", h.Login)
	auth := r.Group("")
	auth.Use(middleware.Auth(jwtMgr))
	auth.GET("/app/get-info", h.GetInfo)
	return r, jwtMgr
}

func TestLoginEndpoint_Success(t *testing.T) {
	r, _ := setupRouter()
	body := `{"username":"admin","password":"admin123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/app/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP 状态应为 200, got %d", w.Code)
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Token  string `json:"token"`
			Socket string `json:"socket"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Code != 0 { // Vben 约定成功码为 0
		t.Fatalf("成功码应为 0, got %d", resp.Code)
	}
	if resp.Data.Token == "" {
		t.Fatal("应返回 token")
	}
}

func TestLoginEndpoint_WrongPassword(t *testing.T) {
	r, _ := setupRouter()
	body := `{"username":"admin","password":"nope"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/app/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code == 0 {
		t.Fatalf("密码错误时业务码不应为 0, body=%s", w.Body.String())
	}
}

func TestGetInfo_RequiresAuth(t *testing.T) {
	r, _ := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/app/get-info", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("无 token 应 401, got %d", w.Code)
	}
}

func TestGetInfo_WithToken(t *testing.T) {
	r, jwtMgr := setupRouter()
	token, _ := jwtMgr.Generate(1, "admin", 0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/app/get-info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("带 token 应 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Username string `json:"username"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Code != 0 || resp.Data.Username != "admin" {
		t.Fatalf("用户信息异常: %s", w.Body.String())
	}
}
