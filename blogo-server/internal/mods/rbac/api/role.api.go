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

type Role struct {
	RoleBIZ *biz.Role
}

// @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary Query role list
// @Param current query int true "pagination index" default(1)
// @Param pageSize query int true "pagination size" default(10)
// @Param name query string false "Display name of role"
// @Param status query string false "Status of role (disabled, enabled)"
// @Success 200 {object} util.ResponseResult{data=[]schema.Role}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/roles [get]
func (r *Role) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.RoleQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := r.RoleBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary Get role record by ID
// @Param id path string true "unique id"
// @Success 200 {object} util.ResponseResult{data=schema.Role}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/roles/{id} [get]
func (r *Role) Get(c *gin.Context) {
	ctx := c.Request.Context()
	role, err := r.RoleBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, role)
}

// @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary Create role record
// @Param body body schema.RoleForm true "Request body"
// @Success 200 {object} util.ResponseResult{data=schema.Role}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/roles [post]
func (r *Role) Create(c *gin.Context) {
	ctx := c.Request.Context()
	role := new(schema.RoleForm)
	if err := util.ParseJSON(c, role); err != nil {
		util.ResError(c, err)
		return
	} else if err := role.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := r.RoleBIZ.Create(ctx, role)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary Update role record by ID
// @Param id path string true "unique id"
// @Param body body schema.RoleForm true "Request body"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/roles/{id} [put]
func (r *Role) Update(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.RoleForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	} else if err := item.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	err := r.RoleBIZ.Update(ctx, c.Param("id"), item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary Delete role record by ID
// @Param id path string true "unique id"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/roles/{id} [delete]
func (r *Role) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := r.RoleBIZ.Delete(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
