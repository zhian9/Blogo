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
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/zhian9/blogo-server/internal/config"
	rschema "github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

const (
	ProjectStatusDraft     = "draft"     // 草稿
	ProjectStatusPublished = "published" // 已发布

	ProjectVisibilityPublic         = "public"          // 公开可见
	ProjectVisibilityPrivate        = "private"         // 仅作者可见
	ProjectVisibilityPartialVisible = "partial_visible" // 部分人可见

	ProjectStateDeveloping  = "developing"  // 开发中
	ProjectStateCompleted   = "completed"   // 已完成
	ProjectStateMaintaining = "maintaining" // 维护中
	ProjectStatePaused      = "paused"      // 暂停开发
	ProjectStateArchived    = "archived"    // 停止维护
)

// ProjectVisibilityTypes 所有可见性类型（用于 binding oneof 校验）
var ProjectVisibilityTypes = []string{
	ProjectVisibilityPublic,
	ProjectVisibilityPrivate,
	ProjectVisibilityPartialVisible,
}

// ProjectStateTypes 所有项目状态类型
var ProjectStateTypes = []string{
	ProjectStateDeveloping,
	ProjectStateCompleted,
	ProjectStateMaintaining,
	ProjectStatePaused,
	ProjectStateArchived,
}

// projectMd 项目专用 goldmark 实例（线程安全，全局复用）
var projectMd = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.NewTypographer(),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

// Project 项目表
type Project struct {
	ID             string                `json:"id" gorm:"size:20;primarykey;"`                          // 唯一ID
	Title          string                `json:"title" gorm:"size:255;not null;index"`                   // 项目名称
	Slug           string                `json:"slug" gorm:"size:255;uniqueIndex"`                       // SEO友好URL路径
	Summary        string                `json:"summary" gorm:"type:text;"`                              // 项目简介
	Content        string                `json:"content" gorm:"type:longtext;"`                          // Markdown 详细介绍
	HtmlContent    string                `json:"html_content" gorm:"type:longtext;"`                     // 预渲染 HTML
	CoverImageID   *string               `json:"cover_image_id" gorm:"size:20;index"`                    // 封面图ID
	CoverImage     *Image                `json:"cover_image,omitempty" gorm:"-"`                         // 封面图（手动加载）
	Author         *rschema.User         `json:"author,omitempty" gorm:"-"`                              // 作者信息（手动加载）
	CategoryID     string                `json:"category_id" gorm:"size:20;index"`                       // 分类ID
	AuthorID       string                `json:"author_id" gorm:"size:20;index"`                         // 作者用户ID
	Tags           Tags                  `json:"tags" gorm:"many2many:project_tag;"`                     // 技术栈标签（多对多）
	Views          int64                 `json:"views" gorm:"default:0;"`                                // 浏览量
	LikeCount      int64                 `json:"like_count" gorm:"default:0;"`                           // 点赞数（冗余计数）
	FavoriteCount  int64                 `json:"favorite_count" gorm:"default:0;"`                       // 收藏数（冗余计数）
	CommentCount   int64                 `json:"comment_count" gorm:"default:0;"`                        // 评论数（冗余计数）
	IsTop          bool                  `json:"is_top" gorm:"default:false;"`                           // 是否置顶
	IsFeatured     bool                  `json:"is_featured" gorm:"default:false;index"`                 // 是否精选
	FeaturedOrder  int                   `json:"featured_order" gorm:"default:0;index"`                  // 精选排序（越小越前）
	Status         string                `json:"status" gorm:"size:20;index;default:draft;"`             // 发布状态：draft / published
	Visibility     string                `json:"visibility" gorm:"size:20;index;default:public;"`        // 可见性：public / private / partial_visible
	ProjectState   string                `json:"project_state" gorm:"size:20;index;default:developing;"` // 项目状态
	GitHubURL      string                `json:"github_url" gorm:"size:512;"`                            // GitHub仓库地址
	DemoURL        string                `json:"demo_url" gorm:"size:512;"`                              // 在线演示地址
	PublishedAt    time.Time             `json:"published_at" gorm:"index;"`                             // 发布时间
	SeoTitle       string                `json:"seo_title" gorm:"size:255;"`                             // SEO标题
	SeoKeywords    string                `json:"seo_keywords" gorm:"size:512;"`                          // SEO关键词
	SeoDesc        string                `json:"seo_desc" gorm:"type:text;"`                             // SEO描述
	CreatedAt      time.Time             `json:"created_at" gorm:"index;"`                               // 创建时间
	UpdatedAt      time.Time             `json:"updated_at" gorm:"index;"`                               // 更新时间
	Category       *Category             `json:"category,omitempty" gorm:"foreignKey:CategoryID"`        // 关联分类
	VisibleUsers   []*ProjectVisibleUser `json:"visible_users,omitempty" gorm:"-"`                       // 可见用户列表（手动加载）
	Timeline       []*ProjectTimeline    `json:"timeline,omitempty" gorm:"-"`                            // 项目历程（手动加载）
	Resources      []*ProjectResource    `json:"resources,omitempty" gorm:"-"`                           // 相关资源（手动加载）
}

func (p *Project) TableName() string {
	return config.C.FormatTableName("project")
}

