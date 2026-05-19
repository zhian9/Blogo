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

// SubscriberStatusActive 订阅状态：活跃
const SubscriberStatusActive = "active"

// Subscriber 邮件订阅者
type Subscriber struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`
	Email     string    `json:"email" gorm:"size:128;uniqueIndex;not null"`
	Status    string    `json:"status" gorm:"size:20;index;default:active"`
	CreatedAt time.Time `json:"created_at" gorm:"index;"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`
}

func (s *Subscriber) TableName() string {
	return config.C.FormatTableName("subscriber")
}

// SubscriberQueryParam 订阅者查询参数
type SubscriberQueryParam struct {
	util.PaginationParam
	Email  string `form:"email"`
	Status string `form:"status"`
}

// SubscriberQueryOptions 查询选项
type SubscriberQueryOptions struct {
	util.QueryOptions
}

// SubscriberQueryResult 查询结果
type SubscriberQueryResult struct {
	Data       Subscribers
	PageResult *util.PaginationResult
}

// Subscribers 订阅者切片
type Subscribers []*Subscriber

// SubscriberForm 订阅表单
type SubscriberForm struct {
	Email string `json:"email" binding:"required,max=128"`
}

// Validate 校验
func (f *SubscriberForm) Validate() error {
	if f.Email == "" {
		return errors.BadRequest("", "Email is required")
	}
	if err := validator.New().Var(f.Email, "email"); err != nil {
		return errors.BadRequest("", "Invalid email address")
	}
	return nil
}

// FillTo 填充到模型
func (f *SubscriberForm) FillTo(s *Subscriber) error {
	s.Email = f.Email
	s.Status = SubscriberStatusActive
	return nil
}
