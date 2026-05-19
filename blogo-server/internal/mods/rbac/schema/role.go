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

const (
	RoleStatusEnabled  = "enabled"  // Enabled
	RoleStatusDisabled = "disabled" // Disabled

	RoleResultTypeSelect = "select" // Select
)

// Role 角色表 数据持久化模型
type Role struct {
	ID          string    `json:"id" gorm:"size:20;primarykey;"` // Unique ID
	Code        string    `json:"code" gorm:"size:32;index;"`    // Code of role (unique)
	Name        string    `json:"name" gorm:"size:128;index"`    // Display name of role
	Description string    `json:"description" gorm:"size:1024"`  // Details about role
	Sequence    int       `json:"sequence" gorm:"index"`         // Sequence for sorting
	Status      string    `json:"status" gorm:"size:20;index"`   // Status of role (disabled, enabled)
	CreatedAt   time.Time `json:"created_at" gorm:"index;"`      // Create time
	UpdatedAt   time.Time `json:"updated_at" gorm:"index;"`      // Update time
	Menus       RoleMenus `json:"menus" gorm:"-"`                // Role menu list
}

// TableName 返回 Role 表名
func (r *Role) TableName() string {
	return config.C.FormatTableName("role")
}

// RoleQueryParam 角色查询参数
type RoleQueryParam struct {
	util.PaginationParam
	LikeName    string     `form:"name"`                                       // Display name of role
	Status      string     `form:"status" binding:"oneof=disabled enabled ''"` // Status of role (disabled, enabled)
	ResultType  string     `form:"resultType"`                                 // Result type (options: select)
	InIDs       []string   `form:"-"`                                          // ID list
	GtUpdatedAt *time.Time `form:"-"`                                          // Update time is greater than
}

// RoleQueryOptions 角色查询选项
type RoleQueryOptions struct {
	util.QueryOptions
}

// RoleQueryResult 角色查询结果
type RoleQueryResult struct {
	Data       Roles
	PageResult *util.PaginationResult
}

// Roles Role 指针切片
type Roles []*Role

// RoleForm 角色表 用于数据验证
type RoleForm struct {
	Code        string    `json:"code" binding:"required,max=32"`                   // Code of role (unique)
	Name        string    `json:"name" binding:"required,max=128"`                  // Display name of role
	Description string    `json:"description"`                                      // Details about role
	Sequence    int       `json:"sequence"`                                         // Sequence for sorting
	Status      string    `json:"status" binding:"required,oneof=disabled enabled"` // Status of role (enabled, disabled)
	Menus       RoleMenus `json:"menus"`                                            // Role menu list
}

// Validate 角色验证
func (rf *RoleForm) Validate() error {
	return nil
}

// FillTo 将 RoleForm 表单数据填充到 Role 模型对象中
func (rf *RoleForm) FillTo(role *Role) error {
	role.Code = rf.Code
	role.Name = rf.Name
	role.Description = rf.Description
	role.Sequence = rf.Sequence
	role.Status = rf.Status
	return nil
}
