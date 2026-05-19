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

// Setting 系统配置表（键值对）
type Setting struct {
	Key         string    `json:"key" gorm:"size:100;primarykey;not null;"` // 配置键（如 site_title）
	Value       string    `json:"value" gorm:"type:text;not null;"`         // 配置值
	Description string    `json:"description" gorm:"size:255;"`             // 描述（可选）
	CreatedAt   time.Time `json:"created_at" gorm:"index;"`                 // 创建时间
	UpdatedAt   time.Time `json:"updated_at" gorm:"index;"`                 // 更新时间
}

func (s *Setting) TableName() string {
	return config.C.FormatTableName("setting")
}

// SettingQueryParam 配置查询参数
type SettingQueryParam struct {
	util.PaginationParam
	Key string `form:"key"` // 按 key 精确查询
}

// SettingQueryOptions 查询选项
type SettingQueryOptions struct {
	util.QueryOptions
}

// SettingQueryResult 查询结果
type SettingQueryResult struct {
	Data       Settings
	PageResult *util.PaginationResult
}

// Settings 配置切片
type Settings []*Setting

// ToKeys 返回配置 key 列表
func (s Settings) ToKeys() []string {
	var keys []string
	for _, setting := range s {
		keys = append(keys, setting.Key)
	}
	return keys
}

// SettingForm 配置表单
type SettingForm struct {
	Key         string `json:"key" binding:"required,max=100"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description" binding:"max=255"`
}

// Validate 验证配置表单
func (sf *SettingForm) Validate() error {
	// 可扩展：验证 key 是否合法（如只允许字母、下划线）
	if !util.IsConfigKey(sf.Key) {
		return errors.BadRequest("", "Key must contain only letters, numbers, and underscores")
	}
	return nil
}

// FillTo 将表单数据填充到 Setting 模型
func (sf *SettingForm) FillTo(setting *Setting) error {
	setting.Key = sf.Key
	setting.Value = sf.Value
	setting.Description = sf.Description
	return nil
}
