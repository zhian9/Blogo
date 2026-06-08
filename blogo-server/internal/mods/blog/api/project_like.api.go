// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/blog/biz"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

type ProjectLike struct {
	ProjectLikeBIZ *biz.ProjectLike
}

func (pl *ProjectLike) Like(c *gin.Context) {
	ctx := c.Request.Context()
	uid := util.FromUserID(ctx)
	if uid == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	form := &schema.ProjectLikeForm{ProjectID: c.Param("id")}
	if err := pl.ProjectLikeBIZ.Like(ctx, uid, form); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (pl *ProjectLike) UnLike(c *gin.Context) {
	ctx := c.Request.Context()
	uid := util.FromUserID(ctx)
	if uid == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	if err := pl.ProjectLikeBIZ.UnLike(ctx, uid, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (pl *ProjectLike) GetLikeStatus(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := pl.ProjectLikeBIZ.GetLikeStatus(ctx, util.FromUserID(ctx), c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}
