// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package schema

import (
	"time"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Page 静态页面表（用于“关于我”、“项目介绍”等固定页面）
type Page struct {
	ID          string    `json:"id" gorm:"size:20;primarykey;"`              // 页面唯一ID
	Title       string    `json:"title" gorm:"size:255;not null;"`            // 页面标题
	Slug        string    `json:"slug" gorm:"size:255;uniqueIndex;not null;"` // URL路径（如 /about）
	Content     string    `json:"content" gorm:"type:longtext;not null;"`     // Markdown内容
	IsPublished bool      `json:"is_published" gorm:"default:true;"`          // 是否发布
	CreatedAt   time.Time `json:"created_at" gorm:"index;"`                   // 创建时间
	UpdatedAt   time.Time `json:"updated_at" gorm:"index;"`                   // 更新时间
}

func (p *Page) TableName() string {
	return config.C.FormatTableName("page")
}

// PageQueryParam 页面查询参数
type PageQueryParam struct {
	util.PaginationParam
	Slug string `form:"slug"` // 按 slug 精确查询（用于前端路由）
}

// PageQueryOptions 查询选项
type PageQueryOptions struct {
	util.QueryOptions
}

// PageQueryResult 查询结果
type PageQueryResult struct {
	Data       Pages
	PageResult *util.PaginationResult
}

// Pages 页面切片
type Pages []*Page

// ToIDs 返回页面ID列表
func (p Pages) ToIDs() []string {
	var ids []string
	for _, page := range p {
		ids = append(ids, page.ID)
	}
	return ids
}

// PageForm 页面表单
type PageForm struct {
	Title       string `json:"title" binding:"required,max=255"`
	Slug        string `json:"slug" binding:"required,max=255"`
	Content     string `json:"content" binding:"required"`
	IsPublished bool   `json:"is_published"`
}

// Validate 验证页面表单
func (pf *PageForm) Validate() error {
	if pf.Slug != "" && !util.IsSlug(pf.Slug) {
		return errors.BadRequest("", "Slug must contain only letters, numbers, hyphens, or underscores")
	}
	return nil
}

// FillTo 将表单数据填充到 Page 模型
func (pf *PageForm) FillTo(page *Page) error {
	page.Title = pf.Title
	page.Slug = pf.Slug
	page.Content = pf.Content
	page.IsPublished = pf.IsPublished
	return nil
}
