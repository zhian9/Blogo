// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/mods/rbac/biz"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

type Login struct {
	LoginBIZ *biz.Login
}

// @Tags LoginAPI
// @Summary Get captcha ID
// @Success 200 {object} util.ResponseResult{data=schema.Captcha}
// @Router /api/v1/captcha/id [get]
func (l *Login) GetCaptcha(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := l.LoginBIZ.GetCaptcha(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, data)
}

// @Tags LoginAPI
// @Summary Response captcha image
// @Param id query string true "Captcha ID"
// @Param reload query number false "Reload captcha image (reload=1)"
// @Produce image/png
// @Success 200 "Captcha image"
// @Failure 404 {object} util.ResponseResult
// @Router /api/v1/captcha/image [get]
func (l *Login) ResponseCaptcha(c *gin.Context) {
	ctx := c.Request.Context()
	// 支持两种 query 参数名：id（文档里）与 captcha_id（前端使用）
	id := c.Query("id")
	if id == "" {
		id = c.Query("captcha_id")
	}
	err := l.LoginBIZ.ResponseCaptcha(ctx, c.Writer, id, c.Query("reload") == "1")
	if err != nil {
		util.ResError(c, err, http.StatusBadRequest)
	}
}

// @Tags LoginAPI
// @Summary Login system with username and password
// @Param body body schema.LoginForm true "Request body"
// @Success 200 {object} util.ResponseResult{data=schema.LoginToken}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/login [post]
func (l *Login) Login(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.LoginForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}

	data, err := l.LoginBIZ.Login(ctx, item.Trim(), c.ClientIP(), c.Request.UserAgent(), util.GetClientIP(c))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, data)
}

// @Tags LoginAPI
// @Summary Register a new user
// @Description Register a new user account. An activation email will be sent to the provided email.
// @Param body body schema.RegisterForm true "Request body"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/register [post]
func (l *Login) Register(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.RegisterForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}

	if err := l.LoginBIZ.Register(ctx, item.Trim(), c.ClientIP(), c.Request.UserAgent(), util.GetClientIP(c)); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags LoginAPI
// @Summary Verify email via activation token
// @Description Activate user account by clicking the activation link sent via email.
// @Param token query string true "Activation token from email"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/verify-email [get]
func (l *Login) VerifyEmail(c *gin.Context) {
	ctx := c.Request.Context()
	token := c.Query("token")
	err := l.LoginBIZ.ActivateAccount(ctx, token)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(activationFailureHTML(err.Error())))
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(activationSuccessHTML))
}

// @Tags LoginAPI
// @Security BearerAuth
// @Summary Logout system
// @Success 200 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/logout [post]
func (l *Login) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	err := l.LoginBIZ.Logout(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags LoginAPI
// @Security BearerAuth
// @Summary Refresh current access token
// @Success 200 {object} util.ResponseResult{data=schema.LoginToken}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/refresh-token [post]
func (l *Login) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := l.LoginBIZ.RefreshToken(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, data)
}

// @Tags LoginAPI
// @Security BearerAuth
// @Summary Get current user info
// @Success 200 {object} util.ResponseResult{data=schema.User}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/user [get]
func (l *Login) GetUserInfo(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := l.LoginBIZ.GetUserInfo(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, data)
}

// @Tags LoginAPI
// @Security BearerAuth
// @Summary Change current user password
// @Param body body schema.UpdateLoginPassword true "Request body"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/password [put]
func (l *Login) UpdatePassword(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.UpdateLoginPassword)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}

	err := l.LoginBIZ.UpdatePassword(ctx, item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags LoginAPI
// @Security BearerAuth
// @Summary Query current user menus based on the current user role
// @Success 200 {object} util.ResponseResult{data=[]schema.Menu}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/menus [get]
func (l *Login) QueryMenus(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := l.LoginBIZ.QueryMenus(ctx)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, data)
}

// @Tags LoginAPI
// @Security BearerAuth
// @Summary Update current user info
// @Param body body schema.UpdateCurrentUser true "Request body"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/user [put]
func (l *Login) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.UpdateCurrentUser)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}

	err := l.LoginBIZ.UpdateUser(ctx, item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}
