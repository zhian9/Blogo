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

// RoleMenu 角色菜单关联表结构定义
// Role-Menu association table structure definition
type RoleMenu struct {
	ID        string    `json:"id" gorm:"size:20;primarykey"` // 唯一标识符 / Unique identifier
	RoleID    string    `json:"role_id" gorm:"size:20;index"` // 角色ID，来自Role表的ID / Role ID, from Role.ID
	MenuID    string    `json:"menu_id" gorm:"size:20;index"` // 菜单ID，来自Menu表的ID / Menu ID, from Menu.ID
	CreatedAt time.Time `json:"created_at" gorm:"index;"`     // 创建时间 / Create timestamp
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`     // 更新时间 / Update timestamp
}

// TableName 指定数据库表名
// TableName specifies the database table name
func (a *RoleMenu) TableName() string {
	return config.C.FormatTableName("role_menu")
}

// RoleMenuQueryParam 角色菜单查询参数
// RoleMenu query parameters
type RoleMenuQueryParam struct {
	util.PaginationParam        // 分页参数 / Pagination parameters
	RoleID               string `form:"-"` // 角色ID，用于按角色查询 / Role ID for filtering by role
}

// RoleMenuQueryOptions 角色菜单查询选项
// RoleMenu query options
type RoleMenuQueryOptions struct {
	util.QueryOptions // 基础查询选项 / Basic query options
}

// RoleMenuQueryResult 角色菜单查询结果
// RoleMenu query result
type RoleMenuQueryResult struct {
	Data       RoleMenus              // 角色菜单关联数据列表 / List of role-menu associations
	PageResult *util.PaginationResult // 分页结果信息 / Pagination result information
}

// RoleMenus 角色菜单关联列表类型
// RoleMenus slice type for role-menu associations
type RoleMenus []*RoleMenu

// RoleMenuForm 角色菜单关联表单结构
// RoleMenu form structure for creating/updating associations
type RoleMenuForm struct {
	// 表单字段可根据实际需求添加
	// Form fields can be added based on actual requirements
}

// Validate 验证角色菜单表单数据
// Validate validates the role-menu form data
func (a *RoleMenuForm) Validate() error {
	// 待实现具体的验证逻辑
	// TODO: Implement specific validation logic
	return nil
}

// FillTo 将表单数据填充到角色菜单模型
// FillTo populates form data into the role-menu model
func (a *RoleMenuForm) FillTo(roleMenu *RoleMenu) error {
	// 待实现具体的数据填充逻辑
	// TODO: Implement specific data population logic
	return nil
}
