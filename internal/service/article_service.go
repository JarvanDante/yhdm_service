package service

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"yhdm_service/internal/crud"
	"yhdm_service/internal/model"
)

// 文章管理：article / article_category / block_position 三个集合的 CRUD。

var (
	ErrArticleFieldsRequired = errors.New("标题、分类、内容不能为空")
	ErrNameCodeRequired      = errors.New("名称和标识(code)不能为空")
	ErrCodeExists            = errors.New("标识(code)已存在")
	ErrRecordNotFound        = errors.New("数据不存在")
)

// ---------- 文章 Article ----------

type ArticleService struct{ repo crud.Repo[model.Article] }

func NewArticleService(repo crud.Repo[model.Article]) *ArticleService {
	return &ArticleService{repo: repo}
}

// ArticleListResult 分页响应。
type ArticleListResult struct {
	List  []model.Article `json:"list"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

func (s *ArticleService) List(ctx context.Context, name, categoryCode string, page, size int) (*ArticleListResult, error) {
	query := bson.M{}
	if name != "" {
		query["title"] = bson.M{"$regex": name, "$options": "i"}
	}
	if categoryCode != "" {
		query["category_code"] = categoryCode
	}
	page, size = normPage(page, size)
	list, total, err := s.repo.List(ctx, query, nil, page, size)
	if err != nil {
		return nil, err
	}
	return &ArticleListResult{List: list, Total: total, Page: page, Size: size}, nil
}

func (s *ArticleService) Detail(ctx context.Context, id int64) (*model.Article, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, ErrRecordNotFound
	}
	return a, nil
}

// ArticleInput 新增/更新入参（ID>0 为更新）。
type ArticleInput struct {
	ID             int64
	Title          string
	CategoryCode   string
	Content        string
	Img            string
	SeoKeywords    string
	SeoDescription string
	URL            string
	IsRecommend    int
	Sort           int
	Click          int
	ShowDialog     int
}

func (s *ArticleService) Save(ctx context.Context, in ArticleInput) (int64, error) {
	if in.Title == "" || in.CategoryCode == "" || in.Content == "" {
		return 0, ErrArticleFieldsRequired
	}
	now := time.Now().Unix()
	fields := bson.M{
		"title": in.Title, "category_code": in.CategoryCode, "content": in.Content,
		"img": in.Img, "seo_keywords": in.SeoKeywords, "seo_description": in.SeoDescription,
		"url": in.URL, "is_recommend": in.IsRecommend, "sort": in.Sort,
		"click": in.Click, "show_dialog": in.ShowDialog, "updated_at": now,
	}
	if in.ID > 0 {
		cur, err := s.repo.FindByID(ctx, in.ID)
		if err != nil {
			return 0, err
		}
		if cur == nil {
			return 0, ErrRecordNotFound
		}
		return in.ID, s.repo.Update(ctx, in.ID, fields)
	}
	return crud.Create(ctx, s.repo, func(id int64) any {
		a := model.Article{ID: id, CreatedAt: now}
		applyArticle(&a, fields)
		return a
	})
}

func (s *ArticleService) Delete(ctx context.Context, id int64) error {
	return simpleDelete(ctx, s.repo, id)
}

func applyArticle(a *model.Article, f bson.M) {
	a.Title, _ = f["title"].(string)
	a.CategoryCode, _ = f["category_code"].(string)
	a.Content, _ = f["content"].(string)
	a.Img, _ = f["img"].(string)
	a.SeoKeywords, _ = f["seo_keywords"].(string)
	a.SeoDescription, _ = f["seo_description"].(string)
	a.URL, _ = f["url"].(string)
	a.IsRecommend, _ = f["is_recommend"].(int)
	a.Sort, _ = f["sort"].(int)
	a.Click, _ = f["click"].(int)
	a.ShowDialog, _ = f["show_dialog"].(int)
	a.UpdatedAt, _ = f["updated_at"].(int64)
}

// ---------- 文章分类 ArticleCategory ----------

type ArticleCategoryService struct{ repo crud.Repo[model.ArticleCategory] }

func NewArticleCategoryService(repo crud.Repo[model.ArticleCategory]) *ArticleCategoryService {
	return &ArticleCategoryService{repo: repo}
}

type ArticleCategoryListResult struct {
	List  []model.ArticleCategory `json:"list"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
}

