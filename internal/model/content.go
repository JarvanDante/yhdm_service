package model

// 内容/文章相关集合模型，字段严格对齐 yhdm-php 各 Model 的 @property 注释。

// Article 对应集合 article（文章）。
type Article struct {
	ID             int64  `bson:"_id" json:"id"`
	Title          string `bson:"title" json:"title"`
	CategoryCode   string `bson:"category_code" json:"category_code"`
	Content        string `bson:"content" json:"content"`
	Img            string `bson:"img" json:"img"`
	SeoKeywords    string `bson:"seo_keywords" json:"seo_keywords"`
	SeoDescription string `bson:"seo_description" json:"seo_description"`
	URL            string `bson:"url" json:"url"`
	IsRecommend    int    `bson:"is_recommend" json:"is_recommend"`
	Sort           int    `bson:"sort" json:"sort"`
	Click          int    `bson:"click" json:"click"`
	ShowDialog     int    `bson:"show_dialog" json:"show_dialog"`
	CreatedAt      int64  `bson:"created_at" json:"created_at"`
	UpdatedAt      int64  `bson:"updated_at" json:"updated_at"`
}

// GetID 实现 crud.Identifiable。
func (a Article) GetID() int64 { return a.ID }

// ArticleCategory 对应集合 article_category（文章分类）。
type ArticleCategory struct {
	ID        int64  `bson:"_id" json:"id"`
	Code      string `bson:"code" json:"code"`
	Name      string `bson:"name" json:"name"`
	Img       string `bson:"img" json:"img"`
	Language  string `bson:"language" json:"language"`
	Sort      int    `bson:"sort" json:"sort"`
	ParentID  int64  `bson:"parent_id" json:"parent_id"`
	CreatedAt int64  `bson:"created_at" json:"created_at"`
	UpdatedAt int64  `bson:"updated_at" json:"updated_at"`
}

// GetID 实现 crud.Identifiable。
func (a ArticleCategory) GetID() int64 { return a.ID }

// BlockPosition 对应集合 block_position（模块位置）。
type BlockPosition struct {
	ID        int64  `bson:"_id" json:"id"`
	Code      string `bson:"code" json:"code"`
	Name      string `bson:"name" json:"name"`
	Sort      int    `bson:"sort" json:"sort"`
	CreatedAt int64  `bson:"created_at" json:"created_at"`
	UpdatedAt int64  `bson:"updated_at" json:"updated_at"`
}

// GetID 实现 crud.Identifiable。
func (a BlockPosition) GetID() int64 { return a.ID }
