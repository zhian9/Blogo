// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"
	"fmt"
	"strings"

	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetLoggerDB 根据上下文返回日志表的 GORM DB 实例
func GetLoggerDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Logger))
}

// Logger 操作日志实体的数据库访问对象 DAO
type Logger struct {
	DB *gorm.DB
}

// Query 根据参数和选项操作日志列表
func (l *Logger) Query(ctx context.Context, params schema.LoggerQueryParam, opts ...schema.LoggerQueryOptions) (*schema.LoggerQueryResult, error) {
	var opt schema.LoggerQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 构建基础查询
	db := l.DB.Table(fmt.Sprintf("%s AS l", new(schema.Logger).TableName()))
	db = db.Joins(fmt.Sprintf("LEFT JOIN %s u ON l.user_id = u.id", new(schema.User).TableName()))
	// 选择字段
	db = db.Select("l.*,u.name AS user_name, u.username AS login_name")

	// 条件查询
	if v := params.Level; v != "" {
		// 数据库存储的 level 为小写（info/warn/error），统一转小写匹配
		db = db.Where("l.level = ?", strings.ToLower(v))
	}
	if v := params.LikeMessage; len(v) > 0 {
		db = db.Where("l.message LIKE ?", "%"+v+"%") // 日志消息模糊查询
	}
	if v := params.TraceID; v != "" {
		db = db.Where("l.trace_id = ?", v) // TraceID 精确匹配
	}
	if v := params.LikeUserName; v != "" {
		// 注意：这里查询的是 user.username（登录名），不是 name（真实姓名）
		db = db.Where("u.username LIKE ?", "%"+v+"%")
	}
	if v := params.Tag; v != "" {
		db = db.Where("l.tag = ?", v) // 日志标签精确匹配
	}

	// 6. 时间范围过滤
	if start, end := params.StartTime, params.EndTime; start != "" && end != "" {
		// 假设 StartTime/EndTime 为字符串格式（如 "2025-04-05 14:30:00"）
		db = db.Where("l.created_at BETWEEN ? AND ?", start, end)
	}

	// 7. 执行分页查询
	var list schema.Loggers
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err) // 保留错误堆栈
	}

	// 8. 返回查询结果
	return &schema.LoggerQueryResult{
		PageResult: pageResult,
		Data:       list,
	}, nil
}
