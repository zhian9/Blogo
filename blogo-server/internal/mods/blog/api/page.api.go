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

type Page struct {
	PageBIZ *biz.Page
}

// @Tags PageAPI
// @Security ApiKeyAuth
// @Summary Query page list
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(10)
// @Param slug query string false "Page slug"
// @Success 200 {object} util.ResponseResult{data=[]schema.Page}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/pages [get]
func (p *Page) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.PageQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := p.PageBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags PageAPI
// @Security ApiKeyAuth
// @Summary Get page by ID
// @Param id path string true "Page ID"
// @Success 200 {object} util.ResponseResult{data=schema.Page}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/pages/{id} [get]
func (p *Page) Get(c *gin.Context) {
	ctx := c.Request.Context()
	page, err := p.PageBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, page)
}

// @Tags PageAPI
// @Summary Get published page by slug (public)
// @Param slug path string true "Page slug"
// @Success 200 {object} util.ResponseResult{data=schema.Page}
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/pages/slug/{slug} [get]
func (p *Page) GetBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	page, err := p.PageBIZ.GetBySlug(ctx, c.Param("slug"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, page)
}

// @Tags PageAPI
// @Security ApiKeyAuth
// @Summary Create page
// @Param body body schema.PageForm true "Page form"
// @Success 200 {object} util.ResponseResult{data=schema.Page}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/pages [post]
func (p *Page) Create(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.PageForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := p.PageBIZ.Create(ctx, form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags PageAPI
// @Security ApiKeyAuth
// @Summary Update page
// @Param id path string true "Page ID"
// @Param body body schema.PageForm true "Page form"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/pages/{id} [put]
func (p *Page) Update(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.PageForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	err := p.PageBIZ.Update(ctx, c.Param("id"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags PageAPI
// @Security ApiKeyAuth
// @Summary Delete page
// @Param id path string true "Page ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/pages/{id} [delete]
func (p *Page) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := p.PageBIZ.Delete(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
