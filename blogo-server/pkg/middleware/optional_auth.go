// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zhian9/blogo-server/pkg/jwtx"
	"github.com/zhian9/blogo-server/pkg/util"
)

// OptionalAuthConfig 可选认证中间件配置。
// 与 Auth 中间件不同：Token 缺失或无效时静默放行，不返回 401。
type OptionalAuthConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	Auth                jwtx.Auther // JWT 认证器（用于签名验证）
	RootID              string
}

// OptionalAuthWithConfig 创建可选认证中间件。
// 从请求中提取 JWT token，若 token 存在且有效，则将用户 ID 注入上下文；
// 若 token 缺失或无效，静默放行（不报错）。
func OptionalAuthWithConfig(config OptionalAuthConfig) gin.HandlerFunc {
	if config.Auth == nil {
		return Empty()
	}

	return func(c *gin.Context) {
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		token := util.GetToken(c)
		if token == "" {
			c.Next()
			return
		}

		userID, err := config.Auth.ParseSubject(c.Request.Context(), token)
		if err != nil || userID == "" {
			c.Next()
			return
		}

		ctx := util.NewUserID(c.Request.Context(), userID)
		if userID == config.RootID {
			ctx = util.NewIsRootUser(ctx)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
