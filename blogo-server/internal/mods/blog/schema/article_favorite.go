// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
// Project repository: https://github.com/zhian9/go-admin

package schema

import (
	"time"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/pkg/util"
)

// ArticleFavorite 用户收藏的文章记录
type ArticleFavorite struct {
	ArticleID string    `json:"article_id" gorm:"size:20;primaryKey"` // 文章ID
	UserID    string    `json:"user_id" gorm:"size:20;primaryKey"`    // 用户ID
	CreatedAt time.Time `json:"created_at" gorm:"index"`              // 收藏时间
}

// TableName 返回带前缀的表名
func (ArticleFavorite) TableName() string {
	return config.C.FormatTableName("article_favorite")
}

// ArticleFavoriteQueryParam 收藏查询参数
type ArticleFavoriteQueryParam struct {
	util.PaginationParam
	UserID       string     `form:"user_id" binding:"required"` // 用户ID（必填）
	ArticleID    string     `form:"article_id"`                 // 文章ID（可选）
	CreatedAtGte *time.Time `form:"created_at_gte"`             // 收藏时间 >=
	CreatedAtLte *time.Time `form:"created_at_lte"`             // 收藏时间 <=
}

// ArticleFavoriteQueryOptions 查询选项
type ArticleFavoriteQueryOptions struct {
	util.QueryOptions
	WithArticle bool
}

// ArticleFavoriteQueryResult 查询结果
type ArticleFavoriteQueryResult struct {
	Data       ArticleFavorites
	PageResult *util.PaginationResult
}

// ArticleFavorites 收藏记录切片
type ArticleFavorites []*ArticleFavorite

// ToArticleIDs 提取文章ID列表
func (f ArticleFavorites) ToArticleIDs() []string {
	ids := make([]string, len(f))
	for i, fav := range f {
		ids[i] = fav.ArticleID
	}
	return ids
}

// ArticleFavoriteForm 收藏/取消收藏表单
type ArticleFavoriteForm struct {
	ArticleID string `json:"article_id" binding:"required"`
}

func (af *ArticleFavoriteForm) Validate() error { return nil }
