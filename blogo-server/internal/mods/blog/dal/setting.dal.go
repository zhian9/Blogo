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

// GetSettingDB 根据上下文返回配置表的 GORM 查询实例
func GetSettingDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Setting))
}

// Setting 配置数据访问对象
type Setting struct {
	DB *gorm.DB
}

// Query 根据查询参数分页查询配置列表
// 支持：按 Key 精确搜索
func (s *Setting) Query(ctx context.Context, params schema.SettingQueryParam, opts ...schema.SettingQueryOptions) (*schema.SettingQueryResult, error) {
	var opt schema.SettingQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetSettingDB(ctx, s.DB)

	// 条件查询
	if v := params.Key; len(v) > 0 {
		db = db.Where("key = ?", v)
	}

	// 执行分页查询
	var list schema.Settings
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schema.SettingQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

// Get 根据 Key 获取单个配置项
func (s *Setting) Get(ctx context.Context, key string, opts ...schema.SettingQueryOptions) (*schema.Setting, error) {
	var opt schema.SettingQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	setting := new(schema.Setting)
	ok, err := util.FindOne(ctx, GetSettingDB(ctx, s.DB).Where("key = ?", key), opt.QueryOptions, setting)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return setting, nil
}

// ExistsKey 检查配置 Key 是否存在
func (s *Setting) ExistsKey(ctx context.Context, key string) (bool, error) {
	ok, err := util.Exists(ctx, GetSettingDB(ctx, s.DB).Where("key = ?", key))
	return ok, errors.WithStack(err)
}

// Create 创建新配置项
func (s *Setting) Create(ctx context.Context, setting *schema.Setting) error {
	result := GetSettingDB(ctx, s.DB).Create(setting)
	return errors.WithStack(result.Error)
}

// Update 更新配置项
// selectFields: 指定更新字段，为空则更新所有字段（除 created_at）
func (s *Setting) Update(ctx context.Context, setting *schema.Setting, selectFields ...string) error {
	db := GetSettingDB(ctx, s.DB).Where("key = ?", setting.Key)

	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("*").Omit("key", "created_at")
	}

	result := db.Updates(setting)
	return errors.WithStack(result.Error)
}

// Delete 根据 Key 删除配置项
func (s *Setting) Delete(ctx context.Context, key string) error {
	result := GetSettingDB(ctx, s.DB).Where("key = ?", key).Delete(new(schema.Setting))
	return errors.WithStack(result.Error)
}

// GetAll 获取所有配置项（不分页，用于初始化或后台管理）
func (s *Setting) GetAll(ctx context.Context) (schema.Settings, error) {
	var list schema.Settings
	err := GetSettingDB(ctx, s.DB).
		Order("created_at ASC").
		Find(&list).Error
	return list, errors.WithStack(err)
}

// GetManyByKeys 根据多个 Key 获取配置项
func (s *Setting) GetManyByKeys(ctx context.Context, keys []string) (schema.Settings, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	var settings schema.Settings
	err := GetSettingDB(ctx, s.DB).Where("key IN ?", keys).Find(&settings).Error
	return settings, errors.WithStack(err)
}