func (s *ArticleCategoryService) List(ctx context.Context, name string, page, size int) (*ArticleCategoryListResult, error) {
	query := bson.M{}
	if name != "" {
		query["name"] = bson.M{"$regex": name, "$options": "i"}
	}
	page, size = normPage(page, size)
	list, total, err := s.repo.List(ctx, query, nil, page, size)
	if err != nil {
		return nil, err
	}
	return &ArticleCategoryListResult{List: list, Total: total, Page: page, Size: size}, nil
}

type ArticleCategoryInput struct {
	ID       int64
	Code     string
	Name     string
	Img      string
	Language string
	Sort     int
	ParentID int64
}

func (s *ArticleCategoryService) Save(ctx context.Context, in ArticleCategoryInput) (int64, error) {
	if in.Name == "" || in.Code == "" {
		return 0, ErrNameCodeRequired
	}
	if err := ensureCodeUnique(ctx, s.repo, in.Code, in.ID); err != nil {
		return 0, err
	}
	now := time.Now().Unix()
	if in.ID > 0 {
		cur, err := s.repo.FindByID(ctx, in.ID)
		if err != nil {
			return 0, err
		}
		if cur == nil {
			return 0, ErrRecordNotFound
		}
		return in.ID, s.repo.Update(ctx, in.ID, bson.M{
			"code": in.Code, "name": in.Name, "img": in.Img, "language": in.Language,
			"sort": in.Sort, "parent_id": in.ParentID, "updated_at": now,
		})
	}
	return crud.Create(ctx, s.repo, func(id int64) any {
		return model.ArticleCategory{ID: id, Code: in.Code, Name: in.Name, Img: in.Img,
			Language: in.Language, Sort: in.Sort, ParentID: in.ParentID, CreatedAt: now, UpdatedAt: now}
	})
}

func (s *ArticleCategoryService) Delete(ctx context.Context, id int64) error {
	return simpleDelete(ctx, s.repo, id)
}

// ---------- 模块位置 BlockPosition ----------

type BlockPositionService struct{ repo crud.Repo[model.BlockPosition] }

func NewBlockPositionService(repo crud.Repo[model.BlockPosition]) *BlockPositionService {
	return &BlockPositionService{repo: repo}
}

type BlockPositionListResult struct {
	List  []model.BlockPosition `json:"list"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Size  int                   `json:"size"`
}

func (s *BlockPositionService) List(ctx context.Context, name string, page, size int) (*BlockPositionListResult, error) {
	query := bson.M{}
	if name != "" {
		query["name"] = bson.M{"$regex": name, "$options": "i"}
	}
	page, size = normPage(page, size)
	list, total, err := s.repo.List(ctx, query, nil, page, size)
	if err != nil {
		return nil, err
	}
	return &BlockPositionListResult{List: list, Total: total, Page: page, Size: size}, nil
}

type BlockPositionInput struct {
	ID   int64
	Code string
	Name string
	Sort int
}

func (s *BlockPositionService) Save(ctx context.Context, in BlockPositionInput) (int64, error) {
	if in.Name == "" || in.Code == "" {
		return 0, ErrNameCodeRequired
	}
	if err := ensureCodeUnique(ctx, s.repo, in.Code, in.ID); err != nil {
		return 0, err
	}
	now := time.Now().Unix()
	if in.ID > 0 {
		cur, err := s.repo.FindByID(ctx, in.ID)
		if err != nil {
			return 0, err
		}
		if cur == nil {
			return 0, ErrRecordNotFound
		}
		return in.ID, s.repo.Update(ctx, in.ID, bson.M{
			"code": in.Code, "name": in.Name, "sort": in.Sort, "updated_at": now,
		})
	}
	return crud.Create(ctx, s.repo, func(id int64) any {
		return model.BlockPosition{ID: id, Code: in.Code, Name: in.Name, Sort: in.Sort, CreatedAt: now, UpdatedAt: now}
	})
}

func (s *BlockPositionService) Delete(ctx context.Context, id int64) error {
	return simpleDelete(ctx, s.repo, id)
}

// ---------- 公共小工具 ----------

func normPage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 15
	}
	return page, size
}

func simpleDelete[T any](ctx context.Context, repo crud.Repo[T], id int64) error {
	n, err := repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// ensureCodeUnique 校验 code 唯一（更新时排除自身）。
func ensureCodeUnique[T crud.Identifiable](ctx context.Context, repo crud.Repo[T], code string, selfID int64) error {
	dup, err := repo.FindOne(ctx, bson.M{"code": code})
	if err != nil {
		return err
	}
	if dup != nil && (*dup).GetID() != selfID {
		return ErrCodeExists
	}
	return nil
}
