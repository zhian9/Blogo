// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/rbac/biz"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

// Menu management for RBAC
type Menu struct {
	MenuBIZ *biz.Menu
}

// @Tags MenuAPI
// @Security BearerAuth
// @Summary Query menu tree data
// @Param code query string false "Code path of menu (like xxx.xxx.xxx)"
// @Param name query string false "Name of menu"
// @Param includeResources query bool false "Whether to include menu resources"
// @Success 200 {object} util.ResponseResult{data=[]schema.Menu}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/menus [get]
func (m *Menu) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.MenuQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := m.MenuBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags MenuAPI
// @Security BearerAuth
// @Summary Get menu record by ID
// @Param id path string true "unique id"
// @Success 200 {object} util.ResponseResult{data=schema.Menu}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/menus/{id} [get]
func (m *Menu) Get(c *gin.Context) {
	ctx := c.Request.Context()
	item, err := m.MenuBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, item)
}

// @Tags MenuAPI
// @Security BearerAuth
// @Summary Create menu record
// @Param body body schema.MenuForm true "Request body"
// @Success 200 {object} util.ResponseResult{data=schema.Menu}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/menus [post]
func (m *Menu) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.MenuForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	} else if err := item.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := m.MenuBIZ.Create(ctx, item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags MenuAPI
// @Security BearerAuth
// @Summary Update menu record by ID
// @Param id path string true "unique id"
// @Param body body schema.MenuForm true "Request body"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/menus/{id} [put]
func (m *Menu) Update(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.MenuForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	} else if err := item.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	err := m.MenuBIZ.Update(ctx, c.Param("id"), item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags MenuAPI
// @Security BearerAuth
// @Summary Delete menu record by ID
// @Param id path string true "unique id"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/menus/{id} [delete]
func (m *Menu) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := m.MenuBIZ.Delete(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
