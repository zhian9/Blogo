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

// GetFriendLinkDB 根据上下文返回友链表的 GORM 查询实例
func GetFriendLinkDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.FriendLink))
}

// FriendLink 友链数据访问对象
type FriendLink struct {
	DB *gorm.DB
}

// Query 根据查询参数分页查询友链列表
// 支持：名称模糊搜索、状态筛选
func (f *FriendLink) Query(ctx context.Context, params schema.FriendLinkQueryParam, opts ...schema.FriendLinkQueryOptions) (*schema.FriendLinkQueryResult, error) {
	var opt schema.FriendLinkQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetFriendLinkDB(ctx, f.DB)

	// 条件查询
	if v := params.Name; len(v) > 0 {
		db = db.Where("name LIKE ?", "%"+v+"%")
	}
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v)
	}

	// 按排序值升序（越小越靠前）
	db = db.Order("sort ASC")

	// 执行分页查询
	var list schema.FriendLinks
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schema.FriendLinkQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

// Get 根据 ID 获取单个友链
func (f *FriendLink) Get(ctx context.Context, id string, opts ...schema.FriendLinkQueryOptions) (*schema.FriendLink, error) {
	var opt schema.FriendLinkQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	link := new(schema.FriendLink)
	ok, err := util.FindOne(ctx, GetFriendLinkDB(ctx, f.DB).Where("id = ?", id), opt.QueryOptions, link)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return link, nil
}

// ExistsID 检查友链 ID 是否存在
func (f *FriendLink) ExistsID(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetFriendLinkDB(ctx, f.DB).Where("id = ?", id))
	return ok, errors.WithStack(err)
}

// ExistsName 检查友链名称是否已存在
func (f *FriendLink) ExistsName(ctx context.Context, name string) (bool, error) {
	ok, err := util.Exists(ctx, GetFriendLinkDB(ctx, f.DB).Where("name = ?", name))
	return ok, errors.WithStack(err)
}

// Create 创建新友链
func (f *FriendLink) Create(ctx context.Context, link *schema.FriendLink) error {
	result := GetFriendLinkDB(ctx, f.DB).Create(link)
	return errors.WithStack(result.Error)
}

// Update 更新友链信息
// selectFields: 指定更新字段，为空则更新所有字段（除 created_at）
func (f *FriendLink) Update(ctx context.Context, link *schema.FriendLink, selectFields ...string) error {
	db := GetFriendLinkDB(ctx, f.DB).Where("id = ?", link.ID)

	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("id", "created_at")
	}

	result := db.Updates(link)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除友链
func (f *FriendLink) Delete(ctx context.Context, id string) error {
	result := GetFriendLinkDB(ctx, f.DB).Where("id = ?", id).Delete(new(schema.FriendLink))
	return errors.WithStack(result.Error)
}

// DeleteByIds 批量删除友链
func (f *FriendLink) DeleteByIds(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	result := GetFriendLinkDB(ctx, f.DB).Where("id IN ?", ids).Delete(new(schema.FriendLink))
	return errors.WithStack(result.Error)
}

// GetAll 获取所有已启用的友链（不分页，用于前端展示）
func (f *FriendLink) GetAll(ctx context.Context) (schema.FriendLinks, error) {
	var list schema.FriendLinks
	err := GetFriendLinkDB(ctx, f.DB).
		Where("status = ?", schema.FriendLinkStatusEnabled).
		Order("sort ASC, created_at ASC").
		Find(&list).Error
	return list, errors.WithStack(err)
}
