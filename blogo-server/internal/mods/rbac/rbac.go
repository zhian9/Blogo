// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by r [MIT/Apache/BSD] style license
// that can be found in the LICENSE file.
// Project repository: https://github.com/zhian9/go-admin

package rbac

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/internal/config"
	"github.com/zhian9/blogo-server/internal/mods/rbac/api"
	"github.com/zhian9/blogo-server/internal/mods/rbac/biz"
	"github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/logging"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RBAC 是 RBAC 模块的核心聚合对象，包含所有子组件。
type RBAC struct {
	DB              *gorm.DB          // 数据库连接
	MenuAPI         *api.Menu         // 菜单 API 控制器
	RoleAPI         *api.Role         // 角色 API 控制器
	UserAPI         *api.User         // 用户 API 控制器
	LoginAPI        *api.Login        // 登录 API 控制器
	LoggerAPI       *api.Logger       // 日志 API 控制器
	OperationLogAPI *api.OperationLog // 操作日志 API 控制器
	FollowAPI       *api.UserFollow   // 关注 API 控制器
	Casbinx         *Casbinx          // Casbin 权限管理器
	OperationLogBIZ *biz.OperationLog // 操作日志 BIZ（供中间件调用）
}

// AutoMigrate 自动创建或更新 RBAC 相关数据库表。
func (r *RBAC) AutoMigrate(ctx context.Context) error {
	return r.DB.AutoMigrate(
		new(schema.Menu),         // 菜单表
		new(schema.MenuResource), // 菜单资源表
		new(schema.Role),         // 角色表
		new(schema.RoleMenu),     // 角色-菜单关联表
		new(schema.User),         // 用户表
		new(schema.UserRole),     // 用户-角色关联表
		new(schema.UserFollow),   // 用户关注表
		new(schema.OperationLog), // 操作日志表
	)
}

// Init 初始化 RBAC 模块。
func (r *RBAC) Init(ctx context.Context) error {
	if config.C.Storage.DB.AutoMigrate {
		if err := r.AutoMigrate(ctx); err != nil {
			return err
		}
	}
	// 种子数据：超级管理员账号和基础角色（幂等）
	if err := r.InitAdminAccount(ctx); err != nil {
		logging.Context(ctx).Error("failed to init admin account", zap.Error(err))
	}
	// 硬编码菜单树种子（替代 menu.json）
	if err := r.InitMenus(ctx); err != nil {
		logging.Context(ctx).Error("failed to init menu seed", zap.Error(err))
	}
	// 确保 super_admin 拥有全部菜单
	if err := r.InitAdminMenuAccess(ctx); err != nil {
		logging.Context(ctx).Error("failed to grant menus to super_admin", zap.Error(err))
	}
	if err := r.Casbinx.Load(ctx); err != nil {
		return err
	}
	return nil
}

// RegisterV1PublicRouters 不需要登录的路由
func (r *RBAC) RegisterV1PublicRouters(ctx context.Context, v1 *gin.RouterGroup) error {
	captcha := v1.Group("captcha")
	{
		captcha.GET("id", r.LoginAPI.GetCaptcha)
		captcha.GET("image", r.LoginAPI.ResponseCaptcha)
	}
	v1.POST("login", r.LoginAPI.Login)
	v1.POST("register", r.LoginAPI.Register)
	v1.GET("verify-email", r.LoginAPI.VerifyEmail)
	return nil
}

// RegisterV1ProtectedRouters 需要登录的路由
func (r *RBAC) RegisterV1ProtectedRouters(ctx context.Context, v1 *gin.RouterGroup) error {
	current := v1.Group("current")
	{
		current.POST("refresh-token", r.LoginAPI.RefreshToken)
		current.GET("user", r.LoginAPI.GetUserInfo)
		current.GET("menus", r.LoginAPI.QueryMenus)
		current.PUT("password", r.LoginAPI.UpdatePassword)
		current.PUT("user", r.LoginAPI.UpdateUser)
		current.POST("logout", r.LoginAPI.Logout)
	}

	menu := v1.Group("menus")
	{
		menu.GET("", r.MenuAPI.Query)
		menu.GET(":id", r.MenuAPI.Get)
		menu.POST("", r.MenuAPI.Create)
		menu.PUT(":id", r.MenuAPI.Update)
		menu.DELETE(":id", r.MenuAPI.Delete)
	}

	role := v1.Group("roles")
	{
		role.GET("", r.RoleAPI.Query)
		role.GET(":id", r.RoleAPI.Get)
		role.POST("", r.RoleAPI.Create)
		role.PUT(":id", r.RoleAPI.Update)
		role.DELETE(":id", r.RoleAPI.Delete)
	}

	user := v1.Group("users")
	{
		user.GET("", r.UserAPI.Query)
		user.GET(":id", r.UserAPI.Get)
		user.POST("", r.UserAPI.Create)
		user.PUT(":id", r.UserAPI.Update)
		user.DELETE(":id", r.UserAPI.Delete)
		user.PATCH(":id/reset-pwd", r.UserAPI.ResetPassword)
	}

	logger := v1.Group("loggers")
	{
		logger.GET("", r.LoggerAPI.Query)
	}

	operationLog := v1.Group("operation-logs")
	{
		operationLog.GET("", r.OperationLogAPI.Query)
		operationLog.GET("/:id", r.OperationLogAPI.Get)
		operationLog.DELETE("/:id", r.OperationLogAPI.Delete)
	}

	// 关注系统
	follow := v1.Group("/users")
	{
		follow.POST("/:id/follow", r.FollowAPI.Follow)
		follow.DELETE("/:id/follow", r.FollowAPI.Unfollow)
		follow.GET("/:id/follow", r.FollowAPI.IsFollowing)
		follow.GET("/:id/followers", r.FollowAPI.ListFollowers)
		follow.GET("/:id/following", r.FollowAPI.ListFollowing)
	}

	return nil
}

// Release 释放 RBAC 模块占用的资源。
func (r *RBAC) Release(ctx context.Context) error {
	if err := r.Casbinx.Release(ctx); err != nil {
		return err
	}
	return nil
}
