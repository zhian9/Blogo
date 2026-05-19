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

// UserContribution 用户每日贡献记录表
type UserContribution struct {
	ID           string    `json:"id" gorm:"size:20;primarykey;"`
	UserID       string    `json:"user_id" gorm:"size:20;uniqueIndex:uk_user_date;index;not null;"`
	Date         string    `json:"date" gorm:"size:10;uniqueIndex:uk_user_date;not null;comment:YYYY-MM-DD"`
	PublishCount int       `json:"publish_count" gorm:"default:0"`
	EditCount    int       `json:"edit_count" gorm:"default:0"`
	LoginCount   int       `json:"login_count" gorm:"default:0"`
	TotalCount   int       `json:"total_count" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *UserContribution) TableName() string {
	return config.C.FormatTableName("user_contribution")
}

// ContributionDay 贡献日（API 返回）
type ContributionDay struct {
	Date         string `json:"date"`
	Count        int    `json:"count"`
	PublishCount int    `json:"publish_count"`
	EditCount    int    `json:"edit_count"`
	LoginCount   int    `json:"login_count"`
}

// ContributionStats 贡献统计（API 返回）
type ContributionStats struct {
	TotalContributions int    `json:"total_contributions"`
	CurrentStreak      int    `json:"current_streak"`
	LongestStreak      int    `json:"longest_streak"`
	ActiveLevel        string `json:"active_level"`
}
