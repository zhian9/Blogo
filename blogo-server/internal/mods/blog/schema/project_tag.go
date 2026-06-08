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
)

// ProjectTag 项目-标签关联表（中间表）
type ProjectTag struct {
	ProjectID string    `json:"project_id" gorm:"size:20;primaryKey"` // 项目ID
	TagID     string    `json:"tag_id" gorm:"size:20;primaryKey"`     // 标签ID
	CreatedAt time.Time `json:"created_at" gorm:"index"`              // 创建时间
}

// TableName 返回格式化后的表名
func (pt *ProjectTag) TableName() string {
	return config.C.FormatTableName("project_tag")
}
