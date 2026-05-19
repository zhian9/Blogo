// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

type Setting struct {
	Trans      *util.Trans  // 事务管理器
	SettingDAL *dal.Setting // 配置数据访问层
}

// Query 查询配置列表（分页）
func (s *Setting) Query(ctx context.Context, params schema.SettingQueryParam) (*schema.SettingQueryResult, error) {
	params.Pagination = true
	return s.SettingDAL.Query(ctx, params, schema.SettingQueryOptions{})
}

// Get 获取单个配置项
func (s *Setting) Get(ctx context.Context, key string) (*schema.Setting, error) {
	setting, err := s.SettingDAL.Get(ctx, key)
	if err != nil {
		return nil, err
	} else if setting == nil {
		return nil, errors.NotFound("", "Setting not found")
	}
	return setting, nil
}

// Create 创建新配置项
func (s *Setting) Create(ctx context.Context, form *schema.SettingForm) (*schema.Setting, error) {
	// 1. Key 唯一性校验
	exists, err := s.SettingDAL.ExistsKey(ctx, form.Key)
	if err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Setting key already exists")
	}

	// 2. 表单验证
	if err := form.Validate(); err != nil {
		return nil, err
	}

	// 3. 事务内创建
	err = s.Trans.Exec(ctx, func(ctx context.Context) error {
		setting := &schema.Setting{}
		form.FillTo(setting)
		return s.SettingDAL.Create(ctx, setting)
	})
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, form.Key)
}

// Update 更新配置项
func (s *Setting) Update(ctx context.Context, key string, form *schema.SettingForm) error {
	// 1. 校验存在性
	exists, err := s.SettingDAL.ExistsKey(ctx, key)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Setting not found")
	}

	// 2. 表单验证
	if err := form.Validate(); err != nil {
		return err
	}

	// 3. 事务内更新
	return s.Trans.Exec(ctx, func(ctx context.Context) error {
		setting := &schema.Setting{Key: key}
		form.FillTo(setting)
		return s.SettingDAL.Update(ctx, setting)
	})
}

// Delete 删除配置项
func (s *Setting) Delete(ctx context.Context, key string) error {
	exists, err := s.SettingDAL.ExistsKey(ctx, key)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Setting not found")
	}

	return s.Trans.Exec(ctx, func(ctx context.Context) error {
		return s.SettingDAL.Delete(ctx, key)
	})
}

// GetAll 获取所有配置项（用于初始化或后台管理）
func (s *Setting) GetAll(ctx context.Context) (schema.Settings, error) {
	return s.SettingDAL.GetAll(ctx)
}

// GetManyByKeys 根据多个 Key 获取配置项
func (s *Setting) GetManyByKeys(ctx context.Context, keys []string) (schema.Settings, error) {
	return s.SettingDAL.GetManyByKeys(ctx, keys)
}
