// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
// Project repository: https://github.com/zhian9/go-admin

package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/blog/biz"
	"github.com/zhian9/blogo-server/pkg/util"
)

type Profile struct {
	ProfileBIZ *biz.Profile
}

// GetDashboard 个人主页聚合接口
// GET /api/v1/users/:id/profile
func (p *Profile) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()
	viewerID := util.FromUserID(ctx)
	targetUserID := c.Param("id")

	dash, err := p.ProfileBIZ.GetDashboard(ctx, viewerID, targetUserID)
	if err != nil {
		util.ResError(c, err)
		return
	}
	if dash == nil {
		util.ResError(c, errors.New("user not found"))
		return
	}
	util.ResSuccess(c, dash)
}
