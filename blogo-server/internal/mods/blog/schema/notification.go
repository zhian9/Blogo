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

const (
	NotificationTypeComment = "comment" // 评论通知
	NotificationTypeSystem  = "system"  // 系统通知
)

// Notification 消息通知表
type Notification struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`          // 通知ID
	UserID    string    `json:"user_id" gorm:"size:20;index;not null;"` // 接收者ID
	Type      string    `json:"type" gorm:"size:20;index;not null;"`    // 通知类型
	Title     string    `json:"title" gorm:"size:255;not null;"`        // 标题
	Content   string    `json:"content" gorm:"type:text;not null;"`     // 内容
	IsRead    bool      `json:"is_read" gorm:"default:false;index;"`    // 是否已读
	RelatedID string    `json:"related_id" gorm:"size:20;"`             // 关联ID（如评论ID）
	CreatedAt time.Time `json:"created_at" gorm:"index;"`               // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`               // 更新时间
}

func (n *Notification) TableName() string {
	return config.C.FormatTableName("notification")
}

// NotificationQueryParam 通知查询参数
type NotificationQueryParam struct {
	util.PaginationParam
	UserID string `form:"user_id" binding:"required"` // 必须指定用户
	Type   string `form:"type" binding:"oneof=comment system ''"`
	IsRead *bool  `form:"is_read"` // 是否已读
}

// NotificationQueryOptions 查询选项
type NotificationQueryOptions struct {
	util.QueryOptions
}

// NotificationQueryResult 查询结果
type NotificationQueryResult struct {
	Data       Notifications
	PageResult *util.PaginationResult
}

// Notifications 通知切片
type Notifications []*Notification

// ToIDs 返回通知ID列表
func (n Notifications) ToIDs() []string {
	var ids []string
	for _, notif := range n {
		ids = append(ids, notif.ID)
	}
	return ids
}

// NotificationForm 通知表单（通常由系统触发，但保留结构）
type NotificationForm struct {
	UserID    string `json:"user_id" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=comment system"`
	Title     string `json:"title" binding:"required,max=255"`
	Content   string `json:"content" binding:"required"`
	RelatedID string `json:"related_id" binding:"omitempty"`
}

// Validate 验证通知表单
func (nf *NotificationForm) Validate() error {
	return nil // 通常由系统生成，无需复杂验证
}

// FillTo 将表单数据填充到 Notification 模型
func (nf *NotificationForm) FillTo(notif *Notification) error {
	notif.UserID = nf.UserID
	notif.Type = nf.Type
	notif.Title = nf.Title
	notif.Content = nf.Content
	notif.RelatedID = nf.RelatedID
	notif.IsRead = false // 新通知默认未读
	return nil
}
