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
	"github.com/zhian9/blogo-server/pkg/util"
)

// ProjectLike 用户点赞的项目记录
type ProjectLike struct {
	ProjectID string    `json:"project_id" gorm:"size:20;primaryKey"` // 项目ID
	UserID    string    `json:"user_id" gorm:"size:20;primaryKey"`    // 用户ID
	CreatedAt time.Time `json:"created_at" gorm:"index"`              // 点赞时间
}

func (ProjectLike) TableName() string {
	return config.C.FormatTableName("project_like")
}

// ProjectLikeForm 点赞/取消点赞表单
type ProjectLikeForm struct {
	ProjectID string `json:"project_id" binding:"required"`
}

func (pf *ProjectLikeForm) Validate() error { return nil }

// ProjectLikeQueryParam 查询参数
type ProjectLikeQueryParam struct {
	util.PaginationParam
	UserID    string `form:"user_id"`    // 用户ID
	ProjectID string `form:"project_id"` // 项目ID
}

// ProjectLikeQueryResult 查询结果
type ProjectLikeQueryResult struct {
	Data       ProjectLikes
	PageResult *util.PaginationResult
}

// ProjectLikes 点赞记录切片
type ProjectLikes []*ProjectLike

// ProjectLikeCountResult 点赞计数结果
type ProjectLikeCountResult struct {
	Count int64 `json:"count"`
	Liked bool  `json:"liked"` // 当前用户是否已点赞
}
