// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetNotificationDB 根据上下文返回通知表的 GORM 查询实例
func GetNotificationDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Notification))
}

// Notification 通知数据访问对象
type Notification struct {
	DB *gorm.DB
}

// Query 根据查询参数分页查询通知列表
// 必须指定 UserID（通知属于特定用户）
func (n *Notification) Query(ctx context.Context, params schema.NotificationQueryParam, opts ...schema.NotificationQueryOptions) (*schema.NotificationQueryResult, error) {
	// UserID 是必填项
	if params.UserID == "" {
		return nil, errors.BadRequest(config.ErrBadRequest, "UserID is required")
	}

	var opt schema.NotificationQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetNotificationDB(ctx, n.DB).Where("user_id = ?", params.UserID)

	// 条件查询
	if v := params.Type; len(v) > 0 {
		db = db.Where("type = ?", v)
	}
	if v := params.IsRead; v != nil {
		db = db.Where("is_read = ?", *v)
	}

	// 按创建时间倒序（最新通知在前）
	db = db.Order("created_at DESC")

	// 执行分页查询
	var list schema.Notifications
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schema.NotificationQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

// Get 根据 ID 获取单条通知
func (n *Notification) Get(ctx context.Context, id string, opts ...schema.NotificationQueryOptions) (*schema.Notification, error) {
	var opt schema.NotificationQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	notification := new(schema.Notification)
	ok, err := util.FindOne(ctx, GetNotificationDB(ctx, n.DB).Where("id = ?", id), opt.QueryOptions, notification)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return notification, nil
}

// Create 创建新通知
func (n *Notification) Create(ctx context.Context, notification *schema.Notification) error {
	result := GetNotificationDB(ctx, n.DB).Create(notification)
	return errors.WithStack(result.Error)
}

// Update 更新通知信息（通常只更新 IsRead）
func (n *Notification) Update(ctx context.Context, notification *schema.Notification, selectFields ...string) error {
	db := GetNotificationDB(ctx, n.DB).Where("id = ?", notification.ID)

	if len(selectFields) > 0 {
		db = db.Select(selectFields)
	} else {
		db = db.Select("is_read", "updated_at") // 默认只允许更新已读状态
	}

	result := db.Updates(notification)
	return errors.WithStack(result.Error)
}

// Delete 根据 ID 删除通知
func (n *Notification) Delete(ctx context.Context, id string) error {
	result := GetNotificationDB(ctx, n.DB).Where("id = ?", id).Delete(new(schema.Notification))
	return errors.WithStack(result.Error)
}

// DeleteByUserID 根据用户 ID 删除所有通知（谨慎使用）
func (n *Notification) DeleteByUserID(ctx context.Context, userID string) error {
	if userID == "" {
		return nil
	}
	result := GetNotificationDB(ctx, n.DB).Where("user_id = ?", userID).Delete(new(schema.Notification))
	return errors.WithStack(result.Error)
}

// ExistsID 检查评论 ID 是否存在
func (n *Notification) ExistsID(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetCommentDB(ctx, n.DB).Where("id = ?", id))
	return ok, errors.WithStack(err)
}

// MarkAllAsRead 将某用户的所有通知标记为已读
func (n *Notification) MarkAllAsRead(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.BadRequest(config.ErrBadRequest, "UserID is required")
	}
	result := GetNotificationDB(ctx, n.DB).
		Where("user_id = ?", userID).
		Where("is_read = ?", false).
		Update("is_read", true)
	return errors.WithStack(result.Error)
}

// CountUnread 统计某用户的未读通知数量
func (n *Notification) CountUnread(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := GetNotificationDB(ctx, n.DB).
		Where("user_id = ?", userID).
		Where("is_read = ?", false).
		Count(&count).Error
	return count, errors.WithStack(err)
}
