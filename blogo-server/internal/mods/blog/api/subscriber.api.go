// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/blog/biz"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Subscriber 订阅者 API
type Subscriber struct {
	SubscriberBIZ *biz.Subscriber
}

// Subscribe 公开订阅接口 POST /api/v1/public/subscribe
func (s *Subscriber) Subscribe(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.SubscriberForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := s.SubscriberBIZ.Subscribe(ctx, form); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// UnsubscribeByEmail 公开退订 GET /api/v1/subscribe/unsubscribe?email=xxx
func (s *Subscriber) UnsubscribeByEmail(c *gin.Context) {
	ctx := c.Request.Context()
	email := c.Query("email")
	if email == "" {
		util.ResError(c, errors.BadRequest("", "Email is required"))
		return
	}
	if err := s.SubscriberBIZ.UnsubscribeByEmail(ctx, email); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// Unsubscribe 取消订阅 POST /api/v1/subscribers/:id/unsubscribe
func (s *Subscriber) Unsubscribe(c *gin.Context) {
	ctx := c.Request.Context()
	if err := s.SubscriberBIZ.Unsubscribe(ctx, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// Query 查询订阅者列表（管理后台）
func (s *Subscriber) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.SubscriberQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := s.SubscriberBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}
