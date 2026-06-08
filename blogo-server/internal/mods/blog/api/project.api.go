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

type Project struct {
	ProjectBIZ *biz.Project
}

func (p *Project) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.ProjectQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := p.ProjectBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

func (p *Project) Get(c *gin.Context) {
	ctx := c.Request.Context()
	project, err := p.ProjectBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, project)
}

func (p *Project) GetBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	project, err := p.ProjectBIZ.GetBySlug(ctx, c.Param("slug"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, project)
}

func (p *Project) GetFeatured(c *gin.Context) {
	ctx := c.Request.Context()
	projects, err := p.ProjectBIZ.GetFeatured(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, projects)
}

func (p *Project) Create(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.ProjectForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := p.ProjectBIZ.Create(ctx, form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

func (p *Project) Update(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.ProjectForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}
	err := p.ProjectBIZ.Update(ctx, c.Param("id"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (p *Project) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := p.ProjectBIZ.Delete(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (p *Project) UpdateStatus(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		IDs    []string `json:"ids" binding:"required,min=1"`
		Status string   `json:"status" binding:"required,oneof=draft published"`
	}
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}
	err := p.ProjectBIZ.UpdateStatus(ctx, req.IDs, req.Status)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (p *Project) ToggleTop(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		IsTop bool `json:"is_top"`
	}
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}
	if err := p.ProjectBIZ.ToggleTop(ctx, c.Param("id"), req.IsTop); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (p *Project) ToggleFeatured(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		IsFeatured bool `json:"is_featured"`
	}
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}
	if err := p.ProjectBIZ.ToggleFeatured(ctx, c.Param("id"), req.IsFeatured); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (p *Project) IncViews(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	num := util.ToInt64(c.DefaultQuery("num", "1"))
	if num <= 0 {
		num = 1
	}
	err := p.ProjectBIZ.IncViews(ctx, id, num)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
