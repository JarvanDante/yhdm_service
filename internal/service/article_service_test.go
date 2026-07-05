package service

import (
	"context"
	"testing"

	"yhdm_service/internal/crud"
	"yhdm_service/internal/model"
)

func TestArticleSave_RequiredFields(t *testing.T) {
	svc := NewArticleService(crud.NewMemStore[model.Article]())
	_, err := svc.Save(context.Background(), ArticleInput{Title: "x", CategoryCode: "", Content: "y"})
	if err != ErrArticleFieldsRequired {
		t.Fatalf("缺分类应报错, got %v", err)
	}
}

func TestArticleSave_Create(t *testing.T) {
	repo := crud.NewMemStore[model.Article]()
	svc := NewArticleService(repo)
	id, err := svc.Save(context.Background(), ArticleInput{
		Title: "标题", CategoryCode: "news", Content: "正文", Sort: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatalf("应自增 id=1, got %d", id)
	}
	got, ok := repo.LastInsert.(model.Article)
	if !ok {
		t.Fatalf("插入类型错误: %T", repo.LastInsert)
	}
	if got.Title != "标题" || got.CategoryCode != "news" || got.Sort != 5 {
		t.Fatalf("字段映射错误: %+v", got)
	}
	if got.CreatedAt == 0 {
		t.Errorf("created_at 应被设置")
	}
}

func TestArticleSave_Update(t *testing.T) {
	repo := crud.NewMemStore[model.Article]()
	repo.FindByIDResult = &model.Article{ID: 7, Title: "旧"}
	svc := NewArticleService(repo)
	id, err := svc.Save(context.Background(), ArticleInput{ID: 7, Title: "新", CategoryCode: "c", Content: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if id != 7 || repo.LastUpdateID != 7 {
		t.Fatalf("应更新 id=7")
	}
	if repo.LastUpdateSet["title"] != "新" {
		t.Errorf("标题未更新")
	}
}

func TestArticleSave_UpdateNotFound(t *testing.T) {
	repo := crud.NewMemStore[model.Article]()
	repo.FindByIDResult = nil
	svc := NewArticleService(repo)
	_, err := svc.Save(context.Background(), ArticleInput{ID: 99, Title: "t", CategoryCode: "c", Content: "x"})
	if err != ErrRecordNotFound {
		t.Fatalf("更新不存在应报错, got %v", err)
	}
}

func TestArticleList_BuildsQuery(t *testing.T) {
	repo := crud.NewMemStore[model.Article]()
	svc := NewArticleService(repo)
	_, err := svc.List(context.Background(), "关键词", "cat1", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if repo.LastQuery["category_code"] != "cat1" {
		t.Errorf("分类筛选未生效: %v", repo.LastQuery)
	}
	if _, ok := repo.LastQuery["title"]; !ok {
		t.Errorf("标题模糊筛选未生效: %v", repo.LastQuery)
	}
}

func TestArticleDelete(t *testing.T) {
	repo := crud.NewMemStore[model.Article]()
	svc := NewArticleService(repo)
	if err := svc.Delete(context.Background(), 3); err != nil {
		t.Fatal(err)
	}
	if repo.LastDeletedID != 3 {
		t.Fatal("应删除 id=3")
	}
	repo.DeleteCount = 0
	if err := svc.Delete(context.Background(), 999); err != ErrRecordNotFound {
		t.Fatalf("删不存在应报错, got %v", err)
	}
}

func TestArticleCategory_CodeUnique(t *testing.T) {
	repo := crud.NewMemStore[model.ArticleCategory]()
	// 预置：已有一个 code=news 的分类(id=5)
	repo.FindOneResult = &model.ArticleCategory{ID: 5, Code: "news", Name: "已有"}
	svc := NewArticleCategoryService(repo)

	// 新增同 code → 冲突
	_, err := svc.Save(context.Background(), ArticleCategoryInput{Code: "news", Name: "新"})
	if err != ErrCodeExists {
		t.Fatalf("重复 code 应报错, got %v", err)
	}
	// 更新自身(id=5) 同 code → 允许
	repo.FindByIDResult = &model.ArticleCategory{ID: 5, Code: "news"}
	_, err = svc.Save(context.Background(), ArticleCategoryInput{ID: 5, Code: "news", Name: "改名"})
	if err != nil {
		t.Fatalf("更新自身不应报重复, got %v", err)
	}
}

func TestArticleCategory_RequiredFields(t *testing.T) {
	svc := NewArticleCategoryService(crud.NewMemStore[model.ArticleCategory]())
	if _, err := svc.Save(context.Background(), ArticleCategoryInput{Name: "", Code: ""}); err != ErrNameCodeRequired {
		t.Fatalf("缺名称/code 应报错, got %v", err)
	}
}

func TestBlockPosition_Create(t *testing.T) {
	repo := crud.NewMemStore[model.BlockPosition]()
	svc := NewBlockPositionService(repo)
	id, err := svc.Save(context.Background(), BlockPositionInput{Code: "home_top", Name: "首页顶部", Sort: 1})
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatalf("应自增 id=1, got %d", id)
	}
	got := repo.LastInsert.(model.BlockPosition)
	if got.Code != "home_top" || got.Name != "首页顶部" {
		t.Fatalf("字段映射错误: %+v", got)
	}
}
