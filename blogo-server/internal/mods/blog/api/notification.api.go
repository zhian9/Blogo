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

type Notification struct {
	NotificationBIZ *biz.Notification
}

// @Tags NotificationAPI
// @Security BearerAuth
// @Summary Query notification list
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(10)
// @Param user_id query string true "User ID"
// @Param type query string false "Notification type" Enums(comment,system)
// @Param is_read query boolean false "Is read"
// @Success 200 {object} util.ResponseResult{data=[]schema.Notification}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/notifications [get]
func (n *Notification) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.NotificationQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	if params.UserID == "" {
		util.ResError(c, errors.BadRequest("", "user_id is required"))
		return
	}

	result, err := n.NotificationBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags NotificationAPI
// @Security BearerAuth
// @Summary Get notification by ID
// @Param id path string true "Notification ID"
// @Success 200 {object} util.ResponseResult{data=schema.Notification}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/notifications/{id} [get]
func (n *Notification) Get(c *gin.Context) {
	ctx := c.Request.Context()
	notification, err := n.NotificationBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, notification)
}

// @Tags NotificationAPI
// @Security BearerAuth
// @Summary Mark notification as read
// @Param id path string true "Notification ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/notifications/{id}/read [patch]
func (n *Notification) MarkAsRead(c *gin.Context) {
	ctx := c.Request.Context()
	err := n.NotificationBIZ.MarkAsRead(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags NotificationAPI
// @Security BearerAuth
// @Summary Mark all notifications as read
// @Param user_id query string true "User ID"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/notifications/read-all [patch]
func (n *Notification) MarkAllAsRead(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Query("user_id")
	if userID == "" {
		util.ResError(c, errors.BadRequest("", "user_id is required"))
		return
	}
	err := n.NotificationBIZ.MarkAllAsRead(ctx, userID)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags NotificationAPI
// @Security BearerAuth
// @Summary Delete notification
// @Param id path string true "Notification ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/notifications/{id} [delete]
func (n *Notification) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := n.NotificationBIZ.Delete(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
