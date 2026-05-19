// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/pkg/logging"
	"github.com/zhian9/blogo-server/pkg/util"
)

// AuthConfig 定义认证中间件的配置。
type AuthConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	RootID              string
	Skipper             func(c *gin.Context) bool
	ParseUserID         func(c *gin.Context) (string, error)

	// CheckUserStatus: 实时校验用户状态（每次请求查库，防封禁后未即时失效）
	// 返回 error 时中间件会返回 401 并附带错误信息。
	CheckUserStatus func(c *gin.Context, userID string) error
}

// AuthWithConfig 根据配置创建认证中间件。
// 功能：
//   - 路径过滤（白名单 + 黑名单 + 自定义 Skipper）
//   - 用户身份解析（通过 ParseUserID）
//   - 上下文注入（userID + isRootUser）
//   - 错误自动响应（调用 util.ResError）
func AuthWithConfig(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {

		// 三种跳过条件（满足任一即跳过）：
		//   a) 不在白名单路径中（AllowedPathPrefixes 非空时）
		//   b) 在黑名单路径中（SkippedPathPrefixes）
		//   c) 自定义 Skipper 返回 true
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			(config.Skipper != nil && config.Skipper(c)) {
			c.Next()
			return
		}

		// 调用自定义解析函数
		userID, err := config.ParseUserID(c)
		if err != nil {
			util.ResError(c, err)
			return
		}

		// 实时状态校验：每次请求查库，封禁即时生效
		if config.CheckUserStatus != nil {
			if err := config.CheckUserStatus(c, userID); err != nil {
				util.ResError(c, err)
				return
			}
		}

		// 注入用户身份到上下文
		// a) 注入到 util 上下文（供业务逻辑使用）
		ctx := util.NewUserID(c.Request.Context(), userID)

		// b) 注入到 logging 上下文（供日志记录使用）
		ctx = logging.NewUserID(ctx, userID)

		// c) 标记超级管理员（用于权限判断）
		if userID == config.RootID {
			ctx = util.NewIsRootUser(ctx)
		}

		// d) 更新请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 继续执行后续中间件/Handler
		c.Next()
	}
}
