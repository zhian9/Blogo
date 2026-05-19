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

// GetTagDB 根据上下文返回标签表的 GORM 查询实例
func GetTagDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Tag))
}

// Tag 标签数据访问对象
type Tag struct {
	DB *gorm.DB
}

// Query 根据查询参数分页查询标签列表
// 支持：名称模糊搜索
func (t *Tag) Query(ctx context.Context, params schema.TagQueryParam, opts ...schema.TagQueryOptions) (*schema.TagQueryResult, error) {
	var opt schema.TagQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetTagDB(ctx, t.DB)

	// 条件查询
	if v := params.Name; len(v) > 0 {
		db = db.Where("name LIKE ?", "%"+v+"%")
	}

	// 执行分页查询
	var list schema.Tags
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schema.TagQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

// Get 根据 ID 获取单个标签
func (t *Tag) Get(ctx context.Context, id string, opts ...schema.TagQueryOptions) (*schema.Tag, error) {
	var opt schema.TagQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	tag := new(schema.Tag)
	ok, err := util.FindOne(ctx, GetTagDB(ctx, t.DB).Where("id = ?", id), opt.QueryOptions, tag)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return tag, nil
}

// GetByName 根据名称获取标签（用于校验唯一性）
func (t *Tag) GetByName(ctx context.Context, name string, opts ...schema.TagQueryOptions) (*schema.Tag, error) {
	var opt schema.TagQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	tag := new(schema.Tag)
	ok, err := util.FindOne(ctx, GetTagDB(ctx, t.DB).Where("name = ?", name), opt.QueryOptions, tag)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return tag, nil
}

// ExistsID 检查标签 ID 是否存在
func (t *Tag) ExistsID(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetTagDB(ctx, t.DB).Where("id = ?", id))
	return ok, errors.WithStack(err)
}

// ExistsName 检查标签名称是否已存在（创建/更新时校验）
func (t *Tag) ExistsName(ctx context.Context, name string) (bool, error) {
	ok, err := util.Exists(ctx, GetTagDB(ctx, t.DB).Where("name = ?", name))
	return ok, errors.WithStack(err)
}

// Create 创建新标签
func (t *Tag) Create(ctx context.Context, tag *schema.Tag) error {
	result := GetTagDB(ctx, t.DB).Create(tag)
	return errors.WithStack(result.Error)
}

// Update 更新标签信息
// selectFields: 指定更新字段，为空则更新所有字段（除 created_at）
func (t *Tag) Update(ctx context.Context, tag *schema.Tag, selectFields ...string) error {
	db := GetTagDB(ctx, t.DB).Where("id = ?", tag.ID)

	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("id", "created_at")
	}

	result := db.Updates(tag)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除标签
// ⚠️ 注意：删除前应在 Service 层检查是否有关联文章
func (t *Tag) Delete(ctx context.Context, id string) error {
	result := GetTagDB(ctx, t.DB).Where("id = ?", id).Delete(new(schema.Tag))
	return errors.WithStack(result.Error)
}

// DeleteByIds 批量删除标签
func (t *Tag) DeleteByIds(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	result := GetTagDB(ctx, t.DB).Where("id IN ?", ids).Delete(new(schema.Tag))
	return errors.WithStack(result.Error)
}

// GetAll 获取所有标签（不分页，用于下拉选择或标签云）
func (t *Tag) GetAll(ctx context.Context) (schema.Tags, error) {
	var list schema.Tags
	err := GetTagDB(ctx, t.DB).
		Order("created_at DESC"). // 按创建时间倒序（最新标签在前）
		Find(&list).Error
	return list, errors.WithStack(err)
}

// GetByNames 根据名称列表获取标签（用于文章创建时批量获取或创建标签）
// 返回：存在的标签列表 + 不存在的名称列表
func (t *Tag) GetByNames(ctx context.Context, names []string) (exists schema.Tags, notExists []string, err error) {
	if len(names) == 0 {
		return nil, nil, nil
	}

	var found schema.Tags
	err = GetTagDB(ctx, t.DB).Where("name IN ?", names).Find(&found).Error
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// 构建已存在名称集合
	existsMap := make(map[string]bool)
	for _, tag := range found {
		existsMap[tag.Name] = true
	}

	// 找出不存在的名称
	for _, name := range names {
		if !existsMap[name] {
			notExists = append(notExists, name)
		}
	}

	return found, notExists, nil
}
