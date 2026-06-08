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

type ProjectResource struct {
	ProjectResourceBIZ *biz.ProjectResource
}

func (pr *ProjectResource) GetByProjectID(c *gin.Context) {
	ctx := c.Request.Context()
	resources, err := pr.ProjectResourceBIZ.GetByProjectID(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, resources)
}

func (pr *ProjectResource) Create(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.ProjectResourceForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := pr.ProjectResourceBIZ.Create(ctx, c.Param("id"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

func (pr *ProjectResource) Update(c *gin.Context) {
	ctx := c.Request.Context()
	form := new(schema.ProjectResourceForm)
	if err := util.ParseJSON(c, form); err != nil {
		util.ResError(c, err)
		return
	}
	if err := form.Validate(); err != nil {
		util.ResError(c, err)
		return
	}
	err := pr.ProjectResourceBIZ.Update(ctx, c.Param("rid"), form)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

func (pr *ProjectResource) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := pr.ProjectResourceBIZ.Delete(ctx, c.Param("id"), c.Param("rid"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
