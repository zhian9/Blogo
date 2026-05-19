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

type Category struct {
	CategoryBIZ *biz.Category
}

// @Tags CategoryAPI
// @Security BearerAuth
// @Summary Query category list
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(10)
// @Param name query string false "Category name (fuzzy)"
// @Success 200 {object} util.ResponseResult{data=[]schema.Category}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/categories [get]
func (c *Category) Query(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	var params schema.CategoryQueryParam
	if err := util.ParseQuery(cxt, &params); err != nil {
		util.ResError(cxt, err)
		return
	}

	result, err := c.CategoryBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResPage(cxt, result.Data, result.PageResult)
}

// @Tags CategoryAPI
// @Security BearerAuth
// @Summary Get category by ID
// @Param id path string true "Category ID"
// @Success 200 {object} util.ResponseResult{data=schema.Category}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/categories/{id} [get]
func (c *Category) Get(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	category, err := c.CategoryBIZ.Get(ctx, cxt.Param("id"))
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResSuccess(cxt, category)
}

// @Tags CategoryAPI
// @Security BearerAuth
// @Summary Create category
// @Param body body schema.CategoryForm true "Category form"
// @Success 200 {object} util.ResponseResult{data=schema.Category}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/categories [post]
func (c *Category) Create(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	form := new(schema.CategoryForm)
	if err := util.ParseJSON(cxt, form); err != nil {
		util.ResError(cxt, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(cxt, err)
		return
	}

	result, err := c.CategoryBIZ.Create(ctx, form)
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResSuccess(cxt, result)
}

// @Tags CategoryAPI
// @Security BearerAuth
// @Summary Update category
// @Param id path string true "Category ID"
// @Param body body schema.CategoryForm true "Category form"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/categories/{id} [put]
func (c *Category) Update(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	form := new(schema.CategoryForm)
	if err := util.ParseJSON(cxt, form); err != nil {
		util.ResError(cxt, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(cxt, err)
		return
	}

	err := c.CategoryBIZ.Update(ctx, cxt.Param("id"), form)
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResOK(cxt)
}

// @Tags CategoryAPI
// @Security BearerAuth
// @Summary Delete category
// @Param id path string true "Category ID"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/categories/{id} [delete]
func (c *Category) Delete(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	err := c.CategoryBIZ.Delete(ctx, cxt.Param("id"))
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResOK(cxt)
}

// @Tags CategoryAPI
// @Summary Get all categories (public)
// @Success 200 {object} util.ResponseResult{data=[]schema.Category}
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/categories/all [get]
func (c *Category) GetAll(cxt *gin.Context) {
	ctx := cxt.Request.Context()
	categories, err := c.CategoryBIZ.GetAll(ctx)
	if err != nil {
		util.ResError(cxt, err)
		return
	}
	util.ResSuccess(cxt, categories)
}
