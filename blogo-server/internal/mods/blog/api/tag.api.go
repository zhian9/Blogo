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

type Tag struct {
	TagBIZ *biz.Tag
}

// @Tags TagAPI
// @Security BearerAuth
// @Summary Query tag list
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(10)
// @Param name query string false "Tag name (fuzzy)"
// @Success 200 {object} util.ResponseResult{data=[]schema.Tag}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/tags [get]
func (t *Tag) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.TagQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := t.TagBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags TagAPI
// @Security BearerAuth
// @Summary Get tag by ID
// @Param id path string true "Tag ID"
// @Success 200 {object} util.ResponseResult{data=schema.Tag}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/tags/{id} [get]
func (t *Tag) Get(c *gin.Context) {
	ctx := c.Request.Context()
	tag, err := t.TagBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, tag)
}

// @Tags TagAPI
// @Security BearerAuth
// @Summary Create tag
// @Param body body schema.TagForm true "Tag form"
// @Success 200 {object} util.ResponseResult{data=schema.Tag}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/tags [post]
func (t *Tag) Create(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.TagForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := t.TagBIZ.Create(ctx, form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags TagAPI
// @Security BearerAuth
// @Summary Update tag
// @Param id path string true "Tag ID"
// @Param body body schema.TagForm true "Tag form"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/tags/{id} [put]
func (t *Tag) Update(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.TagForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	err := t.TagBIZ.Update(ctx, c.Param("id"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags TagAPI
// @Security BearerAuth
// @Summary Delete tag
// @Param id path string true "Tag ID"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/tags/{id} [delete]
func (t *Tag) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := t.TagBIZ.Delete(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags TagAPI
// @Summary Get all tags (public, for tag cloud)
// @Success 200 {object} util.ResponseResult{data=[]schema.Tag}
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/tags/all [get]
func (t *Tag) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	tags, err := t.TagBIZ.GetAll(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, tags)
}
