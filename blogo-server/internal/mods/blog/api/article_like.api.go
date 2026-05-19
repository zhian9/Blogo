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

type ArticleLike struct {
	ArticleLikeBIZ *biz.ArticleLike
}

func (al *ArticleLike) Like(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	if userID == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	form := &schema.ArticleLikeForm{ArticleID: c.Param("id")}
	if err := al.ArticleLikeBIZ.Like(ctx, userID, form); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (al *ArticleLike) UnLike(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	if userID == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	if err := al.ArticleLikeBIZ.UnLike(ctx, userID, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (al *ArticleLike) GetLikeStatus(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := al.ArticleLikeBIZ.GetLikeStatus(ctx, util.FromUserID(ctx), c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}
