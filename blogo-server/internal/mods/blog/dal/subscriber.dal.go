// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package dal

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
	"gorm.io/gorm"
)

// GetSubscriberDB 返回订阅者表的 GORM 查询实例
func GetSubscriberDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Subscriber))
}

// Subscriber 订阅者数据访问对象
type Subscriber struct {
	DB *gorm.DB
}

// Query 分页查询
func (s *Subscriber) Query(ctx context.Context, params schema.SubscriberQueryParam) (*schema.SubscriberQueryResult, error) {
	db := GetSubscriberDB(ctx, s.DB)
	if v := params.Email; len(v) > 0 {
		db = db.Where("email LIKE ?", "%"+v+"%")
	}
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v)
	}
	var list schema.Subscribers
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, util.QueryOptions{
		OrderFields: []util.OrderByParam{{Field: "created_at", Direction: util.DESC}},
	}, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &schema.SubscriberQueryResult{Data: list, PageResult: pageResult}, nil
}

// GetAllActive 获取所有活跃订阅者（用于群发邮件）
func (s *Subscriber) GetAllActive(ctx context.Context) (schema.Subscribers, error) {
	var list schema.Subscribers
	err := GetSubscriberDB(ctx, s.DB).
		Where("status = ?", schema.SubscriberStatusActive).
		Order("created_at ASC").
		Find(&list).Error
	return list, errors.WithStack(err)
}

// GetByEmail 根据邮箱查找
func (s *Subscriber) GetByEmail(ctx context.Context, email string) (*schema.Subscriber, error) {
	sub := new(schema.Subscriber)
	ok, err := util.FindOne(ctx, GetSubscriberDB(ctx, s.DB).Where("email = ?", email), util.QueryOptions{}, sub)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return sub, nil
}

// ExistsByEmail 检查邮箱是否已订阅
func (s *Subscriber) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	ok, err := util.Exists(ctx, GetSubscriberDB(ctx, s.DB).Where("email = ?", email))
	return ok, errors.WithStack(err)
}

// Create 创建订阅者
func (s *Subscriber) Create(ctx context.Context, sub *schema.Subscriber) error {
	result := GetSubscriberDB(ctx, s.DB).Create(sub)
	return errors.WithStack(result.Error)
}

// Get 根据 ID 获取订阅者
func (s *Subscriber) Get(ctx context.Context, id string) (*schema.Subscriber, error) {
	sub := new(schema.Subscriber)
	ok, err := util.FindOne(ctx, GetSubscriberDB(ctx, s.DB).Where("id = ?", id), util.QueryOptions{}, sub)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		return nil, nil
	}
	return sub, nil
}

// DeleteByEmail 根据邮箱删除订阅者（公开退订用）
func (s *Subscriber) DeleteByEmail(ctx context.Context, email string) error {
	result := GetSubscriberDB(ctx, s.DB).Where("email = ?", email).Delete(new(schema.Subscriber))
	return errors.WithStack(result.Error)
}

// Delete 删除订阅者
func (s *Subscriber) Delete(ctx context.Context, id string) error {
	result := GetSubscriberDB(ctx, s.DB).Where("id = ?", id).Delete(new(schema.Subscriber))
	return errors.WithStack(result.Error)
}
