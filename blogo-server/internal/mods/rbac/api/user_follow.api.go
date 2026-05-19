// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
// Project repository: https://github.com/zhian9/go-admin

package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/rbac/biz"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

type UserFollow struct {
	UserFollowBIZ *biz.UserFollow
}

func (uf *UserFollow) Follow(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	if userID == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	form := &schema.UserFollowForm{FollowingID: c.Param("id")}
	if err := uf.UserFollowBIZ.Follow(ctx, userID, form); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (uf *UserFollow) Unfollow(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	if userID == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	if err := uf.UserFollowBIZ.Unfollow(ctx, userID, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (uf *UserFollow) IsFollowing(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	if userID == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	isFollow, err := uf.UserFollowBIZ.IsFollowing(ctx, userID, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, isFollow)
}

// ListFollowers 获取粉丝列表（返回用户详情，含分页）
func (uf *UserFollow) ListFollowers(c *gin.Context) {
	ctx := c.Request.Context()
	var params struct {
		util.PaginationParam
	}
	util.ParseQuery(c, &params)
	users, page, err := uf.UserFollowBIZ.ListFollowers(ctx, c.Param("id"), params.PaginationParam)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, users, page)
}

// ListFollowing 获取关注列表（返回用户详情，含分页）
func (uf *UserFollow) ListFollowing(c *gin.Context) {
	ctx := c.Request.Context()
	var params struct {
		util.PaginationParam
	}
	util.ParseQuery(c, &params)
	users, page, err := uf.UserFollowBIZ.ListFollowing(ctx, c.Param("id"), params.PaginationParam)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, users, page)
}
