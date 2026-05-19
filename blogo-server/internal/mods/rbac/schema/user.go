// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package schema

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/pkg/crypto/hash"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

const (
	UserStatusActivated = "activated"
	UserStatusInactive  = "inactive"
	UserStatusFreezed   = "freezed"
)

// User 用户表 数据持久化
// @name User
type User struct {
	ID              string     `json:"id" gorm:"size:20;primarykey;"`      // Unique ID
	Username        string     `json:"username" gorm:"size:64;index"`      // Username for login
	Name            string     `json:"name" gorm:"size:64;index"`          // Name of user
	Password        string     `json:"-" gorm:"size:64;"`                  // Password for login (encrypted)
	Phone           string     `json:"phone" gorm:"size:32;"`              // Phone number of user
	Email           string     `json:"email" gorm:"size:128;uniqueIndex;"` // Email of user (unique)
	Avatar          string     `json:"avatar" gorm:"size:512;"`            // Avatar URL
	Bio             string     `json:"bio" gorm:"size:512;"`               // Bio / personal signature
	Remark          string     `json:"remark" gorm:"size:1024;"`           // Remark of user
	Status          string     `json:"status" gorm:"size:20;index"`        // Status of user (activated, inactive, freezed)
	ActivationToken string     `json:"-" gorm:"size:64;index"`             // Email activation token (one-time use)
	ActivatedAt     *time.Time `json:"activated_at"`                       // Account activation time
	LastLoginAt     *time.Time `json:"last_login_at" gorm:"index;"`        // Last login time
	LastLoginIP     string     `json:"last_login_ip" gorm:"size:45;"`      // Last login IP (supports IPv6)
	FollowerCount   int64      `json:"follower_count" gorm:"default:0"`    // 粉丝数
	FollowingCount  int64      `json:"following_count" gorm:"default:0"`   // 关注数
	CreatedAt       time.Time  `json:"created_at" gorm:"index;"`           // Create time
	UpdatedAt       time.Time  `json:"updated_at" gorm:"index;"`           // Update time
	Roles           UserRoles  `json:"roles" gorm:"-"`                     // Roles of user
}

func (a *User) TableName() string {
	return config.C.FormatTableName("user")
}

// UserQueryParam 用户表查询参数
type UserQueryParam struct {
	util.PaginationParam
	LikeUsername string   `form:"username"`                                             // Username for login
	LikeName     string   `form:"name"`                                                 // Name of user
	Status       string   `form:"status" binding:"oneof=activated inactive freezed ''"` // Status of user
	InIDs        []string `form:"-"`                                                    // ID 列表查询（内部使用）
}

// UserQueryOptions 用户表查询选项
type UserQueryOptions struct {
	util.QueryOptions
}

// UserQueryResult 用户表查询结果
type UserQueryResult struct {
	Data       Users
	PageResult *util.PaginationResult
}

// Users 用户表指针切片
type Users []*User

// ToIDs 返回 Users ID 切片
func (u Users) ToIDs() []string {
	var ids []string
	for _, user := range u {
		ids = append(ids, user.ID)
	}
	return ids
}

// UserForm 用户表单 用于数据输入验证
// @name UserForm
type UserForm struct {
	Username string    `json:"username" binding:"required,max=64"`                         // Username for login
	Name     string    `json:"name" binding:"required,max=64"`                             // Name of user
	Password string    `json:"password" binding:"max=64"`                                  // Password for login (md5 hash)
	Phone    string    `json:"phone" binding:"max=32"`                                     // Phone number of user
	Email    string    `json:"email" binding:"max=128"`                                    // Email of user
	Avatar   string    `json:"avatar" binding:"max=512"`                                   // Avatar URL
	Bio      string    `json:"bio" binding:"max=512"`                                      // Bio / personal signature
	Remark   string    `json:"remark" binding:"max=1024"`                                  // Remark of user
	Status   string    `json:"status" binding:"required,oneof=activated inactive freezed"` // Status of user
	Roles    UserRoles `json:"roles" binding:"required"`                                   // Roles of user
}

// Validate 验证用户表单 email
func (uf *UserForm) Validate() error {
	if uf.Email != "" && validator.New().Var(uf.Email, "email") != nil {
		return errors.BadRequest("", "Invalid email address")
	}
	return nil
}

// FillTo 将 UserForm 表单数据填充到 User 模型对象中
func (uf *UserForm) FillTo(user *User) error {
	user.Username = uf.Username
	user.Name = uf.Name
	user.Phone = uf.Phone
	user.Email = uf.Email
	user.Avatar = uf.Avatar
	user.Bio = uf.Bio
	user.Remark = uf.Remark
	user.Status = uf.Status

	// 对非空密码进行加密处理，并存入 User 中
	if pass := uf.Password; pass != "" {
		hashPass, err := hash.GeneratePassword(pass)
		if err != nil {
			return errors.BadRequest("", "Failed to generate hash password: %s", err.Error())
		}
		user.Password = hashPass
	}
	return nil
}
