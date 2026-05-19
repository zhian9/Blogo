// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
// Project repository: https://github.com/zhian9/go-admin

package schema

import (
	"time"

	"github.com/zhian9/blogo-server/internal/config"
)

// UserFollow 用户关注关系表（自关联）
type UserFollow struct {
	FollowerID  string    `json:"follower_id" gorm:"size:20;primaryKey"`  // 关注者ID
	FollowingID string    `json:"following_id" gorm:"size:20;primaryKey"` // 被关注者ID
	CreatedAt   time.Time `json:"created_at" gorm:"index"`                // 关注时间
}

func (UserFollow) TableName() string {
	return config.C.FormatTableName("user_follow")
}

// UserFollowForm 关注/取消关注表单
type UserFollowForm struct {
	FollowingID string `json:"following_id" binding:"required"` // 要关注的用户ID
}
