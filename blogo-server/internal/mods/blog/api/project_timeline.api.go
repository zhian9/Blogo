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

type ProjectTimeline struct {
	ProjectTimelineBIZ *biz.ProjectTimeline
}

func (pt *ProjectTimeline) GetByProjectID(c *gin.Context) {
	ctx := c.Request.Context()
	timelines, err := pt.ProjectTimelineBIZ.GetByProjectID(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, timelines)
}

func (pt *ProjectTimeline) Get(c *gin.Context) {
	ctx := c.Request.Context()
	tl, err := pt.ProjectTimelineBIZ.Get(ctx, c.Param("tid"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, tl)
}

func (pt *ProjectTimeline) Create(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.ProjectTimelineForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := pt.ProjectTimelineBIZ.Create(ctx, c.Param("id"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

func (pt *ProjectTimeline) Update(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.ProjectTimelineForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}
	err := pt.ProjectTimelineBIZ.Update(ctx, c.Param("tid"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (pt *ProjectTimeline) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := pt.ProjectTimelineBIZ.Delete(ctx, c.Param("tid"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