// ProjectQueryParam 项目查询参数
type ProjectQueryParam struct {
	util.PaginationParam
	Title        string     `form:"title"`                                                        // 标题模糊搜索
	CategoryID   string     `form:"category_id"`                                                  // 分类ID
	TagName      string     `form:"tag_name"`                                                     // 标签名称精确匹配
	AuthorID     string     `form:"author_id"`                                                    // 作者ID
	Status       string     `form:"status" binding:"oneof=draft published ''"`                    // 发布状态
	Visibility   string     `form:"visibility" binding:"oneof=public private partial_visible ''"` // 可见性
	ProjectState string     `form:"project_state" binding:"omitempty,oneof=developing completed maintaining paused archived"` // 项目状态
	IsTop        *bool      `form:"is_top"`                                                       // 是否置顶
	IsFeatured   *bool      `form:"is_featured"`                                                  // 是否精选
	SortBy       string     `form:"sort_by" binding:"omitempty,oneof=latest hot most_liked"`      // 排序方式
	PublishedAtGte *time.Time `form:"published_at_gte"`                                           // 发布时间 >=
	PublishedAtLte *time.Time `form:"published_at_lte"`                                           // 发布时间 <=
}

// ProjectQueryOptions 查询选项
type ProjectQueryOptions struct {
	util.QueryOptions
	WithCategory     bool   // 是否预加载分类
	WithTags         bool   // 是否预加载标签
	WithCoverImage   bool   // 是否预加载封面图
	WithAuthor       bool   // 是否预加载作者信息
	WithVisibleUsers bool   // 是否预加载可见用户列表
	WithTimeline     bool   // 是否预加载项目历程
	WithResources    bool   // 是否预加载相关资源
	UserID           string // 当前请求用户ID，用于隐私过滤（空=公开访问）
}

// ProjectQueryResult 查询结果
type ProjectQueryResult struct {
	Data       Projects
	PageResult *util.PaginationResult
}

// Projects 项目切片
type Projects []*Project

// ToIDs 返回项目ID列表
func (p Projects) ToIDs() []string {
	var ids []string
	for _, project := range p {
		ids = append(ids, project.ID)
	}
	return ids
}

// ProjectForm 项目表单（用于创建/更新）
type ProjectForm struct {
	Title          string     `json:"title" binding:"required,max=255"`                                    // 项目名称
	Slug           string     `json:"slug" binding:"required,max=255"`                                     // URL路径
	Summary        string     `json:"summary" binding:"max=1000"`                                          // 项目简介
	Content        string     `json:"content" binding:"required"`                                          // Markdown内容
	CoverImageID   *string    `json:"cover_image_id"`                                                      // 封面图ID
	CategoryID     string     `json:"category_id"`                                                         // 分类ID/名称
	TagIDs         []string   `json:"tag_ids"`                                                             // 标签ID/名称列表
	IsTop          bool       `json:"is_top"`                                                              // 是否置顶
	IsFeatured     bool       `json:"is_featured"`                                                         // 是否精选
	FeaturedOrder  int        `json:"featured_order"`                                                      // 精选排序
	Status         string     `json:"status" binding:"required,oneof=draft published"`                     // 发布状态
	Visibility     string     `json:"visibility" binding:"omitempty,oneof=public private partial_visible"` // 可见性
	ProjectState   string     `json:"project_state" binding:"omitempty,oneof=developing completed maintaining paused archived"` // 项目状态
	GitHubURL      string     `json:"github_url" binding:"max=512"`                                        // GitHub地址
	DemoURL        string     `json:"demo_url" binding:"max=512"`                                          // 演示地址
	VisibleUserIDs []string   `json:"visible_user_ids"`                                                    // 可见用户ID列表
	PublishedAt    *time.Time `json:"published_at"`                                                        // 发布时间
	SeoTitle       string     `json:"seo_title" binding:"max=255"`                                         // SEO标题
	SeoKeywords    string     `json:"seo_keywords" binding:"max=512"`                                      // SEO关键词
	SeoDesc        string     `json:"seo_desc" binding:"max=1000"`                                         // SEO描述
}

// Validate 验证表单
func (pf *ProjectForm) Validate() error {
	if pf.Slug != "" && !util.IsSlug(pf.Slug) {
		return errors.BadRequest("", "Slug must contain only letters, numbers, hyphens, or underscores")
	}
	return nil
}

// FillTo 将表单数据填充到 Project 模型
func (pf *ProjectForm) FillTo(project *Project) error {
	project.Title = pf.Title
	project.Slug = pf.Slug
	project.Summary = pf.Summary
	project.Content = pf.Content

	project.Status = pf.Status

	// MD → HTML 预渲染
	var buf bytes.Buffer
	if err := projectMd.Convert([]byte(pf.Content), &buf); err != nil {
		return errors.BadRequest("", "Failed to render markdown: %s", err.Error())
	}
	project.HtmlContent = buf.String()

	// 封面图
	project.CoverImageID = pf.CoverImageID

	// 可见性：默认 public
	project.Visibility = pf.Visibility
	if project.Visibility == "" {
		project.Visibility = ProjectVisibilityPublic
	}

	// 项目状态：默认 developing
	project.ProjectState = pf.ProjectState
	if project.ProjectState == "" {
		project.ProjectState = ProjectStateDeveloping
	}

	project.CategoryID = pf.CategoryID
	project.IsTop = pf.IsTop
	project.IsFeatured = pf.IsFeatured
	project.FeaturedOrder = pf.FeaturedOrder
	project.GitHubURL = pf.GitHubURL
	project.DemoURL = pf.DemoURL
	project.SeoTitle = pf.SeoTitle
	project.SeoKeywords = pf.SeoKeywords
	project.SeoDesc = pf.SeoDesc

	// 发布时间
	if pf.Status == ProjectStatusPublished {
		if pf.PublishedAt != nil {
			project.PublishedAt = *pf.PublishedAt
		} else if project.PublishedAt.IsZero() {
			project.PublishedAt = time.Now()
		}
	}

	return nil
}
