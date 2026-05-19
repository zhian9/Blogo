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

// OperationLog 操作审计日志表（记录所有用户操作）
// @name OperationLog
type OperationLog struct {
	// 主键
	ID string `gorm:"size:20;primaryKey;" json:"id"`

	// 操作人信息
	OperatorID string `gorm:"size:20;index;" json:"operator_id"` // 操作用户ID
	Operator   string `gorm:"size:64;index;" json:"operator"`    // 操作用户名
	OperatorIP string `gorm:"size:45;index;" json:"operator_ip"` // 操作者IP（支持IPv6）

	// 操作内容分类
	Module      string `gorm:"size:64;index;" json:"module"`      // 操作模块（如：用户、文章、评论）
	ActionType  string `gorm:"size:64;index;" json:"action_type"` // 操作类型（如：新增、编辑、删除、审核）
	Description string `gorm:"type:text;" json:"description"`     // 具体行为描述（如：删除了文章「标题」）

	// 资源信息
	ResourceID   string `gorm:"size:256;index;" json:"resource_id"` // 操作的资源ID
	ResourceName string `gorm:"size:256;" json:"resource_name"`     // 操作的资源名称

	// 请求信息
	RequestPath   string `gorm:"size:512;" json:"request_path"`        // 请求路径
	RequestMethod string `gorm:"size:20;index;" json:"request_method"` // HTTP方法
	UserAgent     string `gorm:"size:512;" json:"user_agent"`          // 用户代理

	// 操作结果
	Status     bool   `gorm:"index;" json:"status"`        // 操作是否成功（true/false）
	StatusCode int    `gorm:"index;" json:"status_code"`   // HTTP状态码
	ErrorMsg   string `gorm:"type:text;" json:"error_msg"` // 错误信息（失败时）

	// 时间戳
	CreatedAt time.Time `gorm:"index;" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 返回操作日志表名
func (o *OperationLog) TableName() string {
	return config.C.FormatTableName("operation_log")
}

// OperationLogQueryParam 操作日志查询参数
type OperationLogQueryParam struct {
	util.PaginationParam

	// 精确匹配
	Module        string `form:"module"`        // 操作模块
	ActionType    string `form:"action_type"`   // 操作类型
	RequestMethod string `form:"requestMethod"` // HTTP方法

	// 模糊匹配
	LikeOperator    string `form:"operator"`    // 操作人（模糊）
	LikeDescription string `form:"description"` // 行为描述（模糊）

	// 时间范围
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`

	// 状态过滤
	Status *bool `form:"status"`
}

// OperationLogQueryOptions 查询选项
type OperationLogQueryOptions struct {
	util.QueryOptions
}

// OperationLogQueryResult 查询结果
type OperationLogQueryResult struct {
	Data       OperationLogs
	PageResult *util.PaginationResult
}

// OperationLogs 操作日志指针切片
type OperationLogs []*OperationLog

// OperationLogCreateParam 创建操作日志的参数（供中间件调用）
type OperationLogCreateParam struct {
	OperatorID    string
	Operator      string
	OperatorIP    string
	Module        string
	ActionType    string
	Description   string
	ResourceID    string
	ResourceName  string
	RequestPath   string
	RequestMethod string
	UserAgent     string
	Status        bool
	StatusCode    int
	ErrorMsg      string
}

// ToOperationLog 将参数转换为OperationLog对象
func (p *OperationLogCreateParam) ToOperationLog() *OperationLog {
	return &OperationLog{
		OperatorID:    p.OperatorID,
		Operator:      p.Operator,
		OperatorIP:    p.OperatorIP,
		Module:        p.Module,
		ActionType:    p.ActionType,
		Description:   p.Description,
		ResourceID:    p.ResourceID,
		ResourceName:  p.ResourceName,
		RequestPath:   p.RequestPath,
		RequestMethod: p.RequestMethod,
		UserAgent:     p.UserAgent,
		Status:        p.Status,
		StatusCode:    p.StatusCode,
		ErrorMsg:      p.ErrorMsg,
	}
}
