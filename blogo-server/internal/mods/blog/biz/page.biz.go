// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Page 是页面管理业务的核心对象。
type Page struct {
	Trans   *util.Trans // 事务管理器
	PageDAL *dal.Page   // 页面数据访问层
}

// Query 查询页面列表（分页）。
func (p *Page) Query(ctx context.Context, params schema.PageQueryParam) (*schema.PageQueryResult, error) {
	params.Pagination = true

	result, err := p.PageDAL.Query(ctx, params, schema.PageQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "created_at", Direction: util.DESC},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get 获取单个页面。
func (p *Page) Get(ctx context.Context, id string) (*schema.Page, error) {
	page, err := p.PageDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if page == nil {
		return nil, errors.NotFound("", "Page not found")
	}
	return page, nil
}

// GetBySlug 根据 Slug 获取页面（用于前端路由）。
func (p *Page) GetBySlug(ctx context.Context, slug string) (*schema.Page, error) {
	page, err := p.PageDAL.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	} else if page == nil || !page.IsPublished {
		return nil, errors.NotFound("", "Page not found or not published")
	}
	return page, nil
}

// Create 创建新页面。
func (p *Page) Create(ctx context.Context, form *schema.PageForm) (*schema.Page, error) {
	// 1. Slug 唯一性校验
	exists, err := p.PageDAL.ExistsSlug(ctx, form.Slug)
	if err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Slug already exists")
	}

	// 2. 表单验证
	if err := form.Validate(); err != nil {
		return nil, err
	}

	// 3. 初始化实体
	page := &schema.Page{
		ID:        util.NewXID(),
		CreatedAt: time.Now(),
	}

	// 4. 填充数据
	form.FillTo(page)

	// 5. 事务内创建
	err = p.Trans.Exec(ctx, func(ctx context.Context) error {
		return p.PageDAL.Create(ctx, page)
	})
	if err != nil {
		return nil, err
	}

	return p.Get(ctx, page.ID)
}

// Update 更新页面。
func (p *Page) Update(ctx context.Context, id string, form *schema.PageForm) error {
	// 1. 获取原页面
	page, err := p.PageDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if page == nil {
		return errors.NotFound("", "Page not found")
	}

	// 2. Slug 唯一性校验（仅当修改时）
	if page.Slug != form.Slug {
		exists, err := p.PageDAL.ExistsSlug(ctx, form.Slug)
		if err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Slug already exists")
		}
	}

	// 3. 表单验证
	if err := form.Validate(); err != nil {
		return err
	}

	// 4. 填充数据
	form.FillTo(page)
	page.UpdatedAt = time.Now()

	// 5. 事务内更新
	return p.Trans.Exec(ctx, func(ctx context.Context) error {
		return p.PageDAL.Update(ctx, page)
	})
}

// Delete 删除页面。
func (p *Page) Delete(ctx context.Context, id string) error {
	exists, err := p.PageDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Page not found")
	}

	return p.Trans.Exec(ctx, func(ctx context.Context) error {
		return p.PageDAL.Delete(ctx, id)
	})
}

// GetAll 获取所有已发布的页面（用于导航菜单）。
func (p *Page) GetAll(ctx context.Context) (schema.Pages, error) {
	return p.PageDAL.GetAll(ctx)
}
