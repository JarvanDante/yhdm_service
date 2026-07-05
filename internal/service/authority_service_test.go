package service

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"yhdm_service/internal/model"
	"yhdm_service/internal/repository"
)

// fakeAuthRepo 是 AuthorityMgmtRepo 的内存实现。
type fakeAuthRepo struct {
	byID     map[int64]*model.Authority
	byKey    map[string]*model.Authority
	nextID   int64
	inserted *model.Authority
	updateID int64
	updated  bson.M
	deleted  int64
}

func newFakeAuthRepo() *fakeAuthRepo {
	a := &model.Authority{ID: 1, Name: "活动专区", Key: "/activityLand", ParentID: 0, IsMenu: 1}
	return &fakeAuthRepo{
		byID:   map[int64]*model.Authority{1: a},
		byKey:  map[string]*model.Authority{"/activityLand": a},
		nextID: 100,
	}
}

func (f *fakeAuthRepo) ListAuthorities(_ context.Context, _ repository.AuthorityFilter, _, _ int) ([]model.Authority, int64, error) {
	out := make([]model.Authority, 0, len(f.byID))
	for _, a := range f.byID {
		out = append(out, *a)
	}
	return out, int64(len(out)), nil
}
func (f *fakeAuthRepo) FindAuthorityByID(_ context.Context, id int64) (*model.Authority, error) {
	return f.byID[id], nil
}
func (f *fakeAuthRepo) FindAuthorityByKey(_ context.Context, key string) (*model.Authority, error) {
	return f.byKey[key], nil
}
func (f *fakeAuthRepo) NextID(_ context.Context, _ string) (int64, error) {
	f.nextID++
	return f.nextID, nil
}
func (f *fakeAuthRepo) InsertAuthority(_ context.Context, a *model.Authority) error {
	f.inserted = a
	f.byID[a.ID] = a
	f.byKey[a.Key] = a
	return nil
}
func (f *fakeAuthRepo) UpdateAuthorityFields(_ context.Context, id int64, set bson.M) error {
	f.updateID = id
	f.updated = set
	return nil
}
func (f *fakeAuthRepo) DeleteAuthority(_ context.Context, id int64) (int64, error) {
	if _, ok := f.byID[id]; ok {
		delete(f.byID, id)
		f.deleted = id
		return 1, nil
	}
	return 0, nil
}

func TestAuthoritySave_Create(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthorityService(repo)
	id, err := svc.Save(context.Background(), AuthorityInput{
		Name: "金刚区", Key: "/kingkong", ParentID: 0, IsMenu: 1, Sort: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != 101 {
		t.Fatalf("应自增 id 101, got %d", id)
	}
	if repo.inserted == nil || repo.inserted.Key != "/kingkong" {
		t.Fatal("未正确插入")
	}
}

func TestAuthoritySave_Update(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthorityService(repo)
	id, err := svc.Save(context.Background(), AuthorityInput{
		ID: 1, Name: "活动专区改", Key: "/activityLand", IsMenu: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 || repo.updateID != 1 {
		t.Fatalf("应更新 id=1")
	}
	if repo.updated["name"] != "活动专区改" {
		t.Errorf("name 未更新")
	}
}

func TestAuthoritySave_EmptyNameKey(t *testing.T) {
	svc := NewAuthorityService(newFakeAuthRepo())
	if _, err := svc.Save(context.Background(), AuthorityInput{Name: "", Key: ""}); err != ErrAuthNameKeyRequired {
		t.Fatalf("空名/key 应报错, got %v", err)
	}
}

func TestAuthoritySave_DuplicateKey(t *testing.T) {
	svc := NewAuthorityService(newFakeAuthRepo())
	// 新建时用了已存在的 key
	_, err := svc.Save(context.Background(), AuthorityInput{Name: "重复", Key: "/activityLand"})
	if err != ErrAuthKeyExists {
		t.Fatalf("重复 key 应报错, got %v", err)
	}
}

func TestAuthoritySave_UpdateKeepsOwnKey(t *testing.T) {
	svc := NewAuthorityService(newFakeAuthRepo())
	// 更新自身、key 不变，不应误判重复
	_, err := svc.Save(context.Background(), AuthorityInput{ID: 1, Name: "x", Key: "/activityLand"})
	if err != nil {
		t.Fatalf("更新自身不应报重复 key, got %v", err)
	}
}

func TestAuthorityDetail_NotFound(t *testing.T) {
	svc := NewAuthorityService(newFakeAuthRepo())
	if _, err := svc.Detail(context.Background(), 999); err != ErrAuthNotFound {
		t.Fatalf("不存在应报错, got %v", err)
	}
}

func TestAuthorityDelete(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthorityService(repo)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if repo.deleted != 1 {
		t.Fatal("应删除 id=1")
	}
	if err := svc.Delete(context.Background(), 999); err != ErrAuthNotFound {
		t.Fatalf("删不存在应报错, got %v", err)
	}
}
