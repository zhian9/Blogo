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

// Category 分类表，用于文章归类（如“后端”、“前端”）
type Category struct {
	ID           string    `json:"id" gorm:"size:20;primarykey;"`             // 唯一ID
	Name         string    `json:"name" gorm:"size:100;not null;uniqueIndex"` // 分类名称（唯一）
	Sort         int       `json:"sort" gorm:"default:0;index"`               // 排序值（越小越靠前）
	ArticleCount int64     `json:"article_count" gorm:"-"`                    // 文章数量（非DB字段，查询时填充）
	CreatedAt    time.Time `json:"created_at" gorm:"index;"`                  // 创建时间
	UpdatedAt    time.Time `json:"updated_at" gorm:"index;"`                  // 更新时间
}

func (c *Category) TableName() string {
	return config.C.FormatTableName("category")
}

// CategoryQueryParam 分类查询参数
type CategoryQueryParam struct {
	util.PaginationParam
	Name string `form:"name"` // 分类名称模糊搜索（通常用精确匹配，但保留扩展性）
}

// CategoryQueryOptions 查询选项
type CategoryQueryOptions struct {
	util.QueryOptions
}

// CategoryQueryResult 查询结果
type CategoryQueryResult struct {
	Data       Categories
	PageResult *util.PaginationResult
}

// Categories 分类切片
type Categories []*Category

// ToIDs 返回分类ID列表
func (c Categories) ToIDs() []string {
	var ids []string
	for _, category := range c {
		ids = append(ids, category.ID)
	}
	return ids
}

// CategoryForm 分类表单（用于创建/更新）
type CategoryForm struct {
	Name string `json:"name" binding:"required,max=100"` // 分类名称
	Sort int    `json:"sort" binding:"min=0"`            // 排序值
}

// Validate 验证分类表单
func (cf *CategoryForm) Validate() error {
	// 名称不能包含特殊字符（可选）
	if cf.Name != "" && util.ContainsSpecialChars(cf.Name) {
		return errors.BadRequest("", "Category name cannot contain special characters")
	}
	return nil
}

// FillTo 将表单数据填充到 Category 模型
func (cf *CategoryForm) FillTo(category *Category) error {
	category.Name = cf.Name
	category.Sort = cf.Sort
	return nil
}
