package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"yhdm_service/internal/response"
	"yhdm_service/internal/service"
)

// ArticleHandler 处理文章管理：文章 / 文章分类 / 模块位置。
type ArticleHandler struct {
	article  *service.ArticleService
	category *service.ArticleCategoryService
	block    *service.BlockPositionService
}

func NewArticleHandler(a *service.ArticleService, c *service.ArticleCategoryService, b *service.BlockPositionService) *ArticleHandler {
	return &ArticleHandler{article: a, category: c, block: b}
}

// ---------- 文章 ----------

// ArticleList 文章列表
// @Summary  文章列表(分页)
// @Tags     文章管理
// @Produce  json
// @Security BearerAuth
// @Param    page           query  int     false  "页码"
// @Param    size           query  int     false  "每页"
// @Param    name           query  string  false  "标题模糊"
// @Param    category_code  query  string  false  "分类"
// @Success  200  {object}  response.Body{data=service.ArticleListResult}
// @Router   /app/articles [get]
func (h *ArticleHandler) ArticleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "15"))
	res, err := h.article.List(c.Request.Context(), c.Query("name"), c.Query("category_code"), page, size)
	respond(c, res, err)
}

// ArticleDetail 文章详情
// @Summary  文章详情
// @Tags     文章管理
// @Produce  json
// @Security BearerAuth
// @Param    id   query  int  true  "文章ID"
// @Success  200  {object}  response.Body
// @Router   /app/article-detail [get]
func (h *ArticleHandler) ArticleDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	res, err := h.article.Detail(c.Request.Context(), id)
	respond(c, res, err)
}

type saveArticleReq struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	CategoryCode   string `json:"category_code"`
	Content        string `json:"content"`
	Img            string `json:"img"`
	SeoKeywords    string `json:"seo_keywords"`
	SeoDescription string `json:"seo_description"`
	URL            string `json:"url"`
	IsRecommend    int    `json:"is_recommend"`
	Sort           int    `json:"sort"`
	Click          int    `json:"click"`
	ShowDialog     int    `json:"show_dialog"`
}

// SaveArticle 新增/更新文章
// @Summary  新增/更新文章(id>0 为更新)
// @Tags     文章管理
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body  saveArticleReq  true  "文章内容"
// @Success  200  {object}  response.Body{data=object}  "data: {id, message}"
// @Router   /app/save-article [post]
func (h *ArticleHandler) SaveArticle(c *gin.Context) {
	var req saveArticleReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	id, err := h.article.Save(c.Request.Context(), service.ArticleInput{
		ID: req.ID, Title: req.Title, CategoryCode: req.CategoryCode, Content: req.Content,
		Img: req.Img, SeoKeywords: req.SeoKeywords, SeoDescription: req.SeoDescription,
		URL: req.URL, IsRecommend: req.IsRecommend, Sort: req.Sort, Click: req.Click, ShowDialog: req.ShowDialog,
	})
	respondSave(c, id, err)
}

// DeleteArticle 删除文章
// @Summary  删除文章
// @Tags     文章管理
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body  idReq  true  "文章ID"
// @Success  200  {object}  response.Body
// @Router   /app/delete-article [post]
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	deleteByID(c, h.article.Delete)
}

// ---------- 文章分类 ----------

// CategoryList 文章分类列表
// @Summary  文章分类列表(分页)
// @Tags     文章管理
// @Produce  json
// @Security BearerAuth
// @Param    page  query  int     false  "页码"
// @Param    size  query  int     false  "每页"
// @Param    name  query  string  false  "名称模糊"
// @Success  200  {object}  response.Body{data=service.ArticleCategoryListResult}
// @Router   /app/article-categories [get]
func (h *ArticleHandler) CategoryList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "15"))
	res, err := h.category.List(c.Request.Context(), c.Query("name"), page, size)
	respond(c, res, err)
}

type saveCategoryReq struct {
	ID       int64  `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Img      string `json:"img"`
	Language string `json:"language"`
	Sort     int    `json:"sort"`
	ParentID int64  `json:"parent_id"`
}

// SaveCategory 新增/更新文章分类
// @Summary  新增/更新文章分类(id>0 为更新)
// @Tags     文章管理
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body  saveCategoryReq  true  "分类信息"
// @Success  200  {object}  response.Body{data=object}  "data: {id, message}"
// @Router   /app/save-article-category [post]
func (h *ArticleHandler) SaveCategory(c *gin.Context) {
	var req saveCategoryReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	id, err := h.category.Save(c.Request.Context(), service.ArticleCategoryInput{
		ID: req.ID, Code: req.Code, Name: req.Name, Img: req.Img,
		Language: req.Language, Sort: req.Sort, ParentID: req.ParentID,
	})
	respondSave(c, id, err)
}

// DeleteCategory 删除文章分类
// @Summary  删除文章分类
// @Tags     文章管理
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body  idReq  true  "分类ID"
// @Success  200  {object}  response.Body
// @Router   /app/delete-article-category [post]
func (h *ArticleHandler) DeleteCategory(c *gin.Context) {
	deleteByID(c, h.category.Delete)
}

// ---------- 模块位置 ----------

// BlockList 模块位置列表
// @Summary  模块位置列表(分页)
// @Tags     文章管理
// @Produce  json
// @Security BearerAuth
// @Param    page  query  int     false  "页码"
// @Param    size  query  int     false  "每页"
// @Param    name  query  string  false  "名称模糊"
// @Success  200  {object}  response.Body{data=service.BlockPositionListResult}
// @Router   /app/block-positions [get]
func (h *ArticleHandler) BlockList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "15"))
	res, err := h.block.List(c.Request.Context(), c.Query("name"), page, size)
	respond(c, res, err)
}

type saveBlockReq struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Sort int    `json:"sort"`
}

// SaveBlock 新增/更新模块位置
// @Summary  新增/更新模块位置(id>0 为更新)
// @Tags     文章管理
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body  saveBlockReq  true  "模块位置信息"
// @Success  200  {object}  response.Body{data=object}  "data: {id, message}"
// @Router   /app/save-block-position [post]
func (h *ArticleHandler) SaveBlock(c *gin.Context) {
	var req saveBlockReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	id, err := h.block.Save(c.Request.Context(), service.BlockPositionInput{
		ID: req.ID, Code: req.Code, Name: req.Name, Sort: req.Sort,
	})
	respondSave(c, id, err)
}

// DeleteBlock 删除模块位置
// @Summary  删除模块位置
// @Tags     文章管理
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body  idReq  true  "模块位置ID"
// @Success  200  {object}  response.Body
// @Router   /app/delete-block-position [post]
func (h *ArticleHandler) DeleteBlock(c *gin.Context) {
	deleteByID(c, h.block.Delete)
}
