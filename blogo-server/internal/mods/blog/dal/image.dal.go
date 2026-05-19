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

// GetImageDB 根据上下文返回图片表的 GORM 查询实例
func GetImageDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Image))
}

// Image 图片数据访问对象
type Image struct {
	DB *gorm.DB
}

// Query 根据查询参数分页查询图片列表
// 支持：URL 模糊搜索、分类筛选、MIME 类型筛选
func (i *Image) Query(ctx context.Context, params schema.ImageQueryParam, opts ...schema.ImageQueryOptions) (*schema.ImageQueryResult, error) {
	var opt schema.ImageQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetImageDB(ctx, i.DB)

	// 条件查询
	if v := params.URL; len(v) > 0 {
		db = db.Where("url LIKE ?", "%"+v+"%")
	}
	if v := params.Category; len(v) > 0 {
		db = db.Where("category = ?", v)
	}
	if v := params.Type; len(v) > 0 {
		db = db.Where("type = ?", v)
	}

	// 执行分页查询
	var list schema.Images
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schema.ImageQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

// Get 根据 ID 获取单张图片
func (i *Image) Get(ctx context.Context, id string, opts ...schema.ImageQueryOptions) (*schema.Image, error) {
	var opt schema.ImageQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	image := new(schema.Image)
	ok, err := util.FindOne(ctx, GetImageDB(ctx, i.DB).Where("id = ?", id), opt.QueryOptions, image)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return image, nil
}

// ExistsID 检查图片 ID 是否存在
func (i *Image) ExistsID(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetImageDB(ctx, i.DB).Where("id = ?", id))
	return ok, errors.WithStack(err)
}

// Create 创建新图片记录
func (i *Image) Create(ctx context.Context, image *schema.Image) error {
	result := GetImageDB(ctx, i.DB).Create(image)
	return errors.WithStack(result.Error)
}

// Update 更新图片信息
// selectFields: 指定更新字段，为空则更新所有字段（除 created_at）
func (i *Image) Update(ctx context.Context, image *schema.Image, selectFields ...string) error {
	db := GetImageDB(ctx, i.DB).Where("id = ?", image.ID)

	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("id", "created_at")
	}

	result := db.Updates(image)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除图片（硬删除）
// ⚠️ 注意：应在 Service 层检查是否被其他资源引用
func (i *Image) Delete(ctx context.Context, id string) error {
	result := GetImageDB(ctx, i.DB).Where("id = ?", id).Delete(new(schema.Image))
	return errors.WithStack(result.Error)
}

// DeleteByIds 批量删除图片
func (i *Image) DeleteByIds(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	result := GetImageDB(ctx, i.DB).Where("id IN ?", ids).Delete(new(schema.Image))
	return errors.WithStack(result.Error)
}

// GetByCategory 获取某分类下的所有图片（不分页）
func (i *Image) GetByCategory(ctx context.Context, category string) (schema.Images, error) {
	var list schema.Images
	err := GetImageDB(ctx, i.DB).
		Where("category = ?", category).
		Order("created_at DESC").
		Find(&list).Error
	return list, errors.WithStack(err)
}
