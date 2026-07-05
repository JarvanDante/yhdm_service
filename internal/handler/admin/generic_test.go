package admin_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	adminHandler "yhdm_service/internal/handler/admin"
	"yhdm_service/internal/crud"
)

func setupGeneric(repo crud.Repo[bson.M]) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := adminHandler.NewGenericHandler(repo, adminHandler.ModuleConfig{
		Collection: "demo",
		Search: []adminHandler.SearchField{
			{Param: "name", Field: "name", Regex: true},
			{Param: "status", Field: "status", Int: true},
		},
		Edit:     []string{"name", "code", "sort", "status"},
		Required: []string{"name", "code"},
		Unique:   "code",
	})
	r := gin.New()
	g := r.Group("")
	h.Register(g, "demos", "demo")
	return r
}

func doJSON(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func decode(w *httptest.ResponseRecorder) map[string]any {
	var m map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &m)
	return m
}

func TestGeneric_ListBuildsQueryAndReturns(t *testing.T) {
	repo := crud.NewMemStore[bson.M]()
	repo.Items = []bson.M{{"_id": int64(1), "name": "甲"}, {"_id": int64(2), "name": "乙"}}
	r := setupGeneric(repo)

	w := doJSON(r, http.MethodGet, "/app/demos?name=甲&status=1&page=1&size=10", "")
	if w.Code != 200 {
		t.Fatalf("HTTP %d", w.Code)
	}
	m := decode(w)
	if m["code"].(float64) != 0 {
		t.Fatalf("code!=0: %v", m)
	}
	// 校验查询条件被正确构造
	if _, ok := repo.LastQuery["name"]; !ok {
		t.Errorf("name 模糊查询未构造: %v", repo.LastQuery)
	}
	if repo.LastQuery["status"] != 1 {
		t.Errorf("status 整数查询未构造: %v", repo.LastQuery)
	}
}

func TestGeneric_CreateSuccess(t *testing.T) {
	repo := crud.NewMemStore[bson.M]()
	r := setupGeneric(repo)
	w := doJSON(r, http.MethodPost, "/app/save-demo", `{"name":"新项","code":"c1","sort":3}`)
	m := decode(w)
	if m["code"].(float64) != 0 {
		t.Fatalf("创建应成功: %v", m)
	}
	doc, ok := repo.LastInsert.(bson.M)
	if !ok {
		t.Fatalf("插入类型 %T", repo.LastInsert)
	}
	if doc["name"] != "新项" || doc["code"] != "c1" {
		t.Errorf("字段未写入: %v", doc)
	}
	if doc["_id"] == nil || doc["created_at"] == nil {
		t.Errorf("应有 _id 与 created_at: %v", doc)
	}
}

func TestGeneric_RequiredValidation(t *testing.T) {
	repo := crud.NewMemStore[bson.M]()
	r := setupGeneric(repo)
	w := doJSON(r, http.MethodPost, "/app/save-demo", `{"name":"只有名字"}`)
	m := decode(w)
	if m["code"].(float64) == 0 {
		t.Fatalf("缺 code 应失败: %v", m)
	}
}

func TestGeneric_UniqueConflict(t *testing.T) {
	repo := crud.NewMemStore[bson.M]()
	repo.FindOneResult = &bson.M{"_id": int64(9), "code": "dup"}
	r := setupGeneric(repo)
	// 新增使用已存在 code
	w := doJSON(r, http.MethodPost, "/app/save-demo", `{"name":"x","code":"dup"}`)
	m := decode(w)
	if m["code"].(float64) == 0 {
		t.Fatalf("重复 code 应失败: %v", m)
	}
}

func TestGeneric_UpdateKeepsOwnUnique(t *testing.T) {
	repo := crud.NewMemStore[bson.M]()
	repo.FindOneResult = &bson.M{"_id": int64(9), "code": "dup"}
	repo.FindByIDResult = &bson.M{"_id": int64(9), "code": "dup"}
	r := setupGeneric(repo)
	// 更新自身(id=9)且 code 不变，应允许
	w := doJSON(r, http.MethodPost, "/app/save-demo", `{"id":9,"name":"改","code":"dup"}`)
	m := decode(w)
	if m["code"].(float64) != 0 {
		t.Fatalf("更新自身不应报重复: %v", m)
	}
	if repo.LastUpdateID != 9 {
		t.Errorf("应更新 id=9, got %d", repo.LastUpdateID)
	}
}

func TestGeneric_UpdateNotFound(t *testing.T) {
	repo := crud.NewMemStore[bson.M]()
	repo.FindByIDResult = nil
	r := setupGeneric(repo)
	w := doJSON(r, http.MethodPost, "/app/save-demo", `{"id":999,"name":"x","code":"c"}`)
	m := decode(w)
	if m["code"].(float64) == 0 {
		t.Fatalf("更新不存在应失败: %v", m)
	}
}

func TestGeneric_Delete(t *testing.T) {
	repo := crud.NewMemStore[bson.M]()
	r := setupGeneric(repo)
	w := doJSON(r, http.MethodPost, "/app/delete-demo", `{"id":5}`)
	m := decode(w)
	if m["code"].(float64) != 0 {
		t.Fatalf("删除应成功: %v", m)
	}
	if repo.LastDeletedID != 5 {
		t.Errorf("应删除 id=5, got %d", repo.LastDeletedID)
	}
	// 删除不存在
	repo.DeleteCount = 0
	w = doJSON(r, http.MethodPost, "/app/delete-demo", `{"id":123}`)
	if decode(w)["code"].(float64) == 0 {
		t.Fatalf("删不存在应失败")
	}
}
