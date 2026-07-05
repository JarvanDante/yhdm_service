package service

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"yhdm_service/internal/model"
	"yhdm_service/internal/repository"
)

var (
	ErrAuthNameKeyRequired = errors.New("名称和标识(key)不能为空")
	ErrAuthKeyExists       = errors.New("当前 key 已存在")
	ErrAuthNotFound        = errors.New("数据不存在")
)

// AuthorityMgmtRepo 是权限资源管理所需的数据访问接口。
type AuthorityMgmtRepo interface {
	ListAuthorities(ctx context.Context, f repository.AuthorityFilter, page, size int) ([]model.Authority, int64, error)
	FindAuthorityByID(ctx context.Context, id int64) (*model.Authority, error)
	FindAuthorityByKey(ctx context.Context, key string) (*model.Authority, error)
	NextID(ctx context.Context, collectionName string) (int64, error)
	InsertAuthority(ctx context.Context, a *model.Authority) error
	UpdateAuthorityFields(ctx context.Context, id int64, set bson.M) error
	DeleteAuthority(ctx context.Context, id int64) (int64, error)
}

// AuthorityListItem 是权限资源列表项。
type AuthorityListItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	ParentID  int64  `json:"parent_id"`
	Sort      int    `json:"sort"`
	ClassName string `json:"class_name"`
	IsMenu    int    `json:"is_menu"`
	Link      string `json:"link"`
	CreatedAt string `json:"created_at"`
}

// AuthorityListResult 是分页响应。
type AuthorityListResult struct {
	List  []AuthorityListItem `json:"list"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}

// AuthorityService 承载系统管理里"权限资源(authority)"的增删改查。
type AuthorityService struct {
	repo AuthorityMgmtRepo
}

func NewAuthorityService(repo AuthorityMgmtRepo) *AuthorityService {
	return &AuthorityService{repo: repo}
}

// List 分页查询权限资源。
func (s *AuthorityService) List(ctx context.Context, f repository.AuthorityFilter, page, size int) (*AuthorityListResult, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 15
	}
	auths, total, err := s.repo.ListAuthorities(ctx, f, page, size)
	if err != nil {
		return nil, err
	}
	items := make([]AuthorityListItem, 0, len(auths))
	for _, a := range auths {
		items = append(items, toAuthorityListItem(a))
	}
	return &AuthorityListResult{List: items, Total: total, Page: page, Size: size}, nil
}

// Detail 查询详情。
func (s *AuthorityService) Detail(ctx context.Context, id int64) (*model.Authority, error) {
	a, err := s.repo.FindAuthorityByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, ErrAuthNotFound
	}
	return a, nil
}

// AuthorityInput 是新增/更新权限资源的入参。
type AuthorityInput struct {
	ID        int64
	Name      string
	Key       string
	ParentID  int64
	Sort      int
	ClassName string
	IsMenu    int
	Link      string
}

// Save 新增或更新（ID>0 为更新）。复刻旧 PHP：name/key 必填，key 全局唯一。
func (s *AuthorityService) Save(ctx context.Context, in AuthorityInput) (int64, error) {
	if in.Name == "" || in.Key == "" {
		return 0, ErrAuthNameKeyRequired
	}
	// key 唯一性校验（排除自身）
	dup, err := s.repo.FindAuthorityByKey(ctx, in.Key)
	if err != nil {
		return 0, err
	}
	if dup != nil && dup.ID != in.ID {
		return 0, ErrAuthKeyExists
	}

	now := time.Now().Unix()
	if in.ID > 0 {
		// 更新
		cur, err := s.repo.FindAuthorityByID(ctx, in.ID)
		if err != nil {
			return 0, err
		}
		if cur == nil {
			return 0, ErrAuthNotFound
		}
		set := bson.M{
			"name":       in.Name,
			"key":        in.Key,
			"parent_id":  in.ParentID,
			"sort":       in.Sort,
			"class_name": in.ClassName,
			"is_menu":    in.IsMenu,
			"link":       in.Link,
			"updated_at": now,
		}
		if err := s.repo.UpdateAuthorityFields(ctx, in.ID, set); err != nil {
			return 0, err
		}
		return in.ID, nil
	}

	// 新增（自增 _id，重复键兜底重试）
	a := &model.Authority{
		Name: in.Name, Key: in.Key, ParentID: in.ParentID, Sort: in.Sort,
		ClassName: in.ClassName, IsMenu: in.IsMenu, Link: in.Link,
		CreatedAt: now, UpdatedAt: now,
	}
	var lastErr error
	for i := 0; i < 5; i++ {
		id, err := s.repo.NextID(ctx, "authority")
		if err != nil {
			return 0, err
		}
		a.ID = id
		if err := s.repo.InsertAuthority(ctx, a); err != nil {
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

// Delete 删除权限资源。
func (s *AuthorityService) Delete(ctx context.Context, id int64) error {
	n, err := s.repo.DeleteAuthority(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrAuthNotFound
	}
	return nil
}

func toAuthorityListItem(a model.Authority) AuthorityListItem {
	return AuthorityListItem{
		ID: a.ID, Name: a.Name, Key: a.Key, ParentID: a.ParentID, Sort: a.Sort,
		ClassName: a.ClassName, IsMenu: a.IsMenu, Link: a.Link,
		CreatedAt: fmtUnix(a.CreatedAt),
	}
}
