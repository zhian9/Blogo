// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetPageDB 根据上下文返回页面表的 GORM 查询实例
func GetPageDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Page))
}

// Page 页面数据访问对象
type Page struct {
	DB *gorm.DB
}

// Query 根据查询参数分页查询页面列表
// 支持：按 Slug 精确搜索
func (p *Page) Query(ctx context.Context, params schema.PageQueryParam, opts ...schema.PageQueryOptions) (*schema.PageQueryResult, error) {
	var opt schema.PageQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetPageDB(ctx, p.DB)

	// 条件查询
	if v := params.Slug; len(v) > 0 {
		db = db.Where("slug = ?", v)
	}

	// 执行分页查询
	var list schema.Pages
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schema.PageQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

// Get 根据 ID 获取单个页面
func (p *Page) Get(ctx context.Context, id string, opts ...schema.PageQueryOptions) (*schema.Page, error) {
	var opt schema.PageQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	page := new(schema.Page)
	ok, err := util.FindOne(ctx, GetPageDB(ctx, p.DB).Where("id = ?", id), opt.QueryOptions, page)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return page, nil
}

// GetBySlug 根据 Slug 获取页面（用于前端路由）
func (p *Page) GetBySlug(ctx context.Context, slug string, opts ...schema.PageQueryOptions) (*schema.Page, error) {
	var opt schema.PageQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	page := new(schema.Page)
	ok, err := util.FindOne(ctx, GetPageDB(ctx, p.DB).Where("slug = ?", slug), opt.QueryOptions, page)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return page, nil
}

// ExistsID 检查页面 ID 是否存在
func (p *Page) ExistsID(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetPageDB(ctx, p.DB).Where("id = ?", id))
	return ok, errors.WithStack(err)
}

// ExistsSlug 检查页面 Slug 是否存在
func (p *Page) ExistsSlug(ctx context.Context, slug string) (bool, error) {
	ok, err := util.Exists(ctx, GetPageDB(ctx, p.DB).Where("slug = ?", slug))
	return ok, errors.WithStack(err)
}

// Create 创建新页面
func (p *Page) Create(ctx context.Context, page *schema.Page) error {
	result := GetPageDB(ctx, p.DB).Create(page)
	return errors.WithStack(result.Error)
}

// Update 更新页面信息
// selectFields: 指定更新字段，为空则更新所有字段（除 created_at）
func (p *Page) Update(ctx context.Context, page *schema.Page, selectFields ...string) error {
	db := GetPageDB(ctx, p.DB).Where("id = ?", page.ID)

	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("id", "created_at")
	}

	result := db.Updates(page)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除页面
func (p *Page) Delete(ctx context.Context, id string) error {
	result := GetPageDB(ctx, p.DB).Where("id = ?", id).Delete(new(schema.Page))
	return errors.WithStack(result.Error)
}

// GetAll 获取所有已发布的页面（不分页，用于导航菜单）
func (p *Page) GetAll(ctx context.Context) (schema.Pages, error) {
	var list schema.Pages
	err := GetPageDB(ctx, p.DB).
		Where("is_published = ?", true).
		Order("created_at ASC").
		Find(&list).Error
	return list, errors.WithStack(err)
}
