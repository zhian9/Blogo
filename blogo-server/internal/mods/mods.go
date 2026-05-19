// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package mods

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/blog"
	"github.com/zhian9/blogo-server/internal/mods/rbac"
	rbacapi "github.com/zhian9/blogo-server/internal/mods/rbac/api"
	"github.com/zhian9/blogo-server/pkg/middleware"
	"github.com/zhian9/blogo-server/pkg/util"
)

const (
	apiPrefix = "/api/"
)

// Collection of wirex providers
var Set = wire.NewSet(
	wire.Struct(new(Mods), "*"),
	rbac.Set,
	blog.Set,
)

type Mods struct {
	RBAC *rbac.RBAC
	Blog *blog.Blog
}

func (a *Mods) Init(ctx context.Context) error {
	if err := a.RBAC.Init(ctx); err != nil {
		return err
	}
	if err := a.Blog.Init(ctx); err != nil {
		return err
	}
	return nil
}

func (a *Mods) RouterPrefixes() []string {
	return []string{
		apiPrefix,
	}
}

//func (a *Mods) RegisterRouters(ctx context.Context, e *gin.Engine) error {
//	gAPI := e.Group(apiPrefix)
//	v1 := gAPI.Group("v1")
//
//	if err := a.RBAC.RegisterV1Routers(ctx, v1); err != nil {
//		return err
//	}
//	if err := a.Blog.RegisterV1Routers(ctx, v1); err != nil {
//		return err
//	}
//	return nil
//}

func (a *Mods) RegisterRouters(ctx context.Context, e *gin.Engine) error {
	gAPI := e.Group(apiPrefix)
	v1 := gAPI.Group("v1")

	// 忘记密码（公开接口，直接注册在 v1 上）
	pwdAPI := &rbacapi.Password{DB: a.RBAC.DB, Cache: a.RBAC.LoginAPI.LoginBIZ.Cache}
	{
		authGroup := v1.Group("/auth")
		authGroup.POST("/forgot-password", pwdAPI.ForgotPassword)
		authGroup.POST("/reset-password", pwdAPI.ResetPassword)
		authGroup.GET("/check-email", pwdAPI.CheckEmail)
	}

	// 博客公开接口（无需认证，注册在 Auth 之前）
	if err := a.Blog.RegisterV1PublicRouters(ctx, v1); err != nil {
		return err
	}

	// public
	if err := a.RBAC.RegisterV1PublicRouters(ctx, v1); err != nil {
		return err
	}

	// protected 必须加 Auth
	auth := v1.Group("")
	auth.Use(middleware.AuthWithConfig(middleware.AuthConfig{
		AllowedPathPrefixes: a.RouterPrefixes(),
		SkippedPathPrefixes: config.C.Middleware.Auth.SkippedPathPrefixes,
		ParseUserID:         a.RBAC.LoginAPI.LoginBIZ.ParseUserID,
		RootID:              config.C.General.Root.ID,
		CheckUserStatus:     a.RBAC.CheckUserStatus,
	}))

	// 操作日志中间件（必须在 Auth 之后，以便获取用户身份）
	// ParseUserID 作为 fallback：当 Auth 跳过某路径但请求携带有效 token 时，仍能解析用户身份
	auth.Use(middleware.OperationLogWithConfig(middleware.OperationLogConfig{
		Enabled:    true,
		AsyncWrite: true,
		OperationLogBIZFunc: func() interface{} {
			return a.RBAC.OperationLogBIZ
		},
		ParseUserID: a.RBAC.LoginAPI.LoginBIZ.ParseUserID,
	}))

	// Casbin RBAC（必须在 Auth 之后，确保 UserCache 已填充）
	auth.Use(middleware.CasbinWithConfig(middleware.CasbinConfig{
		AllowedPathPrefixes: a.RouterPrefixes(),
		SkippedPathPrefixes: config.C.Middleware.Casbin.SkippedPathPrefixes,
		Skipper: func(c *gin.Context) bool {
			if config.C.Middleware.Casbin.Disable ||
				util.FromIsRootUser(c.Request.Context()) {
				return true
			}
			return false
		},
		GetEnforcer: func(c *gin.Context) *casbin.Enforcer {
			return a.RBAC.Casbinx.GetEnforcer()
		},
		GetSubjects: func(c *gin.Context) []string {
			userCache := util.FromUserCache(c.Request.Context())
			if len(userCache.RoleIDs) > 0 {
				return userCache.RoleIDs
			}
			return []string{"guest"}
		},
	}))

	auth.PUT("/users/role/:id", a.RBAC.ChangeUserRole)
	auth.PUT("/users/status/:id", a.RBAC.ChangeUserStatus)

	if err := a.RBAC.RegisterV1ProtectedRouters(ctx, auth); err != nil {
		return err
	}

	if err := a.Blog.RegisterV1Routers(ctx, auth); err != nil {
		return err
	}

	return nil
}

func (a *Mods) Release(ctx context.Context) error {
	if err := a.RBAC.Release(ctx); err != nil {
		return err
	}
	if err := a.Blog.Release(ctx); err != nil {
		return err
	}
	return nil
}
