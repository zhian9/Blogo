// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package biz

import (
	"context"
	"time"

	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	"github.com/zhian9/blogo-server/internal/mods/blog/msg"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Subscriber 订阅者业务逻辑
type Subscriber struct {
	SubscriberDAL *dal.Subscriber
	Trans         *util.Trans
	MailWorker    *msg.MailWorker
}

// Query 分页查询订阅者（后台管理用）
func (s *Subscriber) Query(ctx context.Context, params schema.SubscriberQueryParam) (*schema.SubscriberQueryResult, error) {
	params.Pagination = true
	return s.SubscriberDAL.Query(ctx, params)
}

// Subscribe 用户订阅
func (s *Subscriber) Subscribe(ctx context.Context, form *schema.SubscriberForm) error {
	if err := form.Validate(); err != nil {
		return err
	}

	// 检查是否已订阅
	exists, err := s.SubscriberDAL.ExistsByEmail(ctx, form.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.BadRequest("", "该邮箱已订阅，无需重复订阅")
	}

	sub := &schema.Subscriber{
		ID:        util.NewXID(),
		CreatedAt: time.Now(),
	}
	if err := form.FillTo(sub); err != nil {
		return err
	}

	if err := s.Trans.Exec(ctx, func(ctx context.Context) error {
		return s.SubscriberDAL.Create(ctx, sub)
	}); err != nil {
		return err
	}

	// 异步发送欢迎邮件
	s.MailWorker.Send(msg.MailMessage{
		Type:     msg.MailTypeWelcome,
		SubEmail: form.Email,
		SiteURL:  config.C.General.SiteURL,
	})
	return nil
}

// UnsubscribeByEmail 通过邮箱取消订阅（公开接口）
func (s *Subscriber) UnsubscribeByEmail(ctx context.Context, email string) error {
	if err := s.SubscriberDAL.DeleteByEmail(ctx, email); err != nil {
		return err
	}
	// 异步发送退订确认邮件
	s.MailWorker.Send(msg.MailMessage{
		Type:     msg.MailTypeUnsubscribe,
		SubEmail: email,
		SiteURL:  config.C.General.SiteURL,
	})
	return nil
}

// Unsubscribe 取消订阅
func (s *Subscriber) Unsubscribe(ctx context.Context, id string) error {
	sub, err := s.SubscriberDAL.Get(ctx, id)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.NotFound("", "Subscriber not found")
	}
	return s.SubscriberDAL.Delete(ctx, id)
}
