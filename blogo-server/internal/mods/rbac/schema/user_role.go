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

// UserRole 用户角色关联模型
type UserRole struct {
	ID        string    `json:"id" gorm:"size:20;primarykey"`           // Unique ID
	UserID    string    `json:"user_id" gorm:"size:20;index"`           // From User.ID
	RoleID    string    `json:"role_id" gorm:"size:20;index"`           // From Role.ID
	CreatedAt time.Time `json:"created_at" gorm:"index;"`               // Create time
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`               // Update time
	RoleName  string    `json:"role_name" gorm:"<-:false;-:migration;"` // From Role.Name
	RoleCode  string    `json:"role_code" gorm:"<-:false;-:migration;"` // From Role.Code
}

func (ur *UserRole) TableName() string {
	return config.C.FormatTableName("user_role")
}

// UserRoleQueryParam UserRole 查询参数模型
type UserRoleQueryParam struct {
	util.PaginationParam
	InUserIDs []string `form:"-"` // From User.ID
	UserID    string   `form:"-"` // From User.ID
	RoleID    string   `form:"-"` // From Role.ID
}

// UserRoleQueryOptions 用户角色查询选项
type UserRoleQueryOptions struct {
	util.QueryOptions
	JoinRole bool // Join role table
}

// UserRoleQueryResult 用户角色关联表查询结果
type UserRoleQueryResult struct {
	Data       UserRoles
	PageResult *util.PaginationResult
}

// UserRoles UserRole 指针切片
type UserRoles []*UserRole

// ToUserIDMap 从 UserRoles 获取 UserID
func (ur UserRoles) ToUserIDMap() map[string]UserRoles {
	m := make(map[string]UserRoles)
	for _, userRole := range ur {
		m[userRole.UserID] = append(m[userRole.UserID], userRole)
	}
	return m
}

// ToRoleIDs 从 UserRoles 获取 RoleID
func (ur UserRoles) ToRoleIDs() []string {
	var ids []string
	for _, item := range ur {
		ids = append(ids, item.RoleID)
	}
	return ids
}

type UserRoleForm struct {
}

// A validation function for the `UserRoleForm` struct.
func (a *UserRoleForm) Validate() error {
	return nil
}

func (a *UserRoleForm) FillTo(userRole *UserRole) error {
	return nil
}
