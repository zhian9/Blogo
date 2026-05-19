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

type Statistics struct {
	StatisticsBIZ *biz.Statistics
}

// @Tags StatisticsAPI
// @Security ApiKeyAuth
// @Summary Query statistics list
// @Param current query int false "pagination index" default(1)
// @Param pageSize query int false "pagination size" default(10)
// @Param date_gte query string false "Date >= (YYYY-MM-DD)"
// @Param date_lte query string false "Date <= (YYYY-MM-DD)"
// @Success 200 {object} util.ResponseResult{data=[]schema.Statistics}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/statistics [get]
func (s *Statistics) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.StatisticsQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	result, err := s.StatisticsBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags StatisticsAPI
// @Security ApiKeyAuth
// @Summary Get statistics by date
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} util.ResponseResult{data=schema.Statistics}
// @Failure 401 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/statistics/{date} [get]
func (s *Statistics) Get(c *gin.Context) {
	ctx := c.Request.Context()
	stat, err := s.StatisticsBIZ.Get(ctx, c.Param("date"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, stat)
}

// @Tags StatisticsAPI
// @Summary Get latest N days statistics (public)
// @Param days query int false "Number of days" default(7)
// @Success 200 {object} util.ResponseResult{data=[]schema.Statistics}
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/statistics/latest [get]
func (s *Statistics) GetLatest(c *gin.Context) {
	ctx := c.Request.Context()
	days := int(util.ToInt64(c.DefaultQuery("days", "7")))
	if days <= 0 {
		days = 7
	}

	stats, err := s.StatisticsBIZ.GetLatest(ctx, days)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, stats)
}

// GetTraffic 返回控制中心流量趋势数据（最近 N 天的 PV/UV，按日期升序）
func (s *Statistics) GetTraffic(c *gin.Context) {
	s.GetLatest(c)
}
