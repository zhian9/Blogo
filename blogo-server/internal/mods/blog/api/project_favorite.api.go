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

type ProjectFavorite struct {
	ProjectFavoriteBIZ *biz.ProjectFavorite
}

func (pf *ProjectFavorite) Create(c *gin.Context) {
	ctx := c.Request.Context()
	uid := util.FromUserID(ctx)
	if uid == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	form := &schema.ProjectFavoriteForm{ProjectID: c.Param("id")}
	if err := pf.ProjectFavoriteBIZ.Create(ctx, uid, form); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (pf *ProjectFavorite) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uid := util.FromUserID(ctx)
	if uid == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	if err := pf.ProjectFavoriteBIZ.Delete(ctx, uid, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (pf *ProjectFavorite) IsFavorite(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := pf.ProjectFavoriteBIZ.IsFavorite(ctx, util.FromUserID(ctx), c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

func (pf *ProjectFavorite) Count(c *gin.Context) {
	ctx := c.Request.Context()
	count, err := pf.ProjectFavoriteBIZ.Count(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, count)
}
