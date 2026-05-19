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

type Logger struct {
	// 唯一 ID（主键，20 字符）
	ID string `gorm:"size:20;primaryKey;" json:"id"`

	// 日志级别（索引，便于查询）
	Level string `gorm:"size:20;index;" json:"level"`

	// 链路追踪 ID（索引，用于排查问题）
	TraceID string `gorm:"size:64;index;" json:"trace_id"`

	// 操作用户 ID（索引，用于审计）
	UserID string `gorm:"size:20;index;" json:"user_id"`

	// 日志分类标签（索引，如 "login", "system"）
	Tag string `gorm:"size:32;index;" json:"tag"`

	// 日志主消息（最多 1024 字符）
	Message string `gorm:"size:1024;" json:"message"`

	// 错误堆栈（TEXT 类型，存储长文本）
	Stack string `gorm:"type:text;" json:"stack"`

	// 其他结构化字段的 JSON 序列化结果（TEXT 类型）
	Data string `gorm:"type:text;" json:"data"`

	// 创建时间（索引，用于时间范围查询）
	CreatedAt time.Time `gorm:"index;" json:"created_at"`

	// ========== 虚拟字段（不存入数据库） ==========

	// LoginName: 关联用户的登录名（来自 User.Username）
	// gorm:"<-:false" 表示禁止写入数据库
	// gorm:"-:migration" 表示迁移时忽略此字段
	LoginName string `json:"login_name" gorm:"<-:false;-:migration;"`

	// UserName: 关联用户的真实姓名（来自 User.Name）
	UserName string `json:"user_name" gorm:"<-:false;-:migration;"`
}

// TableName 动态返回表名（支持表前缀）。
func (a *Logger) TableName() string {
	return config.C.FormatTableName("logger")
}

// LoggerQueryParam 定义日志查询的参数结构。
// 嵌入 util.PaginationParam 支持分页。
type LoggerQueryParam struct {
	util.PaginationParam // 分页参数（Current, PageSize 等）

	// 日志级别（精确匹配）
	Level string `form:"level"`

	// TraceID（精确匹配）
	TraceID string `form:"traceID"`

	// 用户姓名（模糊匹配，对应 UserName 虚拟字段）
	LikeUserName string `form:"userName"`

	// 日志标签（精确匹配）
	Tag string `form:"tag"`

	// 日志消息（模糊匹配）
	LikeMessage string `form:"message"`

	// 时间范围（字符串格式，如 "2025-04-05 14:30:00"）
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
}

// LoggerQueryOptions 定义日志查询的选项（字段选择、排序等）。
// 嵌入 util.QueryOptions 支持 Select/Omit/Order。
type LoggerQueryOptions struct {
	util.QueryOptions
}

// LoggerQueryResult 定义日志查询的返回结果。
type LoggerQueryResult struct {
	// 日志数据列表
	Data Loggers

	// 分页信息（总记录数、当前页等）
	PageResult *util.PaginationResult
}

// Loggers 是 Logger 指针的切片类型。
// 便于实现自定义方法（如批量处理）。
type Loggers []*Logger
