// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package schema

import (
	"bytes"

	"time"

	"github.com/yuin/goldmark"
	"github.com/zhian9/blogo-server/internal/config"
	rschema "github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

const (
	ArticleStatusDraft     = "draft"     // 草稿
	ArticleStatusPublished = "published" // 已发布

	ArticleVisibilityPublic         = "public"          // 公开可见
	ArticleVisibilityPrivate        = "private"         // 仅作者可见
	ArticleVisibilityPartialVisible = "partial_visible" // 部分人可见
)

// ArticleVisibilityTypes 所有可见性类型（用于 binding oneof 校验）
var ArticleVisibilityTypes = []string{
	ArticleVisibilityPublic,
	ArticleVisibilityPrivate,
	ArticleVisibilityPartialVisible,
}

// goldmark 实例（线程安全，全局复用）
var md = goldmark.New()

// Article 文章表，数据持久化
type Article struct {
	ID           string                `json:"id" gorm:"size:20;primarykey;"`                   // 唯一ID
	Title        string                `json:"title" gorm:"size:255;not null;index"`            // 文章标题
	Slug         string                `json:"slug" gorm:"size:255;uniqueIndex"`                // SEO友好URL路径
	Summary      string                `json:"summary" gorm:"type:text;"`                       // 摘要
	Content      string                `json:"content" gorm:"type:longtext;"`                   // Markdown 原文
	HtmlContent  string                `json:"html_content" gorm:"type:longtext;"`              // 预渲染 HTML
	CoverImageID *string               `json:"cover_image_id" gorm:"size:20;index"`             // 封面图ID
	CoverImage   *Image                `json:"cover_image,omitempty" gorm:"-"`                  // 封面图（手动加载）
	Author       *rschema.User         `json:"author,omitempty" gorm:"-"`                       // 作者信息（手动加载）
	CategoryID   string                `json:"category_id" gorm:"size:20;index"`                // 分类ID
	AuthorID     string                `json:"author_id" gorm:"size:20;index"`                  // 作者用户ID
	Tags         Tags                  `json:"tags" gorm:"many2many:article_tag;"`              // 标签（多对多）
	Views        int64                 `json:"views" gorm:"default:0;"`                         // 浏览量
	IsTop        bool                  `json:"is_top" gorm:"default:false;"`                    // 是否置顶
	Status       string                `json:"status" gorm:"size:20;index;default:draft;"`      // 状态：draft / published
	Visibility   string                `json:"visibility" gorm:"size:20;index;default:public;"` // 可见性：public / private
	PublishedAt  time.Time             `json:"published_at" gorm:"index;"`                      // 发布时间
	SeoTitle     string                `json:"seo_title" gorm:"size:255;"`                      // SEO标题
	SeoKeywords  string                `json:"seo_keywords" gorm:"size:512;"`                   // SEO关键词
	SeoDesc      string                `json:"seo_desc" gorm:"type:text;"`                      // SEO描述
	CreatedAt    time.Time             `json:"created_at" gorm:"index;"`                        // 创建时间
	UpdatedAt    time.Time             `json:"updated_at" gorm:"index;"`                        // 更新时间
	Category     *Category             `json:"category,omitempty" gorm:"foreignKey:CategoryID"` // 关联分类
	VisibleUsers []*ArticleVisibleUser `json:"visible_users,omitempty" gorm:"-"`                // 可见用户列表（手动加载）
}

func (a *Article) TableName() string {
	return config.C.FormatTableName("article")
}

// ArticleQueryParam 文章查询参数
type ArticleQueryParam struct {
	util.PaginationParam
	Title          string     `form:”title”`                                                        // 标题模糊搜索
	CategoryID     string     `form:”category_id”`                                                  // 分类ID
	TagName        string     `form:”tag_name”`                                                     // 标签名称精确匹配
	AuthorID       string     `form:”author_id”`                                                    // 作者ID（用于”我的文章”过滤）
	Status         string     `form:”status” binding:”oneof=draft published ''”`                    // 状态
	Visibility     string     `form:"visibility" binding:"oneof=public private partial_visible ''"` // 可见性
	IsTop          *bool      `form:"is_top"`                                                       // 是否置顶
	PublishedAtGte *time.Time `form:"published_at_gte"`                                             // 发布时间 >=
	PublishedAtLte *time.Time `form:"published_at_lte"`                                             // 发布时间 <=
}

// ArticleQueryOptions 查询选项
type ArticleQueryOptions struct {
	util.QueryOptions
	WithCategory     bool   // 是否预加载分类
	WithTags         bool   // 是否预加载标签
	WithCoverImage   bool   // 是否预加载封面图
	WithAuthor       bool   // 是否预加载作者信息
	WithVisibleUsers bool   // 是否预加载可见用户列表
	UserID           string // 当前请求用户ID，用于隐私过滤（空=公开访问）
}

// ArticleQueryResult 查询结果
type ArticleQueryResult struct {
	Data       Articles
	PageResult *util.PaginationResult
}

// Articles 文章切片
type Articles []*Article

// ToIDs 返回文章ID列表
func (a Articles) ToIDs() []string {
	var ids []string
	for _, article := range a {
		ids = append(ids, article.ID)
	}
	return ids
}

// ArticleForm 文章表单（用于创建/更新）
type ArticleForm struct {
	Title          string     `json:"title" binding:"required,max=255"`                                    // 标题
	Slug           string     `json:"slug" binding:"required,max=255"`                                     // URL路径
	Summary        string     `json:"summary" binding:"max=1000"`                                          // 摘要
	Content        string     `json:"content" binding:"required"`                                          // Markdown内容
	CoverImageID   *string    `json:"cover_image_id"`                                                      // 封面图ID
	CoverImage     *Image     `json:"cover_image"`                                                         // 封面图
	CategoryID     string     `json:"category_id"`                                                         // 分类ID/名称（发布时自动创建）
	TagIDs         []string   `json:"tag_ids"`                                                             // 标签ID/名称列表（发布时自动创建）
	IsTop          bool       `json:"is_top"`                                                              // 是否置顶
	Status         string     `json:"status" binding:"required,oneof=draft published"`                     // 状态
	Visibility     string     `json:"visibility" binding:"omitempty,oneof=public private partial_visible"` // 可见性
	VisibleUserIDs []string   `json:"visible_user_ids"`                                                    // 可见用户ID列表（仅 partial_visible 时使用）
	PublishedAt    *time.Time `json:"published_at"`                                                        // 发布时间
	SeoTitle       string     `json:"seo_title" binding:"max=255"`                                         // SEO标题
	SeoKeywords    string     `json:"seo_keywords" binding:"max=512"`                                      // SEO关键词
	SeoDesc        string     `json:"seo_desc" binding:"max=1000"`                                         // SEO描述
}

// Validate 验证表单
func (af *ArticleForm) Validate() error {
	if af.Slug != "" && !util.IsSlug(af.Slug) {
		return errors.BadRequest("", "Slug must contain only letters, numbers, hyphens, or underscores")
	}
	return nil
}

// FillTo 将表单数据填充到 Article 模型。
// 流程：
//  1. 自动存稿：非 published 状态统一按 draft 处理
//  2. MD → HTML：使用 goldmark 渲染并存入 HtmlContent
//  3. 默认可见性：空则为 public
func (af *ArticleForm) FillTo(article *Article) error {
	article.Title = af.Title
	article.Slug = af.Slug
	article.Summary = af.Summary
	article.Content = af.Content

	article.Status = af.Status

	// 2. MD → HTML 预渲染
	var buf bytes.Buffer
	if err := md.Convert([]byte(af.Content), &buf); err != nil {
		return errors.BadRequest("", "Failed to render markdown: %s", err.Error())
	}
	article.HtmlContent = buf.String()

	// 3. 封面图
	article.CoverImageID = af.CoverImageID
	if af.CoverImageID != nil && *af.CoverImageID != "" {
		article.CoverImage = af.CoverImage
	}

	// 4. 可见性：默认 public
	article.Visibility = af.Visibility
	if article.Visibility == "" {
		article.Visibility = ArticleVisibilityPublic
	}

	article.CategoryID = af.CategoryID
	article.IsTop = af.IsTop
	article.SeoTitle = af.SeoTitle
	article.SeoKeywords = af.SeoKeywords
	article.SeoDesc = af.SeoDesc

	// 5. 发布时间
	if af.Status == ArticleStatusPublished {
		if af.PublishedAt != nil {
			article.PublishedAt = *af.PublishedAt
		} else if article.PublishedAt.IsZero() {
			article.PublishedAt = time.Now()
		}
	}

	return nil
}

// ArchiveItem 文章归档项（按年月分组）
type ArchiveItem struct {
	Year  int `json:"year"`  // 年份
	Month int `json:"month"` // 月份
	Count int `json:"count"` // 该月文章数量
}
