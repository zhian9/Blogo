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
	"github.com/zhian9/blogo-server/pkg/errors"
	"github.com/zhian9/blogo-server/pkg/util"
)

// OperationLog 操作日志API控制器
type OperationLog struct {
	OperationLogBIZ *biz.OperationLog
}

// @Tags OperationLogAPI
// @Security BearerAuth
// @Summary Query operation logs list
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(20)
// @Param module query string false "operation module"
// @Param action_type query string false "action type"
// @Param requestMethod query string false "http method"
// @Param operator query string false "operator name"
// @Param description query string false "action description"
// @Param status query bool false "operation status"
// @Param startTime query string false "start time"
// @Param endTime query string false "end time"
// @Success 200 {object} util.ResponseResult{data=[]schema.OperationLog}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/operation-logs [get]
func (ol *OperationLog) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.OperationLogQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := ol.OperationLogBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags OperationLogAPI
// @Security BearerAuth
// @Summary Get operation log by ID
// @Param id path string true "operation log ID"
// @Success 200 {object} util.ResponseResult{data=schema.OperationLog}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/operation-logs/{id} [get]
func (ol *OperationLog) Get(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	item, err := ol.OperationLogBIZ.GetByID(ctx, id)
	if err != nil {
		util.ResError(c, err)
		return
	}
	if item == nil {
		util.ResError(c, errors.NotFound("", "Operation log not found"))
		return
	}
	util.ResSuccess(c, item)
}

// @Tags OperationLogAPI
// @Security BearerAuth
// @Summary Delete operation log
// @Param id path string true "operation log ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/operation-logs/{id} [delete]
func (ol *OperationLog) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := ol.OperationLogBIZ.DeleteByID(ctx, id); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
