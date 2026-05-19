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

// MenuResource 表示菜单关联的 API 资源。
// 每个资源定义了一个可访问的 API 端点（路径+HTTP 方法）。
// 字段说明：
//   - ID: 唯一标识（20 字符）
//   - MenuID: 关联的菜单 ID（来自 Menu.ID）
//   - Method: HTTP 方法（GET/POST/PUT/DELETE 等）
//   - Path: API 请求路径（支持路径参数，如 /api/v1/users/:id）
//   - CreatedAt/UpdatedAt: 时间戳
type MenuResource struct {
	// 唯一 ID（主键）
	ID string `json:"id" gorm:"size:20;primaryKey"`

	// 关联的菜单 ID（外键，指向 menu.id）
	MenuID string `json:"menu_id" gorm:"size:20;index"`

	// HTTP 方法（如 GET, POST, PUT, DELETE）
	Method string `json:"method" gorm:"size:20;"`

	// API 请求路径（支持路径参数）
	// 示例: /api/v1/users, /api/v1/users/:id
	Path string `json:"path" gorm:"size:255;"`

	// 创建时间
	CreatedAt time.Time `json:"created_at" gorm:"index;"`

	// 更新时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`
}

// TableName 动态返回表名（支持表前缀）。
// 例如：config.C.Storage.DB.TablePrefix = "ga_" → "ga_menu_resource"
func (mr *MenuResource) TableName() string {
	return config.C.FormatTableName("menu_resource")
}

// MenuResourceQueryParam 定义菜单资源查询的参数结构。
type MenuResourceQueryParam struct {
	util.PaginationParam // 分页参数

	// MenuID: 单个菜单 ID（用于查询该菜单的所有资源）
	MenuID string `form:"-"`

	// MenuIDs: 多个菜单 ID 列表（用于批量查询）
	MenuIDs []string `form:"-"`
}

// MenuResourceQueryOptions 定义菜单资源查询的选项（字段选择、排序等）。
type MenuResourceQueryOptions struct {
	util.QueryOptions
}

// MenuResourceQueryResult 定义菜单资源查询的返回结果。
type MenuResourceQueryResult struct {
	// 资源数据列表
	Data MenuResources

	// 分页信息
	PageResult *util.PaginationResult
}

// MenuResources 是 MenuResource 指针的切片类型。
type MenuResources []*MenuResource

// MenuResourceForm 表示菜单资源创建/更新的请求结构。
type MenuResourceForm struct {
	// TODO: 应包含 Method 和 Path 字段
	// Method string `json:"method" binding:"required,oneof=GET POST PUT DELETE"`
	// Path   string `json:"path" binding:"required,max=255"`
}

// Validate 验证 MenuResourceForm 的合法性。
func (a *MenuResourceForm) Validate() error {
	// TODO: 验证 Method 和 Path 的合法性
	return nil
}

// FillTo 将 MenuResourceForm 的数据填充到 MenuResource 实体。
func (a *MenuResourceForm) FillTo(menuResource *MenuResource) error {
	// TODO: 填充 Method 和 Path
	return nil
}
