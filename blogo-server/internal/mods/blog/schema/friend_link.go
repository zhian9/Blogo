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
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

const (
	FriendLinkStatusEnabled  = "enabled"  // 已启用
	FriendLinkStatusDisabled = "disabled" // 已禁用
)

// FriendLink 友情链接表
type FriendLink struct {
	ID          string    `json:"id" gorm:"size:20;primarykey;"`                // 友链唯一ID
	Name        string    `json:"name" gorm:"size:100;not null;index"`          // 站点名称
	URL         string    `json:"url" gorm:"size:512;not null;"`                // 站点链接
	Logo        string    `json:"logo" gorm:"size:512;"`                        // Logo图片URL（可关联 Image 表）
	Description string    `json:"description" gorm:"size:255;"`                 // 站点描述
	Email       string    `json:"email" gorm:"size:128;"`                       // 联系邮箱（用于审核沟通）
	Status      string    `json:"status" gorm:"size:20;index;default:enabled;"` // 状态：enabled / disabled
	Sort        int       `json:"sort" gorm:"default:0;index;"`                 // 排序值（越小越靠前）
	CreatedAt   time.Time `json:"created_at" gorm:"index;"`                     // 创建时间
	UpdatedAt   time.Time `json:"updated_at" gorm:"index;"`                     // 更新时间
}

func (f *FriendLink) TableName() string {
	return config.C.FormatTableName("friend_link")
}

// FriendLinkQueryParam 友链查询参数
type FriendLinkQueryParam struct {
	util.PaginationParam
	Name   string `form:"name"`                                       // 站点名称模糊搜索
	Status string `form:"status" binding:"oneof=enabled disabled ''"` // 状态筛选
}

// FriendLinkQueryOptions 查询选项
type FriendLinkQueryOptions struct {
	util.QueryOptions
}

// FriendLinkQueryResult 查询结果
type FriendLinkQueryResult struct {
	Data       FriendLinks
	PageResult *util.PaginationResult
}

// FriendLinks 友链切片
type FriendLinks []*FriendLink

// ToIDs 返回友链ID列表
func (f FriendLinks) ToIDs() []string {
	var ids []string
	for _, link := range f {
		ids = append(ids, link.ID)
	}
	return ids
}

// FriendLinkForm 友链表单（用于创建/更新）
type FriendLinkForm struct {
	Name        string `json:"name" binding:"required,max=100"`                  // 站点名称
	URL         string `json:"url" binding:"required"`                           // 站点链接
	Logo        string `json:"logo" binding:"max=512"`                           // Logo图片URL
	Description string `json:"description" binding:"max=255"`                    // 描述
	Email       string `json:"email" binding:"max=128"`                          // 联系邮箱
	Status      string `json:"status" binding:"required,oneof=enabled disabled"` // 状态
	Sort        int    `json:"sort" binding:"min=0"`                             // 排序值
}

// Validate 验证友链表单
func (ff *FriendLinkForm) Validate() error {
	// 验证 URL 格式
	if !util.IsURL(ff.URL) {
		return errors.BadRequest("", "Invalid site URL")
	}

	// 验证邮箱格式（如果提供了）
	if ff.Email != "" {
		if err := validator.New().Var(ff.Email, "email"); err != nil {
			return errors.BadRequest("", "Invalid email address")
		}
	}

	// 验证 Logo 是否为合法图片 URL（可选）
	if ff.Logo != "" && !util.IsImageURL(ff.Logo) {
		return errors.BadRequest("", "Logo must be a valid image URL")
	}

	return nil
}

// FillTo 将表单数据填充到 FriendLink 模型
func (ff *FriendLinkForm) FillTo(link *FriendLink) error {
	link.Name = ff.Name
	link.URL = ff.URL
	link.Logo = ff.Logo
	link.Description = ff.Description
	link.Email = ff.Email
	link.Status = ff.Status
	link.Sort = ff.Sort
	return nil
}
