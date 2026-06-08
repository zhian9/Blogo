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

// ProjectVisibleUser 项目部分人可见——可见用户关联表
type ProjectVisibleUser struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`
	ProjectID string    `json:"project_id" gorm:"size:20;uniqueIndex:uk_project_user;not null;"`
	UserID    string    `json:"user_id" gorm:"size:20;uniqueIndex:uk_project_user;index;not null;"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *ProjectVisibleUser) TableName() string {
	return config.C.FormatTableName("project_visible_user")
}
