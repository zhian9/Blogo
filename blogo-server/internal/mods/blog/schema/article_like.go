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

// ArticleLike 用户点赞的文章记录
type ArticleLike struct {
	ArticleID string    `json:"article_id" gorm:"size:20;primaryKey"` // 文章ID
	UserID    string    `json:"user_id" gorm:"size:20;primaryKey"`    // 用户ID
	CreatedAt time.Time `json:"created_at" gorm:"index"`              // 点赞时间
}

func (ArticleLike) TableName() string {
	return config.C.FormatTableName("article_like")
}

// ArticleLikeForm 点赞/取消点赞表单
type ArticleLikeForm struct {
	ArticleID string `json:"article_id" binding:"required"`
}

func (af *ArticleLikeForm) Validate() error { return nil }

// ArticleLikeQueryParam 查询参数
type ArticleLikeQueryParam struct {
	util.PaginationParam
	UserID    string `form:"user_id"`    // 用户ID
	ArticleID string `form:"article_id"` // 文章ID
}

// ArticleLikeQueryResult 查询结果
type ArticleLikeQueryResult struct {
	Data       ArticleLikes
	PageResult *util.PaginationResult
}

// ArticleLikes 点赞记录切片
type ArticleLikes []*ArticleLike

// ArticleLikeCountResult 点赞计数结果
type ArticleLikeCountResult struct {
	Count int64 `json:"count"`
	Liked bool  `json:"liked"` // 当前用户是否已点赞
}
