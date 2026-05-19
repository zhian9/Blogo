// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	rdal "github.com/zhian9/blogo-server/internal/mods/rbac/dal"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Notification 是通知管理业务的核心对象。
type Notification struct {
	Trans           *util.Trans       // 事务管理器
	NotificationDAL *dal.Notification // 通知数据访问层
	UserDAL         *rdal.User        // 用户数据访问层（可选，用于校验用户存在性）
}

// Query 查询通知列表（必须指定用户 ID）。
func (n *Notification) Query(ctx context.Context, params schema.NotificationQueryParam) (*schema.NotificationQueryResult, error) {
	if params.UserID == "" {
		return nil, errors.BadRequest("", "UserID is required")
	}
	params.Pagination = true

	result, err := n.NotificationDAL.Query(ctx, params, schema.NotificationQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "created_at", Direction: util.DESC},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get 获取单条通知。
func (n *Notification) Get(ctx context.Context, id string) (*schema.Notification, error) {
	notification, err := n.NotificationDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if notification == nil {
		return nil, errors.NotFound("", "Notification not found")
	}
	return notification, nil
}

// Create 创建新通知（通常由系统事件触发，如评论、审核）。
func (n *Notification) Create(ctx context.Context, form *schema.NotificationForm) (*schema.Notification, error) {
	// 1. 校验用户存在性（可选）
	if form.UserID != "" {
		exists, err := n.UserDAL.ExistsID(ctx, form.UserID)
		if err != nil {
			return nil, err
		} else if !exists {
			return nil, errors.BadRequest("", "User not found")
		}
	}

	// 2. 表单验证
	if err := form.Validate(); err != nil {
		return nil, err
	}

	// 3. 初始化实体
	notification := &schema.Notification{
		ID:        util.NewXID(),
		CreatedAt: time.Now(),
	}

	// 4. 填充数据
	form.FillTo(notification)

	// 5. 事务内创建
	err := n.Trans.Exec(ctx, func(ctx context.Context) error {
		return n.NotificationDAL.Create(ctx, notification)
	})
	if err != nil {
		return nil, err
	}

	return n.Get(ctx, notification.ID)
}

// MarkAsRead 将通知标记为已读。
func (n *Notification) MarkAsRead(ctx context.Context, id string) error {
	// 1. 校验存在性
	exists, err := n.NotificationDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Notification not found")
	}

	// 2. 更新状态
	return n.Trans.Exec(ctx, func(ctx context.Context) error {
		return n.NotificationDAL.Update(ctx, &schema.Notification{
			ID:        id,
			IsRead:    true,
			UpdatedAt: time.Now(),
		}, "is_read", "updated_at")
	})
}

// MarkAllAsRead 将某用户的所有通知标记为已读。
func (n *Notification) MarkAllAsRead(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.BadRequest("", "UserID is required")
	}

	// 检查用户是否存在（可选）
	exists, err := n.UserDAL.ExistsID(ctx, userID)
	if err != nil {
		return err
	} else if !exists {
		return errors.BadRequest("", "User not found")
	}

	return n.Trans.Exec(ctx, func(ctx context.Context) error {
		return n.NotificationDAL.MarkAllAsRead(ctx, userID)
	})
}

// Delete 删除通知。
func (n *Notification) Delete(ctx context.Context, id string) error {
	exists, err := n.NotificationDAL.ExistsID(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Notification not found")
	}

	return n.Trans.Exec(ctx, func(ctx context.Context) error {
		return n.NotificationDAL.Delete(ctx, id)
	})
}

// DeleteByUserID 删除某用户的所有通知（谨慎使用）。
func (n *Notification) DeleteByUserID(ctx context.Context, userID string) error {
	if userID == "" {
		return nil
	}
	return n.Trans.Exec(ctx, func(ctx context.Context) error {
		return n.NotificationDAL.DeleteByUserID(ctx, userID)
	})
}

// CountUnread 获取用户未读通知数量。
func (n *Notification) CountUnread(ctx context.Context, userID string) (int64, error) {
	if userID == "" {
		return 0, errors.BadRequest("", "UserID is required")
	}
	return n.NotificationDAL.CountUnread(ctx, userID)
}
