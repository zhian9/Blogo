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
	"github.com/zhian9/blogo-server/pkg/util"
)

type FriendLink struct {
	FriendLinkBIZ *biz.FriendLink
}

// @Tags FriendLinkAPI
// @Security ApiKeyAuth
// @Summary Query friend link list
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(10)
// @Param name query string false "Friend link name (fuzzy)"
// @Param status query string false "Status" Enums(enabled,disabled)
// @Success 200 {object} util.ResponseResult{data=[]schema.FriendLink}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/friend-links [get]
func (f *FriendLink) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.FriendLinkQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := f.FriendLinkBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags FriendLinkAPI
// @Security ApiKeyAuth
// @Summary Get friend link by ID
// @Param id path string true "Friend link ID"
// @Success 200 {object} util.ResponseResult{data=schema.FriendLink}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/friend-links/{id} [get]
func (f *FriendLink) Get(c *gin.Context) {
	ctx := c.Request.Context()
	link, err := f.FriendLinkBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, link)
}

// @Tags FriendLinkAPI
// @Summary Get all enabled friend links (public)
// @Success 200 {object} util.ResponseResult{data=[]schema.FriendLink}
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/friend-links/all [get]
func (f *FriendLink) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	links, err := f.FriendLinkBIZ.GetAll(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, links)
}

// @Tags FriendLinkAPI
// @Security ApiKeyAuth
// @Summary Create friend link
// @Param body body schema.FriendLinkForm true "Friend link form"
// @Success 200 {object} util.ResponseResult{data=schema.FriendLink}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/friend-links [post]
func (f *FriendLink) Create(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.FriendLinkForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := f.FriendLinkBIZ.Create(ctx, form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags FriendLinkAPI
// @Security ApiKeyAuth
// @Summary Update friend link
// @Param id path string true "Friend link ID"
// @Param body body schema.FriendLinkForm true "Friend link form"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/friend-links/{id} [put]
func (f *FriendLink) Update(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.FriendLinkForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	err := f.FriendLinkBIZ.Update(ctx, c.Param("id"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags FriendLinkAPI
// @Security ApiKeyAuth
// @Summary Delete friend link
// @Param id path string true "Friend link ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/friend-links/{id} [delete]
func (f *FriendLink) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := f.FriendLinkBIZ.Delete(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
