// /*
//  * SPDX-License-Identifier: MIT
//  *
//  * Copyright (c) 2026-present 李星云 (lxy911)
//  *
//  * Project: Blogo
//  * Repository: https://github.com/zhian9/Blogo
//  */

package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/blog/biz"
	"github.com/zhian9/blogo-server/internal/mods/blog/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

type ArticleFavorite struct {
	ArticleFavoriteBIZ *biz.ArticleFavorite
}

func (af *ArticleFavorite) Query(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	if userID == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	var params schema.ArticleFavoriteQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	params.UserID = userID
	result, err := af.ArticleFavoriteBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

func (af *ArticleFavorite) IsFavorite(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	if userID == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	isFav, err := af.ArticleFavoriteBIZ.IsFavorite(ctx, userID, c.Param("article_id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, isFav)
}

func (af *ArticleFavorite) Create(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	if userID == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	form := new(schema.ArticleFavoriteForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := af.ArticleFavoriteBIZ.Create(ctx, userID, form); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (af *ArticleFavorite) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	if userID == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	if err := af.ArticleFavoriteBIZ.Delete(ctx, userID, c.Param("article_id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (af *ArticleFavorite) Count(c *gin.Context) {
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	if userID == "" {
		util.ResError(c, errors.New("unauthorized"))
		return
	}
	count, err := af.ArticleFavoriteBIZ.CountByUserID(ctx, userID)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, count)
}
