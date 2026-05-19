// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/rbac/dal"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

// OperationLog 操作日志业务逻辑
type OperationLog struct {
	OperationLogDAL *dal.OperationLog
}

// Create 创建操作日志
func (ol *OperationLog) Create(ctx context.Context, item *schema.OperationLog) error {
	return ol.OperationLogDAL.Create(ctx, item)
}

// Query 查询操作日志列表
func (ol *OperationLog) Query(ctx context.Context, params schema.OperationLogQueryParam) (*schema.OperationLogQueryResult, error) {
	params.Pagination = true
	result, err := ol.OperationLogDAL.Query(ctx, params, schema.OperationLogQueryOptions{
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

// GetByID 根据ID获取操作日志
func (ol *OperationLog) GetByID(ctx context.Context, id string) (*schema.OperationLog, error) {
	return ol.OperationLogDAL.GetByID(ctx, id)
}

// DeleteByID 删除操作日志
func (ol *OperationLog) DeleteByID(ctx context.Context, id string) error {
	return ol.OperationLogDAL.DeleteByID(ctx, id)
}

// DeleteOlderThan 删除指定时间之前的日志
func (ol *OperationLog) DeleteOlderThan(ctx context.Context, timestamp string) error {
	return ol.OperationLogDAL.DeleteOlderThan(ctx, timestamp)
}
