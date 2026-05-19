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

// ArticleTag 文章-标签关联表（中间表）
type ArticleTag struct {
	ArticleID string    `json:"article_id" gorm:"size:20;primaryKey"` // 文章ID
	TagID     string    `json:"tag_id" gorm:"size:20;primaryKey"`     // 标签ID
	CreatedAt time.Time `json:"created_at" gorm:"index"`              // 创建时间
}

// TableName 返回格式化后的表名
func (at *ArticleTag) TableName() string {
	return config.C.FormatTableName("article_tag")
}
