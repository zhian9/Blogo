// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetCategoryDB 根据上下文返回分类表的 GORM 查询实例
func GetCategoryDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Category))
}

// Category 分类数据访问对象
type Category struct {
	DB *gorm.DB
}

// Query 根据查询参数分页查询分类列表
// 支持：名称模糊搜索
func (c *Category) Query(ctx context.Context, params schema.CategoryQueryParam, opts ...schema.CategoryQueryOptions) (*schema.CategoryQueryResult, error) {
	var opt schema.CategoryQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetCategoryDB(ctx, c.DB)

	// 条件查询
	if v := params.Name; len(v) > 0 {
		db = db.Where("name LIKE ?", "%"+v+"%")
	}

	// 执行分页查询
	var list schema.Categories
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schema.CategoryQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

// Get 根据 ID 获取单个分类
func (c *Category) Get(ctx context.Context, id string, opts ...schema.CategoryQueryOptions) (*schema.Category, error) {
	var opt schema.CategoryQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	category := new(schema.Category)
	ok, err := util.FindOne(ctx, GetCategoryDB(ctx, c.DB).Where("id = ?", id), opt.QueryOptions, category)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return category, nil
}

// GetByName 根据名称获取分类（用于校验唯一性）
func (c *Category) GetByName(ctx context.Context, name string, opts ...schema.CategoryQueryOptions) (*schema.Category, error) {
	var opt schema.CategoryQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	category := new(schema.Category)
	ok, err := util.FindOne(ctx, GetCategoryDB(ctx, c.DB).Where("name = ?", name), opt.QueryOptions, category)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return category, nil
}

// ExistsID 检查分类 ID 是否存在
func (c *Category) ExistsID(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetCategoryDB(ctx, c.DB).Where("id = ?", id))
	return ok, errors.WithStack(err)
}

// ExistsName 检查分类名称是否已存在（创建/更新时校验）
func (c *Category) ExistsName(ctx context.Context, name string) (bool, error) {
	ok, err := util.Exists(ctx, GetCategoryDB(ctx, c.DB).Where("name = ?", name))
	return ok, errors.WithStack(err)
}

// Create 创建新分类
func (c *Category) Create(ctx context.Context, category *schema.Category) error {
	result := GetCategoryDB(ctx, c.DB).Create(category)
	return errors.WithStack(result.Error)
}

// Update 更新分类信息
// selectFields: 指定更新字段，为空则更新所有字段（除 created_at）
func (c *Category) Update(ctx context.Context, category *schema.Category, selectFields ...string) error {
	db := GetCategoryDB(ctx, c.DB).Where("id = ?", category.ID)

	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("id", "created_at")
	}

	result := db.Updates(category)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除分类
// 注意：删除前应在 Service 层检查是否有关联文章
func (c *Category) Delete(ctx context.Context, id string) error {
	result := GetCategoryDB(ctx, c.DB).Where("id = ?", id).Delete(new(schema.Category))
	return errors.WithStack(result.Error)
}

// DeleteByIds 批量删除分类
func (c *Category) DeleteByIds(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	result := GetCategoryDB(ctx, c.DB).Where("id IN ?", ids).Delete(new(schema.Category))
	return errors.WithStack(result.Error)
}

// FindOrCreateByName 根据名称查找分类，若不存在则自动创建。
// 返回分类ID，用于文章创建时的分类关联。
func (c *Category) FindOrCreateByName(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", nil
	}

	// 1. 尝试按名称查找
	existing, err := c.GetByName(ctx, name)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return existing.ID, nil
	}

	// 2. 不存在则创建
	cat := &schema.Category{
		ID:        util.NewXID(),
		Name:      name,
		Sort:      0,
		CreatedAt: time.Now(),
	}
	if err := c.Create(ctx, cat); err != nil {
		return "", err
	}
	return cat.ID, nil
}

// GetAll 获取所有分类（不分页，用于下拉选择、首页展示），含文章数量
func (c *Category) GetAll(ctx context.Context) (schema.Categories, error) {
	var list schema.Categories
	err := GetCategoryDB(ctx, c.DB).
		Order("sort ASC, created_at ASC").
		Find(&list).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// 统计每个分类下已发布文章的数量
	type countRow struct {
		CategoryID string
		Count      int64
	}
	var counts []countRow
	GetArticleDB(ctx, c.DB).
		Select("category_id, count(*) as count").
		Where("status = ? AND visibility = ?",
			schema.ArticleStatusPublished, schema.ArticleVisibilityPublic).
		Group("category_id").
		Find(&counts)

	m := make(map[string]int64, len(counts))
	for _, row := range counts {
		m[row.CategoryID] = row.Count
	}
	for _, cat := range list {
		cat.ArticleCount = m[cat.ID]
	}

	return list, nil
}
