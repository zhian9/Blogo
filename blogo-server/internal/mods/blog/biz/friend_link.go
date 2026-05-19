// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by a [MIT/Apache/BSD] style license
// that can be found in the LICENSE file.

package biz

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

type FriendLink struct {
	Trans         *util.Trans     // 事务管理器
	FriendLinkDAL *dal.FriendLink // 友链数据访问层
}

// Query 查询友链列表
func (f *FriendLink) Query(ctx context.Context, params schema.FriendLinkQueryParam) (*schema.FriendLinkQueryResult, error) {
	params.Pagination = true
	return f.FriendLinkDAL.Query(ctx, params, schema.FriendLinkQueryOptions{})
}

// Get 获取单个友链
func (f *FriendLink) Get(ctx context.Context, id string) (*schema.FriendLink, error) {
	link, err := f.FriendLinkDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if link == nil {
		return nil, errors.NotFound("", "Friend link not found")
	}
	return link, nil
}

// Create 创建新友链
func (f *FriendLink) Create(ctx context.Context, form *schema.FriendLinkForm) (*schema.FriendLink, error) {
	// 1. 名称唯一性校验
	exists, err := f.FriendLinkDAL.ExistsName(ctx, form.Name)
	if err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Friend link name already exists")
	}

	// 2. 表单验证
	if err := form.Validate(); err != nil {
		return nil, err
	}

	// 3. 事务内创建
	err = f.Trans.Exec(ctx, func(ctx context.Context) error {
		link := &schema.FriendLink{
			ID:        util.NewXID(),
			CreatedAt: time.Now(),
		}
		form.FillTo(link)
		return f.FriendLinkDAL.Create(ctx, link)
	})
	if err != nil {
		return nil, err
	}

	return f.Get(ctx, util.NewXID()) // 同 Tag，建议优化
}

// Update 更新友链
func (f *FriendLink) Update(ctx context.Context, id string, form *schema.FriendLinkForm) error {
	link, err := f.FriendLinkDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if link == nil {
		return errors.NotFound("", "Friend link not found")
	}

	if link.Name != form.Name {
		exists, err := f.FriendLinkDAL.ExistsName(ctx, form.Name)
		if err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Friend link name already exists")
		}
	}

	if err := form.Validate(); err != nil {
		return err
	}

	form.FillTo(link)
	link.UpdatedAt = time.Now()

	return f.Trans.Exec(ctx, func(ctx context.Context) error {
		return f.FriendLinkDAL.Update(ctx, link)
	})
}

// Delete 删除友链
func (f *FriendLink) Delete(ctx context.Context, id string) error {
	exists, err := f.FriendLinkDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Friend link not found")
	}

	return f.Trans.Exec(ctx, func(ctx context.Context) error {
		return f.FriendLinkDAL.Delete(ctx, id)
	})
}

// GetAll 获取所有已启用的友链（用于前端展示）
func (f *FriendLink) GetAll(ctx context.Context) (schema.FriendLinks, error) {
	return f.FriendLinkDAL.GetAll(ctx)
}
