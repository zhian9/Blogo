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

// ProjectFavorite 用户收藏的项目记录
type ProjectFavorite struct {
	ProjectID string    `json:"project_id" gorm:"size:20;primaryKey"` // 项目ID
	UserID    string    `json:"user_id" gorm:"size:20;primaryKey"`    // 用户ID
	CreatedAt time.Time `json:"created_at" gorm:"index"`              // 收藏时间
}

// TableName 返回带前缀的表名
func (ProjectFavorite) TableName() string {
	return config.C.FormatTableName("project_favorite")
}

// ProjectFavoriteQueryParam 收藏查询参数
type ProjectFavoriteQueryParam struct {
	util.PaginationParam
	UserID       string     `form:"user_id" binding:"required"` // 用户ID（必填）
	ProjectID    string     `form:"project_id"`                 // 项目ID（可选）
	CreatedAtGte *time.Time `form:"created_at_gte"`             // 收藏时间 >=
	CreatedAtLte *time.Time `form:"created_at_lte"`             // 收藏时间 <=
}

// ProjectFavoriteQueryOptions 查询选项
type ProjectFavoriteQueryOptions struct {
	util.QueryOptions
	WithProject bool
}

// ProjectFavoriteQueryResult 查询结果
type ProjectFavoriteQueryResult struct {
	Data       ProjectFavorites
	PageResult *util.PaginationResult
}

// ProjectFavorites 收藏记录切片
type ProjectFavorites []*ProjectFavorite

// ToProjectIDs 提取项目ID列表
func (f ProjectFavorites) ToProjectIDs() []string {
	ids := make([]string, len(f))
	for i, fav := range f {
		ids[i] = fav.ProjectID
	}
	return ids
}

// ProjectFavoriteForm 收藏/取消收藏表单
type ProjectFavoriteForm struct {
	ProjectID string `json:"project_id" binding:"required"`
}

func (pf *ProjectFavoriteForm) Validate() error { return nil }
