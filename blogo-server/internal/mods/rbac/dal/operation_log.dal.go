// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// OperationLog 操作日志数据访问对象
type OperationLog struct {
	DB *gorm.DB
}

// Create 创建操作日志记录
func (ol *OperationLog) Create(ctx context.Context, item *schema.OperationLog) error {
	if item.ID == "" {
		item.ID = util.NewXID()
	}
	db := util.GetDB(ctx, ol.DB)
	if err := db.Model(&schema.OperationLog{}).Create(item).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Query 查询操作日志列表（支持多维度筛选）
func (ol *OperationLog) Query(ctx context.Context, params schema.OperationLogQueryParam, opts ...schema.OperationLogQueryOptions) (*schema.OperationLogQueryResult, error) {
	var opt schema.OperationLogQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := util.GetDB(ctx, ol.DB).Model(&schema.OperationLog{})

	// 精确匹配条件
	if v := params.Module; v != "" {
		db = db.Where("module = ?", v)
	}
	if v := params.ActionType; v != "" {
		db = db.Where("action_type = ?", v)
	}
	if v := params.RequestMethod; v != "" {
		db = db.Where("request_method = ?", v)
	}

	// 模糊匹配条件
	if v := params.LikeOperator; v != "" {
		db = db.Where("operator LIKE ?", "%"+v+"%")
	}
	if v := params.LikeDescription; v != "" {
		db = db.Where("description LIKE ?", "%"+v+"%")
	}

	// 状态过滤
	if params.Status != nil {
		db = db.Where("status = ?", *params.Status)
	}

	// 时间范围
	if start := params.StartTime; start != "" {
		db = db.Where("created_at >= ?", start)
	}
	if end := params.EndTime; end != "" {
		db = db.Where("created_at <= ?", end)
	}

	// 执行分页查询
	var list schema.OperationLogs
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &schema.OperationLogQueryResult{
		Data:       list,
		PageResult: pageResult,
	}, nil
}

// GetByID 根据ID获取操作日志
func (ol *OperationLog) GetByID(ctx context.Context, id string) (*schema.OperationLog, error) {
	var item schema.OperationLog
	db := util.GetDB(ctx, ol.DB).Model(&schema.OperationLog{})
	if err := db.Where("id = ?", id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}
	return &item, nil
}

// DeleteByID 删除操作日志
func (ol *OperationLog) DeleteByID(ctx context.Context, id string) error {
	db := util.GetDB(ctx, ol.DB)
	if err := db.Model(&schema.OperationLog{}).Where("id = ?", id).Delete(&schema.OperationLog{}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteOlderThan 删除指定时间之前的日志（用于数据清理）
func (ol *OperationLog) DeleteOlderThan(ctx context.Context, timestamp string) error {
	db := util.GetDB(ctx, ol.DB)
	if err := db.Model(&schema.OperationLog{}).Where("created_at < ?", timestamp).Delete(&schema.OperationLog{}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
