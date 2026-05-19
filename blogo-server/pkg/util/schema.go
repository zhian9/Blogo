// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package util

import "github.com/zhian9/blogo-server/pkg/errors"

// context constant

const ReqBodyKey = "req-body"
const ResBodyKey = "res-body"
const TreePathDelimiter = "." // 分隔符

// ResponseResult 响应规范结构
type ResponseResult struct {
	Success bool          `json:"success"`
	Data    interface{}   `json:"data,omitempty"`
	Total   int64         `json:"total,omitempty"`
	Error   *errors.Error `json:"error,omitempty"`
}

// PaginationResult 分页结构
type PaginationResult struct {
	Total    int64 `json:"total"`    // 总记录数
	Current  int   `json:"current"`  // 当前页码
	PageSize int   `json:"pageSize"` // 每页大小
}

// PaginationParam 是分页查询的请求参数结构。
// 通过 Gin 的 binding 标签自动从 URL Query 或 Form 中解析。
type PaginationParam struct {
	// Pagination 是否启用分页（true=分页，false=返回全部）
	Pagination bool `form:"-"` // 不从请求中解析，通常由业务逻辑控制

	// OnlyCount 是否仅查询总数（不返回数据列表）
	OnlyCount bool `form:"-"` // 同上

	// Current 当前页码（从 1 开始）
	Current int `form:"current"`

	// PageSize 每页大小（最大 100，防止滥用）
	PageSize int `form:"pageSize" binding:"max=1000"`
}

//4. 查询选项（字段选择、排序等）

// QueryOptions 定义通用查询选项，用于灵活控制查询行为。
type QueryOptions struct {
	// SelectFields 指定要查询的字段（白名单），避免 SELECT *
	SelectFields []string

	// OmitFields 指定要排除的字段（黑名单）
	OmitFields []string

	// OrderFields 排序字段列表
	OrderFields OrderByParams
}

// Direction 定义排序方向。
type Direction string

// 预定义排序方向常量。
const (
	ASC  Direction = "ASC"  // 升序
	DESC Direction = "DESC" // 降序
)

// OrderByParam 表示单个排序字段及其方向。
type OrderByParam struct {
	Field     string    // 数据库字段名（注意：需防 SQL 注入！）
	Direction Direction // 排序方向（ASC/DESC）
}

// OrderByParams 是排序参数的切片类型。
type OrderByParams []OrderByParam

// ToSQL 将排序参数转换为 SQL ORDER BY 子句。
// 避免直接使用用户输入导致 SQL 注入。
func (a OrderByParams) ToSQL() string {
	if len(a) == 0 {
		return ""
	}

	var sql string
	for _, v := range a {
		// 拼接 "field direction,"
		sql += v.Field + " " + string(v.Direction) + ","
	}
	// 移除末尾多余的逗号
	return sql[:len(sql)-1]
}
