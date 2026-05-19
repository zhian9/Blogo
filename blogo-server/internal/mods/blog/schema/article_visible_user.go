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

// ArticleVisibleUser 文章部分人可见——可见用户关联表
type ArticleVisibleUser struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`
	ArticleID string    `json:"article_id" gorm:"size:20;uniqueIndex:uk_article_user;not null;"`
	UserID    string    `json:"user_id" gorm:"size:20;uniqueIndex:uk_article_user;index;not null;"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *ArticleVisibleUser) TableName() string {
	return config.C.FormatTableName("article_visible_user")
}
