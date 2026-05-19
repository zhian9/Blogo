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

type Setting struct {
	SettingBIZ *biz.Setting
}

// @Tags SettingAPI
// @Security ApiKeyAuth
// @Summary Query setting list
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(10)
// @Param key query string false "Setting key"
// @Success 200 {object} util.ResponseResult{data=[]schema.Setting}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/settings [get]
func (s *Setting) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.SettingQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := s.SettingBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags SettingAPI
// @Security ApiKeyAuth
// @Summary Get setting by key
// @Param key path string true "Setting key"
// @Success 200 {object} util.ResponseResult{data=schema.Setting}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/settings/{key} [get]
func (s *Setting) Get(c *gin.Context) {
	ctx := c.Request.Context()
	setting, err := s.SettingBIZ.Get(ctx, c.Param("key"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, setting)
}

// @Tags SettingAPI
// @Security ApiKeyAuth
// @Summary Create setting
// @Param body body schema.SettingForm true "Setting form"
// @Success 200 {object} util.ResponseResult{data=schema.Setting}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/settings [post]
func (s *Setting) Create(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.SettingForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := s.SettingBIZ.Create(ctx, form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SettingAPI
// @Security ApiKeyAuth
// @Summary Update setting
// @Param key path string true "Setting key"
// @Param body body schema.SettingForm true "Setting form"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/settings/{key} [put]
func (s *Setting) Update(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.SettingForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}

	err := s.SettingBIZ.Update(ctx, c.Param("key"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags SettingAPI
// @Security ApiKeyAuth
// @Summary Delete setting
// @Param key path string true "Setting key"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/settings/{key} [delete]
func (s *Setting) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := s.SettingBIZ.Delete(ctx, c.Param("key"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags SettingAPI
// @Security ApiKeyAuth
// @Summary Get all settings
// @Success 200 {object} util.ResponseResult{data=[]schema.Setting}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/settings/all [get]
func (s *Setting) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	settings, err := s.SettingBIZ.GetAll(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, settings)
}
